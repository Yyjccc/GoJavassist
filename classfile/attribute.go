package classfile

import (
	"bytes"
	"fmt"
	"reflect"
)

/*
	attribute_info {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u1 info[attribute_length];
	}
*/

// Predefined class file attributes
const (
	ConstantValue                       = "ConstantValue"                        //	1.0.2
	Code                                = "Code"                                 //	1.0.2
	Exceptions                          = "Exceptions"                           //	1.0.2
	SourceFile                          = "SourceFile"                           //	1.0.2
	LineNumberTable                     = "LineNumberTable"                      //	1.0.2
	LocalVariableTable                  = "LocalVariableTable"                   //	1.0.2
	InnerClasses                        = "InnerClasses"                         //	1.1
	Synthetic                           = "Synthetic"                            //	1.1
	Deprecated                          = "Deprecated"                           //	1.1
	EnclosingMethod                     = "EnclosingMethod"                      //	5.0
	Signature                           = "Signature"                            //	5.0
	SourceDebugExtension                = "SourceDebugExtension"                 //	5.0
	LocalVariableTypeTable              = "LocalVariableTypeTable"               //	5.0
	RuntimeVisibleAnnotations           = "RuntimeVisibleAnnotations"            //	5.0
	RuntimeInvisibleAnnotations         = "RuntimeInvisibleAnnotations"          //	5.0
	RuntimeVisibleParameterAnnotations  = "RuntimeVisibleParameterAnnotations"   //	5.0
	RuntimeInvisibleParameterAnnotation = "RuntimeInvisibleParameterAnnotations" //	5.0
	AnnotationDefault                   = "AnnotationDefault"                    //	5.0
	StackMapTable                       = "StackMapTable"                        //	6
	BootstrapMethods                    = "BootstrapMethods"                     //	7
	RuntimeVisibleTypeAnnotations       = "RuntimeVisibleTypeAnnotations"        //	8
	RuntimeInvisibleTypeAnnotations     = "RuntimeInvisibleTypeAnnotations"      //	8
	MethodParameters                    = "MethodParameters"                     //	8
	Module                              = "Module"                               // 9
	ModulePackages                      = "ModulePackages"                       // 9
	ModuleMainClass                     = "ModuleMainClass"                      // 9
	NestHost                            = "NestHost"                             // 11
	NestMembers                         = "NestMembers"                          // 11
)

type AttributeInfo interface{}

type UnparsedAttribute struct {
	Name   string
	Length uint32
	Info   []byte
}

type MarkerAttribute struct{}

/*
	Deprecated_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	}
*/
type DeprecatedAttribute struct {
	MarkerAttribute
}

/*
	Synthetic_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	}
*/
type SyntheticAttribute struct {
	MarkerAttribute
}

func writeUnparsedAttribute(writer *ClassWriter, attr UnparsedAttribute) []byte {
	var buf bytes.Buffer
	nameIndex := writer.cf.GetConstStrIndex(attr.Name)
	buf.Write([]byte{byte(nameIndex >> 8), byte(nameIndex)})                                                       // 属性名称索引
	buf.Write([]byte{byte(attr.Length >> 24), byte(attr.Length >> 16), byte(attr.Length >> 8), byte(attr.Length)}) // 属性长度
	buf.Write(attr.Info)                                                                                           // 属性信息
	return buf.Bytes()                                                                                             // 返回字节切片
}

func readAttributes(reader *ClassReader) []AttributeInfo {
	return reader.readTable(readAttributeInfo).([]AttributeInfo)
}

func writeMarkerAttribute(attr MarkerAttribute, attributeIndex uint16) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(attributeIndex >> 8), byte(attributeIndex)}) // 属性名称索引
	buf.Write([]byte{0, 0, 0, 0})                                      // 属性长度为0
	return buf.Bytes()                                                 // 返回字节切片
}

func writeDeprecatedAttribute(attr DeprecatedAttribute) []byte {
	const DeprecatedIndex = 0x0001                                     // 示例索引值
	return writeMarkerAttribute(attr.MarkerAttribute, DeprecatedIndex) // 返回字节切片
}

func writeSyntheticAttribute(attr SyntheticAttribute) []byte {
	const SyntheticIndex = 0x0002                                     // 示例索引值
	return writeMarkerAttribute(attr.MarkerAttribute, SyntheticIndex) // 返回字节切片
}

func readAttributeInfo(reader *ClassReader) AttributeInfo {
	attrNameIndex := reader.ReadUint16()
	attrLen := reader.ReadUint32()
	attrName := reader.cf.GetRawUTF8(attrNameIndex)

	switch attrName {
	// case AnnotationDefault:
	case BootstrapMethods:
		return readBootstrapMethodsAttribute(reader)
	case Code:
		return readCodeAttribute(reader)
	case ConstantValue:
		return readConstantValueAttribute(reader)
	case Deprecated:
		return DeprecatedAttribute{}
	case EnclosingMethod:
		return readEnclosingMethodAttribute(reader)
	case Exceptions:
		return readExceptionsAttribute(reader)
	case InnerClasses:
		return readInnerClassesAttribute(reader)
	case LineNumberTable:
		return readLineNumberTableAttribute(reader)
	case LocalVariableTable:
		return readLocalVariableTableAttribute(reader)
	case LocalVariableTypeTable:
		return readLocalVariableTypeTableAttribute(reader)
	// case MethodParameters:
	case Module:
		return readModuleAttribute(reader)
	// case RuntimeInvisibleAnnotations:
	// case RuntimeInvisibleParameterAnnotations:
	// case RuntimeInvisibleTypeAnnotations:
	// case RuntimeVisibleAnnotations:
	// case RuntimeVisibleParameterAnnotations:
	// case RuntimeVisibleTypeAnnotations:
	case Signature:
		return readSignatureAttribute(reader)
	case SourceFile:
		return readSourceFileAttribute(reader)
		// case SourceDebugExtension:
	case StackMapTable:
		return readStackMapTableAttribute(reader)
	case Synthetic:
		return SyntheticAttribute{}
	default:
		// undefined attr
		//println("unknown attribute: " + attrName)
		return UnparsedAttribute{
			Name:   attrName,
			Length: attrLen,
			Info:   reader.ReadBytes(int(attrLen)),
		}
	}
}

func readStackMapTableAttribute(reader *ClassReader) StackMapTableAttribute {
	numEntries := reader.ReadUint16()
	entries := make([]StackMapFrame, 0, numEntries)
	for i := 0; i < int(numEntries); i++ {
		frameType := reader.ReadUint8()
		var frame StackMapFrame
		switch {
		case frameType <= 63:
			frame = &SameFrame{FrameTypeVal: frameType}
		case frameType >= 64 && frameType <= 127:
			stack := make([]VerificationTypeInfo, 1)
			stack[0] = readVerificationTypeInfo(reader)
			frame = &SameLocals1StackItemFrame{FrameTypeVal: frameType, Stack: stack}
		case frameType == 247:
			offsetDelta := reader.ReadUint16()
			stack := make([]VerificationTypeInfo, 1)
			stack[0] = readVerificationTypeInfo(reader)
			frame = &SameLocals1StackItemFrameExtended{OffsetDelta: offsetDelta, Stack: stack}
		case frameType >= 248 && frameType <= 250:
			offsetDelta := reader.ReadUint16()
			frame = &ChopFrame{FrameTypeVal: frameType, OffsetDelta: offsetDelta}
		case frameType == 251:
			offsetDelta := reader.ReadUint16()
			frame = &SameFrameExtended{OffsetDelta: offsetDelta}
		case frameType >= 252 && frameType <= 254:
			offsetDelta := reader.ReadUint16()
			nLocals := int(frameType - 251)
			locals := make([]VerificationTypeInfo, nLocals)
			for j := 0; j < nLocals; j++ {
				locals[j] = readVerificationTypeInfo(reader)
			}
			frame = &AppendFrame{FrameTypeVal: frameType, OffsetDelta: offsetDelta, Locals: locals}
		case frameType == 255:
			offsetDelta := reader.ReadUint16()
			numLocals := reader.ReadUint16()
			locals := make([]VerificationTypeInfo, numLocals)
			for j := 0; j < int(numLocals); j++ {
				locals[j] = readVerificationTypeInfo(reader)
			}
			numStack := reader.ReadUint16()
			stack := make([]VerificationTypeInfo, numStack)
			for j := 0; j < int(numStack); j++ {
				stack[j] = readVerificationTypeInfo(reader)
			}
			frame = &FullFrame{
				OffsetDelta: offsetDelta,
				Locals:      locals,
				Stack:       stack,
			}
		}
		entries = append(entries, frame)
	}
	return StackMapTableAttribute{Entries: entries}
}

func readVerificationTypeInfo(reader *ClassReader) VerificationTypeInfo {
	tag := reader.ReadUint8()
	info := VerificationTypeInfo{Tag: tag}
	switch tag {
	case 7:
		info.CPoolIndex = reader.ReadUint16()
	case 8:
		info.Offset = reader.ReadUint16()
	}
	return info
}

func writeStackMapTableAttribute(attribute StackMapTableAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(attribute.Entries) >> 8), byte(len(attribute.Entries))}) // number_of_entries
	for _, entry := range attribute.Entries {
		buf.Write(entry.ToBytes())
	}
	return buf.Bytes()
}

type AttributeTable []AttributeInfo

/* group 1 */

func (at AttributeTable) GetCodeAttribute() (CodeAttribute, bool) {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(CodeAttribute); ok {
			return a, true
		}
	}
	return CodeAttribute{}, false
}

func (at AttributeTable) GetConstantValueIndex() uint16 {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(ConstantValueAttribute); ok {
			return a.ConstantValueIndex
		}
	}
	return 0
}

func (at AttributeTable) GetExceptionIndexTable() []uint16 {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(ExceptionsAttribute); ok {
			return a.ExceptionIndexTable
		}
	}
	return nil
}

func (at AttributeTable) GetBootstrapMethods() []BootstrapMethod {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(BootstrapMethodsAttribute); ok {
			return a.BootstrapMethods
		}
	}
	return nil
}

/* group 2 */

func (at AttributeTable) GetEnclosingMethodAttribute() (EnclosingMethodAttribute, bool) {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(EnclosingMethodAttribute); ok {
			return a, true
		}
	}
	return EnclosingMethodAttribute{}, false
}

func (at AttributeTable) GetSignatureIndex() uint16 {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(SignatureAttribute); ok {
			return a.SignatureIndex
		}
	}
	return 0
}

/* group 3 */

func (at AttributeTable) GetSourceFileIndex() uint16 {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(SourceFileAttribute); ok {
			return a.SourceFileIndex
		}
	}
	return 0
}

func (at AttributeTable) GetLineNumberTable() []LineNumberTableEntry {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(LineNumberTableAttribute); ok {
			return a.LineNumberTable
		}
	}
	return nil
}

func (at AttributeTable) GetModuleAttribute() (ModuleAttribute, bool) {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(ModuleAttribute); ok {
			return a, true
		}
	}
	return ModuleAttribute{}, false
}

/* unparsed */

func (at AttributeTable) GetRuntimeVisibleAnnotationsAttributeData() []byte {
	return at.getUnparsedAttributeData(RuntimeVisibleAnnotations)
}
func (at AttributeTable) GetRuntimeVisibleParameterAnnotationsAttributeData() []byte {
	return at.getUnparsedAttributeData(RuntimeVisibleParameterAnnotations)
}
func (at AttributeTable) GetAnnotationDefaultAttributeData() []byte {
	return at.getUnparsedAttributeData(AnnotationDefault)
}

func (at AttributeTable) getUnparsedAttributeData(name string) []byte {
	for _, attrInfo := range at {
		if a, ok := attrInfo.(UnparsedAttribute); ok && a.Name == name {
			return a.Info
		}
	}
	return nil
}

func (at AttributeTable) Remove(tag string) {

	for i, attrInfo := range at {
		if _, ok := attrInfo.(*ExceptionsAttribute); ok {
			if tag == Exceptions {
				at = append(at[:i], at[i+1:]...)
			}
		}
	}
}

/*
	Code_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 max_stack;
	    u2 max_locals;
	    u4 code_length;
	    u1 code[code_length];
	    u2 exception_table_length;
	    {   u2 start_pc;
	        u2 end_pc;
	        u2 handler_pc;
	        u2 catch_type;
	    } exception_table[exception_table_length];
	    u2 attributes_count;
	    attribute_info attributes[attributes_count];
	}
*/
type CodeAttribute struct {
	MaxStack       uint16
	MaxLocals      uint16
	Code           []byte
	ExceptionTable []ExceptionTableEntry
	AttributeTable
}

// RemoveDebugAttribute 删除调试属性,仅保留运行时必要属性,用于缩短class长度
func (c CodeAttribute) RemoveDebugAttribute(file *ClassFile) CodeAttribute {
	attrs := make([]AttributeInfo, 0)
	for _, a := range c.AttributeTable {
		switch a.(type) {
		case LocalVariableTableAttribute:
			table := a.(LocalVariableTableAttribute)
			cp := file.ConstantPool
			for _, attr := range table.LocalVariableTable {
				name := string(cp[attr.NameIndex].([]byte))
				// 删除不是字段名的局面变量名称
				if name != "this" && name != "super" && name != "" {
					hasExist := false
					names := file.GetFieldNames()
					for _, fieldName := range names {
						if name == fieldName {
							hasExist = true
							break
						}
					}
					if !hasExist {
						cp[attr.NameIndex] = make([]byte, 0)
					}
					continue
				}
			}
			file.ConstantPool = cp
			continue
		case LocalVariableTypeTableAttribute:
			table := a.(LocalVariableTypeTableAttribute)
			cp := file.ConstantPool
			for _, attr := range table.LocalVariableTypeTable {
				sig := string(cp[attr.SignatureIndex].([]byte))
				if sig != "" {
					cp[attr.SignatureIndex] = make([]byte, 0)
				}
			}
			file.ConstantPool = cp
			continue
		case LineNumberTableAttribute, SourceFileAttribute, SignatureAttribute:
			continue
		default:
			attrs = append(attrs, a)
		}

	}
	c.AttributeTable = attrs
	return c
}

func (a *CodeAttribute) GetAttribute(s string) interface{} {
	for _, attrInfo := range a.AttributeTable {
		switch s {
		case LocalVariableTable:
			if localVTable, ok := attrInfo.(LocalVariableTableAttribute); ok {
				return &localVTable
			}
		case Code:
			if codeVTable, ok := attrInfo.(CodeAttribute); ok {
				return &codeVTable
			}
		case LocalVariableTypeTable:
			if localVTable, ok := attrInfo.(LocalVariableTypeTableAttribute); ok {
				return &localVTable
			}
		case LineNumberTable:
			if lineNumberVTable, ok := attrInfo.(LineNumberTableAttribute); ok {
				return &lineNumberVTable
			}
		case InnerClasses:
			if classesVTable, ok := attrInfo.(InnerClassesAttribute); ok {
				return &classesVTable
			}
		case StackMapTable:
			if stackMapVTable, ok := attrInfo.(StackMapTableAttribute); ok {
				return &stackMapVTable
			}
		}

	}
	return nil
}

func (a *CodeAttribute) GetCodeLength() uint16 {
	return uint16(len(a.Code))
}

func (a *CodeAttribute) InsertLocalVar(where int, size int) {
	//TODO
	a.MaxLocals = a.MaxLocals + uint16(size)
}

type ExceptionTableEntry struct {
	StartPc   uint16
	EndPc     uint16
	HandlerPc uint16
	CatchType uint16
}

func readCodeAttribute(reader *ClassReader) CodeAttribute {
	return CodeAttribute{
		MaxStack:       reader.ReadUint16(),
		MaxLocals:      reader.ReadUint16(),
		Code:           reader.ReadBytes(int(reader.ReadUint32())),
		ExceptionTable: readExceptionTable(reader),
		AttributeTable: readAttributes(reader),
	}
}

func writeCodeAttribute(writer *ClassWriter, attribute CodeAttribute) []byte {
	var buf bytes.Buffer
	// 计算 CodeAttribute 的长度
	//totalLength := uint32(2 + 2 + 4 + uint32(len(attribute.Code)) + 2) // MaxStack + MaxLocals + Code Length + exception_table_length
	// 计算附加属性的长度
	attachAttr := writeAttributes(writer, attribute.AttributeTable)
	exceptionTable := writeExceptionTable(attribute.ExceptionTable)
	//totalLength += uint32(len(exceptionTable))
	//totalLength += uint32(len(attachAttr))
	//写入名称和长度
	//writeAttributeNameAndLength(writer, &buf, Code, totalLength)
	buf.Write([]byte{byte(attribute.MaxStack >> 8), byte(attribute.MaxStack)})                                                                     // MaxStack
	buf.Write([]byte{byte(attribute.MaxLocals >> 8), byte(attribute.MaxLocals)})                                                                   // MaxLocals
	buf.Write([]byte{byte(len(attribute.Code) >> 24), byte(len(attribute.Code) >> 16), byte(len(attribute.Code) >> 8), byte(len(attribute.Code))}) // Code Length
	buf.Write(attribute.Code)
	buf.Write(exceptionTable) // 使用返回的字节切片
	buf.Write(attachAttr)     // 使用返回的字节切片
	return buf.Bytes()        // 返回字节切片
}

func readExceptionTable(reader *ClassReader) []ExceptionTableEntry {
	return reader.readTable(func(reader *ClassReader) ExceptionTableEntry {
		return ExceptionTableEntry{
			StartPc:   reader.ReadUint16(),
			EndPc:     reader.ReadUint16(),
			HandlerPc: reader.ReadUint16(),
			CatchType: reader.ReadUint16(),
		}
	}).([]ExceptionTableEntry)
}

func writeExceptionTable(table []ExceptionTableEntry) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(table) >> 8), byte(len(table))}) // exception_table_length
	for _, entry := range table {
		buf.Write([]byte{byte(entry.StartPc >> 8), byte(entry.StartPc)})
		buf.Write([]byte{byte(entry.EndPc >> 8), byte(entry.EndPc)})
		buf.Write([]byte{byte(entry.HandlerPc >> 8), byte(entry.HandlerPc)})
		buf.Write([]byte{byte(entry.CatchType >> 8), byte(entry.CatchType)})
	}
	return buf.Bytes() // 返回字节切片
}

//行号
/*
LineNumberTable_attribute {
    u2 attribute_name_index;
    u4 attribute_length;
    u2 line_number_table_length;
    {   u2 start_pc;
        u2 line_number;
    } line_number_table[line_number_table_length];
}
*/
type LineNumberTableAttribute struct {
	LineNumberTable []LineNumberTableEntry
}

type LineNumberTableEntry struct {
	StartPC    uint16
	LineNumber uint16
}

func readLineNumberTableAttribute(reader *ClassReader) LineNumberTableAttribute {
	return LineNumberTableAttribute{
		LineNumberTable: reader.readTable(func(reader *ClassReader) LineNumberTableEntry {
			return LineNumberTableEntry{
				StartPC:    reader.ReadUint16(),
				LineNumber: reader.ReadUint16(),
			}
		}).([]LineNumberTableEntry),
	}
}

func writeLineNumberTableAttribute(attribute LineNumberTableAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(attribute.LineNumberTable) >> 8), byte(len(attribute.LineNumberTable))}) // line_number_table_length
	for _, entry := range attribute.LineNumberTable {
		buf.Write([]byte{byte(entry.StartPC >> 8), byte(entry.StartPC)})
		buf.Write([]byte{byte(entry.LineNumber >> 8), byte(entry.LineNumber)})
	}
	return buf.Bytes() // 返回字节切片
}

/*
	BootstrapMethods_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 num_bootstrap_methods;
	    {   u2 bootstrap_method_ref;
	        u2 num_bootstrap_arguments;
	        u2 bootstrap_arguments[num_bootstrap_arguments];
	    } bootstrap_methods[num_bootstrap_methods];
	}
*/
type BootstrapMethodsAttribute struct {
	BootstrapMethods []BootstrapMethod
}

type BootstrapMethod struct {
	BootstrapMethodRef uint16
	BootstrapArguments []uint16
}

func readBootstrapMethodsAttribute(reader *ClassReader) BootstrapMethodsAttribute {
	return BootstrapMethodsAttribute{
		BootstrapMethods: reader.readTable(func(reader *ClassReader) BootstrapMethod {
			return BootstrapMethod{
				BootstrapMethodRef: reader.ReadUint16(),
				BootstrapArguments: reader.readUint16s(),
			}
		}).([]BootstrapMethod),
	}
}

func writeBootstrapMethodsAttribute(attribute BootstrapMethodsAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(attribute.BootstrapMethods) >> 8), byte(len(attribute.BootstrapMethods))}) // num_bootstrap_methods
	for _, method := range attribute.BootstrapMethods {
		buf.Write([]byte{byte(method.BootstrapMethodRef >> 8), byte(method.BootstrapMethodRef)})           // BootstrapMethodRef
		buf.Write([]byte{byte(len(method.BootstrapArguments) >> 8), byte(len(method.BootstrapArguments))}) // num_bootstrap_arguments

		// 将 BootstrapArguments 转换为 []byte
		for _, arg := range method.BootstrapArguments {
			buf.Write([]byte{byte(arg >> 8), byte(arg)}) // 将每个 uint16 转换为两个字节
		}
	}
	return buf.Bytes() // 返回字节切片
}

//内部类描述符

/*
	InnerClasses_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 number_of_classes;
	    {   u2 inner_class_info_index;
	        u2 outer_class_info_index;
	        u2 inner_name_index;
	        u2 inner_class_access_flags;
	    } classes[number_of_classes];
	}
*/
type InnerClassesAttribute struct {
	Classes []InnerClassInfo
}

type InnerClassInfo struct {
	InnerClassInfoIndex   uint16
	OuterClassInfoIndex   uint16
	InnerNameIndex        uint16
	InnerClassAccessFlags uint16
}

func readInnerClassesAttribute(reader *ClassReader) InnerClassesAttribute {
	return InnerClassesAttribute{
		Classes: reader.readTable(func(reader *ClassReader) InnerClassInfo {
			return InnerClassInfo{
				InnerClassInfoIndex:   reader.ReadUint16(),
				OuterClassInfoIndex:   reader.ReadUint16(),
				InnerNameIndex:        reader.ReadUint16(),
				InnerClassAccessFlags: reader.ReadUint16(),
			}
		}).([]InnerClassInfo),
	}
}

func writeInnerClassesAttribute(attribute InnerClassesAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(attribute.Classes) >> 8), byte(len(attribute.Classes))}) // number_of_classes
	for _, classInfo := range attribute.Classes {
		buf.Write([]byte{byte(classInfo.InnerClassInfoIndex >> 8), byte(classInfo.InnerClassInfoIndex)})
		buf.Write([]byte{byte(classInfo.OuterClassInfoIndex >> 8), byte(classInfo.OuterClassInfoIndex)})
		buf.Write([]byte{byte(classInfo.InnerNameIndex >> 8), byte(classInfo.InnerNameIndex)})
		buf.Write([]byte{byte(classInfo.InnerClassAccessFlags >> 8), byte(classInfo.InnerClassAccessFlags)})
	}
	return buf.Bytes() // 返回字节切片
}

/*
	Exceptions_attribute {
	    u2 attribute_name_index;
	    u4 attribute_length;
	    u2 number_of_exceptions;
	    u2 exception_index_table[number_of_exceptions];
	}
*/
type ExceptionsAttribute struct {
	ExceptionIndexTable []uint16
}

func readExceptionsAttribute(reader *ClassReader) ExceptionsAttribute {
	return ExceptionsAttribute{
		ExceptionIndexTable: reader.readUint16s(),
	}
}

func writeExceptionsAttribute(attribute ExceptionsAttribute) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{byte(len(attribute.ExceptionIndexTable) >> 8), byte(len(attribute.ExceptionIndexTable))}) // number_of_exceptions

	// 将 ExceptionIndexTable 转换为 []byte
	for _, index := range attribute.ExceptionIndexTable {
		buf.Write([]byte{byte(index >> 8), byte(index)}) // 将每个 uint16 转换为两个字节
	}
	return buf.Bytes() // 返回字节切片
}

func writeAttributeInfo(writer *ClassWriter, attr AttributeInfo) []byte {
	// 如果属性是指针类型，解引用
	if reflect.TypeOf(attr).Kind() == reflect.Ptr {
		attr = reflect.ValueOf(attr).Elem().Interface()
	}
	var data []byte
	var nameIndex uint16
	switch a := attr.(type) {
	case BootstrapMethodsAttribute:
		data = writeBootstrapMethodsAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(BootstrapMethods)
	case CodeAttribute:
		data = writeCodeAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(Code)
	case ConstantValueAttribute:
		data = writeConstantValueAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(ConstantValue)
	case DeprecatedAttribute:
		data = writeDeprecatedAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(Deprecated)
	case EnclosingMethodAttribute:
		data = writeEnclosingMethodAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(EnclosingMethod)
	case ExceptionsAttribute:
		data = writeExceptionsAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(Exceptions)
	case InnerClassesAttribute:
		data = writeInnerClassesAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(InnerClasses)
	case LineNumberTableAttribute:
		data = writeLineNumberTableAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(LineNumberTable)
	case LocalVariableTableAttribute:
		data = writeLocalVariableTableAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(LocalVariableTable)
	case LocalVariableTypeTableAttribute:
		data = writeLocalVariableTypeTableAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(LocalVariableTypeTable)
	case ModuleAttribute:
		data = writeModuleAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(Module)
	case SignatureAttribute:
		data = writeSignatureAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(Signature)
	case SourceFileAttribute:
		data = writeSourceFileAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(SourceFile)
	case SyntheticAttribute:
		data = writeSyntheticAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(Synthetic)
	case StackMapTableAttribute:
		data = writeStackMapTableAttribute(a)
		nameIndex = writer.cf.GetConstStrIndex(StackMapTable)
	case UnparsedAttribute:
		data = writeUnparsedAttribute(writer, a)
		nameIndex = writer.cf.GetConstStrIndex(a.Name)
	default:
		// 其他类型的属性
		panic(fmt.Errorf("unknown attribute type: %T", a))
	}
	var all = make([]byte, 0)
	// 将 uint16 转换为 []byte
	all = append(all, byte(nameIndex>>8), byte(nameIndex)) // 高字节和低字节
	l := uint32(len(data))
	// 将 uint32 转换为 []byte
	all = append(all, byte(l>>24), byte(l>>16), byte(l>>8), byte(l)) // 高字节到低字节
	all = append(all, data...)
	return all
}

func writeAttributes(writer *ClassWriter, attributes []AttributeInfo) []byte {
	if len(attributes) == 0 {
		return []byte{0x00, 0x00}
	}
	var buf bytes.Buffer
	// 写入属性的数量
	buf.Write([]byte{byte(len(attributes) >> 8), byte(len(attributes))}) // 属性数量

	// 遍历每个属性并写入
	for _, attr := range attributes {
		buf.Write(writeAttributeInfo(writer, attr)) // 使用返回的字节切片
	}
	return buf.Bytes() // 返回字节切片
}

type StackMapTableAttribute struct {
	Entries []StackMapFrame
}

type VerificationTypeInfo struct {
	Tag        uint8
	CPoolIndex uint16 // ITEM_Object 用
	Offset     uint16 // ITEM_Uninitialized 用
}

func (v *VerificationTypeInfo) ToBytes() []byte {
	data := []byte{v.Tag}
	switch v.Tag {
	case 7, 8:
		data = append(data, uint16ToBytes(v.CPoolIndex)...)
	}
	return data
}

// StackMapFrame 接口
type StackMapFrame interface {
	// FrameType 返回 frame_type 值（u1）
	FrameType() uint8
	// ToBytes 返回 frame 编码后的 []byte
	ToBytes() []byte
}

// same_frame: frame_type in [0..63]
type SameFrame struct {
	FrameTypeVal uint8 // 必须在0..63
}

func (sf *SameFrame) FrameType() uint8 {
	return sf.FrameTypeVal
}

func (sf *SameFrame) ToBytes() []byte {
	return []byte{sf.FrameTypeVal}
}

// same_locals_1_stack_item_frame: frame_type in [64..127]
type SameLocals1StackItemFrame struct {
	FrameTypeVal uint8 // 64..127
	Stack        []VerificationTypeInfo
}

func (f *SameLocals1StackItemFrame) FrameType() uint8 {
	return f.FrameTypeVal
}

func (f *SameLocals1StackItemFrame) ToBytes() []byte {
	data := []byte{f.FrameTypeVal}
	for _, item := range f.Stack {
		data = append(data, item.ToBytes()...)
	}
	return data
}

// full_frame: frame_type=255
type FullFrame struct {
	OffsetDelta uint16
	Locals      []VerificationTypeInfo
	Stack       []VerificationTypeInfo
}

func (f *FullFrame) FrameType() uint8 {
	return 255
}

func (f *FullFrame) ToBytes() []byte {
	data := []byte{255}
	data = append(data, uint16ToBytes(f.OffsetDelta)...)
	data = append(data, uint16ToBytes(uint16(len(f.Locals)))...)
	for _, local := range f.Locals {
		data = append(data, local.ToBytes()...)
	}
	data = append(data, uint16ToBytes(uint16(len(f.Stack)))...)
	for _, s := range f.Stack {
		data = append(data, s.ToBytes()...)
	}
	return data
}

// same_locals_1_stack_item_frame_extended: frame_type = 247
// 结构：u1 frame_type, u2 offset_delta, verification_type_info stack[1]
type SameLocals1StackItemFrameExtended struct {
	OffsetDelta uint16
	Stack       []VerificationTypeInfo // 1 个元素
}

func (f *SameLocals1StackItemFrameExtended) FrameType() uint8 {
	return 247
}
func (f *SameLocals1StackItemFrameExtended) ToBytes() []byte {
	data := []byte{247}
	data = append(data, uint16ToBytes(f.OffsetDelta)...)
	for _, item := range f.Stack {
		data = append(data, item.ToBytes()...)
	}
	return data
}

// chop_frame: frame_type in [248-250]
type ChopFrame struct {
	FrameTypeVal uint8 // 248-250
	OffsetDelta  uint16
}

func (f *ChopFrame) FrameType() uint8 {
	return f.FrameTypeVal
}
func (f *ChopFrame) ToBytes() []byte {
	return append([]byte{f.FrameTypeVal}, uint16ToBytes(f.OffsetDelta)...)
}

// same_frame_extended: frame_type = 251
type SameFrameExtended struct {
	OffsetDelta uint16
}

func (f *SameFrameExtended) FrameType() uint8 {
	return 251
}
func (f *SameFrameExtended) ToBytes() []byte {
	return append([]byte{251}, uint16ToBytes(f.OffsetDelta)...)
}

// append_frame: frame_type in [252-254]
type AppendFrame struct {
	FrameTypeVal uint8 // 252-254
	OffsetDelta  uint16
	Locals       []VerificationTypeInfo
}

func (f *AppendFrame) FrameType() uint8 {
	return f.FrameTypeVal
}
func (f *AppendFrame) ToBytes() []byte {
	data := []byte{f.FrameTypeVal}
	data = append(data, uint16ToBytes(f.OffsetDelta)...)
	for _, local := range f.Locals {
		data = append(data, local.ToBytes()...)
	}
	return data
}

func uint16ToBytes(n uint16) []byte {
	return []byte{byte(n >> 8), byte(n)}
}

func uint32ToBytes(n uint32) []byte {
	return []byte{
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	}
}

func (smt *StackMapTableAttribute) InsertLocal(where, typeTag, classInfo int) {
	// 遍历所有帧，插入本地变量类型
	for _, frame := range smt.Entries {
		switch f := frame.(type) {
		case *FullFrame:
			// 插入 VerificationTypeInfo
			if where <= len(f.Locals) {
				vt := VerificationTypeInfo{Tag: uint8(typeTag)}
				if typeTag == 7 { // 对象类型
					vt.CPoolIndex = uint16(classInfo)
				}
				// 插入到指定位置
				locals := append(f.Locals, VerificationTypeInfo{})
				copy(locals[where+1:], locals[where:])
				locals[where] = vt
				f.Locals = locals
			}
		case *AppendFrame:
			// 只在追加帧时插入
			if where <= len(f.Locals) {
				vt := VerificationTypeInfo{Tag: uint8(typeTag)}
				if typeTag == 7 {
					vt.CPoolIndex = uint16(classInfo)
				}
				locals := append(f.Locals, VerificationTypeInfo{})
				copy(locals[where+1:], locals[where:])
				locals[where] = vt
				f.Locals = locals
			}
		}
	}
}

func TypeTagOf(typeDesc rune) int {
	switch typeDesc {
	case 'B': // byte
		return 1 // ITEM_Integer
	case 'C': // char
		return 1 // ITEM_Integer
	case 'D': // double
		return 3 // ITEM_Double
	case 'F': // float
		return 2 // ITEM_Float
	case 'I': // int
		return 1 // ITEM_Integer
	case 'J': // long
		return 4 // ITEM_Long
	case 'L': // object
		return 7 // ITEM_Object
	case 'S': // short
		return 1 // ITEM_Integer
	case 'Z': // boolean
		return 1 // ITEM_Integer
	case '[': // array
		return 7 // ITEM_Object
	default:
		return 0 // ITEM_Top
	}
}
