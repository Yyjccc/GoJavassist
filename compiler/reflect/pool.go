package reflect

import (
	"strings"
)

const major = 52

type ClassPool struct {
	Parent       *ClassPool
	ClassLoader  *ClassLoader
	classes      map[string]CtClass
	MajorVersion int
}

func GetVersion() int {
	return major
}

func NewClassPool(parent *ClassPool, loader *ClassLoader) *ClassPool {
	return &ClassPool{
		Parent:      parent,
		ClassLoader: loader,
		classes:     make(map[string]CtClass),
	}
}

func (this *ClassPool) Get(full string) *CtClass {
	if class, ok := this.classes[full]; ok {
		return &class
	}
	if this.Parent != nil {
		return this.Parent.Get(full)
	}
	loader := bootLoader
	if this.ClassLoader != nil {
		loader = this.ClassLoader
	}
	cf := loader.LoadClass(toJvmName(full))
	if cf != nil {
		class := NewClass(cf)
		this.Register(class)
		return class
	}
	return nil
}

func (this *ClassPool) Put(class *CtClass) {
	if class != nil {
		this.classes[class.QualifiedName] = *class
	}
}

func (this *ClassPool) Register(class *CtClass) {
	if this == nil {
		InitClassPool()
		this = DefaultPool
	}
	if class == nil {
		return
	}
	class.loadPool = this
	this.classes[class.QualifiedName] = *class
}

// only make array type
func (this *ClassPool) GetOrMake(name string) *CtClass {
	get := this.Get(name)
	if get != nil {
		get.loadPool = this
		return get
	}
	if strings.Contains(name, "[") {
		return this.CreateCtClass(name)
	}
	if !strings.Contains(name, ".") {
		name = "java.lang." + name
		return this.GetOrMake(name)
	}
	return nil
}

func (cp *ClassPool) CreateCtClass(classname string) *CtClass {
	// 处理数组类型的类名（"[L<classname>;” 转换为正常类名）
	if strings.HasPrefix(classname, "[") {
		classname = DescriptorToClassName(classname)
	}

	// 处理 "xxx[]" 数组类型
	if strings.HasSuffix(classname, "[]") {
		//base := classname[:strings.Index(classname, "[")] // 获取基础类型
		class, err := MakeClass(classname)
		class.arrayDim = GetArrayDimension(classname)
		if err != nil {
			panic(err)
		}
		class.loadPool = cp
		return class // 这里可以创建 CtArray，简化为 CtClass
	}

	panic("unrecognized classname: " + classname)
}

func DescriptorToClassName(descriptor string) string {
	arrayDim := 0
	i := 0
	runes := []rune(descriptor)

	// 统计数组维度
	for i < len(runes) && runes[i] == '[' {
		arrayDim++
		i++
	}

	if i >= len(runes) {
		return ""
	}

	var name string
	switch runes[i] {
	case 'L': // 对象类型，如 "Ljava/lang/String;"
		i++
		i2 := strings.IndexRune(descriptor[i:], ';')
		if i2 == -1 {
			return ""
		}
		name = strings.ReplaceAll(descriptor[i:i+i2], "/", ".")
		i += i2
	case 'V':
		name = "void"
	case 'I':
		name = "int"
	case 'B':
		name = "byte"
	case 'J':
		name = "long"
	case 'D':
		name = "double"
	case 'F':
		name = "float"
	case 'C':
		name = "char"
	case 'S':
		name = "short"
	case 'Z':
		name = "boolean"
	default:
		return ""
	}

	// 确保描述符正确
	if i+1 != len(runes) {
		return ""
	}

	// 如果是数组类型，拼接 "[]"
	if arrayDim > 0 {
		name += strings.Repeat("[]", arrayDim)
	}

	return name
}
