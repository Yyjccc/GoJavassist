package compiler

import (
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
	"strings"
)

type MemberResolver struct {
	pool *reflect.ClassPool
}

func (r *MemberResolver) lookupClass(decl *ast.Declarator) (*reflect.CtClass, error) {
	return r.lookupClass0(decl.GetType(), decl.GetArrayDim(), decl.GetClassName())
}

func (r *MemberResolver) lookupClass0(getType ast.TokenID, dim int, name string) (*reflect.CtClass, error) {
	cname := ""
	var err error
	if getType == ast.Class {
		clazz, err := r.lookupClassByJvmName(name)
		if err != nil {
			return nil, err
		}
		if dim > 0 {

		} else {
			return clazz, nil
		}
	} else {
		cname, err = getTypeName(getType)
		if err != nil {
			return nil, err
		}
	}

	for dim > 0 {
		cname += "[]"
		dim--
	}
	return r.lookupClass1(cname, false)

}

func (r *MemberResolver) lookupClass1(className string, notCheckInner bool) (*reflect.CtClass, error) {
	var cc *reflect.CtClass

	for {
		cc = r.pool.GetOrMake(className)
		if cc != nil {
			break
		}
		i := strings.LastIndex(className, ".")
		if notCheckInner || i < 0 {
			return nil, NewCompileError("not found class :" + className)
		}
		// 替换 '.' 为 '$'
		runes := []rune(className)
		runes[i] = '$'
		className = string(runes)
	}

	return cc, nil
}

func (r *MemberResolver) lookupClassByJvmName(name string) (*reflect.CtClass, error) {
	return r.lookupClass1(JvmToJavaName(name), false)
}

func (r *MemberResolver) lookupClassByName(name *ast.ASTList) (*reflect.CtClass, error) {
	return r.lookupClass1(ast.AstToClassName(name, '.'), false)
}

func (r *MemberResolver) resolveJvmClassName(jvmName string) (string, error) {
	if jvmName == "" {
		return jvmName, nil
	}
	name, err := r.lookupClassByJvmName(jvmName)
	if err != nil {
		return "", err
	}
	return JavaToJvmName(name.QualifiedName), nil
}

func (r *MemberResolver) resolveClassName(name *ast.ASTList) (string, error) {
	if isNil(name) {
		return "", nil
	}
	clazz, err := r.lookupClassByName(name)
	if err != nil {
		return "", err
	}
	return JavaToJvmName(clazz.QualifiedName), nil
}

func (r *MemberResolver) lookupMethod(clazz *reflect.CtClass, currentClass *reflect.CtClass, current *reflect.CtMethod, methodName string, argTypes []ast.TokenID, argDims []int, argClassNames []string) *reflect.CtMethod {
	var maybe *reflect.CtMethod

	// 允许查找递归调用的方法
	if current != nil && clazz == currentClass {
		if current.Name == methodName {
			res := r.CompareSignature(current.GetDescriptor(), argTypes, argDims, argClassNames)
			if res != -1 {
				method := reflect.NewMethod(clazz, current.Member)
				if res == 0 {
					return method
				}
				maybe = method
			}
		}
	}

	// 继续查找方法
	mth, err := r.lookupMethod1(clazz, methodName, argTypes, argDims, argClassNames, maybe != nil)
	if err != nil {
		return nil
	}
	if mth != nil {
		return mth
	}
	return maybe
}

func (r *MemberResolver) CompareSignature(descriptor string, argTypes []ast.TokenID, argDims []int, argClassNames []string) int {
	const (
		Yes = 0
		No  = -1
	)

	result := Yes
	i := 1
	nArgs := len(argTypes)

	argNum, err := NumOfParameters(descriptor)
	if err != nil {
		return No
	}
	if nArgs != argNum {
		return No
	}

	lenDesc := len(descriptor)
	for n := 0; i < lenDesc; n++ {
		c := descriptor[i]
		i++

		if c == ')' {
			if n == nArgs {
				return result
			}
			return No
		} else if n >= nArgs {
			return No
		}

		// 计算数组维度
		dim := 0
		for c == '[' {
			dim++
			c = descriptor[i]
			i++
		}

		if argTypes[n] == ast.Null {
			if dim == 0 && c != 'L' {
				return No
			}
			if c == 'L' {
				i = strings.IndexByte(descriptor[i:], ';') + i + 1
				if i <= 0 {
					return No
				}
			}
		} else if argDims[n] != dim {
			if !(dim == 0 && c == 'L' && strings.HasPrefix(descriptor[i:], "java/lang/Object;")) {
				return No
			}
			i = strings.IndexByte(descriptor[i:], ';') + i + 1
			if i <= 0 {
				return No
			}
			result++
		} else if c == 'L' { // 对比类名
			j := strings.IndexByte(descriptor[i:], ';')
			if j < 0 || argTypes[n] != ast.Class {
				return No
			}

			cname := descriptor[i : i+j]
			if cname != argClassNames[n] {
				clazz, err := r.lookupClassByJvmName(argClassNames[n])
				targetClazz, err := r.lookupClassByJvmName(cname)
				if err != nil {
					return No
				}

				if clazz == nil || targetClazz == nil {
					result++
				} else if clazz.SubtypeOf(targetClazz) {
					result++
				} else {
					return No
				}
			}
			i += j + 1
		} else {
			t := descToType(rune(c))
			at := argTypes[n]
			if t != at {
				if t == reflect.T_INT && (at == reflect.T_SHORT || at == reflect.T_BYTE || at == reflect.T_CHAR) {
					result++
				} else {
					return No
				}
			}
		}
	}

	return No
}

func (r *MemberResolver) lookupMethod1(clazz *reflect.CtClass, methodName string, argTypes []ast.TokenID, argDims []int, argClassNames []string, onlyExact bool) (*reflect.CtMethod, error) {
	var maybe *reflect.CtMethod
	if !clazz.IsPrimitive() {
		// 处理类文件中的方法列表
		for _, method := range clazz.Methods {
			if method.Name == methodName && (method.Acc.Raw&0x40) == 0 {
				res := r.CompareSignature(method.Descriptor, argTypes, argDims, argClassNames)
				if res != -1 {
					//method = reflect.NewMethod(clazz, method.Member)
					if res == 0 {
						return method, nil
					} else if maybe == nil || -1 > res {
						maybe = method
					}
				}
			}
		}
	}

	// 如果只有完全匹配且没有找到方法，则返回 nil
	if onlyExact {
		maybe = nil
	} else if maybe != nil {
		return maybe, nil
	}
	//
	//// 搜索父类中的方法
	//mod := clazz.Acc.Raw
	//isIntf := clazz.IsInterface()
	//
	//// 如果是接口，跳过搜索父类
	//if !isIntf {
	//	pclazz, err := klass.GetSuperclass()
	//	if err == nil && pclazz != nil {
	//		method, err := r.lookupMethod(pclazz, methodName, argTypes, argDims, argClassNames, onlyExact)
	//		if err == nil && method != nil {
	//			return method, nil
	//		}
	//	}
	//}

	//// 搜索接口中的方法
	//ifs, err := clazz
	//if err == nil {
	//	for _, intf := range ifs {
	//		r, err := m.LookupMethod(intf, methodName, argTypes, argDims, argClassNames, onlyExact)
	//		if err == nil && r != nil {
	//			return r, nil
	//		}
	//	}
	//}
	//
	//// 对于接口类型，最终搜索父类
	//if isIntf {
	//	pclazz, err := klass.GetSuperclass()
	//	if err == nil && pclazz != nil {
	//		r, err := m.LookupMethod(pclazz, methodName, argTypes, argDims, argClassNames, onlyExact)
	//		if err == nil && r != nil {
	//			return r, nil
	//		}
	//	}
	//}

	return maybe, nil
}

func (r *MemberResolver) isDotSuper(target ast.Node) *reflect.CtClass {
	return nil
}

func (r *MemberResolver) GetSuperInterface(class *reflect.CtClass, super *reflect.CtClass) *reflect.CtClass {
	return nil
}

func (r *MemberResolver) lookupField(className string, symbol *ast.Symbol) (*reflect.CtField, error) {
	cc, err := r.lookupClass1(className, false)
	if err != nil {
		return nil, err
	}
	return cc.GetField(symbol.Identifier)
}

func (r *MemberResolver) lookupFieldByJvmName(jvmClassName string, fieldName *ast.Symbol) (*reflect.CtField, error) {
	return r.lookupField(JvmToJavaName(jvmClassName), fieldName)
}

func (r *MemberResolver) GetSuperInterface1(class *reflect.CtClass, interfaceName string) (*reflect.CtClass, error) {
	infos := class.Interfaces
	for _, info := range infos {
		if info.QualifiedName == interfaceName {
			return info, nil
		}
	}
	return nil, NewCompileError("cannot find the super interface" + interfaceName + " of " + class.QualifiedName)
}

func (r *MemberResolver) lookupFieldByJvmName2(jvmClassName string, fieldSym *ast.MemberSymbol, expr ast.Node) (*reflect.CtField, error) {
	field := fieldSym.Identifier
	cc, err := r.lookupClass1(JvmToJavaName(jvmClassName), true)
	if err != nil {
		return nil, NewNoFieldError(jvmClassName+"/"+field, expr)
	}
	f, err := cc.GetField(field)
	if err != nil {
		return nil, NewNoFieldError(jvmClassName+"$"+field, expr)
	}
	return f, nil
}

func NewMemberResolver(pool *reflect.ClassPool) *MemberResolver {
	return &MemberResolver{pool}
}
