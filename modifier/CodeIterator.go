package modifier

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type CodeIterator struct {
	Codes        []byte
	Method       reflect.CtMethod
	endPos       int
	CodeAttr     classfile.CodeAttribute
	currentPos   int
	mark         int
	Instruction  []classfile.Instruction
	ConstantPool []classfile.ConstantInfo //常量池
}

type Gap struct {
	position int
	length   int
}

type SearchOption func(iterator *CodeIterator, index int) bool

func NewCodeIterator(pool []classfile.ConstantInfo, code []byte) *CodeIterator {
	it := &CodeIterator{
		ConstantPool: pool,
		Codes:        code,
		Instruction:  make([]classfile.Instruction, 0),
	}
	it.parseCodes()
	return it
}

func (it *CodeIterator) parseCodes() {
	var instructions []classfile.Instruction
	pc := 0
	for pc < len(it.Codes) {
		// 读取操作码
		opcode := classfile.OpCode(it.Codes[pc])
		pc++
		// 根据操作码解析操作数
		var operands []byte
		//获取操作数长度数组
		operandLength := classfile.GetOperandLength(opcode)
		operands = it.Codes[pc : pc+operandLength]
		pc += operandLength
		instructions = append(instructions, classfile.Instruction{
			Opcode:   opcode,
			Operands: operands,
		})

	}
	it.Instruction = instructions
}

func (it *CodeIterator) SearchOne(op classfile.OpCode, options ...SearchOption) int {
	for index, instruction := range it.Instruction {
		if instruction.Opcode == op {
			finished := true
			for _, option := range options {
				if !option(it, index) {
					finished = false
					break
				}
			}
			if finished {
				return index
			}
		}
	}
	return -1
}

func (it *CodeIterator) HasNext() bool {
	return it.currentPos < len(it.Instruction)
}

func (it *CodeIterator) Next() (classfile.Instruction, int) {
	opcode := classfile.OpCode(it.Codes[it.currentPos])
	it.currentPos++
	// 根据操作码解析操作数
	var operands []byte
	//获取操作数长度数组
	operandLength := classfile.GetOperandLength(opcode)
	operands = it.Codes[it.currentPos : it.currentPos+operandLength]
	it.currentPos += operandLength
	instructions := classfile.Instruction{
		Opcode:   opcode,
		Operands: operands,
	}
	return instructions, it.currentPos
}

func (it *CodeIterator) Append(bytes []byte) int {
	var instructions []classfile.Instruction
	pc := 0
	for pc < len(bytes) {
		// 读取操作码
		opcode := classfile.OpCode(bytes[pc])
		pc++
		// 根据操作码解析操作数
		var operands []byte
		//获取操作数长度数组
		operandLength := classfile.GetOperandLength(opcode)
		operands = bytes[pc : pc+operandLength]
		pc += operandLength
		instructions = append(instructions, classfile.Instruction{
			Opcode:   opcode,
			Operands: operands,
		})
	}
	it.Instruction = append(it.Instruction, instructions...)
	it.Codes = append(it.Codes, bytes...)
	return len(it.Codes)
}

func (it *CodeIterator) AppendExceptionTable(table reflect.ExceptionTable, pos int) {

}

func (it *CodeIterator) SetMark(mark int) {
	it.mark = mark
}

// 返回指定位置的字节码
func (it *CodeIterator) ByteAt(index int) byte {
	if index < 0 || index >= len(it.Codes) {
		panic("ByteAt: index out of range")
	}
	return it.Codes[index]
}

// 修改指定位置的字节码
func (it *CodeIterator) WriteByte(value byte, index int) {
	if index < 0 || index >= len(it.Codes) {
		panic("WriteByte: index out of range")
	}
	it.Codes[index] = value
	it.parseCodes() // 修改后重新解析指令
}

// 在指定位置插入length个0x00字节
func (it *CodeIterator) InsertGapAt(pos int, length int) {
	if pos < 0 || pos > len(it.Codes) || length <= 0 {
		return
	}
	gap := make([]byte, length)
	it.Codes = append(it.Codes[:pos], append(gap, it.Codes[pos:]...)...)
	it.parseCodes()
}

// 在指定位置插入字节码
func (it *CodeIterator) InsertAt(pos int, bytes []byte) {
	if pos < 0 || pos > len(it.Codes) || len(bytes) == 0 {
		return
	}
	it.Codes = append(it.Codes[:pos], append(bytes, it.Codes[pos:]...)...)
	it.parseCodes()
}

// 删除指定位置length个字节码
func (it *CodeIterator) DeleteAt(pos int, length int) {
	if pos < 0 || pos+length > len(it.Codes) || length <= 0 {
		return
	}
	it.Codes = append(it.Codes[:pos], it.Codes[pos+length:]...)
	it.parseCodes()
}

// 将currentPos移动到指定位置
func (it *CodeIterator) Move(pos int) {
	if pos < 0 || pos > len(it.Codes) {
		panic("Move: pos out of range")
	}
	it.currentPos = pos
}

// 返回下一个指令的位置
func (it *CodeIterator) LookAhead() int {
	return it.currentPos
}

// 返回字节码长度
func (it *CodeIterator) GetCodeLength() int {
	return len(it.Codes)
}

// 返回当前字节码切片
func (it *CodeIterator) GetCode() []byte {
	return it.Codes
}

//func (it *CodeIterator) InsertGapAt(pos int, length int, exclusive bool) Gap {
//	gap := Gap{}
//	if length < 0 {
//		gap.position = pos
//		gap.length = 0
//		return gap
//	}
//	var c = make([]byte, 0)
//	length2 := 0
//	if len(it.Codes)+length > 32767 {
//
//	} else {
//		cur := it.currentPos
//		c = insertGapCore0(it.Codes, pos, length, exclusive, it.CodeAttr.GetExceptionIndexTable(), it.CodeAttr)
//	}
//}

func WithLastOp(op classfile.OpCode) SearchOption {
	return func(iterator *CodeIterator, index int) bool {
		i := index - 1
		if i < 0 {
			return false
		}
		instruction := iterator.Instruction[i]
		return instruction.Opcode == op
	}
}

func WithNextOp(op classfile.OpCode) SearchOption {
	return func(iterator *CodeIterator, index int) bool {
		i := index + 1
		if i > len(iterator.Instruction) {
			return false
		}
		return iterator.Instruction[i].Opcode == op
	}
}

// 移动到第一条指令
func (it *CodeIterator) Begin() {
	it.currentPos = 0
}

// 获取当前CodeAttribute
func (it *CodeIterator) Get() classfile.CodeAttribute {
	return it.CodeAttr
}

// 获取mark
func (it *CodeIterator) GetMark() int {
	return it.mark
}

// 备用mark2
func (it *CodeIterator) GetMark2() int {
	// TODO: 支持第二个mark
	return 0
}

// 设置备用mark2
func (it *CodeIterator) SetMark2(mark int) {
	// TODO: 支持第二个mark
}

// 返回有符号8位
func (it *CodeIterator) SignedByteAt(index int) int8 {
	if index < 0 || index >= len(it.Codes) {
		panic("SignedByteAt: index out of range")
	}
	return int8(it.Codes[index])
}

// 返回无符号16位
func (it *CodeIterator) U16bitAt(index int) uint16 {
	if index < 0 || index+1 >= len(it.Codes) {
		panic("U16bitAt: index out of range")
	}
	return uint16(it.Codes[index])<<8 | uint16(it.Codes[index+1])
}

// 返回有符号16位
func (it *CodeIterator) S16bitAt(index int) int16 {
	if index < 0 || index+1 >= len(it.Codes) {
		panic("S16bitAt: index out of range")
	}
	return int16(int16(it.Codes[index])<<8 | int16(it.Codes[index+1]))
}

// 返回有符号32位
func (it *CodeIterator) S32bitAt(index int) int32 {
	if index < 0 || index+3 >= len(it.Codes) {
		panic("S32bitAt: index out of range")
	}
	return int32(it.Codes[index])<<24 | int32(it.Codes[index+1])<<16 | int32(it.Codes[index+2])<<8 | int32(it.Codes[index+3])
}

// 写入16位
func (it *CodeIterator) Write16bit(value int, index int) {
	if index < 0 || index+1 >= len(it.Codes) {
		panic("Write16bit: index out of range")
	}
	it.Codes[index] = byte(value >> 8)
	it.Codes[index+1] = byte(value)
	it.parseCodes()
}

// 写入32位
func (it *CodeIterator) Write32bit(value int, index int) {
	if index < 0 || index+3 >= len(it.Codes) {
		panic("Write32bit: index out of range")
	}
	it.Codes[index] = byte(value >> 24)
	it.Codes[index+1] = byte(value >> 16)
	it.Codes[index+2] = byte(value >> 8)
	it.Codes[index+3] = byte(value)
	it.parseCodes()
}

// 写入字节数组
func (it *CodeIterator) Write(code []byte, index int) {
	if index < 0 || index+len(code) > len(it.Codes) {
		panic("Write: index out of range")
	}
	copy(it.Codes[index:], code)
	it.parseCodes()
}

// 末尾追加gap
func (it *CodeIterator) AppendGap(gapLength int) {
	gap := make([]byte, gapLength)
	it.Codes = append(it.Codes, gap...)
	it.parseCodes()
}

// 在当前指令前插入字节码
func (it *CodeIterator) Insert(code []byte) int {
	pos := it.currentPos
	it.InsertAt(pos, code)
	return pos
}

// 在指定位置插入字节码
func (it *CodeIterator) InsertAtPos(pos int, code []byte) {
	it.InsertAt(pos, code)
}

// 独占插入（排除块）
func (it *CodeIterator) InsertEx(code []byte) int {
	// TODO: 独占插入，暂同Insert
	return it.Insert(code)
}

func (it *CodeIterator) InsertExAt(pos int, code []byte) int {
	// TODO: 独占插入，暂同InsertAt
	it.InsertAt(pos, code)
	return pos
}

// 在当前指令前插入gap
func (it *CodeIterator) InsertGap(length int) int {
	pos := it.currentPos
	it.InsertGapAt(pos, length)
	return pos
}

// 在指定位置插入gap
func (it *CodeIterator) InsertGapAtPos(pos int, length int) int {
	it.InsertGapAt(pos, length)
	return pos
}

// 独占gap插入
func (it *CodeIterator) InsertExGap(length int) int {
	// TODO: 独占gap插入，暂同InsertGap
	return it.InsertGap(length)
}

func (it *CodeIterator) InsertExGapAt(pos int, length int) int {
	// TODO: 独占gap插入，暂同InsertGapAt
	it.InsertGapAt(pos, length)
	return pos
}

// 支持inclusive/exclusive gap插入
func (it *CodeIterator) InsertGapAtFull(pos int, length int, exclusive bool) Gap {
	// TODO: 支持exclusive参数
	it.InsertGapAt(pos, length)
	return Gap{position: pos, length: length}
}

// 跳过构造器super/this调用，返回INVOKESPECIAL指令位置，未找到返回-1
func (it *CodeIterator) SkipConstructor() int {
	for idx, inst := range it.Instruction {
		if inst.Opcode == classfile.OpInvokeSpecial && len(inst.Operands) >= 2 {
			index := int(inst.Operands[0])<<8 | int(inst.Operands[1])
			if it.isConstructorInvoke(index) {
				return idx
			}
		}
	}
	return -1
}

// 跳过super构造器调用，返回INVOKESPECIAL指令位置，未找到返回-1
func (it *CodeIterator) SkipSuperConstructor() int {
	for idx, inst := range it.Instruction {
		if inst.Opcode == classfile.OpInvokeSpecial && len(inst.Operands) >= 2 {
			index := int(inst.Operands[0])<<8 | int(inst.Operands[1])
			if it.isSuperConstructorInvoke(index) {
				return idx
			}
		}
	}
	return -1
}

// 跳过this构造器调用，返回INVOKESPECIAL指令位置，未找到返回-1
func (it *CodeIterator) SkipThisConstructor() int {
	for idx, inst := range it.Instruction {
		if inst.Opcode == classfile.OpInvokeSpecial && len(inst.Operands) >= 2 {
			index := int(inst.Operands[0])<<8 | int(inst.Operands[1])
			if it.isThisConstructorInvoke(index) {
				return idx
			}
		}
	}
	return -1
}

// 判断是否为构造器调用
func (it *CodeIterator) isConstructorInvoke(index int) bool {
	if index < 0 || index >= len(it.ConstantPool) {
		return false
	}
	info, ok := it.ConstantPool[index].(classfile.ConstantMethodRefInfo)
	if !ok {
		return false
	}
	nameAndTypeIdx := info.NameAndTypeIndex
	ntInfo, ok := it.ConstantPool[nameAndTypeIdx].(classfile.ConstantNameAndTypeInfo)
	if !ok {
		return false
	}
	name := string(it.ConstantPool[ntInfo.NameIndex].([]byte))
	return name == "<init>"
}

// 判断是否为super构造器调用
func (it *CodeIterator) isSuperConstructorInvoke(index int) bool {
	if index < 0 || index >= len(it.ConstantPool) {
		return false
	}
	info, ok := it.ConstantPool[index].(classfile.ConstantMethodRefInfo)
	if !ok {
		return false
	}
	classIdx := info.ClassIndex
	className := string(it.ConstantPool[classIdx].([]byte))
	// 这里假设CodeAttr有SuperClassName字段，实际需根据classfile结构调整
	if it.CodeAttr.Code == nil {
		return false
	}
	// 需根据实际父类名判断
	// 假设有it.SuperClassName字段
	// return className == it.SuperClassName && it.isConstructorInvoke(index)
	return className == it.Method.Class.QualifiedName && it.isConstructorInvoke(index)
}

// 判断是否为this构造器调用
func (it *CodeIterator) isThisConstructorInvoke(index int) bool {
	if index < 0 || index >= len(it.ConstantPool) {
		return false
	}
	info, ok := it.ConstantPool[index].(classfile.ConstantMethodRefInfo)
	if !ok {
		return false
	}
	classIdx := info.ClassIndex
	className := string(it.ConstantPool[classIdx].([]byte))
	// 这里假设CodeAttr有ThisClassName字段，实际需根据classfile结构调整
	if it.CodeAttr.AttributeTable == nil {
		return false
	}
	// 需根据实际当前类名判断
	// 假设有it.ThisClassName字段
	return className == it.Method.Class.QualifiedName && it.isConstructorInvoke(index)
	//return it.isConstructorInvoke(index) // 简化处理
}
