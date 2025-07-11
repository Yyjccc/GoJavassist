package reflect

import (
	"github.com/Yyjccc/GoJavassist/classfile"
)

type CtField struct {
	Member     *classfile.MemberInfo
	Declaring  *CtClass
	Name       string
	Descriptor string // jvm 描述符
	Acc        *AccDesc
	Value      string
	ValueType  int
	Annotation string
}

func NewField(declaring *CtClass, desc *classfile.MemberInfo) *CtField {
	class := declaring.ClassFile
	field := &CtField{
		Member:     desc,
		Acc:        NewAccDec(desc.AccessFlags),
		Descriptor: class.GetUTF8(desc.DescriptorIndex),
		Name:       class.GetUTF8(desc.NameIndex),
		Declaring:  declaring,
	}
	return field
}

// 逆向寻找索引
func (f CtField) GetFieldIndex() uint16 {
	nameAndTypeIndex := 0
	//遍历常量池
	for i, data := range f.Declaring.ClassFile.ConstantPool {
		info, ok := data.(classfile.ConstantNameAndTypeInfo)
		if ok {
			if info.NameIndex == f.Member.NameIndex && info.DescriptorIndex == f.Member.DescriptorIndex {
				nameAndTypeIndex = i
				break
			}
		} else {
			continue
		}
	}
	if nameAndTypeIndex == 0 {
		return 0
	}
	for i, data := range f.Declaring.ClassFile.ConstantPool {
		info, ok := data.(classfile.ConstantFieldRefInfo)
		if ok {
			if info.NameAndTypeIndex == uint16(nameAndTypeIndex) {
				return uint16(i)
			}
		} else {
			continue
		}
	}
	return 0
}

func (f *CtField) String() string {
	return f.Acc.String() + " " + f.Name + " ;"
}

func (f *CtField) GetConstantValue() interface{} {
	return 0
}

// 获取setter方法
func (f *CtField) GetFieldSetter() (*CtMethod, error) {
	return nil, nil
}
