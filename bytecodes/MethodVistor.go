package bytecodes

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type MethodVisitor struct {
	thisMethod *reflect.CtMethod
	thisClass  *reflect.CtClass
	compiler   *compiler.Javac
}

func NewMethodVisitor(method *reflect.CtMethod) *MethodVisitor {
	return &MethodVisitor{
		thisMethod: method,
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
	//TODO
	return nil
}

func (v *MethodVisitor) InsertAfter(body string) {}

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
