package classfile

import (
	"math"
)

/*
CONSTANT_Class_info {
    u1 tag;
    u2 name_index;
}

CONSTANT_Module_info {
    u1 tag;
    u2 name_index;
}

CONSTANT_Package_info {
    u1 tag;
    u2 name_index;
}
*/

type ConstantClassInfo constantWithNameIdx
type ConstantModuleInfo constantWithNameIdx
type ConstantPackageInfo constantWithNameIdx

type constantWithNameIdx struct {
	NameIndex uint16
}

func readConstantClassInfo(reader *ClassReader) ConstantClassInfo {
	return ConstantClassInfo(readConstantWithNameIdx(reader))
}

func writeConstantClassInfo(writer *ClassWriter, data ConstantClassInfo) {
	writeConstantWithNameIdx(writer, constantWithNameIdx(data))
}

func readConstantModuleInfo(reader *ClassReader) ConstantModuleInfo {
	return ConstantModuleInfo(readConstantWithNameIdx(reader))
}

func writeConstantModuleInfo(writer *ClassWriter, data ConstantModuleInfo) {
	writeConstantWithNameIdx(writer, constantWithNameIdx(data))
}

func readConstantPackageInfo(reader *ClassReader) ConstantPackageInfo {
	return ConstantPackageInfo(readConstantWithNameIdx(reader))
}

func writeConstantPackageInfo(writer *ClassWriter, data ConstantPackageInfo) {
	writeConstantWithNameIdx(writer, constantWithNameIdx(data))
}

func readConstantWithNameIdx(reader *ClassReader) constantWithNameIdx {
	return constantWithNameIdx{
		NameIndex: reader.ReadUint16(),
	}
}

func writeConstantWithNameIdx(writer *ClassWriter, data constantWithNameIdx) {
	writer.WriteUint16(data.NameIndex)
}

/*
	CONSTANT_InvokeDynamic_info {
	    u1 tag;
	    u2 bootstrap_method_attr_index;
	    u2 name_and_type_index;
	}
*/
type ConstantInvokeDynamicInfo struct {
	BootstrapMethodAttrIndex uint16
	NameAndTypeIndex         uint16
}

func readConstantInvokeDynamicInfo(reader *ClassReader) ConstantInvokeDynamicInfo {
	return ConstantInvokeDynamicInfo{
		BootstrapMethodAttrIndex: reader.ReadUint16(),
		NameAndTypeIndex:         reader.ReadUint16(),
	}
}

func writeConstantInvokeDynamicInfo(writer *ClassWriter, data ConstantInvokeDynamicInfo) {
	writer.WriteUint16(data.BootstrapMethodAttrIndex)
	writer.WriteUint16(data.NameAndTypeIndex)
}

/*
	CONSTANT_MethodHandle_info {
	    u1 tag;
	    u1 reference_kind;
	    u2 reference_index;
	}
*/
type ConstantMethodHandleInfo struct {
	ReferenceKind  uint8
	ReferenceIndex uint16
}

func readConstantMethodHandleInfo(reader *ClassReader) ConstantMethodHandleInfo {
	return ConstantMethodHandleInfo{
		ReferenceKind:  reader.ReadUint8(),
		ReferenceIndex: reader.ReadUint16(),
	}
}

func writeConstantMethodHandleInfo(writer *ClassWriter, data ConstantMethodHandleInfo) {
	writer.WriteUint8(data.ReferenceKind)
	writer.WriteUint16(data.ReferenceIndex)
}

/*
	CONSTANT_MethodType_info {
	    u1 tag;
	    u2 descriptor_index;
	}
*/
type ConstantMethodTypeInfo struct {
	DescriptorIndex uint16
}

func readConstantMethodTypeInfo(reader *ClassReader) ConstantMethodTypeInfo {
	return ConstantMethodTypeInfo{
		DescriptorIndex: reader.ReadUint16(),
	}
}

func writeConstantMethodTypeInfo(writer *ClassWriter, data ConstantMethodTypeInfo) {
	writer.WriteUint16(data.DescriptorIndex)
}

/*
CONSTANT_Fieldref_info {
    u1 tag;
    u2 class_index;
    u2 name_and_type_index;
}
CONSTANT_Methodref_info {
    u1 tag;
    u2 class_index;
    u2 name_and_type_index;
}
CONSTANT_InterfaceMethodref_info {
    u1 tag;
    u2 class_index;
    u2 name_and_type_index;
}
*/

type ConstantFieldRefInfo constantMemberRefInfo
type ConstantMethodRefInfo constantMemberRefInfo
type ConstantInterfaceMethodRefInfo constantMemberRefInfo

type constantMemberRefInfo struct {
	ClassIndex       uint16
	NameAndTypeIndex uint16
}

func readConstantFieldRefInfo(reader *ClassReader) ConstantFieldRefInfo {
	return ConstantFieldRefInfo(readConstantMemberRefInfo(reader))
}

func writeConstantFieldRefInfo(writer *ClassWriter, data ConstantFieldRefInfo) {
	writeConstantMemberRefInfo(writer, constantMemberRefInfo(data))
}

func readConstantMethodRefInfo(reader *ClassReader) ConstantMethodRefInfo {
	return ConstantMethodRefInfo(readConstantMemberRefInfo(reader))
}

func writeConstantMethodRefInfo(writer *ClassWriter, data ConstantMethodRefInfo) {
	writeConstantMemberRefInfo(writer, constantMemberRefInfo(data))
}

func readConstantInterfaceMethodRefInfo(reader *ClassReader) ConstantInterfaceMethodRefInfo {
	return ConstantInterfaceMethodRefInfo(readConstantMemberRefInfo(reader))
}

func writeConstantInterfaceMethodRefInfo(writer *ClassWriter, data ConstantInterfaceMethodRefInfo) {
	writeConstantMemberRefInfo(writer, constantMemberRefInfo(data))
}

func readConstantMemberRefInfo(reader *ClassReader) constantMemberRefInfo {
	return constantMemberRefInfo{
		ClassIndex:       reader.ReadUint16(),
		NameAndTypeIndex: reader.ReadUint16(),
	}
}

func writeConstantMemberRefInfo(writer *ClassWriter, data constantMemberRefInfo) {
	writer.WriteUint16(data.ClassIndex)
	writer.WriteUint16(data.NameAndTypeIndex)
}

/*
	CONSTANT_NameAndType_info {
	    u1 tag;
	    u2 name_index;
	    u2 descriptor_index;
	}
*/
type ConstantNameAndTypeInfo struct {
	NameIndex       uint16
	DescriptorIndex uint16
}

func readConstantNameAndTypeInfo(reader *ClassReader) ConstantNameAndTypeInfo {
	return ConstantNameAndTypeInfo{
		NameIndex:       reader.ReadUint16(),
		DescriptorIndex: reader.ReadUint16(),
	}
}

func writeConstantNameAndTypeInfo(writer *ClassWriter, data ConstantNameAndTypeInfo) {
	writer.WriteUint16(data.NameIndex)
	writer.WriteUint16(data.DescriptorIndex)
}

/*
	CONSTANT_Integer_info {
	    u1 tag;
	    u4 bytes;
	}
*/
func readConstantIntegerInfo(reader *ClassReader) int32 {
	return int32(reader.ReadUint32())
}

func writeConstantIntegerInfo(writer *ClassWriter, data int32) {
	writer.WriteUint32(uint32(data))
}

/*
	CONSTANT_Float_info {
	    u1 tag;
	    u4 bytes;
	}
*/
func readConstantFloatInfo(reader *ClassReader) float32 {
	return math.Float32frombits(reader.ReadUint32())
}

func writeConstantFloatInfo(writer *ClassWriter, data float32) {
	writer.WriteUint32(math.Float32bits(data))
}

/*
	CONSTANT_Long_info {
	    u1 tag;
	    u4 high_bytes;
	    u4 low_bytes;
	}
*/
func readConstantLongInfo(reader *ClassReader) int64 {
	return int64(reader.ReadUint64())
}

func writeConstantLongInfo(writer *ClassWriter, data int64) {
	writer.WriteUint64(uint64(data))
}

/*
	CONSTANT_Double_info {
	    u1 tag;
	    u4 high_bytes;
	    u4 low_bytes;
	}
*/
func readConstantDoubleInfo(reader *ClassReader) float64 {
	return math.Float64frombits(reader.ReadUint64())
}
func writeConstantDoubleInfo(writer *ClassWriter, data float64) {
	writer.WriteUint64(math.Float64bits(data))
}

/*
	CONSTANT_Utf8_info {
	    u1 tag;
	    u2 length;
	    u1 bytes[length];
	}
*/
func readConstantUtf8Info(reader *ClassReader) []byte {
	length := int(reader.ReadUint16())
	return reader.ReadBytes(length)
}
func writeConstantUtf8Info(writer *ClassWriter, data []byte) {
	writer.WriteUint16(uint16(len(data)))
	writer.WriteBytes(data)
}

/*
	CONSTANT_String_info {
	    u1 tag;
	    u2 string_index;
	}
*/
type ConstantStringInfo struct {
	StringIndex uint16
}

func readConstantStringInfo(reader *ClassReader) ConstantStringInfo {
	return ConstantStringInfo{
		StringIndex: reader.ReadUint16(),
	}
}
func writeConstantStringInfo(writer *ClassWriter, data ConstantStringInfo) {
	writer.WriteUint16(data.StringIndex)
}
