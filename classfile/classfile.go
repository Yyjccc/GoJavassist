package classfile

import "fmt"

//class 文件

/*
ClassFile {
    u4             magic;
    u2             minor_version;
    u2             major_version;
    u2             constant_pool_count;
    cp_info        constant_pool[constant_pool_count-1];
    u2             access_flags;
    u2             this_class;
    u2             super_class;
    u2             interfaces_count;
    u2             interfaces[interfaces_count];
    u2             fields_count;
    field_info     fields[fields_count];
    u2             methods_count;
    method_info    methods[methods_count];
    u2             attributes_count;
    attribute_info attributes[attributes_count];
}
*/

type ClassFile struct {
	//magic      uint32
	MinorVersion      uint16         //次版本
	MajorVersion      uint16         //主版本
	ConstantPoolCount uint16         // 常量池大小
	ConstantPool      []ConstantInfo //常量池
	AccessFlags       uint16         //类访问控制符
	ThisClass         uint16         //类名
	SuperClass        uint16         //父类名
	InterfacesCount   uint16         //接口数量
	Interfaces        []uint16       // 类实现的接口
	FieldsCount       uint16         //类属性数量
	Fields            []MemberInfo   //类属性表
	MethodsCount      uint16         //方法数量
	Methods           []MemberInfo   //方法表
	AttributesCount   uint16         //属性数量
	AttributeTable                   // 属性表
}

func (cf *ClassFile) ToByteCode() []byte {
	writer := NewClassWriter(cf)
	// 写入魔数
	writer.WriteUint32(0xCAFEBABE)

	// 写入版本信息
	writer.WriteUint16(cf.MinorVersion)
	writer.WriteUint16(cf.MajorVersion)

	// 写入常量池的大小
	writer.WriteUint16(cf.ConstantPoolCount)

	// 写入常量池的内容
	for _, constant := range cf.ConstantPool {
		writeConstantInfo(writer, constant)
	}

	// 写入访问标志
	writer.WriteUint16(cf.AccessFlags)

	// 写入this_class和super_class
	writer.WriteUint16(cf.ThisClass)
	writer.WriteUint16(cf.SuperClass)

	// 写入接口信息
	writer.WriteUint16(cf.InterfacesCount)
	for _, iface := range cf.Interfaces {
		writer.WriteUint16(iface)
	}

	// 写入字段表
	writer.WriteUint16(cf.FieldsCount)
	for _, field := range cf.Fields {
		writeMember(writer, field)
	}

	// 写入方法表
	writer.WriteUint16(cf.MethodsCount)
	for _, method := range cf.Methods {
		writeMember(writer, method)
	}

	// 写入属性表
	//writer.WriteUint16(cf.AttributesCount)
	attr := writeAttributes(writer, cf.AttributeTable)
	writer.WriteBytes(attr)
	return writer.Bytes()
}

func (cf *ClassFile) Read(reader *ClassReader) {
	reader.cf = cf
	cf.readAndCheckMagic(reader)
	cf.readAndCheckVersions(reader)
	cf.ConstantPool = readConstantPool(reader)
	cf.ConstantPoolCount = uint16(len(cf.ConstantPool))
	cf.AccessFlags = reader.ReadUint16()
	cf.ThisClass = reader.ReadUint16()
	cf.SuperClass = reader.ReadUint16()
	cf.Interfaces = reader.readUint16s()
	cf.InterfacesCount = uint16(len(cf.Interfaces))
	cf.Fields = readMembers(reader)
	cf.FieldsCount = uint16(len(cf.Fields))
	cf.Methods = readMembers(reader)
	cf.MethodsCount = uint16(len(cf.Methods))
	cf.AttributeTable = readAttributes(reader)
	cf.AttributesCount = uint16(len(cf.AttributeTable))
}

// 改变常量池,添加内容
func (cf *ClassFile) AddConstInfo(info ConstantInfo) uint16 {
	cf.ConstantPoolCount++
	cf.ConstantPool = append(cf.ConstantPool, info)
	return uint16(len(cf.ConstantPool) - 1)
}

func (cf *ClassFile) readAndCheckMagic(reader *ClassReader) {
	magic := reader.ReadUint32()
	if magic != 0xCAFEBABE {
		panic("Bad magic!") // TODO
	}
}

func (cf *ClassFile) readAndCheckVersions(reader *ClassReader) {
	cf.MinorVersion = reader.ReadUint16()
	cf.MajorVersion = reader.ReadUint16()

	switch cf.MajorVersion {
	case 45:
		return
	case 46, 47, 48, 49, 50, 51, 52,
		53, 54, 55, 56, 57:
		if cf.MinorVersion == 0 {
			return
		}
	}
	panic("java.lang.UnsupportedClassVersionError!")
}

func (cf *ClassFile) GetThisClassName() string {
	return cf.GetClassName(cf.ThisClass)
}
func (cf *ClassFile) GetSuperClassName() string {
	return cf.GetClassName(cf.SuperClass)
}
func (cf *ClassFile) GetInterfaceNames() []string {
	return cf.GetClassNames(cf.Interfaces)
}

func (cf *ClassFile) GetNameAndType(cpIndex uint16) (name, _type string) {
	if cpIndex > 0 {
		ntInfo := cf.GetConstantInfo(cpIndex).(ConstantNameAndTypeInfo)
		name = cf.GetUTF8(ntInfo.NameIndex)
		_type = cf.GetUTF8(ntInfo.DescriptorIndex)
	}
	return
}

func (cf *ClassFile) GetClassName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	classInfo := cf.GetConstantInfo(cpIndex).(ConstantClassInfo)
	return cf.GetUTF8(classInfo.NameIndex)
}
func (cf *ClassFile) GetPackageName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	pkgInfo := cf.GetConstantInfo(cpIndex).(ConstantPackageInfo)
	return cf.GetUTF8(pkgInfo.NameIndex)
}
func (cf *ClassFile) GetModuleName(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	modInfo := cf.GetConstantInfo(cpIndex).(ConstantModuleInfo)
	return cf.GetUTF8(modInfo.NameIndex)
}

func (cf *ClassFile) GetClassNames(cpIndexes []uint16) []string {
	ss := make([]string, len(cpIndexes))
	for i, cpIndex := range cpIndexes {
		ss[i] = cf.GetClassName(cpIndex)
	}
	return ss
}
func (cf *ClassFile) GetModuleNames(cpIndexes []uint16) []string {
	ss := make([]string, len(cpIndexes))
	for i, cpIndex := range cpIndexes {
		ss[i] = cf.GetModuleName(cpIndex)
	}
	return ss
}

func (cf *ClassFile) GetRawUTF8(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	rawBytes := cf.GetConstantInfo(cpIndex).([]byte)
	return string(rawBytes)
}

// GetUTF8 从常量池中读取字符串
func (cf *ClassFile) GetUTF8(cpIndex uint16) string {
	if cpIndex == 0 {
		return ""
	}
	bytes := cf.GetConstantInfo(cpIndex).([]byte)
	return DecodeMUTF8(bytes)
}

func (cf *ClassFile) GetConstantInfo(cpIndex uint16) ConstantInfo {
	if cpInfo := cf.ConstantPool[cpIndex]; cpInfo == nil {
		panic(fmt.Errorf("invalid constant pool index: %d", cpIndex))
	} else {
		return cpInfo
	}
}

func (cf *ClassFile) GetConstantClassInfoIndex(nameIndex uint16) uint16 {
	for i, entry := range cf.ConstantPool {
		if data, ok := entry.(ConstantClassInfo); ok {
			if data.NameIndex == nameIndex {
				return uint16(i)
			}
		}
	}
	return uint16(0)
}

// 获取常量池中现有字符串索引
func (cf *ClassFile) GetConstStrIndex(name string) uint16 {
	for i, entry := range cf.ConstantPool {
		data, ok := entry.([]uint8)
		if ok {
			if string(data) == name {
				return uint16(i)
			}
		}
	}
	return 0
}

func (cf *ClassFile) GetFieldNames() []string {
	var res []string
	for _, field := range cf.Fields {
		s := string(cf.ConstantPool[field.NameIndex].([]byte))
		res = append(res, s)
	}
	return res
}

func NewClassFile() *ClassFile {
	return &ClassFile{
		MajorVersion:   52,
		Interfaces:     make([]uint16, 0),
		Fields:         make([]MemberInfo, 0),
		Methods:        make([]MemberInfo, 0),
		ConstantPool:   make([]ConstantInfo, 0),
		AttributeTable: make([]AttributeInfo, 0),
	}

}
