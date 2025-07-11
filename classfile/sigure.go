package classfile

import (
	"bytes"
)

/*
	Signature_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 signature_index;
	}
*/
type SignatureAttribute struct {
	SignatureIndex uint16
}

func readSignatureAttribute(reader *ClassReader) SignatureAttribute {
	return SignatureAttribute{
		SignatureIndex: reader.ReadUint16(),
	}
}

func writeSignatureAttribute(writer *ClassWriter, data SignatureAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(data.SignatureIndex >> 8), byte(data.SignatureIndex)}) // SignatureIndex
	return buf.Bytes()                                                           // 返回字节切片
}

/*
	SourceFile_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 sourcefile_index;
	}
*/
type SourceFileAttribute struct {
	SourceFileIndex uint16
}

func readSourceFileAttribute(reader *ClassReader) SourceFileAttribute {
	return SourceFileAttribute{SourceFileIndex: reader.ReadUint16()}
}

func writeSourceFileAttribute(writer *ClassWriter, data SourceFileAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(data.SourceFileIndex >> 8), byte(data.SourceFileIndex)}) // SourceFileIndex
	return buf.Bytes()                                                             // 返回字节切片
}
