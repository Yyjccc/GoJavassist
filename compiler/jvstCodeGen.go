package compiler

import (
	"GoJavassist/classfile"
	"GoJavassist/compiler/ast"
	"GoJavassist/compiler/reflect"
	"strconv"
)

const (
	sigName         = "$sig"
	dollarTypeName  = "$type"
	clazzName       = "$class"
	wrapperCastName = "$w"
	cflowName       = "$cflow"
)

type JvtCodeGenerator struct {
	*MemberCodeGenerator
	paramTypeList  []*reflect.CtClass
	returnType     *reflect.CtClass
	returnCastName string
	dollarType     *reflect.CtClass
	paramListName  string
	paramArrayName string
	clazzName      string
	returnVarName  string
	useParam0      bool
	param0Type     string //jvm name
	paramVarBase   int
}

func NewJvtCodeGenerator(class *reflect.CtClass) *JvtCodeGenerator {
	mem := NewMemberCodeGenerator(class)
	generator := &JvtCodeGenerator{
		MemberCodeGenerator: mem,
	}
	mem.SetParent(generator)
	mem.typeChecker.parent = NewJvstTypeChecker(mem.typeChecker, generator)
	return generator
}

func (j *JvtCodeGenerator) recordParams(params []*reflect.CtClass, isStatic bool, prefix, paramVarName, paramsName string, use0 bool, paramBase int, target string, tbl *SymbolTable) int {
	j.paramTypeList = params
	j.paramListName = paramVarName
	j.paramListName = paramsName
	j.paramVarBase = paramBase
	j.useParam0 = use0
	j.isStaticMethod = isStatic
	if target != "" {
		j.param0Type = JvmToJavaName(target)
	}
	varNo := paramBase
	if use0 {
		varName := prefix + "0"
		decl := ast.NewDeclaratorFull(ast.Class, JavaToJvmName(target), 0, varNo, ast.NewSymbol(varName))
		varNo++
		tbl.Append(varName, decl)
	}

	for i, param := range params {
		varNo += j.recordVar(param, prefix+strconv.Itoa(i+1), varNo, tbl)
	}
	if j.bytecodes.MaxLocals < varNo {
		j.bytecodes.MaxLocals = varNo
	}
	return varNo
}

func (j *JvtCodeGenerator) recordVar(cc *reflect.CtClass, varName string, varNo int, tbl *SymbolTable) int {
	if cc == reflect.VoidType {
		j.exprType = ast.Class
		j.arrayDim = 0
		j.className = jvmJavaLangObject
	} else {
		j.SetType(cc, 0)
	}
	decl := ast.NewDeclaratorFull(j.exprType, j.className, j.arrayDim, varNo, ast.NewSymbol(varName))
	tbl.Append(varName, decl)
	if is2word(j.exprType, j.arrayDim) {
		return 2
	}
	return 1
}

func (j *JvtCodeGenerator) SetType(cc *reflect.CtClass, dim int) {
	if cc == nil {
		return
	}
	if cc.IsPrimitive() {
		j.exprType = descToType(cc.PrimitiveType.Descriptor)
		j.arrayDim = dim
		j.className = ""
	} else if cc.IsArray() {
		j.SetType(cc.GetComponentType(), dim+1)
	} else {
		j.exprType = ast.Class
		j.arrayDim = dim
		j.className = JavaToJvmName(cc.QualifiedName)
	}
}

func (j *JvtCodeGenerator) indexOfParam1() int {
	val := 0
	if j.useParam0 {
		val = 1
	}
	return j.paramVarBase + val
}

func (j *JvtCodeGenerator) AtMember(member *ast.MemberSymbol) error {
	name := member.Symbol.Identifier
	if name == j.paramArrayName {
		compileParameterList(j.bytecodes, j.paramTypeList, j.indexOfParam1())
		j.exprType = ast.Class
		j.arrayDim = 1
		j.clazzName = jvmJavaLangObject
	} else if name == sigName {

	} else if name == dollarTypeName {

	} else if name == clazzName {

	} else {
		return j.MemberCodeGenerator.AtMember(member)
	}
	return nil
}

func (j *JvtCodeGenerator) AtCallExpr(expr *ast.CallExpr) error {
	method := expr.Oprand1()
	if m, ok := method.(*ast.MemberSymbol); ok {
		name := m.Symbol.Identifier
		if name == cflowName {

		}

	}
	//if (method instanceof Member) {
	//	String name = ((Member)method).get();
	//	if (procHandler != null && name.equals(proceedName)) {
	//		procHandler.doit(this, bytecode, (ASTList)expr.oprand2());
	//		return;
	//	}
	//	else if (name.equals(cflowName)) {
	//		atCflow((ASTList)expr.oprand2());
	//		return;
	//	}
	//}
	j.ActiveSuper = false
	return j.MemberCodeGenerator.AtCallExpr(expr)
}

func (j *JvtCodeGenerator) recordType(rType *reflect.CtClass) {
	j.dollarType = rType
}

func (j *JvtCodeGenerator) recordReturnType(rType *reflect.CtClass, castName string, resultName string, table *SymbolTable) int {
	j.returnType = rType
	j.returnCastName = castName
	j.returnVarName = resultName

	varNo := j.bytecodes.MaxLocals
	locals := varNo + j.recordVar(rType, resultName, varNo, table)
	j.bytecodes.MaxLocals = locals
	return varNo
}

func compileParameterList(code *ByteCodes, params []*reflect.CtClass, regno int) int {
	if len(params) == 0 {
		code.addIconst(0)
		code.addAnewArray(javaLangObject)
		return 1
	}
	args := make([]*reflect.CtClass, 1)
	n := len(params)
	code.addIconst(n)
	code.addAnewArray(javaLangObject)

	for i := 0; i < n; i++ {
		code.AddOpcode(classfile.OpDup)
		code.addIconst(i)
		if params[i].IsPrimitive() {
			//基本数据类型
			clazz := params[i]
			wrapper := clazz.PrimitiveType.GetWrapperName()
			code.AddNew(wrapper)            // new <wrapper>
			code.AddOpcode(classfile.OpDup) // dup
			s := code.AddLoad(regno, clazz) // ?load <regno>
			regno += s
			args[0] = clazz
			code.AddInvokeSpecial(wrapper, "<init>", OfMethodDescriptor(reflect.VoidType, args))
		} else {
			code.addAload(regno)
			regno++
		}
		code.AddOpcode(classfile.OpAAStore) // aastore
	}
	return 8
}

func OfMethodDescriptor(returnType *reflect.CtClass, paramTypes []*reflect.CtClass) string {
	return ""
}
