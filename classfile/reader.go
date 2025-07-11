package classfile

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

type ClassReader struct {
	BytesReader
	cf *ClassFile
}

func NewClassReader(data []byte) ClassReader {
	br := NewBytesReader(data, binary.BigEndian)
	return ClassReader{BytesReader: br}
}

func (reader *ClassReader) readUint16s() []uint16 {
	n := reader.ReadUint16()
	s := make([]uint16, n)
	for i := range s {
		s[i] = reader.ReadUint16()
	}
	return s
}

// readFn: func(reader *ClassReader) XXX
func (reader *ClassReader) readTable(readFn interface{}) interface{} {
	n := int(reader.ReadUint16())

	itemType := reflect.TypeOf(readFn).Out(0)
	sliceType := reflect.SliceOf(itemType)
	s := reflect.MakeSlice(sliceType, n, n) // make([]x, n, n)

	readFnVal := reflect.ValueOf(readFn)
	args := []reflect.Value{reflect.ValueOf(reader)}

	for i := 0; i < n; i++ {
		x := readFnVal.Call(args)[0]
		s.Index(i).Set(x) // s[i] = x
	}

	return s.Interface()
}

type BytesReader struct {
	byteOrder binary.ByteOrder
	data      []byte
	position  int
}

func NewBytesReader(data []byte, byteOrder binary.ByteOrder) BytesReader {
	return BytesReader{
		byteOrder: byteOrder,
		data:      data,
		position:  0,
	}
}

func (reader *BytesReader) Position() int {
	return reader.position
}

func (reader *BytesReader) ReadUint8() uint8 {
	i := reader.data[reader.position]
	reader.position++
	return i
}

func (reader *BytesReader) ReadUint16() uint16 {
	i := reader.byteOrder.Uint16(reader.data[reader.position:])
	reader.position += 2
	return i
}

func (reader *BytesReader) ReadUint32() uint32 {
	i := reader.byteOrder.Uint32(reader.data[reader.position:])
	reader.position += 4
	return i
}

func (reader *BytesReader) ReadUint64() uint64 {
	i := reader.byteOrder.Uint64(reader.data[reader.position:])
	reader.position += 8
	return i
}

func (reader *BytesReader) ReadBytes(n int) []byte {
	bytes := reader.data[reader.position : reader.position+n]
	reader.position += n
	return bytes
}

func Parse(classData []byte) (cf *ClassFile, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	cr := NewClassReader(classData)
	cf = &ClassFile{}
	cf.Read(&cr)
	return
}
