package modifier

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type FieldVisitor struct {
	thisField *reflect.CtField
	thisClass *reflect.CtClass
}

func NewFieldVisitor(thisField *reflect.CtField) *FieldVisitor {
	return &FieldVisitor{
		thisField: thisField,
		thisClass: thisField.Declaring,
	}
}

func (v *FieldVisitor) SetInitStringValue(value string) error {
	if v.thisField.Descriptor != "Ljava/lang/String;" {
		return fmt.Errorf(v.thisField.Name + " is not String type")
	}
	if v.thisField.Acc.Static {

	} else {
		methods := v.thisClass.GetDeclareMethods("<init>")
		if len(methods) != 0 {
			return fmt.Errorf("cannot find <init> method")
		}
		method := methods[0]
		visitor := NewMethodVisitor(method)
		fmt.Println(visitor)
	}
	return nil

}
