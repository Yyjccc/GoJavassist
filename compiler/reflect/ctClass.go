package reflect

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/classfile"
	"strings"
)

const version = "1.0.0"

type CtPrimitiveType struct {
	wrapper       string
	Descriptor    rune
	mDescriptor   string
	getMethodName string
	returnOp      classfile.OpCode
	arrayType     int
	dataSize      int
}

func (p *CtPrimitiveType) GetReturnOp() classfile.OpCode {
	return p.returnOp
}

func (c *CtPrimitiveType) GetWrapperName() string {
	return c.wrapper
}

func (c *CtPrimitiveType) GetDataSize() int {
	return c.dataSize
}

func (c *CtPrimitiveType) GetDescriptor() rune {
	return c.Descriptor
}

type CtClass struct {
	ClassFile      *classfile.ClassFile
	ClassAttribute classfile.AttributeTable
	arrayDim       int
	editMode       bool
	isPrimitive    bool
	isInterface    bool
	PrimitiveType  *CtPrimitiveType
	hasConstructor bool
	rawClassData   []byte
	edPool         *ConstPool
	Methods        map[string]*CtMethod
	Fields         map[string]*CtField
	PackageName    string
	SimpleName     string
	QualifiedName  string
	wasFrozen      bool
	Acc            *AccDesc
	SuperClassName string
	Interfaces     []*CtClass
	loadPool       *ClassPool
	Imports        []string //导入表
	wasChanged     bool
	wasPruned      bool
}

func NewClass(class *classfile.ClassFile) *CtClass {
	c := &CtClass{
		ClassFile:      class,
		Methods:        make(map[string]*CtMethod),
		Imports:        make([]string, 0),
		Fields:         make(map[string]*CtField),
		Acc:            NewAccDec(class.AccessFlags),
		Interfaces:     make([]*CtClass, 0),
		rawClassData:   make([]byte, 0),
		ClassAttribute: classfile.AttributeTable{},
	}
	name := class.GetThisClassName()
	c.QualifiedName = _fullName(name)
	c.SimpleName = _simpleName(name)
	c.PackageName = _packageName(name)
	c.SuperClassName = _fullName(class.GetSuperClassName())
	c.ClassAttribute = class.AttributeTable
	for _, method := range class.Methods {
		ClassMethod := NewMethod(c, &method)
		c.Methods[ClassMethod.GetFullName()] = ClassMethod
	}
	for _, field := range class.Fields {
		ClassField := NewField(c, &field)
		c.Fields[ClassField.Name] = ClassField
	}
	DefaultPool.Register(c)
	return c
}

func (c *CtClass) ChangeThisClassName(fullName string) error {
	pool := c.loadPool
	if pool == nil {
		pool = DefaultPool
	}
	class := pool.Get(fullName)
	if class != nil {
		return fmt.Errorf(" %s : already has this class.", fullName)
	}
	c.ReplaceSymbolLink(c.QualifiedName, fullName)
	c.QualifiedName = fullName
	c.SimpleName = _simpleName(fullName)
	c.PackageName = _packageName(fullName)
	pool.Register(c)
	return nil
}

// ReplaceSymbolLink 替换常量池中的符号:直接替换,无需进入编辑模式
func (c *CtClass) ReplaceSymbolLink(oldSymbol, newSymbol string) {

	for i, constInfo := range c.ClassFile.ConstantPool {
		if utf8, ok := constInfo.([]byte); ok {
			s := string(utf8)
			if strings.Contains(s, ".") && strings.Contains(newSymbol, ".") {
				res := strings.ReplaceAll(s, oldSymbol, newSymbol)
				c.ClassFile.ConstantPool[i] = []byte(res)
				continue
			}
			if strings.Contains(s, "/") {
				jvmName := strings.ReplaceAll(newSymbol, ".", "/")
				oldjvmName := strings.ReplaceAll(oldSymbol, ".", "/")
				res := strings.ReplaceAll(s, oldjvmName, jvmName)
				c.ClassFile.ConstantPool[i] = []byte(res)
			}
			str := c.ClassFile.ConstantPool[i].([]byte)
			res := strings.ReplaceAll(string(str), oldSymbol, newSymbol)
			c.ClassFile.ConstantPool[i] = []byte(res)
		}
	}
}

func (c *CtClass) GetDescriptor() string {
	if c.isPrimitive {
		return string(c.PrimitiveType.Descriptor)
	}
	if !c.editMode {
		return _toDescriptor(c.ClassFile.GetThisClassName())
	}
	return _toDescriptor(c.QualifiedName)
}

func MakeClass(className string) (*CtClass, error) {
	if DefaultPool.Get(className) != nil {
		return nil, fmt.Errorf("class %s already exists", className)
	}
	if bootLoader.LoadClass(className) != nil {
		return nil, fmt.Errorf("class %s already exists", className)
	}
	pool := MakeConstPool()
	acc := &AccDesc{
		Public: true,
		Raw:    classfile.AccPublic,
	}
	thisIndex := pool.AddString(toJvmName(className))
	thisInfo := classfile.ConstantClassInfo{
		NameIndex: uint16(thisIndex),
	}

	classFile := &classfile.ClassFile{
		MajorVersion:    uint16(major),
		ConstantPool:    make([]classfile.ConstantInfo, 0),
		AccessFlags:     acc.Raw,
		ThisClass:       uint16(pool.AddConstantClassInfo(thisInfo)),
		Interfaces:      make([]uint16, 0),
		Fields:          make([]classfile.MemberInfo, 0),
		MethodsCount:    0,
		Methods:         make([]classfile.MemberInfo, 0),
		AttributesCount: 0,
		AttributeTable:  make([]classfile.AttributeInfo, 0),
	}
	class := &CtClass{
		ClassFile:     classFile,
		edPool:        pool,
		editMode:      true,
		Methods:       make(map[string]*CtMethod),
		Fields:        make(map[string]*CtField),
		PackageName:   _packageName(className),
		SimpleName:    _simpleName(className),
		QualifiedName: className,
		Acc:           acc,
		Interfaces:    make([]*CtClass, 0),
		Imports:       make([]string, 0),
	}
	pool.This = class
	DefaultPool.Register(class)
	return class, nil
}

// not a static method ,defualt public
func (c *CtClass) NewMethod(returnType *CtClass, methodName string, args []*CtClass) (*CtMethod, error) {
	// 构建方法的参数描述符
	var argDesc string
	for _, arg := range args {
		argDesc += arg.GetDescriptor() // 获取参数类型的描述符
	}
	returnDesc := returnType.GetDescriptor() // 获取返回类型的描述符
	methodDesc := "(" + argDesc + ")" + returnDesc
	if c.CheckExistMethod(methodName, methodDesc) {
		return nil, fmt.Errorf("Method %s already exists", methodName)
	}
	acc := &AccDesc{Public: true, Raw: classfile.AccPublic}
	member := &classfile.MemberInfo{
		AccessFlags:     acc.Raw,
		NameIndex:       0,
		DescriptorIndex: 0,
		AttributeTable:  make([]classfile.AttributeInfo, 0),
	}
	method := &CtMethod{
		Class:      c,
		Member:     member,
		Name:       methodName,
		Descriptor: methodDesc,
		Acc:        acc,
	}
	return method, nil
}

func (c *CtClass) check() error {
	return nil
}

// 获取简单的类型名称
func (c *CtClass) GetSimpleType(classpath string) string {
	for k, v := range typeMap {
		if classpath == k {
			return v
		}
		if classpath == v {
			return classpath
		}
	}
	fullName := _fullName(classpath)
	for _, im := range c.Imports {
		if im == fullName {
			return _simpleName(classpath)
		}
	}
	if !strings.Contains(fullName, "java.lang") {
		c.Imports = append(c.Imports, fullName)
	}
	return _simpleName(classpath)
}

//func (c *CtClass) String() string {
//	return "class[" + strings.ReplaceAll(c.ClassFile.GetThisClassName(), "/", ".") + "]"
//}

// 添加 field
func (c *CtClass) AddField(field *CtField) {
	c.Fields[field.Name] = field
	//cf 中添加
	c.ClassFile.Fields = append(c.ClassFile.Fields, *field.Member)
	c.ClassFile.FieldsCount = uint16(len(c.ClassFile.Fields))
}

func (c *CtClass) IsInterface() bool {
	return false
}

func (c *CtClass) IsArray() bool {
	return c.arrayDim == 0
}

// 检查是否已经存在 该方法
func (c *CtClass) CheckExistMethod(name string, methodDesc string) bool {
	fullName := name + methodDesc
	_, ok := c.Methods[fullName]
	return ok
}

func (c *CtClass) IsPrimitive() bool {
	return c.isPrimitive
}

// 如果是数组类型
func (c *CtClass) GetComponentType() *CtClass {
	lastIndex := strings.LastIndex(c.QualifiedName, "[]")
	name := c.QualifiedName
	if lastIndex != -1 {
		name = c.QualifiedName[:lastIndex] // 去掉最后一个 "[]"
	}
	return DefaultPool.GetOrMake(name)
}

func (c *CtClass) GetConstPool() *ConstPool {
	if !c.editMode {
		c.enterEditorMode()
	}
	return c.edPool
}

func (c *CtClass) ToClassFile() *classfile.ClassFile {
	cp := c.GetConstPool()
	cf := classfile.NewClassFile()
	c.editMode = false
	c.wasFrozen = true
	for _, method := range c.Methods {
		cf.Methods = append(cf.Methods, *method.Member)
		cf.MethodsCount++
	}
	for _, field := range c.Fields {
		cf.Fields = append(cf.Fields, *field.Member)
		cf.FieldsCount++
	}
	for _, intf := range c.Interfaces {
		intfIndex := cp.AddString(intf.QualifiedName)
		cf.Interfaces = append(cf.Interfaces, uint16(intfIndex))
		cf.InterfacesCount++
	}
	if c.SuperClassName == "" {
		c.SuperClassName = "java.lang.Object"
	}
	pool := cp.GetPool()
	cf.AccessFlags = c.Acc.toAccessFlags()
	cf.ConstantPool = pool
	cf.ConstantPoolCount = uint16(len(cf.ConstantPool))
	thisIndex := cf.GetConstStrIndex(toJvmName(c.QualifiedName))
	cf.ThisClass = cf.GetConstantClassInfoIndex(thisIndex)
	superClass := c.loadPool.Get(c.SuperClassName)
	if superClass == nil {
		panic("super class not exist")
	}
	superIndex := cf.GetConstStrIndex(toJvmName(superClass.QualifiedName))
	cf.SuperClass = cf.GetConstantClassInfoIndex(superIndex)
	c.ClassFile = cf
	return cf
}

func (c *CtClass) AddMethod(method *CtMethod) error {
	c.CheckModify()
	if method.Class != c {
		return fmt.Errorf("bad declaring class")
	}
	acc := method.Acc.toAccessFlags()
	if acc&classfile.AccInterface != 0 {
		if method.Acc.Protected || method.Acc.Private {
			return fmt.Errorf("an interface method must be public: %s", method.GetFullName())
		}
		acc = acc | classfile.AccPublic
	}
	method.Member.AccessFlags = acc
	c.Methods[method.Name+method.Descriptor] = method
	return nil
}

func (c *CtClass) SubtypeOf(clazz *CtClass) bool {
	//c.SuperClassName
	return true
}

// GetField 获取属性
func (c *CtClass) GetField(name string) (*CtField, error) {
	if field, ok := c.Fields[name]; ok {
		return field, nil
	}
	return nil, fmt.Errorf("Field %s not found", name)
}

func (c *CtClass) GetName() string {
	return c.QualifiedName
}

func (c *CtClass) CheckModify() {
	if c.wasFrozen {
		msg := c.GetName() + " class is frozen"
		if c.wasPruned {
			msg += " and pruned"
		}
		panic(msg)
	}
	c.wasChanged = true
}

func (c *CtClass) assertConstructor() {
	if !c.hasConstructor {
		c.inheritAllConstructors()
		c.hasConstructor = true
	}
}

func (c *CtClass) inheritAllConstructors() {
	//superClazz := ClassForName(c.SuperClassName)

}

func (c *CtClass) GetClassPool() *ClassPool {
	if c.loadPool != nil {
		return c.loadPool
	}
	return DefaultPool
}

func (c *CtClass) GetSuperClass() *CtClass {
	return c.GetClassPool().Get(c.SuperClassName)
}

// SetSuperClass 设置父类
func (c *CtClass) SetSuperClass(pool *ClassPool, fullName string) error {
	c.CheckModify()
	superClass := pool.Get(fullName)
	if superClass == nil {
		return fmt.Errorf("class %s not found", fullName)
	}
	c.checkImport(fullName)

	return nil
}

// AddInterface 添加类实现的接口--不会添加相关实现方法
func (c *CtClass) AddInterface(pool *ClassPool, fullName string) error {
	c.CheckModify()
	c.enterEditorMode()
	interfaceClass := pool.Get(fullName)
	if interfaceClass == nil {
		return fmt.Errorf("class %s not found", fullName)
	}
	if !interfaceClass.isInterface {
		return fmt.Errorf("class %s is not an interface", fullName)
	}
	c.checkImport(fullName)
	c.Interfaces = append(c.Interfaces, interfaceClass)
	c.edPool.AddClassInfo0(interfaceClass)
	return nil
}

// checkImport 检查类是否导入
func (c *CtClass) checkImport(className string) {
	if strings.HasPrefix(className, "java.lang") {
		return
	}
	for _, imp := range c.Imports {
		if imp == className {
			return
		}
	}
	c.Imports = append(c.Imports, className)
}

// enterEditorMode 进入编辑模式
func (c *CtClass) enterEditorMode() {
	if !c.editMode {
		c.editMode = true
		c.wasFrozen = false
		c.wasChanged = true
		c.edPool = NewConstPool(c, c.ClassFile.ConstantPool)
	}
}

func (c *CtClass) Equal(class *CtClass) bool {
	return c.QualifiedName == class.QualifiedName
}

func (c *CtClass) GetDeclareMethods(methodName string) []*CtMethod {
	var methods []*CtMethod
	for _, m := range c.Methods {
		if m.Name == methodName {
			methods = append(methods, m)
		}
	}
	return methods
}

func (c *CtClass) RemoveDebugAttribute() {
	c.CheckModify()
	c.GetConstPool()
	for _, m := range c.Methods {
		attrs := make([]classfile.AttributeInfo, 0)
		for _, a := range m.Member.AttributeTable {
			if codeAtr, ok := a.(classfile.CodeAttribute); ok {
				attrs = append(attrs, codeAtr.RemoveDebugAttribute(c.ClassFile))
				continue
			}
			if codeAtr, ok := a.(*classfile.CodeAttribute); ok {
				attrs = append(attrs, codeAtr.RemoveDebugAttribute(c.ClassFile))
				continue
			}
			attrs = append(attrs, a)
		}
		m.Member.AttributeTable = attrs
	}

	// 删除 class 中属性

	cp := c.ClassFile.ConstantPool
	for _, attr := range c.ClassAttribute {
		if sourceAttr, ok := attr.(classfile.SourceFileAttribute); ok {
			cp[sourceAttr.SourceFileIndex] = make([]byte, 0)
			continue
		}
	}
	c.ClassFile.RemoveStrings([]string{classfile.InnerClasses, classfile.LocalVariableTable, classfile.Signature, classfile.LocalVariableTypeTable, classfile.LineNumberTable, classfile.SourceFile})
}

func (c *CtClass) Unfreeze() {
	c.wasFrozen = false
	c.wasChanged = true
}
