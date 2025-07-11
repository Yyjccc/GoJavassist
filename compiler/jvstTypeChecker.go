package compiler

import (
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

type JvstTypeChecker struct {
	*TypeChecker
	codeGen *JvtCodeGenerator
}

func NewJvstTypeChecker(check *TypeChecker, codeGen *JvtCodeGenerator) *JvstTypeChecker {
	return &JvstTypeChecker{
		TypeChecker: check,
		codeGen:     codeGen,
	}
}

func (j *JvstTypeChecker) atMethodArgs(args *ast.ASTList, types []ast.TokenID, dims []int, cnames []string) error {
	params := j.codeGen.paramTypeList
	pname := j.codeGen.paramListName
	i := 0
	for !isNil(args) {
		a := args.Head()
		member, ok := a.(*ast.MemberSymbol)
		if ok && member.Identifier == pname {
			if params != nil {
				n := len(params)
				for k := 0; k < n; k++ {
					p := params[k]
					j.setType(p, 0)
					types[i] = j.exprType
					dims[i] = j.arrayDim
					cnames[i] = j.className
					i++
				}
			}
		} else {
			if err := a.Accept(j.TypeChecker); err != nil {
				return err
			}
			types[i] = j.exprType
			dims[i] = j.arrayDim
			cnames[i] = j.className
			i++
		}
		args = args.Tail()
	}
	return nil
}

func (j *JvstTypeChecker) setType(p *reflect.CtClass, dim int) {
	if p.IsPrimitive() {
		j.exprType = descToType(rune(p.GetDescriptor()[0]))
		j.arrayDim = dim
		j.className = ""
	} else if p.IsArray() {
		j.setType(p.GetComponentType(), dim+1)
	} else {
		j.exprType = ast.Class
		j.arrayDim = dim
		j.className = JavaToJvmName(p.QualifiedName)
	}
}

func (j *JvstTypeChecker) addNullIfVoid() {
	if j.exprType == ast.Void {
		j.exprType = ast.Class
		j.arrayDim = 0
		j.className = jvmJavaLangObject
	}
}
