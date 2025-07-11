package compiler

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
	"slices"
	"strings"
)

// 类型转化指令
var (
	castOp = []classfile.OpCode{
		/*            D    F    L    I */
		/* double */ classfile.OpNop, classfile.OpD2F, classfile.OpD2L, classfile.OpD2I,
		/* float  */ classfile.OpF2D, classfile.OpPop, classfile.OpF2L, classfile.OpF2I,
		/* long   */ classfile.OpL2D, classfile.OpL2D, classfile.OpNop, classfile.OpL2I,
		/* other  */ classfile.OpI2D, classfile.OpI2F, classfile.OpI2L, classfile.OpNop,
	}

	ifOp = map[ast.TokenID][]classfile.OpCode{
		ast.EQ:  {classfile.OpIfICmpEQ, classfile.OpIfICmpNE},
		ast.NEQ: {classfile.OpIfICmpNE, classfile.OpIfICmpEQ},
		ast.LE:  {classfile.OpIfICmpLE, classfile.OpIfICmpGT},
		ast.GE:  {classfile.OpIfICmpGE, classfile.OpIfICmpLT},
		ast.LT:  {classfile.OpIfICmpLT, classfile.OpIfICmpGE},
		ast.GT:  {classfile.OpIfICmpGT, classfile.OpIfICmpLE},
	}

	ifOp2 = map[ast.TokenID][]classfile.OpCode{
		ast.EQ:  {classfile.OpIfEQ, classfile.OpIfNE},
		ast.NEQ: {classfile.OpIfNE, classfile.OpIfEQ},
		ast.LE:  {classfile.OpIfLE, classfile.OpIfGT},
		ast.GE:  {classfile.OpIfGE, classfile.OpIfLT},
		ast.LT:  {classfile.OpIfLT, classfile.OpIfGE},
		ast.GT:  {classfile.OpIfGT, classfile.OpIfLE},
	}

	binOp = map[ast.TokenID][]classfile.OpCode{
		ast.Plus:     {classfile.OpDAdd, classfile.OpFAdd, classfile.OpLAdd, classfile.OpIAdd},
		ast.Minus:    {classfile.OpDSub, classfile.OpFSub, classfile.OpLSub, classfile.OpISub},
		ast.Multiply: {classfile.OpDMul, classfile.OpFMul, classfile.OpLMul, classfile.OpIMul},
		ast.Divide:   {classfile.OpDDiv, classfile.OpFDiv, classfile.OpLDiv, classfile.OpIDiv},
		ast.Mod:      {classfile.OpDRem, classfile.OpFRem, classfile.OpLRem, classfile.OpIRem},
		ast.Or:       {classfile.OpNop, classfile.OpNop, classfile.OpLOr, classfile.OpIOr},
		ast.Xor:      {classfile.OpNop, classfile.OpNop, classfile.OpLXor, classfile.OpIXor},
		ast.And:      {classfile.OpNop, classfile.OpNop, classfile.OpLAnd, classfile.OpIAnd},
		ast.LSHIFT:   {classfile.OpNop, classfile.OpNop, classfile.OpLShl, classfile.OpIShl},
		ast.RSHIFT:   {classfile.OpNop, classfile.OpNop, classfile.OpLShr, classfile.OpIShr},
		ast.ARSHIFT:  {classfile.OpNop, classfile.OpNop, classfile.OpLUshr, classfile.OpIUshr},
	}
)

// CCodeGenerator 基础的 java代码生成器
type CCodeGenerator struct {
	ast.Visitor
	returnHooks    *ReturnHook
	bytecodes      *ByteCodes
	arrayDim       int
	tempVar        int
	exprType       ast.TokenID
	typeChecker    *TypeChecker
	className      string
	isStaticMethod bool
	hasReturn      bool
	breakList      []int
	continueList   []int
	Parent         *MemberCodeGenerator

	atFieldPlusPlus func()
}

func NewCCodeGenerator(pool *reflect.ConstPool) *CCodeGenerator {
	return &CCodeGenerator{
		bytecodes:   MakeByteCodes(pool),
		typeChecker: &TypeChecker{},
	}
}

func (g *CCodeGenerator) SetParent(parent *MemberCodeGenerator) {
	g.Parent = parent
}

func (g *CCodeGenerator) GetBytecodes() *ByteCodes {
	return g.bytecodes
}

func (g *CCodeGenerator) AtStringL(s *ast.StringLiteral) error {
	g.exprType = ast.Class
	g.arrayDim = 0
	g.bytecodes.addLdc(g.bytecodes.AddStringRef(s.Text))
	return nil
}

func (g *CCodeGenerator) AtIntConst(i *ast.IntConst) error {
	g.arrayDim = 0
	typeId := i.TypeID
	if typeId == ast.IntConstant || typeId == ast.CharConstant {
		if typeId == ast.IntConstant {
			g.exprType = ast.Int
		} else {
			g.exprType = ast.Char
		}
		g.bytecodes.addIconst(int(i.Value))
	} else {
		g.exprType = ast.Long
		g.bytecodes.addLconst(i.Value)
	}
	return nil
}

func (g *CCodeGenerator) AtDoubleConst(d *ast.DoubleConst) error {
	g.arrayDim = 0
	if d.GetType() == ast.DoubleConstant {
		g.exprType = ast.Double
		g.bytecodes.addDconst(d.Value)
	} else {
		g.exprType = ast.Float
		g.bytecodes.addFconst(float32(d.Value))
	}
	return nil
}

func (g *CCodeGenerator) AtKeyword(k *ast.Keyword) error {
	g.arrayDim = 0
	tokenID := k.TokenID
	switch tokenID {
	case ast.True:
		g.bytecodes.addIconst(1)
		g.exprType = ast.Boolean
		break
	case ast.False:
		g.bytecodes.addIconst(0)
		g.exprType = ast.Boolean
		break
	case ast.Null:
		g.bytecodes.AddOpcode(classfile.OpAConstNull)
		g.exprType = ast.Null
		break
	case ast.This:
	case ast.Super:
		if g.isStaticMethod {
			return NewCompileError("not-available:" + tokenID.GetName())
		}
		g.bytecodes.addAload(0)
		g.exprType = ast.Class
		//getThisName()
		//getSuperName()
		break
	default:
		return NewCompileError("fatal,unknown-token:" + tokenID.GetName())
	}
	return nil
}

func (g *CCodeGenerator) AtVariable(v *ast.Variable) error {
	if v == nil {
		return nil
	}
	d := v.Declarator
	g.exprType = d.GetType()
	g.arrayDim = d.GetArrayDim()
	g.className = d.GetClassName()
	varIndex := g.getLocalVar(d)
	if g.arrayDim > 0 {
		g.bytecodes.addAload(varIndex)
	} else {
		switch g.exprType {
		case ast.Class:
			g.bytecodes.addAload(varIndex)
			break
		case ast.Long:
			g.bytecodes.AddLload(varIndex)
			break
		case ast.Float:
			g.bytecodes.addFload(varIndex)
			break
		case ast.Double:
			g.bytecodes.addDload(varIndex)
			break
		default: // BOOLEAN, BYTE, CHAR, SHORT, INT
			g.bytecodes.addIload(varIndex)
			break
		}
	}
	return nil
}

func (g *CCodeGenerator) AtFieldDecl(f *ast.FieldDecl) error {
	return f.GetInit().Accept(g)
}

func (g *CCodeGenerator) AtMethodDecl(m *ast.MethodDecl) error {
	mods := m.GetModifiers()
	g.bytecodes.MaxLocals = 1
	// 遍历 权限限定符
	for !isNil(mods) {
		k := mods.Head().(*ast.Keyword)
		mods = mods.Tail()
		if k.TokenID == ast.Static {
			g.bytecodes.MaxLocals = 0
			g.isStaticMethod = true
		}
	}
	params := m.GetParams()
	for !isNil(params) {
		err := g.AtDeclarator(params.Head().(*ast.Declarator))
		if err != nil {
			return err
		}
		params = params.Tail()
	}
	body := m.GetBody()
	return g.AtMethodBody(body, m.IsConstructor(), m.GetReturn().GetType() == ast.Void)
}

func (g *CCodeGenerator) doTypeCheck(tree ast.Node) error {
	//return tree.Accept(g.typeChecker)

	return tree.Accept(g.typeChecker)
}

func (g *CCodeGenerator) getLocalVar(d *ast.Declarator) int {
	v := d.GetLocalVar()
	if v < 0 {
		v = g.bytecodes.MaxLocals
		d.SetLocalVar(v)
		g.bytecodes.incMaxLocals(1)
	}
	return v
}

func (g *CCodeGenerator) AtDeclarator(d *ast.Declarator) error {
	d.SetLocalVar(g.bytecodes.MaxLocals)
	cname, err := g.Parent.resolveClassName(d.GetClassName())
	if err != nil {
		return err
	}
	d.SetClassName(cname)
	size := 1
	if is2word(d.GetType(), d.GetArrayDim()) {
		size = 2
	}
	g.bytecodes.incMaxLocals(size)
	/*  NOTE: Array initializers has not been supported.
	 */
	init := d.GetInitializer()
	if !isNil(init) {
		err := g.doTypeCheck(init)
		if err != nil {
			return err
		}
		return g.atVariableAssign(nil, ast.Assign, nil, d, init, false)
	}
	return nil
}

func (g *CCodeGenerator) AtMethodBody(body *ast.Statement, isCons, isVoid bool) error {
	if isNil(body) {
		return nil
	}
	if isCons && needsSuperCall(body) {
		g.Parent.insertDefaultSuperCall()
	}
	g.hasReturn = false
	err := body.Accept(g)
	if err != nil {
		return err
	}
	if !g.hasReturn {
		if isVoid {
			g.bytecodes.AddOpcode(classfile.OpReturn)
			g.hasReturn = true
		}
		return NewCompileError("no return statement")
	}
	return nil
}

func (g *CCodeGenerator) AtStatement(s *ast.Statement) error {
	if s == nil {
		return nil
	}
	op := s.GetOperator()
	switch op {
	case ast.Expr:
		expr := s.GetLeft()
		err := g.doTypeCheck(expr)
		if err != nil {
			return err
		}
		if assign, ok := expr.(*ast.AssignExpr); ok {
			return g.atAssignExpr0(assign, false)
		}
		if e, ok := expr.(*ast.Expression); ok {
			op := e.OperatorId
			if op == ast.PLUSPLUS || op == ast.MINUSMINUS {
				return g.atPlusPlus(e.OperatorId, e.Oprand1(), e, false)
			}
		}
		err = expr.Accept(g)
		if err != nil {
			return err
		}
		if is2word(g.exprType, g.arrayDim) {
			g.bytecodes.AddOpcode(classfile.OpPop2)
		} else if g.exprType != ast.Void {
			g.bytecodes.AddOpcode(classfile.OpPop)
		}
		return nil
	case ast.Block, ast.Decl:
		list := ast.Node(s.GetASTList())
		for !isNil(list) {
			h := list.GetLeft()
			list = list.GetRight()
			if !isNil(h) {
				if err := h.Accept(g); err != nil {
					return err
				}
			}
		}
		return nil
	case ast.IF:
		return g.atIfStmnt(s)
	case ast.While, ast.Do:
		return g.atWhileStmnt(s, op == ast.While)
	case ast.For:
		return g.atForStmnt(s)
	case ast.Break, ast.Continue:
		return g.atBreakStmnt(s, op == ast.Break)
	case ast.Return:
		return g.atReturnStmnt(s)
	case ast.Throw:
		return g.atThrowStmnt(s)
	case ast.Try:
		return g.atTryStmnt(s)
	case ast.Switch:
		return g.atSwitchStmnt(s)
	case ast.Synchronized:
		return g.atSyncStmnt(s)
	default:
		// LABEL, SWITCH label statement might be null?.
		g.hasReturn = false
		return NewCompileError(" not supported statement: " + op.GetName())
	}
}

func (g *CCodeGenerator) atPlusPlus(token ast.TokenID, oprand ast.Node, expr *ast.Expression, doDup bool) error {
	isPost := isNil(oprand) // ++i or i++?
	if isPost {
		oprand = expr.Oprand2()
	}
	if v, ok := oprand.(*ast.Variable); ok {
		d := v.Declarator
		t := d.GetType()
		g.exprType = t
		g.arrayDim = d.GetArrayDim()
		varIndex := g.getLocalVar(d)
		if g.arrayDim > 0 {
			return NewCompileError("invalid type for " + expr.GetName())
		}
		switch t {
		case ast.Double:
			g.bytecodes.addDload(varIndex)
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup2)
			}
			g.bytecodes.addDconst(1.0)
			if token == ast.PLUSPLUS {
				g.bytecodes.AddOpcode(classfile.OpDAdd)
			} else {
				g.bytecodes.AddOpcode(classfile.OpDSub)
			}
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup2)
			}
			g.bytecodes.addDstore(varIndex)
			return nil
		case ast.Float:
			g.bytecodes.addFload(varIndex)
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup)
			}
			g.bytecodes.addFconst(1.0)
			if token == ast.PLUSPLUS {
				g.bytecodes.AddOpcode(classfile.OpFAdd)
			} else {
				g.bytecodes.AddOpcode(classfile.OpFSub)
			}
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup)
			}
			g.bytecodes.addFstore(varIndex)
			return nil
		case ast.Long:
			g.bytecodes.AddLload(varIndex)
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup2)
			}
			g.bytecodes.addLconst(1)
			if token == ast.PLUSPLUS {
				g.bytecodes.AddOpcode(classfile.OpLAdd)
			} else {
				g.bytecodes.AddOpcode(classfile.OpLSub)
			}
			if doDup && isPost {
				g.bytecodes.AddOpcode(classfile.OpDup2)
			}
			g.bytecodes.addLstore(varIndex)
			return nil
		case ast.Byte:
		case ast.Char:
		case ast.Short:
		case ast.Int:
			if doDup && isPost {
				g.bytecodes.addIload(varIndex)
			}
			delta := -1
			if token == ast.PLUSPLUS {
				delta = 1
			}
			if varIndex > 0xff {
				g.bytecodes.AddOpcode(classfile.OpWide)
				g.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpIInc).AddIndex(varIndex).AddIndex(delta))
			} else {
				g.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpIInc).Add(varIndex).Add(delta))
			}
			if doDup && !isPost {
			}
			g.bytecodes.addIload(varIndex)
			return nil
		default:
			return NewCompileError("invalid type for " + expr.GetName())
		}
	} else {
		if e, ok := oprand.(*ast.Expression); ok {
			if e.GetOperator() == ast.Array {
				return g.AtArrayPlusPlus(token, isPost, e, doDup)
			}
		}
		// TODO
		// atFieldPlusPlus
	}
	return nil
}

func (g *CCodeGenerator) AtArrayPlusPlus(token ast.TokenID, post bool, e *ast.Expression, dup bool) error {
	return nil
}

func (g *CCodeGenerator) AtExpression(expr *ast.Expression) error {
	token := expr.GetOperator()
	oprand := expr.Oprand1()

	switch token {
	case ast.Dot:
		member := ""
		oprand2 := expr.Oprand2()
		if k, ok := oprand2.(*ast.Symbol); ok {
			member = k.Identifier
		}
		if ms, ok := oprand2.(*ast.MemberSymbol); ok {
			member = ms.Identifier
		}
		if member == "class" {
			return g.atClassObject(expr)
		}
		return g.atFieldRead(expr)

	case ast.Member:
		/* MEMBER ('#') is an extension by Javassist.
		 * The compiler internally uses # for compiling .class
		 * expressions such as "int.class".
		 */
		return g.atFieldRead(expr)

	case ast.Array:
		return g.atArrayRead(oprand, expr.Oprand2())

	case ast.PLUSPLUS, ast.MINUSMINUS:
		return g.atPlusPlus(token, oprand, expr, true)

	case ast.Not:
		booleanExpr, err := g.booleanExpr(false, expr)
		if err != nil {
			return err
		}
		if !booleanExpr {
			g.bytecodes.WriteIndex(7)
			g.bytecodes.addIconst(1)
			g.bytecodes.AddOpcode(classfile.OpGoto)
			g.bytecodes.WriteIndex(4)
		}
		g.bytecodes.addIconst(0)
		return nil

	case ast.Call:
		return NewCompileError("fatal")

	default:
		if err := oprand.Accept(g); err != nil {
			return err
		}
		typePrec := typePrecedence(g.exprType)
		if g.arrayDim > 0 {
			return NewCompileError("invalid types for " + expr.GetName())
		}

		switch token {
		case ast.Minus:
			switch typePrec {
			case P_DOUBLE:
				g.bytecodes.AddOpcode(classfile.OpDNeg)
			case P_FLOAT:
				g.bytecodes.AddOpcode(classfile.OpFNeg)
			case P_LONG:
				g.bytecodes.AddOpcode(classfile.OpLNeg)
			case P_INT:
				g.bytecodes.AddOpcode(classfile.OpINeg)
				g.exprType = ast.Int
			default:
				return NewCompileError("invalid types for " + expr.GetName())
			}

		case ast.BitNot:
			switch typePrec {
			case P_INT:
				g.bytecodes.addIconst(-1)
				g.bytecodes.AddOpcode(classfile.OpIXor)
				g.exprType = ast.Int
			case P_LONG:
				g.bytecodes.addLconst(-1)
				g.bytecodes.AddOpcode(classfile.OpLXor)
			default:
				return NewCompileError("invalid types for " + expr.GetName())
			}

		case ast.Plus:
			if typePrec == P_OTHER {
				return NewCompileError("invalid types for " + expr.GetName())
			}
			// Do nothing, ignore.

		default:
			return NewCompileError("fatal")
		}
	}
	return nil
}

func (g *CCodeGenerator) AtAssignExpr(expr *ast.AssignExpr) error {
	return g.atAssignExpr0(expr, true)
}

func (g *CCodeGenerator) AtBinExpr(expr *ast.BinExpr) error {
	token := expr.GetOperator()
	k := ast.LookupBinOp(token)
	// Arithmetic operators: +, -, *, /, %, |, ^, &, <<, >>, >>>
	if _, ok := binOp[token]; ok {
		if err := expr.Oprand1().Accept(g); err != nil {
			return err
		}
		right := expr.Oprand2()
		if right == nil {
			return nil // see TypeChecker.atBinExpr().
		}

		type1 := g.exprType
		dim1 := g.arrayDim
		cname1 := g.className
		if err := right.Accept(g); err != nil {
			return err
		}

		if dim1 != g.arrayDim {
			return NewCompileError("incompatible array types")
		}

		if token == ast.Plus && dim1 == 0 && (type1 == ast.Class || g.exprType == ast.Class) {
			return g.atStringConcatExpr(expr, type1, dim1, cname1)
		} else {
			return g.atArithBinExpr(expr.Expression, token, k, type1)
		}
	} else {
		// Logical and comparison operators: &&, ||, ==, !=, <=, >=, <, >
		booleanExpr, err := g.booleanExpr(true, expr)
		if err != nil {
			return err
		}
		if !booleanExpr {
			g.bytecodes.WriteIndex(7)
			g.bytecodes.addIconst(0) // false
			g.bytecodes.AddOpcode(classfile.OpGoto)
			g.bytecodes.WriteIndex(4)
		}

		g.bytecodes.addIconst(1) // true
	}

	return nil
}

func (g *CCodeGenerator) AtCastExpr(expr *ast.CastExpr) error {
	cname, err := g.resolveClassName0(expr.GetClassName())
	if err != nil {
		return err
	}
	toClass, err := g.checkCastExpr(expr, cname)
	if err != nil {
		return err
	}
	srcType := g.exprType
	g.exprType = expr.GetType()
	g.arrayDim = expr.GetArrayDim()
	g.className = cname
	if toClass == "" {
		return g.atNumCastExpr(srcType, g.exprType) // built-in type
	} else {
		g.bytecodes.addCheckCast(toClass)
	}
	return nil
}

func (g *CCodeGenerator) AtCondExpr(expr *ast.CondExpr) error {
	booleanExpr, err := g.booleanExpr(false, expr.CondExpr())
	if err != nil {
		return err
	}
	if booleanExpr {
		if err = expr.ElseExpr().Accept(g); err != nil {
			return err
		}
	} else {
		pc := g.bytecodes.currentPc()
		g.bytecodes.WriteIndex(0) // correct later
		if err = expr.ThenExpr().Accept(g); err != nil {
			return err
		}
		dim1 := g.arrayDim
		g.bytecodes.AddOpcode(classfile.OpGoto)
		pc2 := g.bytecodes.currentPc()
		g.bytecodes.WriteIndex(0)
		g.bytecodes.Write16bit(pc, g.bytecodes.currentPc()-pc+1)
		if err = expr.ElseExpr().Accept(g); err != nil {
			return err
		}
		if dim1 != g.arrayDim {
			return NewCompileError("type mismatch in ?:")
		}
		g.bytecodes.Write16bit(pc2, g.bytecodes.currentPc()-pc2+1)
	}
	return nil
}

func (g *CCodeGenerator) AtArrayInit(init *ast.ArrayInit) error {
	if g.Parent != nil {
		return g.Parent.AtArrayInit(init)
	}
	panic("unreachable")
}

func (g *CCodeGenerator) AtPair(p *ast.Pair) error {
	return NewCompileError("fatal at pair")
}

func (g *CCodeGenerator) AtSymbol(s *ast.Symbol) error {
	return NewCompileError("fatal at symbol")
}

func (g *CCodeGenerator) AtNewExpr(expr *ast.NewExpr) error {
	if g.Parent != nil {
		return g.Parent.AtNewExpr(expr)
	}
	panic("unreachable")
}

func (g *CCodeGenerator) AtASTList(l *ast.ASTList) error {
	return NewCompileError("fatal at list")
}

/* op is either =, %=, &=, *=, /=, +=, -=, ^=, |=, <<=, >>=, or >>>=.
 *
 * expr and var can be null.
 */
func (g *CCodeGenerator) atVariableAssign(expr *ast.Expression, op ast.TokenID, v *ast.Variable, d *ast.Declarator, right ast.Node, doDup bool) error {
	varType := d.GetType()
	varArray := d.GetArrayDim()
	varClassName := d.GetClassName()
	varNo := g.getLocalVar(d)
	g.exprType = varType
	g.arrayDim = varArray
	g.className = varClassName
	if op != ast.Assign {
		err := g.AtVariable(v)
		if err != nil {
			return err
		}
	}
	// expr is null if the caller is atDeclarator().
	array, isArrayInit := right.(*ast.ArrayInit)
	if expr == nil && isArrayInit {
		err := g.atArrayVariableAssign(array, varType, varArray, varClassName)
		if err != nil {
			return err
		}
	} else {
		err := g.atAssignCore(expr, op, right, varType, varArray, varClassName)
		if err != nil {
			return err
		}
	}
	if doDup {
		if is2word(varType, varArray) {
			g.bytecodes.AddOpcode(classfile.OpDup2)
		} else {
			g.bytecodes.AddOpcode(classfile.OpDup)
		}
	}
	if varArray > 0 {
		g.bytecodes.addAStore(varNo)
	} else if varType == ast.Double {
		g.bytecodes.addDstore(varNo)
	} else if varType == ast.Float {
		g.bytecodes.addFstore(varNo)
	} else if varType == ast.Long {
		g.bytecodes.addLstore(varNo)
	} else if isRefType(varType) {
		g.bytecodes.addAStore(varNo)
	} else {
		g.bytecodes.addIStore(varNo)
	}
	g.exprType = varType
	g.arrayDim = varArray
	g.className = varClassName
	return nil
}

/* overridden in MemberCodeGenerator
 */
func (g *CCodeGenerator) atTryStmnt(s *ast.Statement) error {
	if g.Parent == nil {
		g.hasReturn = false
		return nil
	}
	return g.Parent.atTryStmnt(s)
}

func (g *CCodeGenerator) atReturnStmnt(s *ast.Statement) error {
	return g.atReturnStmnt2(s.GetLeft())
}

func (g *CCodeGenerator) atReturnStmnt2(result ast.Node) error {
	var op classfile.OpCode
	if result == nil {
		op = classfile.OpReturn
	} else {
		err := g.compileExpr(result)
		if err != nil {
			return err
		}
		if g.arrayDim > 0 {
			op = classfile.OpAReturn
		} else {
			switch g.exprType {
			case ast.Double:
				op = classfile.OpDReturn
				break
			case ast.Float:
				op = classfile.OpFReturn
				break
			case ast.Long:
				op = classfile.OpLReturn
				break
			default:
				if isRefType(g.exprType) {
					op = classfile.OpAReturn
				} else {
					op = classfile.OpIReturn
				}
			}
		}
	}

	var har *ReturnHook
	for har = g.returnHooks; har != nil; har = har.next {
		if har.DoIt(g.Parent.Parent, op) {
			g.hasReturn = true
			return nil
		}
	}
	g.bytecodes.AddOpcode(op)
	g.hasReturn = true
	return nil
}

func (g *CCodeGenerator) compileExpr(expr ast.Node) error {
	err := g.doTypeCheck(expr)
	if err != nil {
		return err
	}
	return expr.Accept(g)
}

func (g *CCodeGenerator) atAssignExpr0(expr *ast.AssignExpr, doDup bool) error {
	op := expr.GetOperator()
	left := expr.Oprand1()
	right := expr.Oprand2()
	if v, ok := left.(*ast.Variable); ok {
		return g.atVariableAssign(expr.Expression, op, v, v.Declarator, right, doDup)
	}
	if e, ok := left.(*ast.Expression); ok {
		if e.GetOperator() == ast.Array {
			return g.atArrayAssign(expr.Expression, op, e, right, doDup)
		}
	}
	return g.atFieldAssign(expr, op, left, right, doDup)
}

func (g *CCodeGenerator) atArrayAssign(expr *ast.Expression, op ast.TokenID, array *ast.Expression, right ast.Node, doDup bool) error {
	err := g.arrayAccess(array.Oprand1(), array.Oprand2())
	if err != nil {
		return err
	}
	if op != ast.Assign {
		g.bytecodes.AddOpcode(classfile.OpDup2)
		g.bytecodes.AddOpcode(getArrayReadOp(g.exprType, g.arrayDim))
	}
	aType := g.exprType
	aDim := g.arrayDim
	cname := g.className
	err = g.atAssignCore(expr, op, right, aType, aDim, cname)
	if err != nil {
		return err
	}
	if doDup {
		if is2word(aType, aDim) {
			g.bytecodes.AddOpcode(classfile.OpDup2X2)
		} else {
			g.bytecodes.AddOpcode(classfile.OpDupX2)
		}
	}
	g.bytecodes.AddOpcode(getArrayWriteOp(aType, aDim))
	g.exprType = aType
	g.arrayDim = aDim
	g.className = cname
	return nil
}

func (g *CCodeGenerator) arrayAccess(array ast.Node, index ast.Node) error {
	err := array.Accept(g)
	if err != nil {
		return err
	}
	exprType := g.exprType
	dim := g.arrayDim
	if g.arrayDim == 0 {
		return NewCompileError("bad array access")
	}
	cname := g.className
	err = index.Accept(g)
	if err != nil {
		return err
	}
	if typePrecedence(g.exprType) != P_INT || g.arrayDim > 0 {
		return NewCompileError("bad array index")
	}
	g.exprType = exprType
	g.arrayDim = dim - 1
	g.className = cname
	return nil
}

func (g *CCodeGenerator) atAssignCore(expr *ast.Expression, op ast.TokenID, right ast.Node, aType ast.TokenID, dim int, cname string) error {
	if op == ast.PLUS_E && dim == 0 && aType == ast.Class {
		err := g.atStringPlusEq(expr, aType, dim, cname, right)
		if err != nil {
			return err
		}
	} else {
		err := right.Accept(g)
		if err != nil {
			return err
		}
		if invalidDim(g.exprType, aType, g.arrayDim, dim, g.className, cname, false) || (op != ast.Assign && dim > 0) {
			return NewCompileError("incompatible type for assignmen")
		}
		if op != ast.Assign {
			token := ast.AssignOps[op-ast.MOD_E]
			k := ast.LookupBinOp(token)
			if k < 0 {
				return NewCompileError("incompatible type for assignment")
			}
			if err = g.atArithBinExpr(expr, token, k, aType); err != nil {
				return err
			}
		}
	}
	if op != ast.Assign || (dim == 0 && !isRefType(aType)) {
		return g.atNumCastExpr(g.exprType, aType)
	}
	// type check should be done here.
	return nil
}

func (g *CCodeGenerator) atNumCastExpr(srcType, destType ast.TokenID) error {
	if srcType == destType {
		return nil
	}
	var op, op2 classfile.OpCode
	stype := typePrecedence(srcType)
	dtype := typePrecedence(destType)
	if 0 <= stype && stype < 3 {
		op = castOp[stype*4+dtype]
	} else {
		op = classfile.OpNop
	}
	switch destType {
	case ast.Double:
		op2 = classfile.OpI2D
		break
	case ast.Float:
		op2 = classfile.OpI2F
		break
	case ast.Long:
		op2 = classfile.OpI2L
		break
	case ast.Short:
		op2 = classfile.OpI2S
		break
	case ast.Char:
		op2 = classfile.OpI2C
		break
	case ast.Byte:
		op2 = classfile.OpI2B
		break
	default:
		op2 = classfile.OpNop
		break
	}
	if op != classfile.OpNop {
		g.bytecodes.AddOpcode(op)
	}
	if op == classfile.OpNop || op2 == classfile.OpL2I || op2 == classfile.OpF2I || op2 == classfile.OpD2I {
		if op2 != classfile.OpNop {
			g.bytecodes.AddOpcode(op2)
		}
	}
	return nil
}

func (g *CCodeGenerator) atStringPlusEq(expr *ast.Expression, aType ast.TokenID, dim int, cname string, right ast.Node) error {
	if jvmJavaLangString != cname {
		return NewCompileError("expected java.lang.string,but " + cname)
	}
	err := g.convToString(aType, dim)
	if err != nil {
		return err
	}
	err = right.Accept(g)
	if err != nil {
		return err
	}
	err = g.convToString(aType, dim)
	if err != nil {
		return err
	}
	g.bytecodes.addInvokeVirtual(javaLangString, "concat", "(Ljava/lang/String;)Ljava/lang/String;")
	g.exprType = ast.Class
	g.arrayDim = 0
	g.className = jvmJavaLangString
	return nil
}

func (g *CCodeGenerator) convToString(aType ast.TokenID, dim int) error {
	method := "valueOf"
	if isRefType(aType) || dim > 0 {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(Ljava/lang/Object;)Ljava/lang/String;")
	} else if aType == ast.Double {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(D)Ljava/lang/String;")
	} else if aType == ast.Float {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(F)Ljava/lang/String;")
	} else if aType == ast.Long {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(J)Ljava/lang/String;")
	} else if aType == ast.Boolean {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(Z)Ljava/lang/String;")
	} else if aType == ast.Char {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(C)Ljava/lang/String;")
	} else if aType == ast.Void {
		return NewCompileError("void type expression")
	} else {
		g.bytecodes.addInvokeStatic(javaLangString, method, "(I)Ljava/lang/String;")
	}
	return nil
}

func (g *CCodeGenerator) atFieldAssign(expr *ast.AssignExpr, op ast.TokenID, left ast.Node, right ast.Node, dup bool) error {
	if g.Parent != nil {
		return g.Parent.atFieldAssign(expr, op, left, right, dup)
	}
	panic("CCodeGenerator panic,parent is nil")
}

func (g *CCodeGenerator) atArrayVariableAssign(array *ast.ArrayInit, varType ast.TokenID, array2 int, name string) error {
	if g.Parent != nil {
		return g.Parent.atArrayVariableAssign(array, varType, array2, name)
	}
	panic("CCodeGenerator panic,parent is nil")
}

func (g *CCodeGenerator) AtMember(m *ast.MemberSymbol) error {
	if g.Parent != nil {
		return g.Parent.AtMember(m)
	}
	panic("CCodeGenerator panic,parent is nil")
}

func (g *CCodeGenerator) AtCallExpr(c *ast.CallExpr) error {
	if g.Parent != nil {
		return g.Parent.AtCallExpr(c)
	}
	panic("CCodeGenerator panic,parent is nil")
}

func (c *CCodeGenerator) atIfStmnt(st *ast.Statement) error {
	expr := st.GetASTList().Head()
	thenp, _ := st.GetASTList().Tail().Head().(*ast.Statement)
	elsep, _ := st.GetASTList().Tail().Tail().Head().(*ast.Statement)
	booleanExpr, err := c.compileBooleanExpr(false, expr)
	if err != nil {
		return err
	}
	if booleanExpr {
		c.hasReturn = false
		if elsep != nil {
			if err := elsep.Accept(c); err != nil {
				return err
			}
		}
		return nil
	}

	pc := c.bytecodes.currentPc()
	var pc2 int
	c.bytecodes.WriteIndex(0) // 占位，稍后修正

	c.hasReturn = false
	if thenp != nil {
		if err := thenp.Accept(c); err != nil {
			return err
		}
	}

	thenHasReturned := c.hasReturn
	c.hasReturn = false

	if elsep != nil && !thenHasReturned {
		c.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpGoto)) // 后续再修正
		pc2 = c.bytecodes.currentPc()
		c.bytecodes.WriteIndex(0)
	}

	// 修改 TODO
	c.bytecodes.Write16bit(pc, c.bytecodes.currentPc()-pc+1)
	if elsep != nil {
		if err := elsep.Accept(c); err != nil {
			return err
		}
		if !thenHasReturned {
			c.bytecodes.Write16bit(pc2, c.bytecodes.currentPc()-pc2+1)
		}
		c.hasReturn = thenHasReturned && c.hasReturn
	}

	return nil
}

func (g *CCodeGenerator) compileBooleanExpr(branchIf bool, expr ast.Node) (bool, error) {
	if err := g.doTypeCheck(expr); err != nil {
		return false, err
	}
	return g.booleanExpr(branchIf, expr)
}

func (c *CCodeGenerator) atWhileStmnt(st *ast.Statement, notDo bool) error {
	// 保存上层循环的 break 和 continue 列表
	prevBreakList := c.breakList
	prevContList := c.continueList
	// 新建当前循环的 break 和 continue 列表
	c.breakList = []int{}
	c.continueList = []int{}

	expr := st.GetASTList().Head()
	var body ast.Node
	if !isNil(st.GetRight()) {
		body = st.GetASTList().GetRight()
	}

	var pc int
	if notDo {
		c.bytecodes.AddOpcode(classfile.OpGoto)
		pc = c.bytecodes.currentPc()
		c.bytecodes.WriteIndex(0) // 占位，待修正跳转偏移
	}

	pc2 := c.bytecodes.currentPc()
	if body != nil {
		if err := body.Accept(c); err != nil {
			return err
		}
	}

	pc3 := c.bytecodes.currentPc()
	if notDo {
		c.bytecodes.Write16bit(pc, c.bytecodes.currentPc()-pc+1)
	}

	alwaysBranch, err := c.compileBooleanExpr(true, expr)
	if err != nil {
		return err
	}
	if alwaysBranch {
		c.bytecodes.AddOpcode(classfile.OpGoto)
		// 如果当前循环中没有 break，则一直分支
		alwaysBranch = len(c.breakList) == 0
	}

	c.bytecodes.WriteIndex(pc2 - c.bytecodes.currentPc() + 1)
	c.patchGoto(c.breakList, c.bytecodes.currentPc())
	c.patchGoto(c.continueList, pc3)

	// 恢复上层循环的 break 和 continue 列表
	c.continueList = prevContList
	c.breakList = prevBreakList
	c.hasReturn = alwaysBranch

	return nil
}

func (g *CCodeGenerator) patchGoto(list []int, pc3 int) {

}

func (g *CCodeGenerator) booleanExpr(branchIf bool, expr ast.Node) (bool, error) {
	var isAndAnd bool
	op := getCompOperator(expr)

	if op == ast.EQ { // ==, !=, ...
		bexpr := expr.(*ast.BinExpr)
		type1, err := g.compileOprands(bexpr)
		if err != nil {
			return false, err
		}
		// here, arrayDim might represent the array dim. of the left operand
		// if the right operand is NULL.
		g.compareExpr(branchIf, bexpr.GetOperator(), type1, bexpr)
	} else if op == ast.Not {
		oprand1 := expr.(*ast.Expression).Oprand1()
		return g.booleanExpr(!branchIf, oprand1)
	} else if (op == ast.ANDAND) || op == ast.OROR {
		isAndAnd = op == ast.ANDAND
		bexpr := expr.(*ast.BinExpr)
		result, err := g.booleanExpr(!isAndAnd, bexpr.Oprand1())
		if err != nil {
			return false, err
		}
		if result {
			g.exprType = ast.Boolean
			g.arrayDim = 0
			return true, nil
		}

		pc := g.bytecodes.currentPc()
		g.bytecodes.WriteIndex(0) // correct later
		result, err = g.booleanExpr(isAndAnd, bexpr.Oprand2())
		if err != nil {
			return false, err
		}
		if result {
			g.bytecodes.AddOpcode(classfile.OpGoto)
		}

		g.bytecodes.Write16bit(pc, g.bytecodes.currentPc()-pc+3)
		if branchIf != isAndAnd {
			g.bytecodes.WriteIndex(6) // skip GOTO instruction
			g.bytecodes.AddOpcode(classfile.OpGoto)
		}
	} else if isAlwaysBranch(expr, branchIf) {
		// Opcode.GOTO is not added here. The caller must add it.
		g.exprType = ast.Boolean
		g.arrayDim = 0
		return true, nil // always branch
	} else { // others
		err := expr.Accept(g)
		if err != nil {
			return false, err
		}
		if g.exprType != ast.Boolean || g.arrayDim != 0 {
			return false, fmt.Errorf("boolean expr is required")
		}

		g.bytecodes.AddOpcode(classfile.OpIfNE)
		if !branchIf {
			g.bytecodes.AddOpcode(classfile.OpIfEQ)
		}
	}

	g.exprType = ast.Boolean
	g.arrayDim = 0
	return false, nil
}

func (c *CCodeGenerator) atForStmnt(st *ast.Statement) error {
	// 保存上层循环的 break 和 continue 列表
	prevBreakList := c.breakList
	prevContList := c.continueList
	// 新建当前循环的 break 和 continue 列表
	c.breakList = []int{}
	c.continueList = []int{}

	// 解析 for 语句的各个部分
	init := st.GetASTList().Head()
	p := st.GetASTList().Tail()
	expr := p.Head()
	p = p.Tail()
	update := p.Head()
	body := p.Tail()

	// 处理初始化语句
	if init != nil {
		if err := init.Accept(c); err != nil {
			return err
		}
	}

	pc := c.bytecodes.currentPc()
	var pc2 int

	// 处理循环条件
	if expr != nil {
		alwaysFalse, err := c.compileBooleanExpr(false, expr)
		if err != nil {
			return err
		}
		if alwaysFalse {
			// for (...; false; ...) 直接退出
			c.continueList = prevContList
			c.breakList = prevBreakList
			c.hasReturn = false
			return nil
		}

		pc2 = c.bytecodes.currentPc()
		c.bytecodes.WriteIndex(0) // 占位，待修正跳转偏移
	}

	// 处理循环体
	if body != nil {
		if err := body.Accept(c); err != nil {
			return err
		}
	}

	pc3 := c.bytecodes.currentPc()
	if update != nil {
		if err := update.Accept(c); err != nil {
			return err
		}
	}

	// 跳回循环开始
	c.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpGoto).AddIndex(pc - c.bytecodes.currentPc() + 1))

	pc4 := c.bytecodes.currentPc()
	if expr != nil {
		c.bytecodes.Write16bit(pc2, pc4-pc2+1)
	}

	// 修正 break 和 continue 语句的跳转目标
	c.patchGoto(c.breakList, pc4)
	c.patchGoto(c.continueList, pc3)

	// 恢复上层循环的 break 和 continue 列表
	c.continueList = prevContList
	c.breakList = prevBreakList
	c.hasReturn = false

	return nil
}

func (g *CCodeGenerator) atBreakStmnt(st *ast.Statement, notCont bool) error {
	if !isNil(st.GetASTList().Head()) {
		// TO DO
		return NewCompileError("sorry, not support labeled break or continue")
	}
	g.bytecodes.AddInstruction(classfile.NewInstruction(classfile.OpGoto).AddIndex(0))
	pc := g.bytecodes.currentPc()

	if notCont {
		g.breakList = append(g.breakList, pc)
	} else {
		g.continueList = append(g.continueList, pc)
	}

	return nil
}

func (g *CCodeGenerator) atThrowStmnt(st *ast.Statement) error {
	e := st.GetLeft()
	if err := g.compileExpr(e); err != nil {
		return err
	}
	if g.exprType != ast.Class || g.arrayDim > 0 {
		return NewCompileError("bad throw statement")
	}
	g.bytecodes.AddOpcode(classfile.OpAThrow)
	g.hasReturn = true
	return nil
}

func (c *CCodeGenerator) atSwitchStmnt(st *ast.Statement) error {
	isString := false
	if c.typeChecker != nil {
		if err := c.doTypeCheck(st.GetASTList().Head()); err != nil {
			return err
		}
		isString = c.typeChecker.exprType == ast.Class &&
			c.typeChecker.arrayDim == 0 &&
			c.typeChecker.className == jvmJavaLangString
	}

	if err := c.compileExpr(st.GetASTList().Head()); err != nil {
		return err
	}
	tmpVar := -1
	if isString {
		tmpVar = c.bytecodes.MaxLocals
		c.bytecodes.incMaxLocals(1)
		c.bytecodes.addAStore(tmpVar)
		c.bytecodes.addAload(tmpVar)
		c.bytecodes.addInvokeVirtual(jvmJavaLangString, "hashCode", "()I")
	}

	prevBreakList := c.breakList
	c.breakList = []int{}

	opcodePc := c.bytecodes.currentPc()
	c.bytecodes.AddOpcode(classfile.OpLookupSwitch)
	npads := 3 - (opcodePc & 3)
	for npads > 0 {
		c.bytecodes.AddByte(0)
		npads--
	}

	body := st.GetASTList().Tail()
	npairs := 0
	for list := body; list != nil; list = list.Tail() {
		if list.Head().(*ast.Statement).GetOperator() == ast.Case {
			npairs++
		}
	}

	opcodePc2 := c.bytecodes.currentPc()
	c.bytecodes.AddGap(4)
	c.bytecodes.add32bit(npairs)
	c.bytecodes.AddGap(npairs * 8)

	pairs := make([]int64, npairs)
	gotoDefaults := []int{}
	iPairs := 0
	defaultPc := -1

	for list := body; list != nil; list = list.Tail() {
		label := list.Head().(*ast.Statement)
		op := label.GetOperator()
		if op == ast.Default {
			defaultPc = c.bytecodes.currentPc()
		} else if op != ast.Case {
			return fmt.Errorf("invalid switch case")
		} else {
			curPos := c.bytecodes.currentPc()
			var caseLabel int64
			var err error
			if isString {
				caseLabel, err = c.computeStringLabel(label.GetASTList().Head(), tmpVar, &gotoDefaults)
				if err != nil {
					return err
				}
			} else {
				caseLabel, err = c.computeLabel(label.GetASTList().Head())
				if err != nil {
					return err
				}
			}

			pairs[iPairs] = (caseLabel << 32) + int64(curPos-opcodePc)
			iPairs++
		}

		c.hasReturn = false
		if err := label.GetASTList().Tail().Accept(c); err != nil {
			return err
		}
	}

	slices.Sort(pairs)
	pc := opcodePc2 + 8
	for i := 0; i < npairs; i++ {
		c.bytecodes.Write32bit(pc, int(pairs[i]>>32))
		c.bytecodes.Write32bit(pc+4, int(pairs[i]))
		pc += 8
	}

	if defaultPc < 0 || len(c.breakList) > 0 {
		c.hasReturn = false
	}

	endPc := c.bytecodes.currentPc()
	if defaultPc < 0 {
		defaultPc = endPc
	}

	c.bytecodes.Write32bit(opcodePc2, defaultPc-opcodePc)
	for _, addr := range gotoDefaults {
		c.bytecodes.Write16bit(addr, defaultPc-addr+1)
	}

	c.patchGoto(c.breakList, endPc)
	c.breakList = prevBreakList
	return nil
}

func (g *CCodeGenerator) atSyncStmnt(st *ast.Statement) error {
	nbreaks := len(g.breakList)
	ncontinues := len(g.continueList)

	if err := g.compileExpr(st.GetASTList().Head()); err != nil {
		return err
	}

	if g.exprType != ast.Class && g.arrayDim == 0 {
		return NewCompileError("bad type expr for synchronized block")
	}

	bc := g.bytecodes
	varIndex := bc.MaxLocals
	bc.incMaxLocals(1)
	bc.AddOpcode(classfile.OpDup)
	bc.addAStore(varIndex)
	bc.AddOpcode(classfile.OpMonitorEnter)

	//rh := &ReturnHook{g: g, DoitFunc: func(b *Bytecode, opcode int) bool {
	//	b.AddAload(var)
	//	b.AddOpcode(MONITOREXIT)
	//	return false
	//}}

	pc := bc.currentPc()
	body := st.GetRight().(*ast.Statement) // Assuming Tail() returns an appropriate type
	if body != nil {
		if err := body.Accept(g); err != nil {
			return err
		}
	}

	pc2 := bc.currentPc()
	pc3 := 0
	if !g.hasReturn {
		//rh.Doit(bc, 0) // the 2nd arg is ignored
		bc.AddOpcode(classfile.OpGoto)
		pc3 = bc.currentPc()
		bc.WriteIndex(0)
	}

	if pc < pc2 { // if the body is not empty
		pc4 := bc.currentPc()
		//rh.Doit(bc, 0) // the 2nd arg is ignored
		bc.AddOpcode(classfile.OpAThrow)
		bc.addExceptionHandler(pc, pc2, pc4, "")
	}

	if !g.hasReturn {
		bc.Write16bit(pc3, bc.currentPc()-pc3+1)
	}

	//rh.Remove(g)

	if len(g.breakList) != nbreaks || len(g.continueList) != ncontinues {
		return fmt.Errorf("sorry, cannot break/continue in synchronized block")
	}

	return nil
}

func (g *CCodeGenerator) compileOprands(expr *ast.BinExpr) (ast.TokenID, error) {
	if err := expr.Oprand1().Accept(g); err != nil {
		return ast.Empty, err
	}
	type1 := g.exprType
	dim1 := g.arrayDim
	if err := expr.Oprand2().Accept(g); err != nil {
		return ast.Empty, err
	}
	if dim1 != g.arrayDim {
		if type1 != ast.Null && g.exprType != ast.Null {
			return ast.Empty, NewCompileError("incompatible array types")
		} else if g.exprType == ast.Null {
			g.arrayDim = dim1
		}

	}
	if type1 == ast.Null {
		return g.exprType, nil
	}
	return type1, nil
}

/* Produces the opcode to branch if the condition is true.
 * The oprands are not produced.
 *
 * Parameter expr - compare expression ==, !=, <=, >=, <, >
 */
func (g *CCodeGenerator) compareExpr(branchIf bool, token, type1 ast.TokenID, expr *ast.BinExpr) error {
	if g.arrayDim == 0 {
		if err := g.convertOprandTypes(type1, g.exprType, expr.Expression); err != nil {
			return err
		}
	}
	p := typePrecedence(g.exprType)
	if p == P_OTHER || g.arrayDim > 0 {
		if token == ast.EQ {
			op := classfile.OpIfACmpNE
			if branchIf {
				op = classfile.OpIfACmpEQ
			}
			g.bytecodes.AddOpcode(op)
		} else if token == ast.NEQ {
			op := classfile.OpIfACmpEQ
			if branchIf {
				op = classfile.OpIfACmpNE
			}
			g.bytecodes.AddOpcode(op)
		} else {
			return NewCompileError("invalid types for " + expr.GetName())
		}
	} else {
		if p == P_INT {
			index := 1
			if branchIf {
				index = 0
			}
			for k, v := range ifOp {
				if token == k {
					g.bytecodes.AddOpcode(v[index])
					return nil
				}
			}
			return NewCompileError("invalid types for " + expr.GetName())
		} else {
			if p == P_DOUBLE {
				if token == '<' || token == ast.LE {
					g.bytecodes.AddOpcode(classfile.OpDCmpG)
				} else {
					g.bytecodes.AddOpcode(classfile.OpDCmpL)
				}
			} else if p == P_FLOAT {
				if token == '<' || token == ast.LE {
					g.bytecodes.AddOpcode(classfile.OpFCmpG)
				} else {
					g.bytecodes.AddOpcode(classfile.OpFCmpL)
				}
			} else if p == P_LONG {
				g.bytecodes.AddOpcode(classfile.OpLCmp) // 1: >, 0: =, -1: <
			} else {
				return NewCompileError("fatal type")
			}
			index := 1
			if branchIf {
				index = 0
			}
			for k, v := range ifOp2 {
				if token == k {
					g.bytecodes.AddOpcode(v[index])
					return nil
				}
			}

			return NewCompileError("invalid types for " + expr.GetName())
		}
	}
	return nil
}

func (g *CCodeGenerator) computeStringLabel(expr ast.Node, tmpVar int, gotoDefaults *[]int) (int64, error) {
	if err := g.doTypeCheck(expr); err != nil {
		return 0, err
	}
	expr = StripPlusExpr(expr)

	switch e := expr.(type) {
	case *ast.StringLiteral:
		label := e.Text
		g.bytecodes.addAload(tmpVar)
		g.bytecodes.addLdc(g.bytecodes.AddStringRef(label))
		g.bytecodes.addInvokeVirtual(jvmJavaLangString, "equals", "(Ljava/lang/Object;)Z")
		g.bytecodes.AddOpcode(classfile.OpIfEQ)
		pc := g.bytecodes.currentPc()
		g.bytecodes.WriteIndex(0)
		*gotoDefaults = append(*gotoDefaults, pc)
		return Hash(label), nil
	default:
		return 0, fmt.Errorf("bad case label")
	}
}

func (g *CCodeGenerator) computeLabel(expr ast.Node) (int64, error) {
	if err := g.doTypeCheck(expr); err != nil {
		return 0, err
	}
	e := StripPlusExpr(expr)
	if intVal, ok := e.(*ast.IntConst); ok {
		return intVal.Value, nil
	}
	return 0, NewCompileError("bad case label")
}

/* do implicit type conversion.
 * arrayDim values of the two oprands must be zero.
 */
func (g *CCodeGenerator) convertOprandTypes(type1, type2 ast.TokenID, expr *ast.Expression) error {
	var rightStrong bool
	type1P := typePrecedence(type1)
	type2P := typePrecedence(type2)

	if type2P < 0 && type1P < 0 { // not primitive types
		return nil
	}

	if type2P < 0 || type1P < 0 { // either is not a primitive type
		return NewCompileError("bad type for " + expr.GetName())
	}

	var resultType int
	var op classfile.OpCode
	if type1P <= type2P {
		rightStrong = false
		g.exprType = type1
		op = castOp[type2P*4+type1P]
		resultType = type1P
	} else {
		rightStrong = true
		op = castOp[type1P*4+type2P]
		resultType = type2P
	}

	if rightStrong {
		if resultType == P_DOUBLE || resultType == P_LONG {
			if type1P == P_DOUBLE || type1P == P_LONG {
				g.bytecodes.AddOpcode(classfile.OpDup2X2)
			} else {
				g.bytecodes.AddOpcode(classfile.OpDup2X1)
			}

			g.bytecodes.AddOpcode(classfile.OpPop2)
			g.bytecodes.AddOpcode(op)
			g.bytecodes.AddOpcode(classfile.OpDup2X2)
			g.bytecodes.AddOpcode(classfile.OpPop2)
		} else if resultType == P_FLOAT {
			if type1P == P_LONG {
				g.bytecodes.AddOpcode(classfile.OpDupX2)
				g.bytecodes.AddOpcode(classfile.OpPop)
			} else {
				g.bytecodes.AddOpcode(classfile.OpSwap)
			}

			g.bytecodes.AddOpcode(op)
			g.bytecodes.AddOpcode(classfile.OpSwap)
		} else {
			return NewCompileError("fatal type on convert")
		}
	} else if op != classfile.OpNop {
		g.bytecodes.AddOpcode(op)
	}

	return nil
}

func (g *CCodeGenerator) atClassObject(expr *ast.Expression) error {

	op1 := expr.Oprand1()
	sym, ok := op1.(*ast.Symbol)
	if !ok {
		return NewCompileError("fatal error: badly parsed .class expr")
	}

	cname := sym.Identifier
	if strings.HasPrefix(cname, "[") {
		i := strings.Index(cname, "[L")
		if i >= 0 {
			name := cname[i+2 : len(cname)-1]
			name2, err := g.resolveClassName(name)
			if err != nil {
				return err
			}
			if name != name2 {
				name2 = JvmToJavaName(name2)
				var sbuf strings.Builder
				for i >= 0 {
					sbuf.WriteByte('[')
					i--
				}
				sbuf.WriteString("L" + name2 + ";")
				cname = sbuf.String()
			}
		}
	} else {
		var err error
		cname, err = g.resolveClassName(JavaToJvmName(cname))
		if err != nil {
			return err
		}
		cname = JvmToJavaName(cname)
	}

	g.atClassObject2(cname)
	g.exprType = ast.Class
	g.arrayDim = 0
	g.className = "java/lang/Class"

	return nil
}

func (g *CCodeGenerator) atClassObject2(cname string) {
	if reflect.GetVersion() < 48 {
		start := g.bytecodes.currentPc()
		g.bytecodes.addLdc(g.bytecodes.AddStringRef(cname))
		g.bytecodes.addInvokeStatic("java/lang/Class", "forName", "(Ljava/lang/String;)Ljava/lang/Class;")
		end := g.bytecodes.currentPc()
		g.bytecodes.AddOpcode(classfile.OpGoto)
		pc := g.bytecodes.currentPc()
		g.bytecodes.WriteIndex(0) // correct later

		g.bytecodes.addExceptionHandler(start, end, g.bytecodes.currentPc(), "java/lang/ClassNotFoundException")

		// Handle exception with DotClass.fail()
		/* -- the following code is for inlining a call to DotClass.fail().

		   int var = getMaxLocals();
		   incMaxLocals(1);
		   bytecode.growStack(1);
		   bytecode.addAstore(var);

		   bytecode.addNew("java.lang.NoClassDefFoundError");
		   bytecode.addOpcode(DUP);
		   bytecode.addAload(var);
		   bytecode.addInvokevirtual("java.lang.ClassNotFoundException",
		                             "getMessage", "()Ljava/lang/String;");
		   bytecode.addInvokespecial("java.lang.NoClassDefFoundError", "<init>",
		                             "(Ljava/lang/String;)V");
		*/

		g.bytecodes.growStack(1)
		g.bytecodes.addInvokeStatic("javassist/runtime/DotClass", "fail", "(Ljava/lang/ClassNotFoundException;)Ljava/lang/NoClassDefFoundError;")
		g.bytecodes.AddOpcode(classfile.OpAThrow)
		g.bytecodes.Write16bit(pc, g.bytecodes.currentPc()-pc+1)
	} else {
		g.bytecodes.addLdc(g.bytecodes.pool.AddClassInfo(cname))
	}
}

func (g *CCodeGenerator) resolveClassName(jvmClassName string) (string, error) {
	if g.Parent != nil {
		return g.Parent.resolveClassName(jvmClassName)
	}
	panic("unreachable")
}

func (g *CCodeGenerator) atFieldRead(expr *ast.Expression) error {
	if g.Parent != nil {
		return g.Parent.atFieldRead(expr)
	}
	panic("unreachable")
}

func (g *CCodeGenerator) atArrayRead(array ast.Node, index ast.Node) error {
	if err := g.arrayAccess(array, index); err != nil {
		return err
	}
	g.bytecodes.AddOpcode(getArrayReadOp(g.exprType, g.arrayDim))
	return nil
}

func (g *CCodeGenerator) atStringConcatExpr(expr *ast.BinExpr, type1 ast.TokenID, dim1 int, cname1 string) error {
	type2 := g.exprType
	dim2 := g.arrayDim
	type2Is2 := is2word(type2, dim2)
	type2IsString := (type2 == ast.Class && g.className == jvmJavaLangString)
	if type2Is2 {
		if err := g.convToString(type2, dim2); err != nil {
			return err
		}
	}

	if is2word(type1, dim1) {
		g.bytecodes.AddOpcode(classfile.OpDupX2)
		g.bytecodes.AddOpcode(classfile.OpPop)
	} else {
		g.bytecodes.AddOpcode(classfile.OpSwap)
	}

	// Even if type1 is String, the left operand might be null.
	if err := g.convToString(type1, dim1); err != nil {
		return err
	}
	g.bytecodes.AddOpcode(classfile.OpSwap)

	if !type2Is2 && !type2IsString {
		if err := g.convToString(type2, dim2); err != nil {
			return err
		}
	}

	g.bytecodes.addInvokeVirtual("java/lang/String", "concat", "(Ljava/lang/String;)Ljava/lang/String;")
	g.exprType = ast.Class
	g.arrayDim = 0
	g.className = jvmJavaLangString

	return nil
}

func (g *CCodeGenerator) atArithBinExpr(expr *ast.Expression, op ast.TokenID, index int, type1 ast.TokenID) error {
	if g.arrayDim != 0 {
		return NewCompileError("bad type for " + expr.GetName())
	}

	type2 := g.exprType
	if op == ast.LSHIFT || op == ast.RSHIFT || op == ast.ARSHIFT {
		if type2 == ast.Int || type2 == ast.Short || type2 == ast.Char || type2 == ast.Byte {
			g.exprType = type1
		} else {
			return NewCompileError("bad type for " + expr.GetName())
		}
	} else {
		if err := g.convertOprandTypes(type1, type2, expr); err != nil {
			return err
		}
	}
	p := typePrecedence(g.exprType)
	if p >= 0 {
		opcode := binOp[op][p]
		if opcode != classfile.OpNop {
			if p == P_INT && g.exprType != ast.Boolean {
				g.exprType = ast.Int // type1 may be BYTE, ...
			}
			g.bytecodes.AddOpcode(opcode)
			return nil
		}
	}
	return NewCompileError("bad type for " + expr.GetName())
}

func (g *CCodeGenerator) resolveClassName0(name *ast.ASTList) (string, error) {
	if g.Parent != nil {
		return g.Parent.resolveClassName0(name)
	}
	panic("unreachable")

}

func (g *CCodeGenerator) checkCastExpr(expr *ast.CastExpr, name string) (string, error) {
	const msg = "invalid cast"
	oprand := expr.GetOprand()
	dim := expr.GetArrayDim()
	typeID := expr.GetType()
	if err := oprand.Accept(g); err != nil {
		return "", err
	}
	srcType := g.exprType
	srcDim := g.arrayDim
	// TODO
	//if invalidDim(srcType, typeID, dim, g.arrayDim, g.className, name, true) || srcType == ast.Void || typeID == ast.Void {
	//	return "", NewCompileError(msg)
	//}

	if typeID == ast.Class {
		if !isRefType(srcType) && srcDim == 0 {
			return "", NewCompileError(msg)
		}
		return toJvmArrayName(name, dim), nil
	} else {
		if dim > 0 {
			return toJvmTypeName(typeID, dim), nil
		} else {
			return "", nil // built-in type
		}
	}
}

func (j *CCodeGenerator) atMethodBody(s *ast.Statement, isCons bool, isVoid bool) error {
	if s == nil {
		return nil
	}
	if isCons && j.needsSuperCall(s) {
		j.Parent.insertDefaultSuperCall()
	}
	j.hasReturn = false
	if err := s.Accept(j); err != nil {
		return err
	}
	if !j.hasReturn {
		if !isVoid {
			return NewCompileError("no return statement")
		}

		j.bytecodes.AddOpcode(classfile.OpReturn)
		j.hasReturn = true
	}
	return nil
}

func (g *CCodeGenerator) needsSuperCall(body *ast.Statement) bool {
	if body.GetOperator() == ast.Block {
		body = body.GetLeft().(*ast.Statement)
	}
	if body != nil && body.GetOperator() == ast.Expr {
		expr := body.GetLeft()
		if !isNil(expr) {
			if e, ok := expr.(*ast.Expression); ok && e.GetOperator() == ast.Call {
				target := e.Head()
				if k, ok := target.(*ast.Keyword); ok {
					token := k.TokenID
					return token != ast.Super && token != ast.This
				}
			}
		}
	}
	return true
}
