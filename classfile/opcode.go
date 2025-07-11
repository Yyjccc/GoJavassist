package classfile

import (
	"bytes"
	"encoding/binary"
)

type OpCode byte

const (
	OpNop             = OpCode(0x00)
	OpAConstNull      = OpCode(0x01)
	OpIConstM1        = OpCode(0x02)
	OpIConst0         = OpCode(0x03)
	OpIConst1         = OpCode(0x04)
	OpIConst2         = OpCode(0x05)
	OpIConst3         = OpCode(0x06)
	OpIConst4         = OpCode(0x07)
	OpIConst5         = OpCode(0x08)
	OpLConst0         = OpCode(0x09)
	OpLConst1         = OpCode(0x0a)
	OpFConst0         = OpCode(0x0b)
	OpFConst1         = OpCode(0x0c)
	OpFConst2         = OpCode(0x0d)
	OpDConst0         = OpCode(0x0e)
	OpDConst1         = OpCode(0x0f)
	OpBIPush          = OpCode(0x10)
	OpSIPush          = OpCode(0x11)
	OpLDC             = OpCode(0x12)
	OpLDCw            = OpCode(0x13)
	OpLDC2w           = OpCode(0x14)
	OpILoad           = OpCode(0x15)
	OpLLoad           = OpCode(0x16)
	OpFLoad           = OpCode(0x17)
	OpDLoad           = OpCode(0x18)
	OpALoad           = OpCode(0x19)
	OpILoad0          = OpCode(0x1a)
	OpILoad1          = OpCode(0x1b)
	OpILoad2          = OpCode(0x1c)
	OpILoad3          = OpCode(0x1d)
	OpLLoad0          = OpCode(0x1e)
	OpLLoad1          = OpCode(0x1f)
	OpLLoad2          = OpCode(0x20)
	OpLLoad3          = OpCode(0x21)
	OpFLoad0          = OpCode(0x22)
	OpFLoad1          = OpCode(0x23)
	OpFLoad2          = OpCode(0x24)
	OpFLoad3          = OpCode(0x25)
	OpDLoad0          = OpCode(0x26)
	OpDLoad1          = OpCode(0x27)
	OpDLoad2          = OpCode(0x28)
	OpDLoad3          = OpCode(0x29)
	OpALoad0          = OpCode(0x2a)
	OpALoad1          = OpCode(0x2b)
	OpALoad2          = OpCode(0x2c)
	OpALoad3          = OpCode(0x2d)
	OpIALoad          = OpCode(0x2e)
	OpLALoad          = OpCode(0x2f)
	OpFALoad          = OpCode(0x30)
	OpDALoad          = OpCode(0x31)
	OpAALoad          = OpCode(0x32)
	OpBALoad          = OpCode(0x33)
	OpCALoad          = OpCode(0x34)
	OpSALoad          = OpCode(0x35)
	OpIStore          = OpCode(0x36)
	OpLStore          = OpCode(0x37)
	OpFStore          = OpCode(0x38)
	OpDStore          = OpCode(0x39)
	OpAStore          = OpCode(0x3a)
	OpIStore0         = OpCode(0x3b)
	OpIStore1         = OpCode(0x3c)
	OpIStore2         = OpCode(0x3d)
	OpIStore3         = OpCode(0x3e)
	OpLStore0         = OpCode(0x3f)
	OpLStore1         = OpCode(0x40)
	OpLStore2         = OpCode(0x41)
	OpLStore3         = OpCode(0x42)
	OpFStore0         = OpCode(0x43)
	OpFStore1         = OpCode(0x44)
	OpFStore2         = OpCode(0x45)
	OpFStore3         = OpCode(0x46)
	OpDStore0         = OpCode(0x47)
	OpDStore1         = OpCode(0x48)
	OpDStore2         = OpCode(0x49)
	OpDStore3         = OpCode(0x4a)
	OpAStore0         = OpCode(0x4b)
	OpAStore1         = OpCode(0x4c)
	OpAStore2         = OpCode(0x4d)
	OpAStore3         = OpCode(0x4e)
	OpIAStore         = OpCode(0x4f)
	OpLAStore         = OpCode(0x50)
	OpFAStore         = OpCode(0x51)
	OpDAStore         = OpCode(0x52)
	OpAAStore         = OpCode(0x53)
	OpBAStore         = OpCode(0x54)
	OpCAStore         = OpCode(0x55)
	OpSAStore         = OpCode(0x56)
	OpPop             = OpCode(0x57)
	OpPop2            = OpCode(0x58)
	OpDup             = OpCode(0x59)
	OpDupX1           = OpCode(0x5a)
	OpDupX2           = OpCode(0x5b)
	OpDup2            = OpCode(0x5c)
	OpDup2X1          = OpCode(0x5d)
	OpDup2X2          = OpCode(0x5e)
	OpSwap            = OpCode(0x5f)
	OpIAdd            = OpCode(0x60)
	OpLAdd            = OpCode(0x61)
	OpFAdd            = OpCode(0x62)
	OpDAdd            = OpCode(0x63)
	OpISub            = OpCode(0x64)
	OpLSub            = OpCode(0x65)
	OpFSub            = OpCode(0x66)
	OpDSub            = OpCode(0x67)
	OpIMul            = OpCode(0x68)
	OpLMul            = OpCode(0x69)
	OpFMul            = OpCode(0x6a)
	OpDMul            = OpCode(0x6b)
	OpIDiv            = OpCode(0x6c)
	OpLDiv            = OpCode(0x6d)
	OpFDiv            = OpCode(0x6e)
	OpDDiv            = OpCode(0x6f)
	OpIRem            = OpCode(0x70)
	OpLRem            = OpCode(0x71)
	OpFRem            = OpCode(0x72)
	OpDRem            = OpCode(0x73)
	OpINeg            = OpCode(0x74)
	OpLNeg            = OpCode(0x75)
	OpFNeg            = OpCode(0x76)
	OpDNeg            = OpCode(0x77)
	OpIShl            = OpCode(0x78)
	OpLShl            = OpCode(0x79)
	OpIShr            = OpCode(0x7a)
	OpLShr            = OpCode(0x7b)
	OpIUshr           = OpCode(0x7c)
	OpLUshr           = OpCode(0x7d)
	OpIAnd            = OpCode(0x7e)
	OpLAnd            = OpCode(0x7f)
	OpIOr             = OpCode(0x80)
	OpLOr             = OpCode(0x81)
	OpIXor            = OpCode(0x82)
	OpLXor            = OpCode(0x83)
	OpIInc            = OpCode(0x84)
	OpI2L             = OpCode(0x85)
	OpI2F             = OpCode(0x86)
	OpI2D             = OpCode(0x87)
	OpL2I             = OpCode(0x88)
	OpL2F             = OpCode(0x89)
	OpL2D             = OpCode(0x8a)
	OpF2I             = OpCode(0x8b)
	OpF2L             = OpCode(0x8c)
	OpF2D             = OpCode(0x8d)
	OpD2I             = OpCode(0x8e)
	OpD2L             = OpCode(0x8f)
	OpD2F             = OpCode(0x90)
	OpI2B             = OpCode(0x91)
	OpI2C             = OpCode(0x92)
	OpI2S             = OpCode(0x93)
	OpLCmp            = OpCode(0x94)
	OpFCmpL           = OpCode(0x95)
	OpFCmpG           = OpCode(0x96)
	OpDCmpL           = OpCode(0x97)
	OpDCmpG           = OpCode(0x98)
	OpIfEQ            = OpCode(0x99)
	OpIfNE            = OpCode(0x9a)
	OpIfLT            = OpCode(0x9b)
	OpIfGE            = OpCode(0x9c)
	OpIfGT            = OpCode(0x9d)
	OpIfLE            = OpCode(0x9e)
	OpIfICmpEQ        = OpCode(0x9f)
	OpIfICmpNE        = OpCode(0xa0)
	OpIfICmpLT        = OpCode(0xa1)
	OpIfICmpGE        = OpCode(0xa2)
	OpIfICmpGT        = OpCode(0xa3)
	OpIfICmpLE        = OpCode(0xa4)
	OpIfACmpEQ        = OpCode(0xa5)
	OpIfACmpNE        = OpCode(0xa6)
	OpGoto            = OpCode(0xa7)
	OpJSR             = OpCode(0xa8)
	OpRET             = OpCode(0xa9)
	OpTableSwitch     = OpCode(0xaa)
	OpLookupSwitch    = OpCode(0xab)
	OpIReturn         = OpCode(0xac)
	OpLReturn         = OpCode(0xad)
	OpFReturn         = OpCode(0xae)
	OpDReturn         = OpCode(0xaf)
	OpAReturn         = OpCode(0xb0)
	OpReturn          = OpCode(0xb1)
	OpGetStatic       = OpCode(0xb2)
	OpPutStatic       = OpCode(0xb3)
	OpGetField        = OpCode(0xb4)
	OpPutField        = OpCode(0xb5)
	OpInvokeVirtual   = OpCode(0xb6)
	OpInvokeSpecial   = OpCode(0xb7)
	OpInvokeStatic    = OpCode(0xb8)
	OpInvokeInterface = OpCode(0xb9)
	OpInvokeDynamic   = OpCode(0xba)
	OpNew             = OpCode(0xbb)
	OpNewArray        = OpCode(0xbc)
	OpANewArray       = OpCode(0xbd)
	OpArrayLength     = OpCode(0xbe)
	OpAThrow          = OpCode(0xbf)
	OpCheckCast       = OpCode(0xc0)
	OpInstanceOf      = OpCode(0xc1)
	OpMonitorEnter    = OpCode(0xc2)
	OpMonitorExit     = OpCode(0xc3)
	OpWide            = OpCode(0xc4)
	OpMultiANewArray  = OpCode(0xc5)
	OpIfNull          = OpCode(0xc6)
	OpIfNonNull       = OpCode(0xc7)
	OpGotoW           = OpCode(0xc8)
	OpJSRw            = OpCode(0xc9)
	OpBreakpoint      = OpCode(0xca)
	OpInvokeNative    = OpCode(0xfe) // impdep1
	OpBootstrap       = OpCode(0xff) // impdep2
)

func ReadOpCode(opcode byte) OpCode {
	return OpCode(opcode)
}

// Instruction 表示一个 JVM 指令
type Instruction struct {
	Opcode   OpCode // 操作码
	Operands []byte // 操作数
}

func NewInstruction(op OpCode) Instruction {
	return Instruction{
		Opcode:   OpCode(op),
		Operands: []byte{},
	}
}

func (i *Instruction) PutUint8(u uint8) Instruction {
	i.Operands = append(i.Operands, byte(u))
	return *i
}
func (i *Instruction) PutUint16(u uint16) Instruction {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], u)
	i.Operands = append(i.Operands, buf[0], buf[1])
	return *i
}
func (i Instruction) Add(index int) Instruction {
	ptr := &i
	ptr.PutUint8(uint8(index))
	return i
}

func (i Instruction) AddIndex(index int) Instruction {
	ptr := &i
	ptr.PutUint16(uint16(index))
	return i
}

// 返回指令长度
func (i *Instruction) Length() int16 {
	return int16(len(i.Operands)) + 1
}

func GetOperandLength(opcode OpCode) int {
	// 根据操作码决定操作数的长度
	switch opcode {
	case OpBIPush:
		return 1 // 一个字节，代表常量值
	case OpSIPush:
		return 2 // 两个字节，代表常量值
	case OpLDC:
		return 1 // 一个字节，常量池索引
	case OpRET, OpNewArray:
		return 1

	case OpLDCw, OpLDC2w:
		return 2 // 两个字节，常量池索引
	case OpILoad, OpLLoad, OpFLoad, OpDLoad, OpALoad:
		return 1 // 一个字节，局部变量索引
	case OpILoad0, OpILoad1, OpILoad2, OpILoad3,
		OpLLoad0, OpLLoad1, OpLLoad2, OpLLoad3,
		OpFLoad0, OpFLoad1, OpFLoad2, OpFLoad3,
		OpDLoad0, OpDLoad1, OpDLoad2, OpDLoad3,
		OpALoad0, OpALoad1, OpALoad2, OpALoad3:
		return 0 // 不需要操作数（直接加载特定局部变量）
	case OpIStore, OpLStore, OpFStore, OpDStore, OpAStore:
		return 1 // 一个字节，局部变量索引
	case OpIStore0, OpIStore1, OpIStore2, OpIStore3,
		OpLStore0, OpLStore1, OpLStore2, OpLStore3,
		OpFStore0, OpFStore1, OpFStore2, OpFStore3,
		OpDStore0, OpDStore1, OpDStore2, OpDStore3,
		OpAStore0, OpAStore1, OpAStore2, OpAStore3:
		return 0 // 不需要操作数（直接存储特定局部变量）
	case OpGoto, OpJSR:
		return 2 // 跳转指令需要一个 2 字节偏移量
	case OpInvokeVirtual, OpInvokeSpecial, OpInvokeStatic:
		return 2 // 方法引用索引（2字节），额外的参数等
	case OpInvokeInterface:
		return 4 // Opcode | index (2 bytes) | count (1 byte) | zero (1 byte)
	case OpIfEQ, OpIfNE, OpIfLT, OpIfGE, OpIfGT, OpIfLE, OpIfNull:
		return 2 // 两个字节，偏移量
	case OpIfICmpEQ, OpIfICmpNE, OpIfICmpLT, OpIfICmpGE, OpIfICmpGT, OpIfICmpLE:
		return 2 // 两个字节，偏移量
	case OpIfACmpEQ, OpIfACmpNE:
		return 2 // 两个字节，偏移量
	case OpLookupSwitch, OpTableSwitch:
		return -1 // 表或查找表的长度是动态的，额外处理需要
	case OpInvokeDynamic:
		return 2 // 2 字节的操作数
	case OpMultiANewArray:
		return 2 // 1 字节的维度 + 1 字节的类型索引
	case OpNew, OpANewArray, OpCheckCast, OpInstanceOf:
		return 2 // 1 字节的类型索引
	case OpGetStatic, OpPutStatic, OpGetField, OpPutField:
		return 2 // 2 字节的字段引用
	case OpIInc:
		return 2 // 1 字节的局部变量索引 + 1 字节的增量
	case OpWide:
		return 2 // 可能需要扩展的操作数
	case OpPop, OpPop2, OpDup, OpDupX1, OpDupX2, OpDup2, OpDup2X1, OpDup2X2, OpSwap:
		return 0 // 不需要操作数
	case OpIAdd, OpLAdd, OpFAdd, OpDAdd:
		return 0 // 不需要操作数
	case OpISub, OpLSub, OpFSub, OpDSub:
		return 0 // 不需要操作数
	case OpIMul, OpLMul, OpFMul, OpDMul:
		return 0 // 不需要操作数
	case OpIDiv, OpLDiv, OpFDiv, OpDDiv:
		return 0 // 不需要操作数
	case OpIRem, OpLRem, OpFRem, OpDRem:
		return 0 // 不需要操作数
	case OpINeg, OpLNeg, OpFNeg, OpDNeg:
		return 0 // 不需要操作数
	case OpIShl, OpLShl, OpIShr, OpLShr, OpIUshr, OpLUshr:
		return 0 // 不需要操作数
	case OpIAnd, OpLAnd, OpIOr, OpLOr, OpIXor, OpLXor:
		return 0 // 不需要操作数
	case OpI2L, OpI2F, OpI2D, OpL2I, OpL2F, OpL2D, OpF2I, OpF2L, OpF2D, OpD2I, OpD2L, OpD2F:
		return 0 // 不需要操作数
	case OpLCmp, OpFCmpL, OpFCmpG, OpDCmpL, OpDCmpG:
		return 0 // 不需要操作数
	case OpArrayLength, OpAThrow:
		return 0
	default:
		return 0 // 不需要操作数
	}
}

// IsJumpInstruction 判断是否为跳转指令
func IsJumpInstruction(opcode OpCode) bool {
	switch opcode {
	case OpGoto, OpGotoW, OpJSR, OpJSRw,
		OpIfEQ, OpIfNE, OpIfLT, OpIfGE, OpIfGT, OpIfLE,
		OpIfICmpEQ, OpIfICmpNE, OpIfICmpLT, OpIfICmpGE, OpIfICmpGT, OpIfICmpLE,
		OpIfACmpEQ, OpIfACmpNE,
		OpIfNull, OpIfNonNull:
		return true
	default:
		return false
	}
}

// 转为byte 数组
func InstructionsToBytes(instructions []Instruction) []byte {
	var buf bytes.Buffer
	for _, inst := range instructions {
		buf.WriteByte(byte(inst.Opcode)) // 写入 Opcode
		buf.Write(inst.Operands)         // 写入 Operands
	}
	return buf.Bytes()
}

func (i Instruction) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(i.Opcode)) // 写入 Opcode
	buf.Write(i.Operands)         // 写入 Operands
	return buf.Bytes()
}

// 添加偏移量处理
func (i Instruction) AddOffset(offset int) Instruction {
	return i.PutUint16(uint16(offset))
}

//// 添加带宽指令处理
//func (i Instruction) AddWide(index, constValue int) Instruction {
//	return i.PutUint16(uint16(index)).PutUint16(uint16(constValue))
//}

// 添加类型索引处理
func (i Instruction) AddTypeIndex(index int) Instruction {
	return i.PutUint16(uint16(index))
}

// 添加方法索引处理
func (i Instruction) AddMethodIndex(index int) Instruction {
	return i.PutUint16(uint16(index))
}
