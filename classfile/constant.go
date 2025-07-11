package classfile

import (
	"bytes"
	"fmt"
)

//常量池

// Constant pool tags
const (
	ConstantUtf8               = 1  // Java 1.0.2
	ConstantInteger            = 3  // Java 1.0.2
	ConstantFloat              = 4  // Java 1.0.2
	ConstantLong               = 5  // Java 1.0.2
	ConstantDouble             = 6  // Java 1.0.2
	ConstantClass              = 7  // Java 1.0.2
	ConstantString             = 8  // Java 1.0.2
	ConstantFieldRef           = 9  // Java 1.0.2
	ConstantMethodRef          = 10 // Java 1.0.2
	ConstantInterfaceMethodRef = 11 // Java 1.0.2
	ConstantNameAndType        = 12 // Java 1.0.2
	ConstantMethodHandle       = 15 // Java 7
	ConstantMethodType         = 16 // Java 7
	ConstantInvokeDynamic      = 18 // Java 7
	ConstantModule             = 19 // Java 9
	ConstantPackage            = 20 // Java 9
	ConstantDynamic            = 17 // Java 11
)

/*
	ConstantValue_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 constantvalue_index;
	}
*/
type ConstantValueAttribute struct {
	ConstantValueIndex uint16
}

func readConstantValueAttribute(reader *ClassReader) ConstantValueAttribute {
	return ConstantValueAttribute{
		ConstantValueIndex: reader.ReadUint16(),
	}
}

func writeConstantValueAttribute(attribute ConstantValueAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(attribute.ConstantValueIndex >> 8), byte(attribute.ConstantValueIndex)}) // ConstantValueIndex
	return buf.Bytes()                                                                             // 返回字节切片
}

/*
	cp_info {
	    u1 tag;
	    u1 info[];
	}
*/
type ConstantInfo interface{}

func readConstantInfo(reader *ClassReader) ConstantInfo {
	tag := reader.ReadUint8()
	switch tag {
	case ConstantInteger:
		return readConstantIntegerInfo(reader)
	case ConstantFloat:
		return readConstantFloatInfo(reader)
	case ConstantLong:
		return readConstantLongInfo(reader)
	case ConstantDouble:
		return readConstantDoubleInfo(reader)
	case ConstantUtf8:
		return readConstantUtf8Info(reader)
	case ConstantString:
		return readConstantStringInfo(reader)
	case ConstantClass:
		return readConstantClassInfo(reader)
	case ConstantModule:
		return readConstantModuleInfo(reader)
	case ConstantPackage:
		return readConstantPackageInfo(reader)
	case ConstantFieldRef:
		return readConstantFieldRefInfo(reader)
	case ConstantMethodRef:
		return readConstantMethodRefInfo(reader)
	case ConstantInterfaceMethodRef:
		return readConstantInterfaceMethodRefInfo(reader)
	case ConstantNameAndType:
		return readConstantNameAndTypeInfo(reader)
	case ConstantMethodType:
		return readConstantMethodTypeInfo(reader)
	case ConstantMethodHandle:
		return readConstantMethodHandleInfo(reader)
	case ConstantInvokeDynamic:
		return readConstantInvokeDynamicInfo(reader)
	default: // TODO
		panic(fmt.Errorf("invalid constant pool tag: %d", tag))
	}
}

func readConstantPool(reader *ClassReader) []ConstantInfo {
	cpCount := int(reader.ReadUint16())
	cp := make([]ConstantInfo, cpCount)

	// The constant_pool table is indexed from 1 to constant_pool_count - 1.
	for i := 1; i < cpCount; i++ {
		cp[i] = readConstantInfo(reader)
		// http://docs.oracle.com/javase/specs/jvms/se8/html/jvms-4.html#jvms-4.4.5
		// All 8-byte constants take up two entries in the constant_pool table of the class file.
		// If a CONSTANT_Long_info or CONSTANT_Double_info structure is the item in the constant_pool
		// table at index n, then the next usable item in the pool is located at index n+2.
		// The constant_pool index n+1 must be valid but is considered unusable.
		switch cp[i].(type) {
		case int64, float64:
			i++
		}
	}

	return cp
}

func writeConstantInfo(writer *ClassWriter, constant ConstantInfo) {
	if constant == nil {
		return
	}
	// 根据常量类型的 tag 来写入
	switch constant := constant.(type) {
	case int32:
		writer.WriteUint8(ConstantInteger) // 写入常量类型 tag
		writeConstantIntegerInfo(writer, constant)
	case float32:
		writer.WriteUint8(ConstantFloat) // 写入常量类型 tag
		writeConstantFloatInfo(writer, constant)
	case int64:
		writer.WriteUint8(ConstantLong) // 写入常量类型 tag
		writeConstantLongInfo(writer, constant)
	case float64:
		writer.WriteUint8(ConstantDouble) // 写入常量类型 tag
		writeConstantDoubleInfo(writer, constant)
	case []byte:
		writer.WriteUint8(ConstantUtf8) // 写入常量类型 tag
		writeConstantUtf8Info(writer, constant)
	case ConstantStringInfo:
		writer.WriteUint8(ConstantString) // 写入常量类型 tag
		writeConstantStringInfo(writer, constant)
	case ConstantClassInfo:
		writer.WriteUint8(ConstantClass) // 写入常量类型 tag
		writeConstantClassInfo(writer, constant)
	case ConstantModuleInfo:
		writer.WriteUint8(ConstantModule) // 写入常量类型 tag
		writeConstantModuleInfo(writer, constant)
	case ConstantPackageInfo:
		writer.WriteUint8(ConstantPackage) // 写入常量类型 tag
		writeConstantPackageInfo(writer, constant)
	case ConstantFieldRefInfo:
		writer.WriteUint8(ConstantFieldRef) // 写入常量类型 tag
		writeConstantFieldRefInfo(writer, constant)
	case ConstantMethodRefInfo:
		writer.WriteUint8(ConstantMethodRef) // 写入常量类型 tag
		writeConstantMethodRefInfo(writer, constant)
	case ConstantInterfaceMethodRefInfo:
		writer.WriteUint8(ConstantInterfaceMethodRef) // 写入常量类型 tag
		writeConstantInterfaceMethodRefInfo(writer, constant)
	case ConstantNameAndTypeInfo:
		writer.WriteUint8(ConstantNameAndType) // 写入常量类型 tag
		writeConstantNameAndTypeInfo(writer, constant)
	case ConstantMethodTypeInfo:
		writer.WriteUint8(ConstantMethodType) // 写入常量类型 tag
		writeConstantMethodTypeInfo(writer, constant)
	case ConstantMethodHandleInfo:
		writer.WriteUint8(ConstantMethodHandle) // 写入常量类型 tag
		writeConstantMethodHandleInfo(writer, constant)
	case ConstantInvokeDynamicInfo:
		writer.WriteUint8(ConstantInvokeDynamic) // 写入常量类型 tag
		writeConstantInvokeDynamicInfo(writer, constant)
	default:
		panic(fmt.Errorf("unknown constant type: %T", constant))
	}
}
