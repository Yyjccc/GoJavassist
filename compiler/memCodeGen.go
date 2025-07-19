package compiler

import (
	"errors"
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
	"strings"
)

var nullClass = &reflect.CtClass{
	QualifiedName: "null",
}
var nullMethod = &reflect.CtMethod{
	Name:  "null",
	Class: nullClass,
}

const initMethodName = "<init>"

type MemberCodeGenerator struct {
	*CCodeGenerator
	Parent         *JvtCodeGenerator
	ActiveSuper    bool
	thisClass      *reflect.CtClass
	thisMethod     *reflect.CtMethod
	resolver       *MemberResolver
	inStaticMethod bool
	resultStatic   bool
}

func NewMemberCodeGenerator(class *reflect.CtClass) *MemberCodeGenerator {
	resolver := NewMemberResolver(reflect.DefaultPool)
	m := &MemberCodeGenerator{
		CCodeGenerator: NewCCodeGenerator(class.GetConstPool()),
		ActiveSuper:    true,
		thisClass:      class,
		thisMethod:     nullMethod,
		resolver:       resolver,
	}
	m.CCodeGenerator.typeChecker.thisClass = class
	m.CCodeGenerator.SetParent(m)
	m.CCodeGenerator.typeChecker.resolver = resolver
	return m
}

func (m *MemberCodeGenerator) SetParent(parent *JvtCodeGenerator) {
	m.Parent = parent
}

func (m *MemberCodeGenerator) SetThisClass(class *reflect.CtClass) {
	m.thisClass = class
	m.CCodeGenerator.typeChecker.thisClass = class
}
func (m *MemberCodeGenerator) SetThisMethod(method *reflect.CtMethod) {
	m.thisMethod = method
	m.CCodeGenerator.typeChecker.thisMethod = method
}

func (m *MemberCodeGenerator) GetThisName() string {
	return JavaToJvmName(m.thisClass.QualifiedName)
}

func (m *MemberCodeGenerator) resolveClassName(jvmClassName string) (string, error) {
	return m.resolver.resolveJvmClassName(jvmClassName)
}

func (m *MemberCodeGenerator) resolveClassName0(name *ast.ASTList) (string, error) {
	return m.resolver.resolveClassName(name)
}

// TODO
func (m *MemberCodeGenerator) getAccessibleConstructor(des string, declClass *reflect.CtClass, method *reflect.CtMethod) string {
	return ""
}

func (m *MemberCodeGenerator) atFieldRead(expr ast.Node) error {
	f, err := m.fieldAccess(expr, true)
	if err != nil {
		return err
	}
	if f == nil {
		return m.atArrayLength(expr)
	}
	is_static := m.resultStatic
	cexpr := getConstantFieldValue0(f)
	if cexpr == nil {
		if _, err := m.atFieldRead1(f, is_static); err != nil {
			return err
		}
	} else {
		if err = cexpr.Accept(m); err != nil {
			return err
		}
		m.setFieldType(f)
	}
	return nil
}

func (m *MemberCodeGenerator) patchGoto(list []int, targetPc int) {
	for _, pc := range list {
		m.bytecodes.Write16bit(pc, targetPc-pc+1)
	}
}

func (m *MemberCodeGenerator) atFieldAssign(expr *ast.AssignExpr, op ast.TokenID, left ast.Node, right ast.Node, dup bool) error {
	f, err := m.fieldAccess(left, false)
	if err != nil {
		return err
	}
	is_static := m.resultStatic
	if op != ast.Assign && !is_static {
		m.bytecodes.AddOpcode(classfile.OpDup)
	}
	fi := 0
	if op == ast.Assign {
		m.setFieldType(f)
		cp := m.bytecodes.pool
		fi = cp.AddFieldRefInfo(cp.AddClassInfo0(f.Declaring), f.Name, f.Descriptor)
	} else {
		fi, err = m.atFieldRead1(f, is_static)
		if err != nil {
			return err
		}
	}
	fType := m.exprType
	fDim := m.arrayDim
	cname := m.className
	if err = m.atAssignCore(expr.Expression, op, right, fType, fDim, cname); err != nil {
		return err
	}

	is2w := is2word(fType, fDim)
	var dup_code classfile.OpCode
	if dup {
		if is_static {
			dup_code = classfile.OpDup
			if is2w {
				dup_code = classfile.OpDup2
			}
		} else {
			dup_code = classfile.OpDupX1
			if is2w {
				dup_code = classfile.OpDup2X1
			}
		}
		m.bytecodes.AddOpcode(dup_code)
	}
	if err = m.atFieldAssignCore(f, is_static, fi, is2w); err != nil {
		return err
	}
	m.exprType = fType
	m.arrayDim = fDim
	m.className = cname
	return nil
}

func (m *MemberCodeGenerator) atArrayVariableAssign(init *ast.ArrayInit, varType ast.TokenID, varArray int, varClass string) error {
	return m.atNewArrayExpr2(varType, nil, varClass, init)
}

func (m *MemberCodeGenerator) AtNewExpr(expr *ast.NewExpr) error {
	if expr.IsArray() {
		return m.AtNewArrayExpr(expr)
	}
	clazz, err := m.resolver.lookupClassByName(expr.GetClassName())
	if err != nil {
		return err
	}
	cname := clazz.QualifiedName
	args := expr.GetArguments()
	m.bytecodes.AddNew(cname)
	m.bytecodes.AddOpcode(classfile.OpDup)
	if err = m.atMethodCallCore(clazz, initMethodName, args, false, true, -1, nil); err != nil {
		return err
	}
	m.exprType = ast.Class
	m.arrayDim = 0
	m.className = JavaToJvmName(cname)
	return nil
}

func (m *MemberCodeGenerator) AtArrayInit(init *ast.ArrayInit) error {
	return NewCompileError("array initializer is not supported")
}

func (m *MemberCodeGenerator) AtMember(member *ast.MemberSymbol) error {
	if member.Identifier == "null" {
		m.bytecodes.AddOpcode(classfile.OpAConstNull)
	}
	return m.atFieldRead(member)

}

func (m *MemberCodeGenerator) AtCallExpr(expr *ast.CallExpr) error {
	if m.ActiveSuper && m.Parent != nil {
		return m.Parent.AtCallExpr(expr)
	}
	m.ActiveSuper = true
	var targetClass *reflect.CtClass
	targetClass = m.thisClass
	mname := ""
	method := expr.Oprand1()
	args := expr.Oprand2().(*ast.ASTList)
	isStatic := false
	isSpecial := false
	aload0pos := -1
	var err error

	var cached *reflect.CtMethod = expr.GetMethod()
	if mem, ok := method.(*ast.MemberSymbol); ok {
		mname = mem.Identifier
		targetClass = m.thisClass
		if m.CCodeGenerator.isStaticMethod || (cached != nil && cached.IsStatic()) {
			isStatic = true // should be static
		} else {
			aload0pos = m.bytecodes.currentPc()
			m.bytecodes.AddAload(0) // load this
		}
	}
	if k, ok := method.(*ast.Keyword); ok { // constructor

		isSpecial = true
		mname = "<init>"
		targetClass = m.thisClass
		if m.CCodeGenerator.isStaticMethod {
			return NewCompileError("a constructor cannot be static")
		}
		m.bytecodes.AddAload(0) // load this
		if k.Name == "super" {
			targetClass, err = m.resolver.lookupClassByJvmName(targetClass.SuperClassName)
			if err != nil {
				return err
			}
		}
	}
	if e, ok := method.(*ast.Expression); ok {
		oprand2 := e.Oprand2()
		switch oprand2.(type) {
		case *ast.Symbol:
			mname = oprand2.(*ast.Symbol).Identifier
			break
		case *ast.MemberSymbol:
			mname = oprand2.(*ast.MemberSymbol).Symbol.Identifier
			break
		default:
			panic("no finish")
		}
		op := e.GetOperator()
		if op == ast.Member {
			targetClass, err = m.resolver.lookupClassByJvmName(e.Oprand1().(*ast.Symbol).Identifier)
			if err != nil {
				return err
			}
			isStatic = true
		} else if op == ast.Dot {
			target := e.Oprand1()
			classFollowedByDotSuper := isDotSuper(target)
			if classFollowedByDotSuper != "" {
				isSpecial = true
				targetClass, err = m.resolver.GetSuperInterface1(m.thisClass, classFollowedByDotSuper)
				if err != nil {
					return err
				}
				if m.inStaticMethod || (cached != nil && cached.Acc.Static) {
					isStatic = true // should be static
				} else {
					aload0pos = m.bytecodes.currentPc()
					m.bytecodes.AddAload(0) // this
				}
			} else {
				if k, ok := target.(*ast.Keyword); ok {
					if k.Name == "super" {
						isSpecial = true
					}
				}
				if err = target.Accept(m); err != nil {
					if nfe, ok := err.(*NoFieldError); ok {
						if nfe.expr != target {
							return nfe
						}
						m.exprType = ast.Class
						m.arrayDim = 0
						m.className = nfe.fieldName
						isStatic = true
					}
					return err
				}
				if m.arrayDim > 0 {
					targetClass, err = m.resolver.lookupClass1(jvmJavaLangObject, true)
					if err != nil {
						return err
					}
				} else if m.exprType == ast.Class {
					targetClass, err = m.resolver.lookupClassByJvmName(m.className)
					if err != nil {
						return err
					}
				} else {
					return NewCompileError("bad method")
				}
			}
		} else {
			return NewCompileError("bad method")
		}
	}

	return m.atMethodCallCore(targetClass, mname, args, isStatic, isSpecial, aload0pos, cached)
}

func (m *MemberCodeGenerator) atTryStmnt(s *ast.Statement) error {
	body := s.GetLeft().(*ast.Statement)
	if body == nil {
		return nil
	}
	catchList := s.GetRight().GetLeft()
	finallyBlock := s.GetRight().GetRight().GetLeft().(*ast.Statement)
	gotoList := make([]int, 0)
	bc := m.bytecodes
	var jsrHook *JsrHook
	if !isNil(finallyBlock) {
		jsrHook = NewJsrHook(m.CCodeGenerator)
	}
	start := bc.currentPc()
	err := body.Accept(m)
	if err != nil {
		return err
	}
	end := bc.currentPc()
	if start == end {
		return NewCompileError("empty try block")
	}
	tryNotReturn := !m.hasReturn
	if tryNotReturn {
		gotoList = append(gotoList, bc.currentPc())
		bc.AddInstruction(classfile.NewInstruction(classfile.OpGoto).AddIndex(0))
	}
	varIndex := bc.MaxLocals
	bc.incMaxLocals(1)
	for !isNil(catchList) {
		// catch clause
		ptr := catchList.GetLeft()
		if isNil(ptr) {
			catchList = catchList.GetRight()
			continue
		}
		p := ptr.(*ast.Pair)
		catchList = catchList.GetRight()
		d := p.GetLeft().(*ast.Declarator)
		block := p.GetRight().(*ast.Statement)
		d.SetLocalVar(varIndex)
		varType, err := m.resolver.lookupClassByJvmName(d.GetClassName())
		if err != nil {
			return err
		}
		cname := varType.QualifiedName
		d.SetClassName(JavaToJvmName(cname))
		bc.AddExceptionHandler(start, end, bc.currentPc(), cname)
		bc.growStack(1)
		bc.AddAStore(varIndex)
		m.hasReturn = false
		if block != nil {
			err := block.Accept(m)
			if err != nil {
				return err
			}
		}
		if !m.hasReturn {
			gotoList = append(gotoList, bc.currentPc())
			bc.AddInstruction(classfile.NewInstruction(classfile.OpGoto).AddIndex(0))
			tryNotReturn = true
		}
	}

	if finallyBlock != nil {
		jsrHook.Remove(m.CCodeGenerator)
		// catch (any) clause
		pcAnyCatch := bc.currentPc()
		bc.tryblocks.Add(start, pcAnyCatch, pcAnyCatch, 0)
		bc.growStack(1)
		bc.AddAStore(varIndex)
		m.hasReturn = false
		err := finallyBlock.Accept(m)
		if err != nil {
			return err
		}
		if !m.hasReturn {
			bc.AddAload(varIndex)
			bc.AddOpcode(classfile.OpAThrow)
		}

		m.addFinally(jsrHook.jsrList, finallyBlock)
	}
	pcEnd := bc.currentPc()
	m.patchGoto(gotoList, pcEnd)
	m.hasReturn = !tryNotReturn
	if finallyBlock != nil {
		if tryNotReturn {
			return finallyBlock.Accept(m)
		}
	}
	return nil
}

func (m *MemberCodeGenerator) atMethodCallCore(class *reflect.CtClass, mname string, args *ast.ASTList, isStatic, isSpecial bool, aload0pos int, found *reflect.CtMethod) error {
	nargs := ast.ASTListLength(args)
	types := make([]ast.TokenID, nargs)
	dims := make([]int, nargs)
	cnames := make([]string, nargs)
	if !isStatic && found != nil && found.IsStatic() {
		m.bytecodes.AddOpcode(classfile.OpPop)
		isStatic = true
	}
	//stack := m.bytecodes.stackDepth
	// generate code for evaluating arguments.
	err := m.atMethodArgs(args, types, dims, cnames)
	if err != nil {
		return err
	}
	if found == nil {
		found = m.resolver.lookupMethod(class, m.thisClass, m.thisMethod, mname, types, dims, cnames)
	}
	if found == nil {
		return NewCompileError("class not found2")
	}
	return m.atMethodCallCore2(class, mname, isStatic, isSpecial, aload0pos, found)
}

func (m *MemberCodeGenerator) atMethodCallCore2(targetClass *reflect.CtClass, mname string, isStatic, isSpecial bool, aload0pos int, found *reflect.CtMethod) error {
	des := found.Descriptor
	declClass := found.Class
	acc := found.Acc
	if mname == initMethodName {
		isSpecial = true
		if !declClass.Equal(targetClass) {
			return NewCompileError("no such constructor: " + targetClass.QualifiedName)
		}
		if declClass != m.thisClass && acc.Private {
			// <= java 11
			if declClass.ClassFile.MajorVersion <= 55 && isFromSameDeclaringClass(declClass, m.thisClass) {
				des = m.getAccessibleConstructor(des, declClass, found)
				m.bytecodes.AddOpcode(classfile.OpAConstNull) // the last parameter
			}
		}
	} else if acc.Private {
		if declClass == m.thisClass {
			isSpecial = true
		} else {
			isSpecial = false
			isStatic = true
			//pass
			//if ((acc & AccessFlag.STATIC) == 0)
			//	desc = Descriptor.insertParameter(declClass.getName(),
			//	origDesc);
			//
			//acc = AccessFlag.setPackage(acc) | AccessFlag.STATIC;
			//mname = getAccessiblePrivate(mname, origDesc, desc,
			//	minfo, declClass);
		}
	}
	popTarget := false
	if acc.Static {
		if !isStatic {
			/* this method is static but the target object is
			   on stack.  It must be popped out.  If aload0pos >= 0,
			   then the target object was pushed by aload_0.  It is
			   overwritten by NOP.
			*/
			isStatic = true
			if aload0pos >= 0 {
				m.bytecodes.Write(aload0pos, classfile.OpNop)
			} else {
				popTarget = true
			}
		}
		m.bytecodes.AddInvokeStatic2(targetClass, mname, des)
	} else if isSpecial {
		m.bytecodes.AddInvokeSpecial(targetClass.QualifiedName, mname, des)
	} else {
		if declClass.Acc.Public || declClass.IsInterface() != targetClass.IsInterface() {
			declClass = targetClass
			if declClass.IsInterface() {
				nargs := ParamSize(des) + 1
				m.bytecodes.AddInvokeInterface(declClass, mname, des, nargs)
			} else {
				if isStatic {
					return NewCompileError(mname + " is not static")
				} else {
					m.bytecodes.AddInvokeVirtual2(declClass, mname, des)
				}

			}
		}
	}
	return m.setReturnType(des, isStatic, popTarget)

}

func ParamSize(des string) int {
	return 0
}

func (m *MemberCodeGenerator) atMethodArgs(args *ast.ASTList, types []ast.TokenID, dims []int, cnames []string) error {
	i := 0
	for args != nil {
		a := args.Head()
		err := a.Accept(m)
		if err != nil {
			return err
		}
		types[i] = m.exprType
		dims[i] = m.arrayDim
		cnames[i] = m.className
		i++
		args = args.Tail()
	}
	return nil
}

func (m *MemberCodeGenerator) setReturnType(des string, isStatic, popTarget bool) error {
	i := strings.Index(des, ")")
	if i < 0 {
		return errors.New("invalid method descriptor")
	}
	i++ // 进入返回值部分

	// 解析数组维度
	dim := 0
	for i < len(des) && des[i] == '[' {
		dim++
		i++
	}

	m.arrayDim = dim

	// 解析返回类型
	if i < len(des) && des[i] == 'L' {
		// 解析类名，如 "Ljava/lang/String;"
		j := strings.Index(des[i+1:], ";")
		if j < 0 {
			return NewCompileError("invalid class type in descriptor")
		}
		m.exprType = ast.Class
		m.className = des[i+1 : i+1+j]
	} else if i < len(des) {
		m.exprType = descToType(rune(des[i])) // 转换基础类型
		m.className = ""
	} else {
		return NewCompileError("unexpected end of descriptor")
	}

	etype := m.exprType

	// 处理静态方法的返回值调整
	if isStatic && popTarget {
		if is2word(etype, dim) {
			m.bytecodes.AddOpcode(classfile.OpDup2X1)
			m.bytecodes.AddOpcode(classfile.OpPop2)
			m.bytecodes.AddOpcode(classfile.OpPop)
		} else if etype == ast.Void {
			m.bytecodes.AddOpcode(classfile.OpPop)
		} else {
			m.bytecodes.AddOpcode(classfile.OpSwap)
			m.bytecodes.AddOpcode(classfile.OpPop)
		}
	}

	return nil
}

func (m *MemberCodeGenerator) makeParamList(md *ast.MethodDecl) ([]*reflect.CtClass, error) {
	plist := md.GetParams()
	if isNil(plist) {
		return make([]*reflect.CtClass, 0), nil
	} else {
		i := 0
		params := make([]*reflect.CtClass, plist.Length())
		for plist != nil {
			param, err := m.resolver.lookupClass(plist.Head().(*ast.Declarator))
			if err != nil {
				return nil, err
			}
			params[i] = param
			i++
			plist = plist.Tail()
		}
		return params, nil
	}
}

func (m *MemberCodeGenerator) makeThrowsList(md *ast.MethodDecl) ([]*reflect.CtClass, error) {
	clist := make([]*reflect.CtClass, 0)
	list := md.GetThrows()
	if isNil(list) {
		return make([]*reflect.CtClass, 0), nil
	}
	i := 0
	for !isNil(list) {
		clazz, err := m.resolver.lookupClassByName(list.Head().(*ast.ASTList))
		if err != nil {
			return nil, err
		}
		clist[i] = clazz
		i++
		list = list.Tail()
	}
	return clist, nil
}

func (m *MemberCodeGenerator) atNewArrayExpr2(varType ast.TokenID, sizeExpr ast.Node, jvmClassname string, init *ast.ArrayInit) error {
	if init == nil {
		if sizeExpr == nil {
			return NewCompileError("no array size")
		}
		if err := sizeExpr.Accept(m); err != nil {
			return err
		}
	} else {
		if sizeExpr == nil {
			s := init.Size()
			m.bytecodes.AddIconst(s)
		} else {
			return NewCompileError("unnecessary array size specified for new")
		}
	}

	var elementClass string
	var err error
	if varType == ast.Class {
		elementClass, err = m.resolveClassName(jvmClassname)
		if err != nil {
			return err
		}
		m.bytecodes.addAnewArray(JvmToJavaName(elementClass))
	} else {
		var atype int
		switch varType {
		case ast.Boolean:
			atype = reflect.T_BOOLEAN
			break
		case ast.Char:
			atype = reflect.T_CHAR
			break
		case ast.Float:
			atype = reflect.T_FLOAT
			break
		case ast.Double:
			atype = reflect.T_DOUBLE
		case ast.Byte:
			atype = reflect.T_BYTE
		case ast.Short:
			atype = reflect.T_SHORT
		case ast.Int:
			atype = reflect.T_INT
		case ast.Long:
			atype = reflect.T_LONG
		default:
			return NewCompileError("invalid type for new array")
		}
		m.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpNewArray).Add(atype))
	}

	if init != nil {
		s := init.Size()
		list := init.ASTList
		for i := 0; i < s; i++ {
			m.bytecodes.AddOpcode(classfile.OpDup)
			m.bytecodes.AddIconst(i)
			if err = list.Head().Accept(m); err != nil {
				return err
			}
			if !isRefType(varType) {
				if err = m.atNumCastExpr(m.exprType, varType); err != nil {
					return err
				}
			}
			m.bytecodes.AddOpcode(getArrayWriteOp(varType, 0))
			list = list.Tail()
		}
	}

	m.exprType = varType
	m.arrayDim = 1
	m.className = elementClass
	return nil
}

func (m *MemberCodeGenerator) AtNewArrayExpr(expr *ast.NewExpr) error {
	varType := expr.GetArrayType()
	size := expr.GetArraySize()
	className := expr.GetClassName()
	init := expr.GetInitializer()

	if size.Length() > 1 {
		if !isNil(init) {
			return NewCompileError("multi-dimensional array initializer for new is not supported")
		}
		return m.atMultiNewArray(varType, className, size)
	}

	sizeExpr := size.Head()
	return m.atNewArrayExpr2(varType, sizeExpr, ast.AstToClassName(className, '/'), init)
}

func (m *MemberCodeGenerator) atMultiNewArray(varType ast.TokenID, className *ast.ASTList, size *ast.ASTList) error {
	dim := size.Length()
	count := 0

	for cur := size; cur != nil; cur = cur.Tail() {
		s := cur.Head()
		if isNil(s) {
			break // 处理 int[][][] a = new int[3][4][]
		}

		count++
		if err := s.Accept(m); err != nil {
			return err
		}
		if m.exprType != ast.Int {
			return NewCompileError("bad type for array size")
		}
	}

	var descriptor string
	m.exprType = varType
	m.arrayDim = dim

	if varType == ast.Class {
		name, err := m.resolver.resolveClassName(className)
		if err != nil {
			return err
		}
		m.className = name
		descriptor = toJvmArrayName(m.className, dim)
	} else {
		descriptor = toJvmTypeName(varType, dim)
	}

	m.bytecodes.AddMultiNewArray(descriptor, count)
	return nil
}

func (m *MemberCodeGenerator) fieldAccess(expr ast.Node, acceptLength bool) (*reflect.CtField, error) {
	// 如果是 成员访问表达式
	if memberExpr, ok := expr.(*ast.MemberSymbol); ok {
		name := memberExpr.Symbol.Identifier
		f, err := m.thisClass.GetField(name)
		if err != nil {
			return nil, err
		}
		is_static := f.Acc.Static
		if m.inStaticMethod {
			return nil, NewCompileError("not available in a static method: " + name)
		} else {
			m.bytecodes.AddAload(0) // load this
		}
		m.resultStatic = is_static
		return f, nil
	} else if expression, ok := expr.(*ast.Expression); ok {
		op := expression.GetOperator()
		if op == ast.Member {
			name := expression.Oprand1().(*ast.Symbol).Identifier
			symbol := expression.Oprand2().(*ast.Symbol)
			f, err := m.resolver.lookupField(name, symbol)
			if err != nil {
				return nil, err
			}
			m.resultStatic = true
			return f, nil
		} else if op == ast.Dot {
			var f *reflect.CtField
			var err error
			if err = expression.Oprand1().Accept(m); err != nil {
				return nil, err
			}
			if m.exprType == ast.Class && m.arrayDim == 0 {
				var symbol *ast.Symbol
				var okk bool
				if symbol, okk = expression.Oprand2().(*ast.Symbol); !okk {
					symbol = expression.Oprand2().(*ast.MemberSymbol).Symbol
				}
				f, err = m.resolver.lookupFieldByJvmName(m.className, symbol)
			} else if acceptLength && m.arrayDim > 0 && (expression.Oprand2().(*ast.Symbol).Identifier == "length") {
				// expr is an array length.
				return f, nil
			} else {
				return nil, badLvalue()
			}
			if err != nil {
				return nil, err
			}
			is_static := f.Acc.Static
			if is_static {
				m.bytecodes.AddOpcode(classfile.OpPop)
			}
			m.resultStatic = is_static
			return f, nil
		} else {
			return nil, badLvalue()
		}
	} else {
		return nil, badLvalue()
	}
}

func (m *MemberCodeGenerator) atArrayLength(expr ast.Node) error {
	if m.arrayDim == 0 {
		return NewCompileError(".length applied to a non array")
	}
	m.bytecodes.AddOpcode(classfile.OpArrayLength)
	m.exprType = ast.Int
	m.arrayDim = 0
	return nil
}

func (m *MemberCodeGenerator) atFieldRead1(f *reflect.CtField, isStatic bool) (int, error) {
	is2byte := m.setFieldType(f)
	if err := m.IsAccessibleField(f); err != nil {
		return 0, err
	}
	// TODO
	//if (maker != null) {
	//	MethodInfo minfo = maker.getFieldGetter(finfo, isStatic);
	//	bytecode.addInvokestatic(f.getDeclaringClass(), minfo.getName(),
	//		minfo.getDescriptor());
	//	return 0;
	// }
	fi := m.addFieldRefInfo(f)
	if isStatic {
		val := 1
		if is2byte {
			val = 2
		}
		m.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpGetStatic).AddIndex(fi))
		m.bytecodes.growStack(val)
	} else {
		val := 0
		if is2byte {
			val = 1
		}
		m.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpGetField).AddIndex(fi))
		m.bytecodes.growStack(val)
	}
	return fi, nil
}

func (m *MemberCodeGenerator) setFieldType(f *reflect.CtField) bool {
	descriptor := f.Descriptor
	dim := 0
	i := 0
	var c rune
	for i, c = range descriptor {
		if c == '[' {
			dim++
		} else {
			break
		}
	}
	m.arrayDim = dim
	m.exprType = descToType(c)
	if c == 'L' {
		m.className = descriptor[i+1 : strings.Index(descriptor, ";")]
	} else {
		m.className = ""
	}
	return dim == 0 && (c == 'J' || c == 'D')
}

func (m *MemberCodeGenerator) IsAccessibleField(f *reflect.CtField) error {
	if f.Acc.Private && f.Declaring != m.thisClass {
		return NewCompileError("Field " + f.Name + " in " + f.Declaring.QualifiedName + " is private.")
		// pass check Enclosing
	}
	return nil
}

func (m *MemberCodeGenerator) addFieldRefInfo(f *reflect.CtField) int {
	cp := m.bytecodes.pool
	cname := f.Declaring.QualifiedName
	ci := cp.AddClassInfo(cname)
	return cp.AddFieldRefInfo(ci, f.Name, f.Descriptor)
}

func (m *MemberCodeGenerator) addFinally(returnList [][]int, finallyBlock *ast.Statement) error {
	bc := m.bytecodes
	for _, ret := range returnList {
		pc := ret[0]
		bc.Write16bit(pc, bc.currentPc()-pc+1)
		//ReturnHook hook = new JsrHook2(this, ret);
		if err := finallyBlock.Accept(m); err != nil {
			return err
		}
		// hook.remove(this);
		if !m.hasReturn {
			bc.AddOpcode(classfile.OpGoto)
			bc.WriteIndex(pc + 3 - bc.currentPc())
		}
	}
	return nil
}

func (m *MemberCodeGenerator) atFieldAssignCore(f *reflect.CtField, is_static bool, fi int, is2byte bool) error {
	if fi != 0 {
		if is_static {
			m.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpPutStatic).AddIndex(fi))
			val := -1
			if is2byte {
				val = -2
			}
			m.bytecodes.growStack(val)
		} else {
			m.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpPutField).AddIndex(fi))
			val := -2
			if is2byte {
				val = -3
			}
			m.bytecodes.growStack(val)
		}
	} else {
		declClass := f.Declaring
		setter, err := f.GetFieldSetter()
		if err != nil {
			return err
		}
		m.bytecodes.AddInvokeStatic(declClass.QualifiedName, setter.Name, setter.Descriptor)
	}
	return nil
}

func (m *MemberCodeGenerator) insertDefaultSuperCall() {
	m.bytecodes.AddAload(0)
	m.bytecodes.AddInvokeSpecial(m.thisClass.SuperClassName, "<init>", "()V")
}

func (m *MemberCodeGenerator) GetSuperName() string {
	return JavaToJvmName(m.thisClass.SuperClassName)
}

func badLvalue() error {
	return NewCompileError("bad lvalue")
}
