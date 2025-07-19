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
	code := method.GetCodeAttribute().Code
	return &MethodVisitor{
		thisMethod: method,
		thisClass:  method.Class,
		Iterator:   NewCodeIterator(method.Class.ClassFile.ConstantPool, code),
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
	iterator := NewCodeIterator(pool.GetPool(), ca.Code)

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
	return nil
}

func (v *MethodVisitor) InsertAfter(body string) error {
	cc := v.thisMethod.Class
	m := v.thisMethod
	cc.CheckModify()
	pool := cc.GetConstPool()
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return compiler.NewCompileError("no method body")
	}
	iterator := NewCodeIterator(pool.GetPool(), ca.Code)
	retAddr := ca.MaxLocals
	b := compiler.NewByteCodes(pool, 0, int(retAddr+1))
	b.StackDepth = int(ca.MaxStack + 1)
	jv := compiler.NewJavacWithByteCodes(b, cc)
	jv.RecordParams(m.GetParameterTypes(), m.Acc.Static)
	retType := m.GetReturnType()
	varNo := jv.RecordReturnType(retType, true)
	jv.RecordLocalVariables(ca, 0)
	handlerLen, err := v.InsertAfterHandler(false, b, retType, varNo, jv, body)
	if err != nil {
		return err
	}
	handlerPos := len(iterator.Codes)
	adviceLen := 0
	advicePos := 0
	noReturn := true
	for iterator.HasNext() {
		inst, pos := iterator.Next()
		c := inst.Opcode
		if c == classfile.OpAReturn || c == classfile.OpIReturn || c == classfile.OpFReturn || c == classfile.OpLReturn || c == classfile.OpDReturn || c == classfile.OpReturn {

		} else {
			if noReturn {
				adviceLen, err = v.InsertAfterAdvice(b, jv, body, pool, retType, varNo)
				if err != nil {
					return err
				}
				handlerPos = iterator.Append(b.Get())
				iterator.AppendExceptionTable(*b.GetExceptionTable(), handlerPos)
				advicePos = len(iterator.Codes) - adviceLen
				handlerLen = advicePos - handlerPos
				noReturn = false
			}
			v.InsertGoto(iterator, advicePos, pos)
			advicePos = len(iterator.Codes) - adviceLen
			handlerPos = advicePos - handlerLen
		}
	}
	if noReturn {
		handlerPos = iterator.Append(b.Get())
		iterator.AppendExceptionTable(*b.GetExceptionTable(), handlerPos)
	}
	ca.MaxStack = uint16(b.MaxStack)
	ca.MaxLocals = uint16(b.MaxLocals)
	ca.Code = iterator.Codes
	return nil
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
	if err := v.addParameter2(v.thisMethod.Acc.Static, pType, desc); err != nil {
		return err
	}
	v.thisMethod.Descriptor = desc2
	return nil
}

func (v *MethodVisitor) addParameter2(where bool, pType *reflect.CtClass, desc string) error {
	ca := v.thisMethod.GetCodeAttribute()
	if ca == nil {
		return nil
	}
	size := 1
	//TODO
	//typeDesc := 'L'
	//classInfo := 0
	if pType.IsPrimitive() {
		cpt := pType.PrimitiveType
		size = cpt.GetDataSize()
		//typeDesc = cpt.GetDescriptor()
	} else {
		//classInfo = v.thisClass.GetConstPool().AddClassInfo0(pType)
	}
	ca.InsertLocalVar(where, size)
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
