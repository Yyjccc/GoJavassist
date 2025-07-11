package compiler

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type Int = int32
type Long = int64
type Float = float32
type Double = float64

// ByteCodes class文件中 Code 属性
type ByteCodes struct {
	pool         *reflect.ConstPool
	Instructions []classfile.Instruction // 指令数组
	codes        []uint8
	MaxStack,
	MaxLocals int
	stackDepth int
	tryblocks  *reflect.ExceptionTable
}

func MakeByteCodes(pool *reflect.ConstPool) *ByteCodes {
	return &ByteCodes{
		pool:         pool,
		Instructions: make([]classfile.Instruction, 0),
		codes:        make([]uint8, 0),
		tryblocks:    reflect.NewExceptionTable(pool),
	}
}

// 添加空隙 ,方便之后填充
func (b *ByteCodes) AddGap(length int) {
	// 确保容量足够
	if len(b.codes)+length > cap(b.codes) {
		newSize := len(b.codes) * 2
		if newSize < len(b.codes)+length {
			newSize = len(b.codes) + length
		}

		newBuf := make([]uint8, len(b.codes), newSize)
		copy(newBuf, b.codes)
		b.codes = newBuf
	}

	// 增加 size
	b.codes = append(b.codes, make([]uint8, length)...)
}

func (b *ByteCodes) incMaxLocals(i int) {
	b.MaxLocals += i
}

// AddOpcode 添加 无操作数指令
func (b *ByteCodes) AddOpcode(op classfile.OpCode) {
	b.AddInstruction(classfile.NewInstruction(op))
}

func (b *ByteCodes) currentPc() int {
	return len(b.codes)
}

func (b *ByteCodes) AddInstruction(i classfile.Instruction) {
	b.Instructions = append(b.Instructions, i)
	b.codes = append(b.codes, i.Bytes()...)
}

// addAStore 添加ASore指令
func (b *ByteCodes) addAStore(n int) {
	if n < 4 {
		b.AddOpcode(classfile.OpCode(75 + n)) // astore_0 +n
	} else if n < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpAStore).Add(n))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpAStore).AddIndex(n))
	}
}

func (b *ByteCodes) addAload(n int) {
	if n < 4 {
		b.AddOpcode(classfile.OpCode(42 + n)) // aload_0 + n
	} else if n < 0x100 { // 256
		b.AddInstruction(classfile.NewInstruction(classfile.OpALoad).Add(n))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpALoad).AddIndex(n))
	}
}

// addIconst add ICONST
func (b *ByteCodes) addIconst(n int) {
	if n < 6 && -2 < n {
		b.AddOpcode(classfile.OpCode(3 + n)) // iconst_<i>   -1..5
	} else if n <= 127 && -128 <= n {
		b.AddInstruction(classfile.NewInstruction(classfile.OpBIPush).Add(n)) // bipush
	} else if n <= 32767 && -32768 <= n {
		b.AddInstruction(classfile.NewInstruction(classfile.OpSIPush).Add(n >> 8).Add(n)) // sipush
	} else {
		b.addLdc(b.pool.AddItem(int32(n)))
	}
}

// addLdc add  LDC or LDC_W
func (b *ByteCodes) addLdc(i int) {
	if i > 0xff {
		b.AddInstruction(classfile.NewInstruction(classfile.OpLDCw).AddIndex(i))
	} else {
		b.AddInstruction(classfile.NewInstruction(classfile.OpLDC).Add(i))
	}
}

// 添加加载指令 (iload/lload/fload/dload/aload)
func (b *ByteCodes) addIload(n int) {
	if n < 4 {
		b.AddOpcode(classfile.OpCode(0x1a + n)) // iload_0 ~ iload_3
	} else if n < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpILoad).Add(n))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpILoad).AddIndex(n))
	}
}

func (b *ByteCodes) AddLload(n int) {
	if n < 4 {
		b.AddOpcode(classfile.OpCode(0x1e + n)) // lload_0 ~ lload_3
	} else if n < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpLLoad).Add(n))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpLLoad).AddIndex(n))
	}
}

// 添加存储指令 (istore/lstore/fstore/dstore/astore)
func (b *ByteCodes) addIStore(n int) {
	if n < 4 {
		b.AddOpcode(classfile.OpCode(0x3b + n)) // istore_0 ~ istore_3
	} else if n < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpIStore).Add(n))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpIStore).AddIndex(n))
	}
}

// addLconst Long 常量 加载
func (b *ByteCodes) addLconst(v int64) {
	switch v {
	case 0:
		b.AddOpcode(classfile.OpLConst0)
	case 1:
		b.AddOpcode(classfile.OpLConst1)
	default:
		index := b.pool.AddItem(v)
		b.addLdc2(index)
	}
}

func (b *ByteCodes) addLdc2(index int) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpLDC2w).AddIndex(index))
}

// 添加控制流指令
func (b *ByteCodes) AddGoto(offset int) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpGoto).AddOffset(offset))
}

//// 添加方法调用指令
//func (b *ByteCodes) AddInvokeVirtual(index int) {
//	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeVirtual).AddIndex(index))
//}

// 添加字段操作指令
func (b *ByteCodes) AddGetStatic(index int) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpGetStatic).AddIndex(index))
}

// 添加对象操作指令
func (b *ByteCodes) AddNew(className string) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpNew).AddIndex(b.pool.AddClassInfo(className)))
}

// 添加数组操作指令
func (b *ByteCodes) AddANewArray(index int) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpANewArray).AddIndex(index))
}

// addDconst DCONST or DCONST_<n  加载 double 常量
func (b *ByteCodes) addDconst(d float64) {
	if d == 0.0 || d == 1.0 {
		b.AddOpcode(classfile.OpCode(int(d) + 14))
	} else {
		b.addLdc2w(d)
	}
}

// addFconst FCONST or FCONST_<n>  加载float 常量
func (b *ByteCodes) addFconst(f float32) {
	if f == 0.0 || f == 1.0 || f == 2.0 {
		b.AddOpcode(classfile.OpCode(int(f) + 11)) // fconst_<n>
	} else {
		b.addLdc(b.pool.AddItem(f))
	}
}

func (b *ByteCodes) addLdc2w(v interface{}) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpLDC2w).AddIndex(b.pool.AddItem(v)))
}

func (b *ByteCodes) addFload(i int) {
	if i < 4 {
		b.AddOpcode(classfile.OpCode(34 + i)) //fload_<n>
	} else if i < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpFLoad).Add(i)) //fload
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpFLoad).AddIndex(i))
	}
}

// addDload 加载 Double 类型的变量
func (b *ByteCodes) addDload(i int) {
	if i < 4 {
		b.AddOpcode(classfile.OpCode(38 + i))
	} else if i < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpDLoad).Add(i))
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpDLoad).AddIndex(i))
	}
}

// addDstore 存储 Double 类型变量
func (b *ByteCodes) addDstore(i int) {
	if i < 4 {
		b.AddOpcode(classfile.OpCode(71 + i)) // dstore_<n>
	} else if i < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpDStore).Add(i)) // dstore
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpDStore).AddIndex(i))
	}
}

// addFstore 存储 Float 类型变量
func (b *ByteCodes) addFstore(i int) {
	if i < 4 {
		b.AddOpcode(classfile.OpCode(67 + i)) // fstore_<n>
	} else if i < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpFStore).Add(i)) // fstore
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpFStore).AddIndex(i))
	}
}

// addLstore 存储 Long 类型变量
func (b *ByteCodes) addLstore(i int) {
	if i < 4 {
		b.AddOpcode(classfile.OpCode(63 + i)) // lstore_<n>
	} else if i < 0x100 {
		b.AddInstruction(classfile.NewInstruction(classfile.OpLStore).Add(i)) // lstore
	} else {
		b.AddOpcode(classfile.OpWide)
		b.AddInstruction(classfile.NewInstruction(classfile.OpLStore).AddIndex(i))
	}
}

// TODO jvmType is Class struct
func (b *ByteCodes) addExceptionHandler(start int, end int, pc int, jvmType string) {
	exceptIndex := 0
	if jvmType != "" {
		exceptIndex = b.pool.AddClassInfo(jvmType)
	}
	b.tryblocks.Add(start, end, pc, exceptIndex)
}

func (b *ByteCodes) growStack(diff int) {
	depth := b.stackDepth + diff
	b.stackDepth = depth
	if b.stackDepth > b.MaxStack {
		b.MaxStack = b.stackDepth
	}

}

func (b *ByteCodes) WriteIndex(index int) {
	b.codes = append(b.codes, byte(index>>8), byte(index))
}

func (b *ByteCodes) Write16bit(offset int, value int) {
	value = value - 1
	if offset < 0 || offset+1 >= len(b.codes) {
		panic("offset out of range")
	}
	b.codes[offset+1] = uint8(value >> 8) // 高 8 位
	b.codes[offset+2] = uint8(value)      // 低 8 位
}

// Write32bit 在 `codes` 指定偏移位置写入 32 位整数
func (b *ByteCodes) Write32bit(offset int, value int) {
	if offset < 0 || offset+3 >= len(b.codes) {
		panic("offset out of range")
	}
	b.Write16bit(offset, value>>16) // 高 16 位
	b.Write16bit(offset+2, value)   // 低 16 位
}

func (b *ByteCodes) addInvokeVirtual(classname string, name string, desc string) {
	b.addInvokeVirtual0(b.AddStringInfo(classname), name, desc)
}

func (b *ByteCodes) addInvokeVirtual0(clazzIndex int, name string, d string) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeVirtual).AddIndex(b.pool.AddMethodRefInfo(clazzIndex, name, d)))
	b.growStack(1)
	// TODO
	// growStack(Descriptor.dataSize(desc) - 1);
}

func (b *ByteCodes) addInvokeStatic(classname string, method string, s string) {
	b.addInvokeStatic0(b.pool.AddClassInfo(classname), method, s)
}

func (b *ByteCodes) addInvokeStatic0(clazz int, name string, s string) {
	b.addInvokeStatic1(clazz, name, s, false)
}

func (b *ByteCodes) addInvokeStatic1(clazz int, name string, s string, isInterface bool) {
	var index int
	if isInterface {
		index = b.pool.AddInterfaceMethodRefInfo(clazz, name, s)
	} else {
		index = b.pool.AddMethodRefInfo(clazz, name, s)
	}
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeStatic).AddIndex(index))

	//growStack(Descriptor.dataSize(desc));
}

func (b *ByteCodes) addInvokeStatic2(class *reflect.CtClass, mname string, des string) {
	if class == nil {
		panic("class cannot be nil")
	}
	// 判断是否是接口
	isInterface := false
	if class.QualifiedName == b.pool.This.QualifiedName {
		isInterface = false
	} else {
		isInterface = class.IsInterface()
	}

	var index int
	if isInterface {
		index = b.pool.AddInterfaceMethodRefInfo(b.pool.AddClassInfo0(class), mname, des)
	} else {
		index = b.pool.AddMethodRefInfo(b.pool.AddClassInfo0(class), mname, des)
	}
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeStatic).AddIndex(index))
	b.growStack(DescriptorDataSize(des))
}

func (b *ByteCodes) addInvokeVirtual2(class *reflect.CtClass, mname string, des string) {
	if class == nil {
		panic("class cannot be nil")
	}
	index := b.pool.AddMethodRefInfo(b.pool.AddClassInfo0(class), mname, des)
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeVirtual).AddIndex(index))
	///b.GrowStack(reflect.DescriptorDataSize(des) - 1)
}
func (b *ByteCodes) Write(pos int, nop classfile.OpCode) {
	// 确保位置有效
	if pos < 0 || pos >= len(b.codes) {
		return
	}
	b.codes[pos] = byte(nop)
}

func (b *ByteCodes) AddByte(v int) *ByteCodes {
	b.codes = append(b.codes, byte(v))
	return b
}

func (b *ByteCodes) add32bit(value int) {
	b.AddByte(value >> 24).AddByte(value >> 16).AddByte(value >> 8).AddByte(value)
}

func (b *ByteCodes) addCheckCast(class string) {
	info := classfile.ConstantClassInfo{
		NameIndex: uint16(b.AddStringInfo(class)),
	}
	b.AddInstruction(classfile.NewInstruction(classfile.OpCheckCast).AddIndex(b.pool.AddConstantClassInfo(info)))
}

func (b *ByteCodes) ToCodeAttribute() *classfile.CodeAttribute {
	return &classfile.CodeAttribute{
		MaxStack:       uint16(b.MaxStack),
		MaxLocals:      uint16(b.MaxLocals),
		Code:           b.codes,
		ExceptionTable: b.tryblocks.ToExceptionTable(),
		AttributeTable: make([]classfile.AttributeInfo, 0),
	}
}

func (b *ByteCodes) AddStringInfo(s string) int {
	return b.pool.AddString(s)
}

func (b *ByteCodes) AddStringRef(s string) int {
	return b.pool.AddStringRef(s)
}

func (b *ByteCodes) addAnewArray(className string) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpANewArray).AddIndex(b.pool.AddClassInfo(className)))
}

func (b *ByteCodes) AddMultiNewArray(descriptor string, dim int) int {
	b.AddInstruction(classfile.NewInstruction(classfile.OpMultiANewArray).AddIndex(b.pool.AddClassInfo(descriptor)).Add(dim))
	b.growStack(1)
	return dim
}

func (b *ByteCodes) AddLoad(n int, clazz *reflect.CtClass) int {
	if clazz.IsPrimitive() {
		if clazz == reflect.BooleanType || clazz == reflect.CharType || clazz == reflect.ByteType || clazz == reflect.ShortType || clazz == reflect.IntType {
			b.AddLload(n)
		} else if clazz == reflect.LongType {
			b.AddLload(n)
			return 2
		} else if clazz == reflect.FloatType {
			b.addFload(n)
		} else if clazz == reflect.DoubleType {
			b.addDload(n)
			return 2
		} else {
			panic("void type ?")
		}

	} else {
		b.addAload(n)
	}
	return 1
}

func (b *ByteCodes) AddInvokeSpecial(clazz string, name string, desc string) {
	b.addInvokeSpecial0(false, b.pool.AddClassInfo(clazz), name, desc)
}

func (b *ByteCodes) addInvokeSpecial0(isInterface bool, clazz int, name string, desc string) {
	var index int
	if isInterface {
		index = b.pool.AddInterfaceMethodRefInfo(clazz, name, desc)
	} else {
		index = b.pool.AddMethodRefInfo(clazz, name, desc)
	}
	b.addInvokeSpecial1(index, desc)
}

func (b *ByteCodes) addInvokeSpecial1(index int, d string) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeSpecial).AddIndex(index))
	b.growStack(DescriptorDataSize(d) - 1)
}

func (b *ByteCodes) AddInvokeInterface(class *reflect.CtClass, mname string, des string, count int) {
	b.AddInvokeInterface0(b.pool.AddClassInfo0(class), mname, des, count)
}

func (b *ByteCodes) AddInvokeInterface0(clazz int, mname string, des string, count int) {
	b.AddInstruction(classfile.NewInstruction(classfile.OpInvokeInterface).AddIndex(b.pool.AddInterfaceMethodrefInfo(clazz, mname, des)).Add(count).Add(0))
	b.growStack(DescriptorDataSize(des) - 1)
}
