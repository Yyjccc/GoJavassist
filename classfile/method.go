package classfile

import "bytes"

/*
	field_info {
	    u2             access_flags;
	    u2             name_index;
	    u2             descriptor_index;
	    u2             attributes_count;
	    attribute_info attributes[attributes_count];
	}

	method_info {
	    u2             access_flags;
	    u2             name_index;
	    u2             descriptor_index;
	    u2             attributes_count;
	    attribute_info attributes[attributes_count];
	}
*/
type MemberInfo struct {
	AccessFlags     uint16
	NameIndex       uint16
	DescriptorIndex uint16
	AttributeTable
}

// read field or method table
func readMembers(reader *ClassReader) []MemberInfo {
	return reader.readTable(readMember).([]MemberInfo)
}

// writeMembers writes the list of members (fields or methods).
func writeMembers(writer *ClassWriter, members []MemberInfo) {
	writer.WriteUint16(uint16(len(members)))
	for _, member := range members {
		writeMember(writer, member)
	}
}

func readMember(reader *ClassReader) MemberInfo {
	return MemberInfo{
		AccessFlags:     reader.ReadUint16(),
		NameIndex:       reader.ReadUint16(),
		DescriptorIndex: reader.ReadUint16(),
		AttributeTable:  readAttributes(reader),
	}
}
func writeMember(writer *ClassWriter, data MemberInfo) {
	writer.WriteUint16(data.AccessFlags)
	writer.WriteUint16(data.NameIndex)
	writer.WriteUint16(data.DescriptorIndex)
	attributes := writeAttributes(writer, data.AttributeTable) // 假设 writeAttributes 写入属性
	writer.WriteBytes(attributes)
}

/*
	EnclosingMethod_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 class_index;
	    u2 method_index;
	}
*/
type EnclosingMethodAttribute struct {
	ClassIndex  uint16
	MethodIndex uint16
}

func readEnclosingMethodAttribute(reader *ClassReader) EnclosingMethodAttribute {
	return EnclosingMethodAttribute{
		ClassIndex:  reader.ReadUint16(),
		MethodIndex: reader.ReadUint16(),
	}
}
func writeEnclosingMethodAttribute(attribute EnclosingMethodAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(attribute.ClassIndex >> 8), byte(attribute.ClassIndex)})   // ClassIndex
	buf.Write([]byte{byte(attribute.MethodIndex >> 8), byte(attribute.MethodIndex)}) // MethodIndex
	return buf.Bytes()                                                               // 返回字节切片
}
