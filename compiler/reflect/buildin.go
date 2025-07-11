package reflect

import (
	"github.com/Yyjccc/GoJavassist/classfile"
)

const (
	/* array-type code for the newarray instruction */

	T_BOOLEAN = 4
	T_CHAR    = 5
	T_FLOAT   = 6
	T_DOUBLE  = 7
	T_BYTE    = 8
	T_SHORT   = 9
	T_INT     = 10
	T_LONG    = 11
)

var (
	DefaultAcc = AccDesc{
		Public: true,
		Raw:    classfile.AccPublic,
	}
	VoidType    *CtClass = NewPrimitiveType("void", 'V', "java.lang.Void", "", "", classfile.OpReturn, 0, 0)
	BooleanType *CtClass = NewPrimitiveType("boolean", 'Z', "java.lang.Boolean", "booleanValue", "()Z", classfile.OpIReturn, T_BOOLEAN, 1)
	CharType    *CtClass = NewPrimitiveType("char", 'C', "java.lang.Character", "charValue", "()C", classfile.OpIReturn, T_CHAR, 1)
	ByteType    *CtClass = NewPrimitiveType("byte", 'B', "java.lang.Byte", "byteValue", "()B", classfile.OpIReturn, T_BYTE, 1)
	ShortType   *CtClass = NewPrimitiveType("short", 'S', "java.lang.Short", "shortValue", "()S", classfile.OpIReturn, T_SHORT, 1)
	IntType     *CtClass = NewPrimitiveType("int", 'I', "java.lang.Integer", "intValue", "()I", classfile.OpIReturn, T_INT, 1)
	LongType    *CtClass = NewPrimitiveType("long", 'J', "java.lang.Long", "longValue", "()J", classfile.OpLReturn, T_LONG, 2)
	FloatType   *CtClass = NewPrimitiveType("float", 'F', "java.lang.Float", "floatValue", "()F", classfile.OpFReturn, T_FLOAT, 1)
	DoubleType  *CtClass = NewPrimitiveType("double", 'D', "java.lang.Double", "doubleValue", "()D", classfile.OpDReturn, T_DOUBLE, 2)
)

func NewPrimitiveType(name string, descriptor rune, wrapper string, methodName string, mDesc string, opcode classfile.OpCode, aType, size int) *CtClass {
	primitiveType := &CtPrimitiveType{
		wrapper:       wrapper,
		Descriptor:    descriptor,
		getMethodName: methodName,
		mDescriptor:   mDesc,
		returnOp:      opcode,
		arrayType:     aType,
		dataSize:      size,
	}
	class := &CtClass{
		ClassFile:      nil,
		isPrimitive:    true,
		PrimitiveType:  primitiveType,
		Methods:        make(map[string]*CtMethod),
		Fields:         make(map[string]*CtField),
		PackageName:    "",
		SimpleName:     name,
		QualifiedName:  name,
		wasFrozen:      true,
		Acc:            &DefaultAcc,
		SuperClassName: "",
		Interfaces:     make([]*CtClass, 0),
		Imports:        make([]string, 0),
	}
	DefaultPool.Register(class)
	return class
}

// 更好描述访问控制符
type AccDesc struct {
	Public    bool
	Private   bool
	Protected bool
	Static    bool
	Final     bool
	Abstract  bool
	Raw       uint16
}

func NewAccDec(acc uint16) *AccDesc {
	return &AccDesc{
		Raw:       acc,
		Public:    acc&classfile.AccPublic != 0,
		Private:   acc&classfile.AccPrivate != 0,
		Protected: acc&classfile.AccProtected != 0,
		Static:    acc&classfile.AccStatic != 0,
		Final:     acc&classfile.AccFinal != 0,
		Abstract:  acc&classfile.AccAbstract != 0,
	}
}

func (flags *AccDesc) toAccessFlags() uint16 {
	var result uint16 = 0
	if flags.Public {
		result |= 0x0001 // ACC_PUBLIC
	}
	if flags.Private {
		result |= 0x0002 // ACC_PRIVATE
	}
	if flags.Protected {
		result |= 0x0004 // ACC_PROTECTED
	}
	if flags.Static {
		result |= 0x0008 // ACC_STATIC
	}
	if flags.Final {
		result |= 0x0010 // ACC_FINAL
	}
	if flags.Abstract {
		result |= 0x0400 // ACC_ABSTRACT
	}
	flags.Raw = result
	return result
}

func (a *AccDesc) String() string {
	res := ""
	if a.Public {
		res += "public "
	} else if a.Private {
		res += "private "
	} else if a.Protected {
		res += "protected "
	}
	if a.Static {
		res += "static "
	}
	if a.Final {
		res += "final "
	}
	if a.Abstract {
		res += "abstract "
	}
	return res
}
