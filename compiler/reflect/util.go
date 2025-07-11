package reflect

import (
	"strings"
)

func GetReturnType(desc string, cp *ClassPool) *CtClass {
	i := strings.IndexByte(desc, ')')
	if i < 0 {
		return nil
	} else {
		aType := make([]*CtClass, 1)
		ToCtClass(cp, desc, i+1, aType, 0)
		return aType[0]
	}
}

func GetParameterTypes(desc string, cp *ClassPool) []*CtClass {
	if desc[0] != '(' {
		return nil
	}
	num := NumOfParameters(desc)
	args := make([]*CtClass, num)
	n := 0
	i := 1
	for i > 0 {
		i = ToCtClass(cp, desc, i, args, n)
		n++
	}
	return args
}

func ToCtClass(cp *ClassPool, desc string, i int, args []*CtClass, n int) int {
	arrayDim := 0
	var c byte
	for c = desc[i]; c == '['; c = desc[i] {
		arrayDim++
		i++
	}
	var i2 int
	var name string
	if c == 'L' {
		i++
		i2 = strings.IndexByte(desc, ';')
		name = strings.ReplaceAll(desc[i:i2], "/", ".")
		i2++
	} else {
		aType := ToPrimitiveClass(c)
		if aType == nil {
			return -1
		}
		i2 = i + 1
		if arrayDim == 0 {
			args[n] = aType
			return i2
		}
		name = aType.GetName()
	}
	if arrayDim > 0 {
		var sbuf strings.Builder
		sbuf.WriteString(name)
		arrayDim--
		for ; arrayDim > 0; arrayDim-- {
			sbuf.WriteString("[]")
		}
		name = sbuf.String()
	}
	args[n] = cp.Get(name)
	return i2
}

func ToPrimitiveClass(c byte) *CtClass {
	switch c {
	case 'B':
		return ByteType
	case 'C':
		return CharType
	case 'D':
		return DoubleType
	case 'F':
		return FloatType
	case 'I':
		return IntType
	case 'J':
		return LongType
	case 'S':
		return ShortType
	case 'V':
		return VoidType
	case 'Z':
		return BooleanType
	default:
		return nil
	}
}

func NumOfParameters(desc string) int {
	n := 0
	i := 1

	for {
		if i >= len(desc) {
			panic("bad descriptor")
		}

		c := desc[i]

		if c == ')' {
			break
		}

		// 处理数组维度
		for c == '[' {
			i++
			if i >= len(desc) {
				panic("bad descriptor")
			}
			c = desc[i]
		}

		// 处理类类型
		if c == 'L' {
			i = strings.IndexByte(desc[i:], ';') + i + 1
			if i <= 0 {
				panic("bad descriptor")
			}
		} else {
			i++
		}
		n++
	}
	return n
}

func DescriptorOf(pType *CtClass) string {
	var sbuf strings.Builder
	ToDescriptor(&sbuf, pType)
	return sbuf.String()
}

func DescriptorInsertParameter(pType *CtClass, desc string) string {
	if desc[0] != '(' {
		return desc
	}
	return "(" + DescriptorOf(pType) + desc[1:]
}
