package reflect

import (
	"fmt"
	"strings"
)

func toJvmName(name string) string {

	return strings.ReplaceAll(name, ".", "/")
}

func _fullName(name string) string {
	return strings.ReplaceAll(name, "/", ".")
}

func _packageName(name string) string {
	classPath := strings.ReplaceAll(name, "/", ".")
	// 找到最后一个 '.' 的位置
	lastDotIndex := strings.LastIndex(classPath, ".")
	if lastDotIndex != -1 {
		return classPath[:lastDotIndex]
	} else {
		// 如果没有包名
		return ""
	}
}

func _simpleName(name string) string {
	classPath := strings.ReplaceAll(name, "/", ".")
	// 找到最后一个 '.' 的位置
	lastDotIndex := strings.LastIndex(classPath, ".")
	if lastDotIndex != -1 {
		return classPath[lastDotIndex+1:]
	} else {
		// 如果没有包名
		return classPath
	}
}

// _toDescriptor 将类路径转为类型描述符
func _toDescriptor(classPath string) string {
	// 如果是基本类型，直接返回描述符
	switch classPath {
	case "byte":
		return "B"
	case "char":
		return "C"
	case "double":
		return "D"
	case "float":
		return "F"
	case "int":
		return "I"
	case "long":
		return "J"
	case "short":
		return "S"
	case "boolean":
		return "Z"
	case "void":
		return "V"
	}

	if strings.Contains(classPath, "[]") {
		count := strings.Count(classPath, "[]")
		return strings.Repeat("[", count) + _toDescriptor(strings.ReplaceAll(classPath, "[]", ""))

	}
	// 如果是数组类型，例如 [Ljava/lang/String;
	if strings.HasPrefix(classPath, "[") {
		return classPath
	}

	// 否则是普通类名，添加前缀 "L" 和后缀 ";"
	return "L" + classPath + ";"
}

// 类型映射表，用于将 JVM 类型标记转换为 Java 类型
var typeMap = map[string]string{
	"V": "void",
	"I": "int",
	"B": "byte",
	"C": "char",
	"D": "double",
	"F": "float",
	"J": "long",
	"S": "short",
	"Z": "boolean",
}

// _descriptor 解析描述符，返回返回参数类型，和参数类型
func _descriptor(class *CtClass, descriptor string) (string, string, error) {
	if len(descriptor) < 2 || descriptor[0] != '(' {
		return "", "", fmt.Errorf("无效的描述符格式")
	}

	// 找到参数结束的 ')' 位置
	endIndex := strings.Index(descriptor, ")")
	if endIndex == -1 {
		return "", "", fmt.Errorf("无效的描述符格式，缺少 ')'")
	}

	// 提取参数和返回值部分
	paramPart := descriptor[1:endIndex]
	returnTypePart := descriptor[endIndex+1:]

	// 解析参数类型
	params, err := parseTypes(class, paramPart)
	if err != nil {
		return "", "", err
	}

	// 解析返回值类型
	returnType, _, err := parseType(class, returnTypePart)
	if err != nil {
		return "", "", err
	}

	// 拼接完整的 Java 方法签名
	return _simpleName(returnType), strings.Join(params, ", "), nil
}

// parseTypes 解析多个类型（如参数列表）
func parseTypes(class *CtClass, typeStr string) ([]string, error) {
	var types []string
	for len(typeStr) > 0 {
		typ, newTypeStr, err := parseType(class, typeStr)
		if err != nil {
			return nil, err
		}
		types = append(types, _simpleName(typ))
		typeStr = newTypeStr
	}
	return types, nil
}

// parseType 解析单一类型
func parseType(class *CtClass, typeStr string) (string, string, error) {
	if len(typeStr) == 0 {
		return "", "", fmt.Errorf("无效的类型描述符")
	}

	// 处理数组类型
	if typeStr[0] == '[' {
		arrayType := "[]"
		for typeStr[0] == '[' {
			arrayType += "[]"
			typeStr = typeStr[1:]
		}
		typ, restTypeStr, err := parseType(class, typeStr)
		if err != nil {
			return "", "", err
		}
		return typ + arrayType, restTypeStr, nil
	}

	// 处理对象类型
	if typeStr[0] == 'L' {
		semicolonIndex := strings.Index(typeStr, ";")
		if semicolonIndex == -1 {
			return "", "", fmt.Errorf("无效的对象类型描述符，缺少 ';'")
		}
		// 转换路径为 Java 类名
		return strings.ReplaceAll(typeStr[1:semicolonIndex], "/", "."), typeStr[semicolonIndex+1:], nil
	}

	// 基本类型
	if javaType, exists := typeMap[string(typeStr[0])]; exists {
		return javaType, typeStr[1:], nil
	}
	return "", "", fmt.Errorf("未知的类型描述符: %c", typeStr[0])
}

// toTypeDescriptor: 将 Java 类型转换为 `.class` 类型描述符
func _toTypeDescriptor(javaType string) string {
	// 检查是否是数组类型（以 [] 结尾）
	if strings.HasSuffix(javaType, "[]") {
		// 递归转换数组元素类型
		return "[" + _toTypeDescriptor(strings.TrimSuffix(javaType, "[]"))
	}

	// 检查是否是基本类型
	if descriptor, exists := typeMap[javaType]; exists {
		return descriptor
	}

	// 处理类名 (L<classname>; 格式)
	return "L" + strings.ReplaceAll(javaType, ".", "/") + ";"
}

// GetArrayDimension 计算 JVM 描述符的数组维度
func GetArrayDimension(descriptor string) int {
	arrayDim := 0
	for _, c := range descriptor {
		if c == '[' {
			arrayDim++
		}
	}
	return arrayDim
}

func ToMethodDescriptor(returnType *CtClass, paramTypes []*CtClass) string {
	var sb strings.Builder
	sb.WriteString("(")

	if paramTypes != nil {
		for _, param := range paramTypes {
			ToDescriptor(&sb, param)
		}
	}

	sb.WriteString(")")
	ToDescriptor(&sb, returnType)

	return sb.String()
}

// ToDescriptor 将 CtClass 转换为 JVM 描述符格式
func ToDescriptor(sb *strings.Builder, ctClass *CtClass) {
	if ctClass != nil {
		sb.WriteString(_toDescriptor(ctClass.QualifiedName))
	}
}
