package classfile

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf16"
	"unsafe"
)

// Java 基本类型对应的 Go 类型别名
type JByte int8
type JShort int16
type JInt int32
type JLong int64
type JFloat float32
type JDouble float64
type JChar rune
type JBoolean bool

func DotToSlash(name string) string {
	return strings.ReplaceAll(name, ".", "/")
}
func SlashToDot(name string) string {
	return strings.ReplaceAll(name, "/", ".")
}

func UTF8ToUTF16(s string) []uint16 {
	runes := []rune(s)
	return utf16.Encode(runes) // func Encode(s []rune) []uint16
}

func UTF16ToUTF8(s []uint16) string {
	runes := utf16.Decode(s) // func Decode(s []uint16) []rune
	return string(runes)
}

// mutf8 -> utf16 -> utf32 -> string
// see java.io.DataInputStream.readUTF(DataInput)
func DecodeMUTF8(bytearr []byte) string {
	utflen := len(bytearr)
	chararr := make([]uint16, utflen)

	var c, char2, char3 uint16
	count := 0
	chararr_count := 0

	for count < utflen {
		c = uint16(bytearr[count])
		if c > 127 {
			break
		}
		count++
		chararr[chararr_count] = c
		chararr_count++
	}

	for count < utflen {
		c = uint16(bytearr[count])
		switch c >> 4 {
		case 0, 1, 2, 3, 4, 5, 6, 7:
			/* 0xxxxxxx*/
			count++
			chararr[chararr_count] = c
			chararr_count++
		case 12, 13:
			/* 110x xxxx   10xx xxxx*/
			count += 2
			if count > utflen {
				panic("malformed input: partial character at end")
			}
			char2 = uint16(bytearr[count-1])
			if char2&0xC0 != 0x80 {
				panic(fmt.Errorf("malformed input around byte %v", count))
			}
			chararr[chararr_count] = c&0x1F<<6 | char2&0x3F
			chararr_count++
		case 14:
			/* 1110 xxxx  10xx xxxx  10xx xxxx*/
			count += 3
			if count > utflen {
				panic("malformed input: partial character at end")
			}
			char2 = uint16(bytearr[count-2])
			char3 = uint16(bytearr[count-1])
			if char2&0xC0 != 0x80 || char3&0xC0 != 0x80 {
				panic(fmt.Errorf("malformed input around byte %v", count-1))
			}
			chararr[chararr_count] = c&0x0F<<12 | char2&0x3F<<6 | char3&0x3F<<0
			chararr_count++
		default:
			/* 10xx xxxx,  1111 xxxx */
			panic(fmt.Errorf("malformed input around byte %v", count))
		}
	}
	// The number of chars produced may be less than utflen
	chararr = chararr[0:chararr_count]
	runes := utf16.Decode(chararr)
	return string(runes)
}

// []int8 <-> []byte
func CastInt8sToBytes(s []int8) []byte {
	ptr := unsafe.Pointer(&s)
	return *((*[]byte)(ptr))
}
func CastBytesToInt8s(s []byte) []int8 {
	ptr := unsafe.Pointer(&s)
	return *((*[]int8)(ptr))
}

// []int8 <-> []uint16
func CastInt8sToUint16s(s []int8) []uint16 {
	ptr := unsafe.Pointer(&s)
	(*reflect.SliceHeader)(ptr).Len /= 2
	return *((*[]uint16)(ptr))
}
func CastUint16sToInt8s(s []uint16) []int8 {
	ptr := unsafe.Pointer(&s)
	(*reflect.SliceHeader)(ptr).Len *= 2
	return *((*[]int8)(ptr))
}

func CastBytesToUint32s(s []byte) []uint32 {
	ptr := unsafe.Pointer(&s)
	(*reflect.SliceHeader)(ptr).Len /= 4
	return *((*[]uint32)(ptr))
}

func CastBytesToInt32s(s []byte) []int32 {
	ptr := unsafe.Pointer(&s)
	(*reflect.SliceHeader)(ptr).Len /= 4
	return *((*[]int32)(ptr))
}
