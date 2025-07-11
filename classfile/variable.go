package classfile

import (
	"bytes"
	"sort"
	"strconv"
)

// 解析字节码 , store指令会改变局部变量表
func (table *LocalVariableTableAttribute) Store(cf *ClassFile, index uint16, varIndex uint16) uint16 {
	varName := "var_" + strconv.Itoa(int(varIndex))
	bytes := []uint8(varName)
	strIndex := cf.AddConstInfo(bytes)
	entry := LocalVariableTableEntry{
		StartPc:         0,
		Length:          0,
		NameIndex:       strIndex,
		DescriptorIndex: 0,
		Index:           index,
	}
	// 如果 index 超出当前表的长度，就扩展 LocalVariableTable
	if int(index) >= len(table.LocalVariableTable) {
		// 扩展表并填充至指定的 index
		// 通过扩展表的方式确保 index 不会越界
		for len(table.LocalVariableTable) <= int(index) {
			table.LocalVariableTable = append(table.LocalVariableTable, LocalVariableTableEntry{})
		}
	}
	// 添加新项到指定位置
	table.LocalVariableTable[index] = entry
	return uint16(strIndex)
}

// Sort 方法：排序 LocalVariableTable 按照 Index 字段
func (table *LocalVariableTableAttribute) Sort() {
	sort.Slice(table.LocalVariableTable, func(i, j int) bool {
		return table.LocalVariableTable[i].Index < table.LocalVariableTable[j].Index
	})
}

//局部变量

/*
	LocalVariableTable_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 local_variable_table_length;
	    {   u2 start_pc;
	        u2 length;
	        u2 name_index;
	        u2 descriptor_index;
	        u2 index;
	    } local_variable_table[local_variable_table_length];
	}
*/
type LocalVariableTableAttribute struct {
	LocalVariableTable []LocalVariableTableEntry
}

type LocalVariableTableEntry struct {
	StartPc         uint16
	Length          uint16
	NameIndex       uint16
	DescriptorIndex uint16
	Index           uint16
}

func readLocalVariableTableAttribute(reader *ClassReader) LocalVariableTableAttribute {
	return LocalVariableTableAttribute{
		LocalVariableTable: reader.readTable(func(reader *ClassReader) LocalVariableTableEntry {
			return LocalVariableTableEntry{
				StartPc:         reader.ReadUint16(),
				Length:          reader.ReadUint16(),
				NameIndex:       reader.ReadUint16(),
				DescriptorIndex: reader.ReadUint16(),
				Index:           reader.ReadUint16(),
			}
		}).([]LocalVariableTableEntry),
	}
}

// 写入 LocalVariableTableAttribute
func writeLocalVariableTableAttribute(writer *ClassWriter, data LocalVariableTableAttribute) []byte {
	var buf bytes.Buffer
	// 写入属性名称和属性长度
	//writeAttributeNameAndLength(writer, &buf, "LocalVariableTable", uint32(len(data.LocalVariableTable))*10)

	// 写入 local_variable_table 的长度
	buf.Write([]byte{byte(len(data.LocalVariableTable) >> 8), byte(len(data.LocalVariableTable))}) // local_variable_table_length

	// 写入每个 local variable table 项
	for _, entry := range data.LocalVariableTable {
		buf.Write([]byte{byte(entry.StartPc >> 8), byte(entry.StartPc)})
		buf.Write([]byte{byte(entry.Length >> 8), byte(entry.Length)})
		buf.Write([]byte{byte(entry.NameIndex >> 8), byte(entry.NameIndex)})
		buf.Write([]byte{byte(entry.DescriptorIndex >> 8), byte(entry.DescriptorIndex)})
		buf.Write([]byte{byte(entry.Index >> 8), byte(entry.Index)})
	}
	return buf.Bytes() // 返回字节切片
}

/*
	LocalVariableTypeTable_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 local_variable_type_table_length;
	    {   u2 start_pc;
	        u2 length;
	        u2 name_index;
	        u2 signature_index;
	        u2 index;
	    } local_variable_type_table[local_variable_type_table_length];
	}
*/
type LocalVariableTypeTableAttribute struct {
	LocalVariableTypeTable []LocalVariableTypeTableEntry
}

type LocalVariableTypeTableEntry struct {
	StartPc        uint16
	Length         uint16
	NameIndex      uint16
	SignatureIndex uint16
	Index          uint16
}

func readLocalVariableTypeTableAttribute(reader *ClassReader) LocalVariableTypeTableAttribute {
	return LocalVariableTypeTableAttribute{
		LocalVariableTypeTable: reader.readTable(func(reader *ClassReader) LocalVariableTypeTableEntry {
			return LocalVariableTypeTableEntry{
				StartPc:        reader.ReadUint16(),
				Length:         reader.ReadUint16(),
				NameIndex:      reader.ReadUint16(),
				SignatureIndex: reader.ReadUint16(),
				Index:          reader.ReadUint16(),
			}
		}).([]LocalVariableTypeTableEntry),
	}
}

// 写入 LocalVariableTypeTableAttribute
func writeLocalVariableTypeTableAttribute(writer *ClassWriter, data LocalVariableTypeTableAttribute) []byte {
	var buf bytes.Buffer
	// 写入属性名称和属性长度
	//writeAttributeNameAndLength(writer, &buf, "LocalVariableTypeTable", uint32(len(data.LocalVariableTypeTable))*10)

	// 写入 local_variable_type_table 的长度
	buf.Write([]byte{byte(len(data.LocalVariableTypeTable) >> 8), byte(len(data.LocalVariableTypeTable))}) // local_variable_type_table_length

	// 写入每个 local variable type table 项
	for _, entry := range data.LocalVariableTypeTable {
		buf.Write([]byte{byte(entry.StartPc >> 8), byte(entry.StartPc)})
		buf.Write([]byte{byte(entry.Length >> 8), byte(entry.Length)})
		buf.Write([]byte{byte(entry.NameIndex >> 8), byte(entry.NameIndex)})
		buf.Write([]byte{byte(entry.SignatureIndex >> 8), byte(entry.SignatureIndex)})
		buf.Write([]byte{byte(entry.Index >> 8), byte(entry.Index)})
	}
	return buf.Bytes() // 返回字节切片
}

// 写入属性名称和长度（修改为接受 *bytes.Buffer）
func writeAttributeNameAndLength(writer *ClassWriter, buf *bytes.Buffer, attributeName string, attributeLength uint32) {
	// 写入属性名称索引
	nameIndex := writer.cf.GetConstStrIndex(attributeName)
	if nameIndex == 0 {
		nameIndex = writer.cf.AddConstInfo([]uint8(attributeName))
	}
	buf.Write([]byte{byte(nameIndex >> 8), byte(nameIndex)})

	// 写入属性长度
	buf.Write([]byte{byte(attributeLength >> 24), byte(attributeLength >> 16), byte(attributeLength >> 8), byte(attributeLength)})
}
