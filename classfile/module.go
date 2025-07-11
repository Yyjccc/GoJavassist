package classfile

/*
Module_attribute {
    u2 attribute_name_index;
    u4 attribute_length;
    u2 module_name_index;
    u2 module_flags;
    u2 module_version_index;
    u2 requires_count;
    {   u2 requires_index;
        u2 requires_flags;
        u2 requires_version_index;
    } requires[requires_count];
    u2 exports_count;
    {   u2 exports_index;
        u2 exports_flags;
        u2 exports_to_count;
        u2 exports_to_index[exports_to_count];
    } exports[exports_count];
    u2 opens_count;
    {   u2 opens_index;
        u2 opens_flags;
        u2 opens_to_count;
        u2 opens_to_index[opens_to_count];
    } opens[opens_count];
    u2 uses_count;
    u2 uses_index[uses_count];
    u2 provides_count;
    {   u2 provides_index;
        u2 provides_with_count;
        u2 provides_with_index[provides_with_count];
    } provides[provides_count];
}
*/

import (
	"bytes"
)

type ModuleAttribute struct {
	ModuleNameIndex    uint16
	ModuleFlags        uint16
	ModuleVersionIndex uint16
	RequiresTable      []ModuleRequires
	ExportsTable       []ModuleExports
	OpensTable         []ModuleOpens
	UsesIndexTable     []uint16
	ProvidesTable      []ModuleProvides
}

type ModuleRequires struct {
	RequiresIndex        uint16
	RequiresFlags        uint16
	RequiresVersionIndex uint16
}

type ModuleExports struct {
	ExportsIndex        uint16
	ExportsFlags        uint16
	ExportsToIndexTable []uint16
}

type ModuleOpens struct {
	OpensIndex        uint16
	OpensFlags        uint16
	OpensToIndexTable []uint16
}

type ModuleProvides struct {
	ProvidesIndex          uint16
	ProvidesWithIndexTable []uint16
}

func readModuleAttribute(reader *ClassReader) ModuleAttribute {
	return ModuleAttribute{
		ModuleNameIndex:    reader.ReadUint16(),
		ModuleFlags:        reader.ReadUint16(),
		ModuleVersionIndex: reader.ReadUint16(),
		RequiresTable:      readRequiresTable(reader),
		ExportsTable:       readExportsTable(reader),
		OpensTable:         readOpensTable(reader),
		UsesIndexTable:     reader.readUint16s(),
		ProvidesTable:      readProvidesTable(reader),
	}
}

func writeModuleAttribute(data ModuleAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(data.ModuleNameIndex >> 8), byte(data.ModuleNameIndex)})       // ModuleNameIndex
	buf.Write([]byte{byte(data.ModuleFlags >> 8), byte(data.ModuleFlags)})               // ModuleFlags
	buf.Write([]byte{byte(data.ModuleVersionIndex >> 8), byte(data.ModuleVersionIndex)}) // ModuleVersionIndex

	// Write RequiresTable
	buf.Write(writeRequiresTable(data.RequiresTable))

	// Write ExportsTable
	buf.Write(writeExportsTable(data.ExportsTable))

	// Write OpensTable
	buf.Write(writeOpensTable(data.OpensTable))

	// Write UsesIndexTable
	buf.Write([]byte{byte(len(data.UsesIndexTable) >> 8), byte(len(data.UsesIndexTable))}) // UsesIndexTable length
	for _, usesIndex := range data.UsesIndexTable {
		buf.Write([]byte{byte(usesIndex >> 8), byte(usesIndex)}) // UsesIndex
	}

	// Write ProvidesTable
	buf.Write(writeProvidesTable(data.ProvidesTable))
	return buf.Bytes() // 返回字节切片
}

func writeRequiresTable(requiresTable []ModuleRequires) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(requiresTable) >> 8), byte(len(requiresTable))}) // RequiresTable length
	for _, require := range requiresTable {
		buf.Write([]byte{byte(require.RequiresIndex >> 8), byte(require.RequiresIndex)})               // RequiresIndex
		buf.Write([]byte{byte(require.RequiresFlags >> 8), byte(require.RequiresFlags)})               // RequiresFlags
		buf.Write([]byte{byte(require.RequiresVersionIndex >> 8), byte(require.RequiresVersionIndex)}) // RequiresVersionIndex
	}
	return buf.Bytes() // 返回字节切片
}

func writeExportsTable(exportsTable []ModuleExports) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(exportsTable) >> 8), byte(len(exportsTable))}) // ExportsTable length
	for _, export := range exportsTable {
		buf.Write([]byte{byte(export.ExportsIndex >> 8), byte(export.ExportsIndex)})                         // ExportsIndex
		buf.Write([]byte{byte(export.ExportsFlags >> 8), byte(export.ExportsFlags)})                         // ExportsFlags
		buf.Write([]byte{byte(len(export.ExportsToIndexTable) >> 8), byte(len(export.ExportsToIndexTable))}) // ExportsToIndexTable length
		for _, toIndex := range export.ExportsToIndexTable {
			buf.Write([]byte{byte(toIndex >> 8), byte(toIndex)}) // ExportsToIndex
		}
	}
	return buf.Bytes() // 返回字节切片
}

func writeOpensTable(opensTable []ModuleOpens) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(opensTable) >> 8), byte(len(opensTable))}) // OpensTable length
	for _, open := range opensTable {
		buf.Write([]byte{byte(open.OpensIndex >> 8), byte(open.OpensIndex)})                         // OpensIndex
		buf.Write([]byte{byte(open.OpensFlags >> 8), byte(open.OpensFlags)})                         // OpensFlags
		buf.Write([]byte{byte(len(open.OpensToIndexTable) >> 8), byte(len(open.OpensToIndexTable))}) // OpensToIndexTable length
		for _, toIndex := range open.OpensToIndexTable {
			buf.Write([]byte{byte(toIndex >> 8), byte(toIndex)}) // OpensToIndex
		}
	}
	return buf.Bytes() // 返回字节切片
}

func writeProvidesTable(providesTable []ModuleProvides) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(providesTable) >> 8), byte(len(providesTable))}) // ProvidesTable length
	for _, provide := range providesTable {
		buf.Write([]byte{byte(provide.ProvidesIndex >> 8), byte(provide.ProvidesIndex)})                             // ProvidesIndex
		buf.Write([]byte{byte(len(provide.ProvidesWithIndexTable) >> 8), byte(len(provide.ProvidesWithIndexTable))}) // ProvidesWithIndexTable length
		for _, withIndex := range provide.ProvidesWithIndexTable {
			buf.Write([]byte{byte(withIndex >> 8), byte(withIndex)}) // ProvidesWithIndex
		}
	}
	return buf.Bytes() // 返回字节切片
}

func readRequiresTable(reader *ClassReader) []ModuleRequires {
	return reader.readTable(func(reader *ClassReader) ModuleRequires {
		return ModuleRequires{
			RequiresIndex:        reader.ReadUint16(),
			RequiresFlags:        reader.ReadUint16(),
			RequiresVersionIndex: reader.ReadUint16(),
		}
	}).([]ModuleRequires)
}

func readExportsTable(reader *ClassReader) []ModuleExports {
	return reader.readTable(func(reader *ClassReader) ModuleExports {
		return ModuleExports{
			ExportsIndex:        reader.ReadUint16(),
			ExportsFlags:        reader.ReadUint16(),
			ExportsToIndexTable: reader.readUint16s(),
		}
	}).([]ModuleExports)
}

func readOpensTable(reader *ClassReader) []ModuleOpens {
	return reader.readTable(func(reader *ClassReader) ModuleOpens {
		return ModuleOpens{
			OpensIndex:        reader.ReadUint16(),
			OpensFlags:        reader.ReadUint16(),
			OpensToIndexTable: reader.readUint16s(),
		}
	}).([]ModuleOpens)
}

func readProvidesTable(reader *ClassReader) []ModuleProvides {
	return reader.readTable(func(reader *ClassReader) ModuleProvides {
		return ModuleProvides{
			ProvidesIndex:          reader.ReadUint16(),
			ProvidesWithIndexTable: reader.readUint16s(),
		}
	}).([]ModuleProvides)
}
