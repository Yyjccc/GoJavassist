package reflect

import (
	"github.com/Yyjccc/GoJavassist/classfile"
)

type CtMethod struct {
	Class           *CtClass
	Member          *classfile.MemberInfo
	Name            string
	Descriptor      string
	Acc             *AccDesc
	throwExceptions *classfile.ExceptionsAttribute

	//local      *classfile.LocalVariableTableAttribute
}

func NewMethod(c *CtClass, method *classfile.MemberInfo) *CtMethod {
	class := c.ClassFile
	m := &CtMethod{
		Member:     method,
		Name:       class.GetUTF8(method.NameIndex),
		Descriptor: class.GetUTF8(method.DescriptorIndex),
		Acc:        NewAccDec(method.AccessFlags),
		Class:      c,
	}
	return m
}

func MakeCtMethod(declare *CtClass, src string) *CtMethod {
	return &CtMethod{
		Class: declare,
	}
}

func NewCtMethodInfo(returnType *CtClass, mname string, parameters []*CtClass, declaring *CtClass) *CtMethod {
	if !declaring.editMode {
		return nil
	}
	descriptor := ToMethodDescriptor(returnType, parameters)
	pool := declaring.GetConstPool()
	member := &classfile.MemberInfo{
		AccessFlags:     DefaultAcc.toAccessFlags(),
		NameIndex:       uint16(pool.AddString(mname)),
		DescriptorIndex: uint16(pool.AddString(descriptor)),
		AttributeTable:  make([]classfile.AttributeInfo, 0),
	}

	m := &CtMethod{
		Class:      declaring,
		Member:     member,
		Name:       mname,
		Descriptor: descriptor,
		Acc:        &DefaultAcc,
	}
	return m
}

func (m *CtMethod) GetSimpleName() string {
	return m.Class.ClassFile.GetUTF8(m.Member.NameIndex)
}

func (m CtMethod) GetDescriptor() string {
	return m.Class.ClassFile.GetUTF8(m.Member.DescriptorIndex)
}

func (m *CtMethod) GetFullName() string {
	return m.GetSimpleName() + m.GetDescriptor()
}

func (m *CtMethod) IsStatic() bool {
	return m.Acc.Static
}

func (m *CtMethod) GetParameterTypes() []*CtClass {
	return GetParameterTypes(m.Descriptor, m.Class.GetClassPool())
}

func (m *CtMethod) SetModifiers(acc *AccDesc) {
	m.Acc = acc
}

func (m *CtMethod) AddAttribute(attribute classfile.AttributeInfo) {
	attrName := ""
	switch attribute.(type) {
	case *classfile.CodeAttribute:
		attrName = classfile.Code
		break
	default:
		break
	}

	m.Class.edPool.AddString(attrName)
	m.Member.AttributeTable = append(m.Member.AttributeTable, attribute)
}

func (m *CtMethod) SetExceptionTypes(types []*CtClass) {
	m.Class.CheckModify()
	if types == nil || len(types) == 0 {
		m.Member.AttributeTable.Remove(classfile.Exceptions)
		return
	}
	n := len(types)
	names := make([]string, n)
	for i := 0; i < n; i++ {
		names[i] = types[i].GetName()
	}
	ea := NewExceptionAttribute()
	m.throwExceptions = ea
	m.Class.edPool.SetExceptions(ea, names)
}

func (m *CtMethod) GetReturnType() *CtClass {
	return GetReturnType(m.Descriptor, m.Class.GetClassPool())
}

func (m *CtMethod) GetCodeAttribute() *classfile.CodeAttribute {
	for _, attr := range m.Member.AttributeTable {
		if codeAttr, ok := attr.(*classfile.CodeAttribute); ok {
			return codeAttr
		}
	}
	return nil
}
