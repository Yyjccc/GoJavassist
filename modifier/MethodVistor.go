package modifier

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type MethodVisitor struct {
	thisMethod *reflect.CtMethod
	thisClass  *reflect.CtClass
	compiler   *compiler.Javac
	Iterator   *CodeIterator
}

func NewMethodVisitor(method *reflect.CtMethod) *MethodVisitor {
	codeAttr := method.GetCodeAttribute()
	return &MethodVisitor{
		thisMethod: method,
		thisClass:  method.Class,
		Iterator:   NewCodeIterator(method, codeAttr),
	}
}

func (v *MethodVisitor) SetBody(body string) error {
	cc := v.thisMethod.Class
	cc.CheckModify()
	jv := compiler.NewJavac(cc)
	if err := jv.CompileBody(v.thisMethod, body); err != nil {
		return err
	}
	v.thisMethod.AddAttribute(jv.GetGen().GetBytecodes().ToCodeAttribute())
	v.thisMethod.Acc = reflect.NewAccDec(uint16(int(v.thisMethod.Acc.Raw) & -1025))

	return nil
}

func (v *MethodVisitor) InsertBefore(body string) error {
	cc := v.thisMethod.Class
	cc.CheckModify()
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return compiler.NewCompileError("no method body")
	}

	pool := cc.GetConstPool()
	iterator := NewCodeIterator(v.thisMethod, ca)

	// 1. 编译 body 代码为字节码
	b := compiler.NewByteCodes(pool, 0, int(ca.MaxLocals))
	b.StackDepth = int(ca.MaxStack)
	jv := compiler.NewJavacWithByteCodes(b, cc)
	jv.RecordParams(v.thisMethod.GetParameterTypes(), v.thisMethod.Acc.Static)
	jv.RecordLocalVariables(ca, 0)
	err := jv.CompileStmnt(body)
	if err != nil {
		return err
	}
	insertCodes := b.Get()

	// 2. 在方法体开头插入字节码
	iterator.InsertAt(0, insertCodes)

	// 3. 更新异常表入口（如果有异常表，需调整异常处理入口偏移）
	for i := range ca.ExceptionTable {
		ca.ExceptionTable[i].StartPc += uint16(len(insertCodes))
		ca.ExceptionTable[i].EndPc += uint16(len(insertCodes))
		ca.ExceptionTable[i].HandlerPc += uint16(len(insertCodes))
	}

	// 4. 更新 CodeAttribute 的 Code 字段
	ca.Code = iterator.Codes
	// 5. 更新最大栈和局部变量数
	if b.MaxStack > int(ca.MaxStack) {
		ca.MaxStack = uint16(b.MaxStack)
	}
	if b.MaxLocals > int(ca.MaxLocals) {
		ca.MaxLocals = uint16(b.MaxLocals)
	}
	v.ReLoadCodeAttr(*ca)
	return nil
}

// 替换指令为goto/goto_w，自动判断偏移
func replaceWithGoto(iterator *CodeIterator, fromInstIdx, toInstIdx int) {
	from := iterator.InstructionIndexToCodeOffset(fromInstIdx)
	to := iterator.InstructionIndexToCodeOffset(toInstIdx)
	// 如果advice代码就在return后面，直接顺序执行，无需插入goto
	if to == from+1 {
		// 可选：将原return指令替换为NOP，或直接跳过
		nopInst := classfile.NewInstruction(classfile.OpNop)
		iterator.Instruction[fromInstIdx] = nopInst
		copy(iterator.Codes[from:], nopInst.Bytes())
		return
	}
	offset := to - (from + 3)
	if offset < -32768 || offset > 32767 {
		gotoW := classfile.NewInstruction(classfile.OpGotoW).AddIndex(to - (from + 5))
		iterator.Instruction[fromInstIdx] = gotoW
		copy(iterator.Codes[from:], gotoW.Bytes())
	} else {
		gotoInst := classfile.NewInstruction(classfile.OpGoto).AddIndex(offset)
		iterator.Instruction[fromInstIdx] = gotoInst
		copy(iterator.Codes[from:], gotoInst.Bytes())
	}
}

func (v *MethodVisitor) InsertAfter(body string) error {
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return compiler.NewCompileError("no method body")
	}
	pool := v.thisClass.GetConstPool()
	iterator := NewCodeIterator(v.thisMethod, ca)

	b := compiler.NewByteCodes(pool, 0, int(ca.MaxLocals)+1)
	jv := compiler.NewJavacWithByteCodes(b, v.thisClass)
	jv.RecordParams(v.thisMethod.GetParameterTypes(), v.thisMethod.Acc.Static)
	jv.RecordLocalVariables(ca, 0)
	retType := v.thisMethod.GetReturnType()
	_ = jv.RecordReturnType(retType, true) // 保证副作用但不产生未使用警告
	err := jv.CompileStmnt(body)
	if err != nil {
		return err
	}

	// 记录第一个return指令类型
	var firstReturnOpcode classfile.OpCode = 0
	for _, inst := range iterator.Instruction {
		if classfile.IsReturnOpcode(inst.Opcode) {
			firstReturnOpcode = inst.Opcode
			break
		}
	}

	adviceCodes := b.Get()
	// adviceCodes 结尾补上原return类型的return指令
	if firstReturnOpcode != 0 {
		adviceCodes = append(adviceCodes, byte(firstReturnOpcode))
	}

	returnPositions := []int{}
	jumpInstructions := []struct{ pos, target int }{}
	switchJumps := []struct {
		pos     int
		targets []int
	}{}
	for i, inst := range iterator.Instruction {
		if classfile.IsReturnOpcode(inst.Opcode) {
			returnPositions = append(returnPositions, i)
		}
		if classfile.IsJumpInstruction(inst.Opcode) {
			target := classfile.CalcJumpTarget(i, inst)
			if target >= 0 {
				jumpInstructions = append(jumpInstructions, struct{ pos, target int }{i, target})
			}
		}
		if inst.Opcode == classfile.OpTableSwitch || inst.Opcode == classfile.OpLookupSwitch {
			targets, _ := classfile.ParseSwitchTargets(iterator.Codes, i)
			switchJumps = append(switchJumps, struct {
				pos     int
				targets []int
			}{i, targets})
		}
	}
	advicePos := len(iterator.Instruction)
	iterator.Append(adviceCodes)
	for _, retPos := range returnPositions {
		replaceWithGoto(iterator, retPos, advicePos)
	}

	for _, jump := range jumpInstructions {
		for _, retPos := range returnPositions {
			if jump.target == retPos {
				replaceWithGoto(iterator, jump.pos, advicePos)
			}
		}
	}

	for _, sw := range switchJumps {
		for idx, t := range sw.targets {
			for _, retPos := range returnPositions {
				if t == retPos {
					// 重新写入switch表项为advicePos
					opcode := iterator.Codes[sw.pos]
					pad := (4 - ((sw.pos + 1) % 4)) % 4
					idx2 := sw.pos + 1 + pad
					if opcode == byte(classfile.OpTableSwitch) {
						idx2 += 4 // default
						low := int(int32(iterator.Codes[idx2])<<24 | int32(iterator.Codes[idx2+1])<<16 | int32(iterator.Codes[idx2+2])<<8 | int32(iterator.Codes[idx2+3]))
						high := int(int32(iterator.Codes[idx2+4])<<24 | int32(iterator.Codes[idx2+5])<<16 | int32(iterator.Codes[idx2+6])<<8 | int32(iterator.Codes[idx2+7]))
						_ = high - low + 1 // 保证副作用但不产生未使用警告
						idx2 += 8
						if idx == 0 {
							adviceOffset := advicePos - sw.pos
							iterator.Write32bit(adviceOffset, sw.pos+1+pad)
						} else {
							adviceOffset := advicePos - sw.pos
							iterator.Write32bit(adviceOffset, idx2+4*(idx-1))
						}
					} else {
						idx2 += 4 // default
						if idx == 0 {
							adviceOffset := advicePos - sw.pos
							iterator.Write32bit(adviceOffset, sw.pos+1+pad)
						} else {
							adviceOffset := advicePos - sw.pos
							iterator.Write32bit(adviceOffset, idx2+8*(idx-1)+4)
						}
					}
				}
			}
		}
	}

	// 修正异常表
	for i := range ca.ExceptionTable {
		if int(ca.ExceptionTable[i].StartPc) >= advicePos {
			ca.ExceptionTable[i].StartPc += uint16(len(adviceCodes))
		}
		if int(ca.ExceptionTable[i].EndPc) >= advicePos {
			ca.ExceptionTable[i].EndPc += uint16(len(adviceCodes))
		}
		if int(ca.ExceptionTable[i].HandlerPc) >= advicePos {
			ca.ExceptionTable[i].HandlerPc += uint16(len(adviceCodes))
		}
	}

	ca.Code = iterator.Codes
	if b.MaxStack > int(ca.MaxStack) {
		ca.MaxStack = uint16(b.MaxStack)
	}
	if b.MaxLocals > int(ca.MaxLocals) {
		ca.MaxLocals = uint16(b.MaxLocals)
	}
	v.ReLoadCodeAttr(*ca)
	return nil
}

func (v *MethodVisitor) ReLoadCodeAttr(ca classfile.CodeAttribute) {
	newTable := make([]classfile.AttributeInfo, 0)
	table := v.thisMethod.Member.AttributeTable
	for _, attr := range table {
		if _, ok := attr.(classfile.CodeAttribute); ok {
			newTable = append(newTable, ca)
			continue
		}
		newTable = append(newTable, attr)
	}
	v.thisMethod.Member.AttributeTable = newTable
}

// AddLocalVariable 添加局部变量
func (v *MethodVisitor) AddLocalVariable(name string, pType *reflect.CtClass) error {
	v.thisClass.CheckModify()
	cp := v.thisClass.GetConstPool()
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return compiler.NewCompileError("no method body")
	} else {
		var va *classfile.LocalVariableTableAttribute
		vaP := ca.GetAttribute("LocalVariableTable")
		if vaP == nil {
			va = cp.NewLocalVariableAttribute()
			ca.AttributeTable = append(ca.AttributeTable, va)
		} else {
			va = vaP.(*classfile.LocalVariableTableAttribute)
		}

		maxLocals := ca.MaxLocals
		desc := reflect.DescriptorOf(pType)
		entry := classfile.LocalVariableTableEntry{
			StartPc:         0,
			Length:          ca.GetCodeLength(),
			NameIndex:       uint16(cp.AddString(name)),
			DescriptorIndex: uint16(cp.AddString(desc)),
			Index:           maxLocals,
		}
		va.LocalVariableTable = append(va.LocalVariableTable, entry)

		ca.MaxLocals = maxLocals + uint16(compiler.DescriptorDataSize(desc))
	}
	return nil
}

func (v *MethodVisitor) InsertParameter(pType *reflect.CtClass) error {
	v.thisClass.CheckModify()
	desc := v.thisMethod.Descriptor
	desc2 := reflect.DescriptorInsertParameter(pType, desc)
	where := 0
	if !v.thisMethod.Acc.Static {
		where = 1
	}
	if err := v.addParameter2(where, pType, desc); err != nil {
		return err
	}
	v.thisMethod.Descriptor = desc2
	return nil
}

func (v *MethodVisitor) addParameter2(where int, pType *reflect.CtClass, desc string) error {
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return nil
	}
	size := 1
	typeDesc := 'L'
	classInfo := 0
	if pType.IsPrimitive() {
		cpt := pType.PrimitiveType
		size = cpt.GetDataSize()
		typeDesc = cpt.GetDescriptor()
	} else {
		// 这里假设 AddClassInfo0 返回 int
		classInfo = v.thisClass.GetConstPool().AddClassInfo0(pType)
	}
	ca.InsertLocalVar(where, size)

	// LocalVariableTable
	vaP := ca.GetAttribute(classfile.LocalVariableTable)
	if vaP != nil {
		va := vaP.(*classfile.LocalVariableTableAttribute)
		va.ShiftIndex(where, size)
	}

	// LocalVariableTypeTable
	lvtaP := ca.GetAttribute(classfile.LocalVariableTypeTable)
	if lvtaP != nil {
		lvta := lvtaP.(*classfile.LocalVariableTypeTableAttribute)
		lvta.ShiftIndex(where, size)
	}

	// StackMapTable
	smtP := ca.GetAttribute(classfile.StackMapTable)
	if smtP != nil {
		smt := smtP.(*classfile.StackMapTableAttribute)
		smt.InsertLocal(where, classfile.TypeTagOf(typeDesc), classInfo)
	}

	return nil
}

func (v *MethodVisitor) InsertAfterHandler(asFinally bool, b *compiler.ByteCodes, retType *reflect.CtClass, returnVarNo int, javac *compiler.Javac, src string) (int, error) {
	if !asFinally {
		return 0, nil
	}
	varIndex := b.MaxLocals
	b.MaxLocals = b.MaxLocals + 1
	pc := b.CurrentPc()
	b.AddAStore(varIndex)
	if retType.IsPrimitive() {
		c := retType.GetDescriptor()
		if c == "D" {
			b.AddDconst(0.0)
			b.AddDstore(returnVarNo)
		} else if c == "F" {
			b.AddFconst(0)
			b.AddFstore(returnVarNo)
		} else if c == "J" {
			b.AddLconst(0)
			b.AddLstore(returnVarNo)
		} else if c == "V" {
			b.AddOpcode(classfile.OpAConstNull)
			b.AddAStore(returnVarNo)
		} else {
			b.AddIconst(0)
			b.AddLstore(returnVarNo)
		}
	} else {
		b.AddOpcode(classfile.OpAConstNull)
		b.AddAStore(returnVarNo)
	}
	err := javac.CompileStmnt(src)
	if err != nil {
		return 0, err
	}
	b.AddAload(varIndex)
	b.AddOpcode(classfile.OpAThrow)
	return b.CurrentPc() - pc, nil
}

func (v *MethodVisitor) InsertAfterAdvice(b *compiler.ByteCodes, jv *compiler.Javac, src string, pool *reflect.ConstPool, retType *reflect.CtClass, varNo int) (int, error) {
	pc := b.CurrentPc()
	if retType == reflect.VoidType {
		b.AddOpcode(classfile.OpAConstNull)
		b.AddAStore(varNo)
		err := jv.CompileStmnt(src)
		if err != nil {
			return 0, err
		}
		b.AddOpcode(classfile.OpReturn)
		if b.MaxLocals < 1 {
			b.MaxLocals = 1
		}
	} else {
		b.AddStore(varNo, retType)
		err2 := jv.CompileStmnt(src)
		if err2 != nil {
			return 0, err2
		}
		b.AddLoad(varNo, retType)
		if retType.IsPrimitive() {
			b.AddOpcode(retType.PrimitiveType.GetReturnOp())
		} else {
			b.AddOpcode(classfile.OpAReturn)
		}
	}

	return b.CurrentPc() - pc, nil
}

// 在 pos 位置插入 goto 指令，跳转到 subr
func (v *MethodVisitor) InsertGoto(iterator *CodeIterator, subr, pos int) {
	// 计算跳转偏移量
	offset := subr - (pos + 3) // goto指令长度为3字节（1字节opcode+2字节offset）

	// 判断是否需要用 goto_w
	if offset < -32768 || offset > 32767 {
		// 使用 goto_w（4字节偏移）
		gotoW := classfile.NewInstruction(classfile.OpGotoW).AddIndex(subr - (pos + 5))
		iterator.InsertAt(pos, gotoW.Bytes())
	} else {
		// 使用 goto（2字节偏移）
		gotoInst := classfile.NewInstruction(classfile.OpGoto).AddIndex(offset)
		iterator.InsertAt(pos, gotoInst.Bytes())
	}
}
