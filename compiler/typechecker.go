package compiler

import (
	"GoJavassist/compiler/ast"
	"GoJavassist/compiler/reflect"
	"errors"
	"fmt"
	"strings"
)

const (
	javaLangObject    = "java.lang.Object"
	jvmJavaLangObject = "java/lang/Object"
	jvmJavaLangString = "java/lang/String"
	jvmJavaLangClass  = "java/lang/Class"
	javaLangString    = "java.lang.String"
)

// TypeChecker java 类型检查器
type TypeChecker struct {
	ast.Visitor
	resolver   *MemberResolver
	thisClass  *reflect.CtClass
	thisMethod *reflect.CtMethod
	exprType   ast.TokenID
	arrayDim   int
	parent     *JvstTypeChecker
	className  string
}

func (j *TypeChecker) AtCondExpr(e *ast.CondExpr) error {
	if err := j.booleanExpr(e.CondExpr()); err != nil {
		return err
	}
	if err := e.ThenExpr().Accept(j); err != nil {
		return err
	}
	type1 := j.exprType
	dim1 := j.arrayDim
	if err := e.ElseExpr().Accept(j); err != nil {
		return err
	}
	if dim1 == 0 && dim1 == j.arrayDim {
		if rightIsStrong(type1, j.exprType) {
			e.SetThenExpr(ast.NewCastExprWithType(j.exprType, 0, e.ThenExpr()))
		} else if rightIsStrong(j.exprType, type1) {
			e.SetElseExpr(ast.NewCastExprWithType(type1, 0, e.ElseExpr()))
			j.exprType = type1
		}
	}
	return nil
}

func (j *TypeChecker) AtBinExpr(expr *ast.BinExpr) error {
	token := expr.GetOperator()
	k := ast.LookupBinOp(token)
	if k > 0 {
		/* arithmetic operators: +, -, *, /, %, |, ^, &, <<, >>, >>>
		 */
		if token == ast.Plus {
			e, err := j.atPlusExpr(expr)
			if err != nil {
				return err
			}
			if !isNil(e) {
				/* String concatenation has been translated into
				 * an expression using StringBuffer.
				 */
				e = ast.MakeCall(ast.MakeExpression(ast.Dot, e, ast.NewMemberSymbol("toString")), nil).Expression
				expr.SetOprand1(e)
				expr.SetOprand2(nil) // <---- look at this!
				j.className = jvmJavaLangClass
			}
		} else {
			left := expr.Oprand1()
			right := expr.Oprand2()
			if err := left.Accept(j); err != nil {
				return err
			}
			type1 := j.exprType
			if err := right.Accept(j); err != nil {
				return err
			}
			isConstant, err := j.isConstant(expr, token, left, right)
			if err != nil {
				return err
			}
			if !isConstant {
				return j.computeBinExprType(expr, token, type1)
			}
		}
	} else {
		return j.booleanExpr(expr)
	}
	return nil
}

func (j *TypeChecker) AtIntConst(i *ast.IntConst) error {
	j.arrayDim = 0
	atype := i.TypeID
	if atype == ast.IntConstant {
		j.exprType = ast.Int
	} else if atype == ast.CharConstant {
		j.exprType = ast.Char
	} else {
		j.exprType = ast.Long
	}
	return nil
}

func (j *TypeChecker) AtDoubleConst(d *ast.DoubleConst) error {
	j.arrayDim = 0
	if d.GetType() == ast.DoubleConstant {
		j.exprType = ast.Double
	} else {
		j.exprType = ast.Float
	}
	return nil
}

func (j *TypeChecker) AtInstanceOfExpr(i *ast.InstanceOfExpr) error {
	if err := i.GetOprand().Accept(j); err != nil {
		return err
	}
	j.exprType = ast.Boolean
	j.arrayDim = 0
	return nil
}

func (j *TypeChecker) AtNewExpr(expr *ast.NewExpr) error {
	if expr.IsArray() {
		return j.AtNewArrayExpr(expr)
	} else {
		clazz, err := j.resolver.lookupClassByName(expr.GetClassName())
		if err != nil {
			return err
		}
		cname := clazz.QualifiedName
		args := expr.GetArguments()
		if _, err = j.atMethodCallCore(clazz, "<init>", args); err != nil {
			return err
		}
		j.exprType = ast.Expr
		j.arrayDim = 0
		j.className = JavaToJvmName(cname)
	}
	return nil
}

func (j *TypeChecker) AtNewArrayExpr(expr *ast.NewExpr) error {
	typeID := expr.GetArrayType()
	size := expr.GetArraySize()
	className := expr.GetClassName()
	init := expr.GetInitializer()
	if init != nil {
		err := init.Accept(j)
		if err != nil {
			return err
		}
	}
	if size.Length() > 1 {
		err := j.AtMultiNewArray(typeID, className, size)
		if err != nil {
			return err
		}
	} else {
		sizeExpr := size.Head()
		if !isNil(sizeExpr) {
			if err := sizeExpr.Accept(j); err != nil {
				return err
			}
		}
		j.exprType = typeID
		j.arrayDim = 1
		if typeID == ast.Class {
			name, err := j.resolver.resolveClassName(className)
			if err != nil {
				return err
			}
			j.className = name
		} else {
			j.className = ""
		}
	}
	return nil
}

func (j *TypeChecker) AtMultiNewArray(id ast.TokenID, name *ast.ASTList, size *ast.ASTList) error {
	dim := size.Length()
	v := ast.Node(size)
	for i := 0; !isNil(v); v = v.GetRight() {
		s := v.(*ast.ASTList).Head().(*ast.ASTList)
		if isNil(s) {
			break
		}
		i++
		err := s.Accept(j)
		if err != nil {
			return err
		}
		j.exprType = id
		j.arrayDim = dim
		if id == ast.Class {
			cname, err := j.resolver.resolveClassName(name)
			if err != nil {
				return err
			}
			j.className = cname
		}
	}
	return nil
}

func (j *TypeChecker) AtStringL(s *ast.StringLiteral) error {
	j.exprType = ast.Class
	j.arrayDim = 0
	j.className = jvmJavaLangString
	return nil
}

func (j *TypeChecker) AtVariable(v *ast.Variable) error {
	d := v.Declarator
	j.exprType = d.GetType()
	j.arrayDim = d.GetArrayDim()
	j.className = d.GetClassName()
	return nil
}

func (j *TypeChecker) AtArrayInit(init *ast.ArrayInit) error {
	list := init.ASTList
	var h ast.Node
	for !isNil(list) {
		h = list.Head()
		list = list.Tail()
		if !isNil(h) {
			if err := h.Accept(j); err != nil {
				return err
			}
		}
	}
	return nil
}

func (j *TypeChecker) AtCallExpr(expr *ast.CallExpr) error {
	var (
		mname       string
		targetClass *reflect.CtClass
	)
	method := expr.Oprand1()
	args := expr.Oprand2().(*ast.ASTList)
	var err error
	switch method.(type) {
	case *ast.MemberSymbol:
		m := method.(*ast.MemberSymbol)
		mname = m.Symbol.Identifier
		targetClass = j.thisClass
	case *ast.Keyword:
		mname = "<init>" // "<init>"
		key := method.(*ast.Keyword)
		if key.TokenID == ast.Super {
			targetClass, err = j.resolver.lookupClassByJvmName(j.thisClass.SuperClassName)
			if err != nil {
				return err
			}
		} else {
			targetClass = j.thisClass
		}
	case *ast.Expression:

		e := method.(*ast.Expression)

		mname = ast.GetIdentifier(e.Oprand2())
		op := e.GetOperator()
		switch op {
		case ast.Member:
			targetClass, err = j.resolver.lookupClass1(e.Oprand1().(*ast.Symbol).Identifier, false)
			if err != nil {
				return err
			}
			break
		case ast.Dot:
			target := e.Oprand1()
			if classFollowedByDotSuper := j.resolver.isDotSuper(target); classFollowedByDotSuper != nil {
				targetClass = j.resolver.GetSuperInterface(j.thisClass, classFollowedByDotSuper)
			} else {
				if err := target.Accept(j); err != nil {
					if nfe, ok := err.(*NoFieldError); ok {
						if nfe.Expr() != target {
							return nfe
						}
						j.exprType = ast.Class
						j.arrayDim = 0
						j.className = nfe.fieldName
						e.SetOperator(ast.Member)
						e.SetOprand1(ast.NewSymbol(JvmToJavaName(j.className)))
					} else {
						return err
					}
				}

				if j.arrayDim > 0 {
					targetClass, err = j.resolver.lookupClass1(javaLangObject, true)
				} else if j.exprType == ast.Class {
					targetClass, err = j.resolver.lookupClassByJvmName(j.className)
				} else {
					return NewCompileError("bad method")
				}
			}
		default:
			return NewCompileError("bad method")
		}
	default:
		return NewCompileError("bad method")
	}
	minfo, err := j.atMethodCallCore(targetClass, mname, args)
	if err != nil {
		return err
	}
	expr.SetMethod(minfo)
	return nil
}

func (t *TypeChecker) AtMember(m *ast.MemberSymbol) error {
	return t.atFieldRead(m)
}

func (t *TypeChecker) atMethodCallCore(targetClass *reflect.CtClass, mname string, args *ast.ASTList) (*reflect.CtMethod, error) {
	nargs := args.Length()
	types := make([]ast.TokenID, nargs)
	dims := make([]int, nargs)
	cnames := make([]string, nargs)
	if err := t.parent.atMethodArgs(args, types, dims, cnames); err != nil {
		return nil, err
	}

	found := t.resolver.lookupMethod(targetClass, t.thisClass, t.thisMethod, mname, types, dims, cnames)
	if found == nil {
		clazzName := targetClass.QualifiedName
		signature := targetClass.GetDescriptor() //argTypesToString(types, dims, cnames)
		var msg string
		if mname == "<init>" {
			msg = "cannot find constructor " + clazzName + signature
		} else {
			msg = mname + " not found in " + clazzName
		}
		return nil, fmt.Errorf(msg)
	}

	desc := found.Descriptor
	if err := t.setReturnType(desc); err != nil {
		return nil, err
	}
	return found, nil
}

func (j *TypeChecker) atMethodArgs(args *ast.ASTList, types []ast.TokenID, dims []int, cnames []string) error {
	i := 0
	for !isNil(args) {
		a := args.Head()
		if err := a.Accept(j); err != nil {
			return err
		}
		types[i] = j.exprType
		dims[i] = j.arrayDim
		cnames[i] = j.className
		i++
		args = args.Tail()
	}
	return nil
}

func (t *TypeChecker) setReturnType(desc string) error {
	i := strings.IndexByte(desc, ')')
	if i < 0 {
		return NewCompileError("bad method")
	}

	i++
	c := desc[i]
	dim := 0
	for c == '[' {
		dim++
		i++
		c = desc[i]
	}

	t.arrayDim = dim
	if c == 'L' {
		j := strings.IndexByte(desc[i+1:], ';')
		if j < 0 {
			return NewCompileError("bad method")
		}
		t.exprType = ast.Class
		t.className = desc[i+1 : i+1+j]
	} else {
		t.exprType = descToType(rune(c))
		t.className = ""
	}
	return nil
}

func isDotSuper(target ast.Node) string {
	if expr, ok := target.(*ast.Expression); ok {
		if expr.GetOperator() == ast.Dot {
			right := expr.Oprand2()
			if k, ok := right.(*ast.Keyword); ok {
				if k.Name == "super" {
					return expr.Oprand1().(*ast.Symbol).Identifier
				}
			}
		}
	}
	return ""
}

func (t *TypeChecker) AtKeyword(k *ast.Keyword) error {
	t.arrayDim = 0
	token := k.TokenID
	switch token {
	case ast.True, ast.False:
		t.exprType = ast.Boolean
		break
	case ast.Null:
		t.exprType = ast.Null
		break
	case ast.This:
		t.exprType = ast.Class
		t.className = t.getThisName()
		break
	case ast.Super:
		t.exprType = ast.Class
		t.className = t.getSuperName()
		break
	default:
		return NewCompileError("fatal")
	}
	return nil
}

func (j *TypeChecker) getThisName() string {
	return JavaToJvmName(j.thisClass.QualifiedName)
}

func (j *TypeChecker) getSuperName() string {
	return JavaToJvmName(j.thisClass.SuperClassName)
}

func (t *TypeChecker) AtExpression(e *ast.Expression) error {
	token := e.GetOperator()
	oprand := e.Oprand1()
	if token == ast.Dot {
		member := ast.GetIdentifier(e.Oprand2())
		if member == "length" {
			err := t.atArrayLength(e)
			if err != nil {
				if err2 := t.atFieldRead(e); err2 != nil {
					return err2
				}
			}
		} else if member == "class" {
			return t.atClassObject(e) // .class
		} else {
			return t.atFieldRead(e)
		}
	} else if token == ast.Member {
		member := ast.GetIdentifier(e.Oprand2())
		if member == "class" {
			return t.atClassObject(e) // .class
		} else {
			return t.atFieldRead(e)
		}
	} else if token == ast.Array {
		return t.atArrayRead(oprand, e.Oprand2())
	} else if token == ast.PLUSPLUS || token == ast.MINUSMINUS {
		return t.atPlusPlus(token, oprand, e)
	} else if token == ast.Not {
		return t.booleanExpr(e)
	} else if token == ast.Call {
		return NewCompileError("fatal")
	} else {
		if err := oprand.Accept(t); err != nil {
			return err
		}
		if !t.isConstant2(e, token, oprand) {
			if token == ast.Plus || token == ast.BitNot {
				if typePrecedence(t.exprType) == P_INT {
					t.exprType = ast.Int // type may be BYTE, ...
				}
			}
		}
	}
	return nil
}

func (j *TypeChecker) AtAssignExpr(a *ast.AssignExpr) error {
	op := a.GetOperator()
	left := a.Oprand1()
	right := a.Oprand2()
	if v, ok := left.(*ast.Variable); ok {
		if err := j.atVariableAssign(a.Expression, op, v, v.Declarator, right); err != nil {
			return err
		}
	} else {
		if expr, ok := left.(*ast.Expression); ok {
			if expr.GetOperator() == ast.Array {
				return j.atArrayAssign(expr, op, expr, right)
			}
		}
		return j.atFieldAssign(a.Expression, op, left, right)
	}
	return nil
}

func (j *TypeChecker) atVariableAssign(expr *ast.Expression, op ast.TokenID, v *ast.Variable, declarator *ast.Declarator, right ast.Node) error {
	varType := declarator.GetType()
	varArray := declarator.GetArrayDim()
	varClass := declarator.GetClassName()
	if op != ast.Assign {
		return j.AtVariable(v)
	}
	if err := right.Accept(j); err != nil {
		return err
	}
	j.exprType = varType
	j.arrayDim = varArray
	j.className = varClass
	return nil
}

func (j *TypeChecker) AtCastExpr(e *ast.CastExpr) error {
	cname, err := j.resolver.resolveClassName(e.GetClassName())
	if err != nil {
		return err
	}
	if err := e.GetOprand().Accept(j); err != nil {
		return err
	}
	j.exprType = e.GetType()
	j.arrayDim = e.GetArrayDim()
	j.className = cname
	return nil
}

func (j *TypeChecker) atArrayAssign(expr *ast.Expression, op ast.TokenID, array *ast.Expression, right ast.Node) error {
	if err := j.atArrayRead(array.Oprand1(), array.Oprand2()); err != nil {
		return err
	}
	aType := j.exprType
	aDim := j.arrayDim
	cname := j.className
	if err := right.Accept(j); err != nil {
		return err
	}
	j.exprType = aType
	j.arrayDim = aDim
	j.className = cname
	return nil
}

func (j *TypeChecker) atArrayRead(array ast.Node, index ast.Node) error {
	if err := array.Accept(j); err != nil {
		return err
	}
	atype := j.exprType
	dim := j.arrayDim
	cname := j.className
	if err := index.Accept(j); err != nil {
		return err
	}
	j.exprType = atype
	j.arrayDim = dim - 1
	j.className = cname
	return nil
}

func (j *TypeChecker) atFieldAssign(expr *ast.Expression, op ast.TokenID, left ast.Node, right ast.Node) error {
	f, err := j.fieldAccess(left)
	if err != nil {
		return err
	}
	j.atFieldRead1(f)
	fType := j.exprType
	fDim := j.arrayDim
	cname := j.className
	if err := right.Accept(j); err != nil {
		return err
	}
	j.exprType = fType
	j.arrayDim = fDim
	j.className = cname
	return nil
}

func (j *TypeChecker) fieldAccess(expr ast.Node) (*reflect.CtField, error) {
	if mem, ok := expr.(*ast.MemberSymbol); ok {
		name := mem.Identifier
		field, err := j.thisClass.GetField(name)
		if err != nil {
			return nil, NewNoFieldError(name, expr)
		}
		if field.Acc.Static {
			mem.Field = field
		}
		return field, nil
	}
	if e, ok := expr.(*ast.Expression); ok {
		op := e.GetOperator()
		if op == ast.Member {
			mem := e.Oprand2().(*ast.MemberSymbol)
			f, err := j.resolver.lookupField(ast.GetIdentifier(e.Oprand1()), mem.Symbol)
			if err != nil {
				return nil, err
			}
			mem.Field = f
			return f, nil
		} else if op == ast.Dot {
			if err := e.Oprand1().Accept(j); err != nil {
				var nfe *NoFieldError
				if errors.As(err, &nfe) {
					if nfe.Expr() != e.Oprand1() {
						return nil, err
					}
					return j.fieldAccess2(e, nfe.fieldName)
				}
				return nil, err
			}
			if j.exprType == ast.Class && j.arrayDim == 0 {
				return j.resolver.lookupFieldByJvmName(j.className, e.Oprand2().(*ast.Symbol))
			}
			oprnd1 := e.Oprand1()
			if s, ok := oprnd1.(*ast.Symbol); ok {
				return j.fieldAccess2(e, s.Identifier)
			}
		}
	}
	return nil, NewCompileError("not found field")
}

func (j *TypeChecker) atFieldRead1(f *reflect.CtField) {
	descriptor := f.Descriptor
	i := 0
	dim := 0
	c := descriptor[0]

	for c == '[' {
		dim++
		i++
		c = descriptor[i]
	}

	j.arrayDim = dim
	j.exprType = descToType(rune(c))

	if c == 'L' {
		j.className = descriptor[i:strings.Index(descriptor, ";")]
	} else {
		j.className = ""
	}
}

func (j *TypeChecker) atFieldRead(node ast.Node) error {
	access, err := j.fieldAccess(node)
	if err != nil {
		return err
	}
	j.atFieldRead1(access)
	return nil
}

func (j *TypeChecker) fieldAccess2(e *ast.Expression, jvmClassName string) (*reflect.CtField, error) {
	fname := e.Oprand2().(*ast.MemberSymbol)
	f, err := j.resolver.lookupFieldByJvmName2(jvmClassName, fname, e)
	if err != nil {
		return nil, err
	}
	e.OperatorId = ast.Member
	e.SetOprand1(ast.NewSymbol(JvmToJavaName(jvmClassName)))
	fname.Field = f
	return f, nil
}

func (j *TypeChecker) atArrayLength(expr *ast.Expression) error {
	if err := expr.Oprand1().Accept(j); err != nil {
		return err
	}
	if j.arrayDim == 0 {
		return NewNoFieldError("length", expr)
	}
	j.exprType = ast.Int
	j.arrayDim = 0
	return nil
}

func (j *TypeChecker) atClassObject(e *ast.Expression) error {
	j.exprType = ast.Class
	j.arrayDim = 0
	j.className = jvmJavaLangClass
	return nil
}

func (j *TypeChecker) atPlusPlus(token ast.TokenID, oprand ast.Node, expr *ast.Expression) error {
	isPost := isNil(oprand)
	if isPost {
		oprand = expr.Oprand2()
	}
	if v, ok := oprand.(*ast.Variable); ok {
		d := v.Declarator
		j.exprType = d.GetType()
		j.arrayDim = d.GetArrayDim()
	} else {
		if e, ok := oprand.(*ast.Expression); ok {
			if e.GetOperator() == ast.Array {
				if err := j.atArrayRead(e.Oprand1(), e.Oprand2()); err != nil {
					return err
				}
				t := j.exprType
				if t == ast.Int || t == ast.Byte || t == ast.Char || t == ast.Short {
					j.exprType = ast.Int
				}
				return nil
			}
		}
		return j.atFieldPlusPlus(oprand)
	}
	return nil
}

func (j *TypeChecker) booleanExpr(e ast.Node) error {
	op := getCompOperator(e)
	if op == ast.EQ {
		bexpr := e.(*ast.BinExpr)
		if err := bexpr.Oprand1().Accept(j); err != nil {
			return err
		}
		type1 := j.exprType
		dim1 := j.arrayDim
		if err := bexpr.Oprand2().Accept(j); err != nil {
			return err
		}
		if dim1 == 0 && j.arrayDim == 0 {
			j.insertCast(bexpr, type1, j.exprType)
		}
	} else if op == ast.Not {
		if err := e.(*ast.Expression).Oprand1().Accept(j); err != nil {
			return err
		}
	} else if op == ast.ANDAND || op == ast.OROR {
		bexpr := e.(*ast.BinExpr)
		if err := bexpr.Oprand1().Accept(j); err != nil {
			return err
		}
		if err := bexpr.Oprand2().Accept(j); err != nil {
			return err
		}
	} else {
		//others
		if err := e.Accept(j); err != nil {
			return err
		}
	}
	j.exprType = ast.Boolean
	j.arrayDim = 0
	return nil
}
func rightIsStrong(type1, type2 ast.TokenID) bool {
	type1_p := typePrecedence(type1)
	type2_p := typePrecedence(type2)
	return type1_p >= 0 && type2_p >= 0 && type1_p > type2_p
}

func (j *TypeChecker) isConstant(expr *ast.BinExpr, op ast.TokenID, left, right ast.Node) (bool, error) {
	left = StripPlusExpr(left)
	right = StripPlusExpr(right)
	var newExpr ast.Node
	if str, ok := left.(*ast.StringLiteral); ok && op == ast.Plus {
		if str2, ok := right.(*ast.StringLiteral); ok {
			newExpr = ast.NewStringL(str.Text + str2.Text)
		}
	} else if intConst, ok := left.(*ast.IntConst); ok {
		newExpr = intConst.Compute(ast.TokenMap[op], right)
	} else if doubleConst, ok := left.(*ast.DoubleConst); ok {
		newExpr = doubleConst.Compute(ast.TokenMap[op], right)
	}
	if isNil(newExpr) {
		return false, nil
	}
	expr.SetOperator(ast.Plus)
	expr.SetOprand1(newExpr)
	expr.SetOprand2(nil)
	if err := newExpr.Accept(j); err != nil {
		return false, err
	}
	return true, nil
}

func (j *TypeChecker) isConstant2(expr *ast.Expression, op ast.TokenID, oprand ast.Node) bool {
	oprand = StripPlusExpr(oprand)
	if intConst, ok := oprand.(*ast.IntConst); ok {
		v := intConst.Value
		if op == ast.Minus {
			v = -v
		} else if op == ast.BitNot {
			v = ^v
		} else {
			return false
		}
		intConst.Value = v
	} else if doubleConst, ok := oprand.(*ast.DoubleConst); ok {
		if op == ast.Minus {
			doubleConst.Value = -doubleConst.Value
		} else {
			return false
		}
	} else {
		return false
	}
	expr.SetOperator(ast.Plus)
	return true
}

func (j *TypeChecker) insertCast(bexpr *ast.BinExpr, type1 ast.TokenID, type2 ast.TokenID) {
	if rightIsStrong(type1, type2) {
		bexpr.SetLeft(ast.NewCastExprWithType(type2, 0, bexpr.Oprand1()))
	} else {
		j.exprType = type1
	}
}

func (j *TypeChecker) computeBinExprType(expr *ast.BinExpr, token ast.TokenID, type1 ast.TokenID) error {
	type2 := j.exprType
	if token == ast.LSHIFT || token == ast.RSHIFT || token == ast.ARSHIFT {
		j.exprType = type1
	} else {
		j.insertCast(expr, type1, type2)
	}
	if typePrecedence(j.exprType) == P_INT && j.exprType != ast.Boolean {
		j.exprType = ast.Int
	}
	return nil
}

func isPlusExpr(expr ast.Node) bool {
	if bexpr, ok := expr.(*ast.BinExpr); ok {
		token := bexpr.GetOperator()
		return token == ast.Plus
	}
	return false
}

func (j *TypeChecker) atPlusExpr(expr *ast.BinExpr) (*ast.Expression, error) {
	left := expr.Oprand1()
	right := expr.Oprand2()
	if isNil(right) {
		return nil, left.Accept(j)
	}
	if isPlusExpr(left) {
		newExpr, err := j.atPlusExpr(left.(*ast.BinExpr))
		if err != nil {
			return nil, err
		}
		if newExpr != nil {
			if err := right.Accept(j); err != nil {
				return nil, err
			}
			j.exprType = ast.Class
			j.exprType = 0
			j.className = "java/lang/StringBuffer"
			return makeAppendCall(newExpr, right), nil
		}
	} else {
		if err := left.Accept(j); err != nil {
			return nil, err
		}
	}
	type1 := j.exprType
	dim1 := j.arrayDim
	cname := j.className
	if err := right.Accept(j); err != nil {
		return nil, err
	}
	isConstant, err := j.isConstant(expr, ast.Plus, left, right)
	if err != nil || isConstant {
		return nil, err
	}
	if (type1 == ast.Class && dim1 == 0 && jvmJavaLangString == cname) || (j.exprType == ast.Class && j.arrayDim == 0 && jvmJavaLangString == j.className) {
		sbufClass := ast.MakeASTList(ast.NewSymbol("java"), ast.NewSymbol("lang"), ast.NewSymbol("StringBuffer"))
		e := ast.NewNewExprWithClass(sbufClass, nil)
		j.exprType = ast.Class
		j.arrayDim = 0
		j.className = "java/lang/StringBuffer"
		return makeAppendCall(makeAppendCall(e, left), right), nil
	}
	return nil, j.computeBinExprType(expr, ast.Plus, type1)
}

func (j *TypeChecker) atFieldPlusPlus(oprand ast.Node) error {
	f, err := j.fieldAccess(oprand)
	if err != nil {
		return err
	}
	j.atFieldRead1(f)
	t := j.exprType
	if t == ast.Int || t == ast.Byte || t == ast.Char || t == ast.Short {
		j.exprType = ast.Int
	}
	return nil
}

func makeAppendCall(target, arg ast.Node) *ast.Expression {
	return ast.MakeCall(ast.MakeExpression(ast.Dot, target, ast.NewMemberSymbol("append")), ast.NewASTListSingle(arg)).Expression
}
