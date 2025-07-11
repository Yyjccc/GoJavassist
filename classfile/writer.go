package classfile

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type BytesWriter struct {
	byteOrder binary.ByteOrder
	buffer    *bytes.Buffer
}

func NewBytesWriter(byteOrder binary.ByteOrder) BytesWriter {
	return BytesWriter{
		byteOrder: byteOrder,
		buffer:    new(bytes.Buffer),
	}
}

func (writer *BytesWriter) WriteUint8(i uint8) {
	writer.buffer.WriteByte(i)
}

func (writer *BytesWriter) WriteUint16(i uint16) {
	b := make([]byte, 2)
	writer.byteOrder.PutUint16(b, i)
	writer.buffer.Write(b)
}

func (writer *ClassWriter) WriteUint16s(values []uint16) {
	for _, value := range values {
		writer.WriteUint16(value)
	}
}

func (writer *BytesWriter) WriteUint32(i uint32) {
	b := make([]byte, 4)
	writer.byteOrder.PutUint32(b, i)
	writer.buffer.Write(b)
}

func (writer *BytesWriter) WriteUint64(i uint64) {
	b := make([]byte, 8)
	writer.byteOrder.PutUint64(b, i)
	writer.buffer.Write(b)
}

func (writer *BytesWriter) WriteBytes(data []byte) {
	writer.buffer.Write(data)
}

func (writer *BytesWriter) Bytes() []byte {
	return writer.buffer.Bytes()
}

type ClassWriter struct {
	BytesWriter
	cf *ClassFile
}

func NewClassWriter(cf *ClassFile) *ClassWriter {
	bw := NewBytesWriter(binary.BigEndian)
	return &ClassWriter{BytesWriter: bw, cf: cf}
}

func (writer *ClassWriter) writeUint16s(data []uint16) {
	writer.WriteUint16(uint16(len(data)))
	for _, v := range data {
		writer.WriteUint16(v)
	}
}

// writeFn: func(writer *ClassWriter, item XXX)
func (writer *ClassWriter) writeTable(writeFn interface{}, data interface{}) {
	sliceVal := reflect.ValueOf(data)
	writer.WriteUint16(uint16(sliceVal.Len()))

	writeFnVal := reflect.ValueOf(writeFn)

	for i := 0; i < sliceVal.Len(); i++ {
		writeFnVal.Call([]reflect.Value{reflect.ValueOf(writer), sliceVal.Index(i)})
	}
}

func (writer *ClassWriter) Bytes() []byte {
	return writer.BytesWriter.Bytes()
}
