package compiler

import (
	"errors"
	"fmt"
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"strconv"
	"strings"
)

type Parser struct {
	lex       *Lexer
	tokenList []Token
	pos       int
}

func NewParser(input string) *Parser {
	return &Parser{
		lex:       NewLexer(input),
		tokenList: []Token{},
		pos:       -1,
	}
}

func (p *Parser) parseMember1(table *SymbolTable) (*ast.ASTList, bool, error) {
	var d *ast.Declarator
	var err error
	mods := p.parseMemberMods(table)
	var isConstructor = false
	// 构造方法的情况: 修饰符后 是 标识符 和 (
	if p.lookAHeadToken().ID == ast.Identifier && p.lookHeadNToken(2).Value == "(" {
		isConstructor = true
		// 无参 返回参数类型为void
		d = ast.NewDeclaratorTypeDim(ast.Void, 0)
	} else {
		//普通的方法
		d, err = p.parseFormalType(table)
		if err != nil {
			return nil, false, err
		}
	}
	if p.getToken().ID != ast.Identifier {
		return nil, false, fmt.Errorf(p.syanxError())
	} else {
		//获取声明的名称
		name := ""
		if isConstructor {
			name = "<init>"
		} else {
			name = p.getString()
		}
		d.SetVariable(ast.NewSymbol(name))
	}
	if !isConstructor && p.lookAHeadToken().Value != "(" {
		//后面没有参数列表，为属性声明
		field, err := p.parseField(table, mods, d)
		if err != nil {
			return nil, false, err
		}
		return field.ASTList, false, err
	} else {
		//方法声明
		method, err := p.parseMethod1(table, isConstructor, mods, d)
		if err != nil {
			return nil, true, err
		}
		return method.ASTList, true, err
	}
}

// parseMemberMods parse Modifiers
func (p *Parser) parseMemberMods(table *SymbolTable) *ast.ASTList {
	var list *ast.ASTList = nil
	for {
		lookToken := p.lookAHeadToken()
		// 如果不是修饰符，返回
		if !lookToken.ID.IsModifier() {
			return list
		}
		token := p.getToken()
		list = ast.NewASTList(ast.NewKeyword(token.Value, token.ID), list)
	}
}

// parse type
func (p *Parser) parseFormalType(table *SymbolTable) (*ast.Declarator, error) {
	t := p.lookAHeadToken().ID

	if !t.IsBasicType() && t != ast.Void {
		name, err := p.parseClassType(table)
		if err != nil {
			return nil, err
		}
		dim, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		return ast.NewDeclaratorWithClassName(name, dim), nil
	} else {
		//基本数据类型 or void
		p.getToken()
		dim, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		return ast.NewDeclaratorTypeDim(t, dim), nil
	}
}

// 解析java 类型
func (p *Parser) parseClassType(table *SymbolTable) (*ast.ASTList, error) {
	var list *ast.ASTList = nil
	for {
		if p.getToken().ID != ast.Identifier {
			return nil, fmt.Errorf(p.syanxError())
		}
		list = ast.AppendASTList(list, ast.NewSymbol(p.getString())).(*ast.ASTList)
		if p.lookAHeadToken().Value == "." {
			p.getToken()
		} else {
			break
		}
	}
	return list, nil
}

func (p *Parser) parseArrayDimension() (int, error) {
	arrayDim := 0
	for {
		if p.lookAHeadToken().Value != "[" {
			return arrayDim, nil
		}
		arrayDim++
		//消耗 [
		p.getToken()
		if p.getToken().Value != "]" {
			// 处理错误情况，例如抛出异常或返回错误
			// 这里假设存在一个 handleError 函数来处理错误
			return 0, fmt.Errorf(p.syanxError() + "; Expected ']' after '['")

		}
	}
}

func (p *Parser) syanxError() string {
	//	panic("")
	return fmt.Sprintf("syntax error near  \"%s\"  in line:%d and column: %d. ", p.lex.getTextAround(), p.lex.line, p.lex.column)
}

// get current token string value
func (p *Parser) getString() string {
	return p.tokenList[p.pos].Value
}

func (p *Parser) getToken() Token {
	if p.pos == len(p.tokenList)-1 {
		if !p.lex.HasNextToken() {
			return EOFToken
		}
		token := p.lex.NextToken()
		p.tokenList = append(p.tokenList, token)
		p.pos++
		return token
	}
	p.pos++
	return p.tokenList[p.pos]
}

// only look the current coast token of next token
func (p *Parser) lookAHeadToken() Token {
	if !p.lex.HasNextToken() {
		//if read all tokens
		return EOFToken
	}
	if p.pos == len(p.tokenList)-1 {
		token := p.lex.NextToken()
		p.tokenList = append(p.tokenList, token)
		return token
	}
	return p.tokenList[p.pos+1]
}

func (p *Parser) lookHeadNToken(n int) Token {
	i := p.pos + n
	if i <= len(p.tokenList)-1 {
		return p.tokenList[i]
	}

	for len(p.tokenList)-1 < i {
		token := p.lex.NextToken()
		p.tokenList = append(p.tokenList, token)
	}
	return p.tokenList[i]
}

func (p *Parser) parseField(table *SymbolTable, mods *ast.ASTList, d *ast.Declarator) (*ast.FieldDecl, error) {
	var expr ast.Node
	var err error
	if p.lookAHeadToken().Value == "=" {
		p.getToken()
		expr, err = p.parseExpression(table)
		if err != nil {
			return nil, err
		}
	}
	t := p.getToken()
	if t.Value == ";" {
		return ast.NewFieldDecl(mods, ast.NewASTList(d, ast.NewASTListSingle(expr))), nil
	} else if t.Value == "," {
		return nil, fmt.Errorf(p.syanxError() + "cause: only one field can be declared in one declaration.")
	}
	return nil, fmt.Errorf(p.syanxError())
}

func (p *Parser) parseMethod1(table *SymbolTable, isConstructor bool, mods *ast.ASTList, d *ast.Declarator) (*ast.MethodDecl, error) {
	if p.getToken().Value != "(" {
		return nil, fmt.Errorf(p.syanxError())
	}
	var params *ast.ASTList
	if p.lookAHeadToken().Value != ")" {
		for {
			arg, err := p.parseFormalParam(table)
			if err != nil {
				return nil, err
			}
			params = ast.AppendASTList(params, arg).(*ast.ASTList)
			t := p.lookAHeadToken()
			if t.Value == "," {
				p.getToken()
			} else if t.Value == ")" {
				break
			}
		}
	}
	p.getToken() // ')'
	dimension, err := p.parseArrayDimension()
	if err != nil {
		return nil, err
	}
	d.AddArrayDim(dimension)
	if isConstructor && d.GetArrayDim() > 0 {
		return nil, fmt.Errorf(p.syanxError())
	}

	//解析throws 列表
	var throwsList *ast.ASTList = nil
	if p.lookAHeadToken().ID == ast.Throws {
		p.getToken()
		for {
			classType, err := p.parseClassType(table)
			if err != nil {
				return nil, err
			}
			throwsList = ast.AppendASTList(throwsList, classType).(*ast.ASTList)
			if p.lookAHeadToken().Value == "," {
				p.getToken()
			} else {
				break
			}
		}
	}
	return ast.NewMethodDecl(mods, ast.NewASTList(d, ast.MakeASTList(params, throwsList, nil))), nil
}

// 解析参数声明
func (p *Parser) parseFormalParam(table *SymbolTable) (*ast.Declarator, error) {
	// 参数类型
	d, err := p.parseFormalType(table)
	if err != nil {
		return nil, err
	}
	if p.getToken().ID != ast.Identifier {
		return nil, fmt.Errorf(p.syanxError())
	}
	name := p.getString()
	d.SetVariable(ast.NewSymbol(name))
	dimension, err := p.parseArrayDimension()
	if err != nil {
		return nil, err
	}
	d.AddArrayDim(dimension)
	table.Append(name, d)
	return d, nil
}

func (p *Parser) parseMethod2(table *SymbolTable, md *ast.MethodDecl) (*ast.MethodDecl, error) {
	var body *ast.Statement
	var err error
	if p.lookAHeadToken().Value == ";" {
		p.getToken()
	} else {
		body, err = p.parseBlock(table)
		if err != nil {
			return nil, err
		}
		if body == nil {
			body = ast.NewStatementEmpty(ast.Block)
		}
	}
	md.Sublist(4).(*ast.ASTList).SetHead(body)
	return md, nil
}

func (p *Parser) parseBlock(table *SymbolTable) (*ast.Statement, error) {
	if p.getToken().Value != "{" {
		return nil, fmt.Errorf(p.syanxError())
	}
	var body *ast.Statement
	//局部变量表
	tbl := NewSymbolTable(table)
	for p.lookAHeadToken().Value != "}" {
		s, err := p.parseStatement(tbl)
		if err != nil {
			return nil, err
		}
		if s != nil {
			body = ast.ConcatASTList(body, ast.NewStatementSingle(ast.Block, s)).(*ast.Statement)
		}
	}
	p.getToken() // '}'
	if body == nil {
		//empty block
		return ast.NewStatementEmpty(ast.Block), nil
	}
	return body, nil
}

func (p *Parser) parseStatement(tbl *SymbolTable) (*ast.Statement, error) {
	token := p.lookAHeadToken()
	switch {
	case token.Value == "{":
		block, err := p.parseBlock(tbl)
		return block, err
	case token.Value == ";":
		p.getToken()
		return ast.NewStatementEmpty(ast.Block), nil
	case token.ID == ast.Identifier && p.lookHeadNToken(2).Value == ":":
		//label
		p.getToken()
		labelName := p.getString()
		//:
		p.getToken()

		statement, err := p.parseStatement(tbl)
		if err != nil {
			return nil, err
		}
		return ast.MakeStatement(ast.Label, ast.NewSymbol(labelName), statement), nil
	case token.ID == ast.IF:
		stat, err := p.parseIf(tbl)
		return stat, err
	case token.ID == ast.While:
		stat, err := p.parseWhile(tbl)
		return stat, err
	case token.ID == ast.Do:
		stat, err := p.parseDo(tbl)
		return stat, err
	case token.ID == ast.For:
		stat, err := p.parseFor(tbl)
		return stat, err
	case token.ID == ast.Try:
		stat, err := p.parseTry(tbl)
		return stat, err
	case token.ID == ast.Switch:
		stat, err := p.parseSwitch(tbl)
		return stat, err
	case token.ID == ast.Synchronized:
		stat, err := p.parseSynchronized(tbl)
		return stat, err
	case token.ID == ast.Return:
		stat, err := p.parseReturn(tbl)
		return stat, err
	case token.ID == ast.Throw:
		stat, err := p.parseThrow(tbl)
		return stat, err
	case token.ID == ast.Break:
		stat, err := p.parseBreak(tbl)
		return stat, err
	case token.ID == ast.Continue:
		stat, err := p.parseContinue(tbl)
		return stat, err
	default:
		stat, err := p.parseDeclarationOrExpression(tbl, false)
		return stat, err
	}
}

// 解析if 语句
func (p *Parser) parseIf(tbl *SymbolTable) (*ast.Statement, error) {
	token := p.getToken()
	expr, err := p.parseParExpression(tbl)
	if err != nil {
		return nil, err
	}
	thenp, err := p.parseStatement(tbl)
	if err != nil {
		return nil, err
	}
	var elsep *ast.Statement
	if p.lookAHeadToken().ID == ast.Else {
		p.getToken()
		elsep, err = p.parseStatement(tbl)
		if err != nil {
			return nil, err
		}
	}
	return ast.NewStatement(token.ID, expr, ast.NewASTList(thenp, ast.NewASTListSingle(elsep))), nil
}

func (p *Parser) parseWhile(tbl *SymbolTable) (*ast.Statement, error) {
	t := p.getToken()
	expr, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	body, err := p.parseStatement(tbl)
	if err != nil {
		return nil, err
	}
	return ast.NewStatement(t.ID, expr, body), nil
}

func (p *Parser) parseDo(tbl *SymbolTable) (*ast.Statement, error) {
	t := p.getToken()
	body, err := p.parseStatement(tbl)
	if err != nil {
		return nil, err
	}
	if p.getToken().ID == ast.While && p.lookAHeadToken().Value == "(" {
		expr, err := p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		if p.getToken().Value == ")" && p.lookAHeadToken().Value == ";" {
			return ast.NewStatement(t.ID, expr, body), nil
		}
		return nil, fmt.Errorf(p.syanxError())
	}
	return nil, fmt.Errorf(p.syanxError())
}

func (p *Parser) parseFor(tbl *SymbolTable) (*ast.Statement, error) {
	t := p.getToken()
	//新建局部变量表
	table := NewSymbolTable(tbl)
	if p.getToken().Value != "(" {
		return nil, fmt.Errorf(p.syanxError() + "cause: '(' is missing.")
	}
	var expr *ast.Statement
	var err error
	if p.lookAHeadToken().Value == ";" {
		p.getToken()
	} else {
		expr, err = p.parseDeclarationOrExpression(table, true)
		if err != nil {
			return nil, err
		}
	}
	var node ast.Node
	if p.lookAHeadToken().Value != ";" {
		node, err = p.parseExpression(table)
		if err != nil {
			return nil, err
		}
	}
	if p.getToken().Value != ";" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ';' is missing.")
	}
	var s *ast.Statement
	if p.lookAHeadToken().Value != ")" {
		s, err = p.parseStatement(table)
		if err != nil {
			return nil, err
		}
	}
	if p.getToken().Value != ")" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
	}
	body, err := p.parseStatement(table)
	if err != nil {
		return nil, err
	}
	return ast.NewStatement(t.ID, expr, ast.AppendASTList(node.(*ast.ASTList), ast.ConcatASTList(s.GetASTList(), body.GetASTList()))), nil
}

func (p *Parser) parseTry(tbl *SymbolTable) (*ast.Statement, error) {
	p.getToken()
	block, err := p.parseBlock(tbl)
	if err != nil {
		return nil, err
	}
	var catchList *ast.ASTList
	var d *ast.Declarator
	var b *ast.Statement
	var newErr error
	for p.lookAHeadToken().ID == ast.Catch {
		p.getToken() // catch
		if p.getToken().Value != "(" {
			return nil, fmt.Errorf(p.syanxError() + "cause: '(' is missing.")
		}
		// 新建局部变量表
		table := NewSymbolTable(tbl)
		d, newErr = p.parseFormalParam(tbl)
		if newErr != nil {
			return nil, newErr
		}
		if d.GetArrayDim() > 0 || d.GetType() != ast.Class {
			return nil, fmt.Errorf(p.syanxError())
		}
		if p.getToken().Value != ")" {
			return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
		}
		b, newErr = p.parseBlock(table)
		if newErr != nil {
			return nil, newErr
		}
		catchList = ast.AppendASTList(catchList, ast.NewPair(d, b)).(*ast.ASTList)
	}
	var finallyBlock *ast.Statement
	if p.lookAHeadToken().ID == ast.Finally {
		p.getToken()
		finallyBlock, newErr = p.parseBlock(tbl)
		if newErr != nil {
			return nil, newErr
		}
	}
	return ast.MakeStatementThree(ast.Try, block, catchList, finallyBlock), nil
}

func (p *Parser) parseSwitch(tbl *SymbolTable) (*ast.Statement, error) {
	t := p.getToken()
	expr, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	body, err := p.parseSwitchBlock(tbl)
	if err != nil {
		return nil, err
	}
	return ast.NewStatement(t.ID, expr, body.GetASTList()), nil
}

func (p *Parser) parseSwitchBlock(tbl *SymbolTable) (*ast.Statement, error) {
	if p.getToken().Value != "{" {
		return nil, fmt.Errorf(p.syanxError())
	}
	//创建局部变量表
	table := NewSymbolTable(tbl)
	orCase, err := p.parseStmntOrCase(table)
	if err != nil {
		return nil, err
	}
	if orCase == nil {
		return nil, fmt.Errorf(p.syanxError() + "cause: empty switch block")
	}
	op := orCase.GetOperator()
	if op != ast.Case && op != ast.Default {
		return nil, fmt.Errorf(p.syanxError() + "case: no case or default in a switch block")
	}
	body := ast.NewStatementSingle(ast.Block, orCase)
	for {
		for {
			var s *ast.Statement
			var err error
			for s == nil {
				if p.lookAHeadToken().Value == "}" {
					p.getToken()
					return body, nil
				}
				s, err = p.parseStmntOrCase(table)
				if err != nil {
					return nil, err
				}
			}

			operator := s.GetOperator()
			if operator != ast.Case && operator != ast.Default {
				orCase = ast.ConcatASTList(orCase.GetASTList(), ast.NewStatementSingle(ast.Block, s).GetASTList()).(*ast.Statement)
			} else {
				body = ast.ConcatASTList(body.GetASTList(), ast.NewStatementSingle(ast.Block, s).GetASTList()).(*ast.Statement)

				orCase = s
			}
		}
	}
}

func (p *Parser) parseStmntOrCase(table *SymbolTable) (*ast.Statement, error) {
	t := p.lookAHeadToken()
	if t.ID != ast.Case && t.ID != ast.Default {
		statement, err := p.parseStatement(table)
		return statement, err
	}
	p.getToken()
	var s *ast.Statement
	if t.ID == ast.Case {
		expr, e := p.parseExpression(table)
		if e != nil {
			return nil, e
		}
		s = ast.NewStatementSingle(t.ID, expr)
	} else {
		s = ast.NewStatementEmpty(ast.Default)
	}
	if p.getToken().Value != ":" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ':' is missing.")
	}
	return s, nil
}

func (p *Parser) parseSynchronized(tbl *SymbolTable) (*ast.Statement, error) {
	t := p.getToken()
	if p.getToken().Value != "(" {
		return nil, fmt.Errorf(p.syanxError() + "cause: '(' is missing.")
	}
	expr, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	if p.getToken().Value != ")" {
		return nil, fmt.Errorf(p.syanxError())
	}
	body, err := p.parseBlock(tbl)
	if err != nil {
		return nil, err
	}
	return ast.NewStatement(t.ID, expr, body.GetASTList()), nil
}

func (p *Parser) parseReturn(tbl *SymbolTable) (*ast.Statement, error) {
	token := p.getToken()
	s := ast.NewStatementEmpty(token.ID)
	if p.lookAHeadToken().Value != ";" {
		expression, err := p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		s.SetLeft(expression)
	}
	if p.getToken().Value != ";" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ';' is missing.")
	}
	return s, nil
}

func (p *Parser) parseThrow(tbl *SymbolTable) (*ast.Statement, error) {
	token := p.getToken()
	expression, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	if p.getToken().Value != ";" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ';' is missing.")
	}
	return ast.NewStatementSingle(token.ID, expression), nil
}

func (p *Parser) parseBreak(tbl *SymbolTable) (*ast.Statement, error) {
	return p.parseContinue(tbl)
}

func (p *Parser) parseContinue(tbl *SymbolTable) (*ast.Statement, error) {
	t1 := p.getToken()
	s := ast.NewStatementEmpty(t1.ID)
	t2 := p.getToken()
	if t2.ID == ast.Identifier {
		s.SetLeft(ast.NewSymbol(p.getString()))
		t2 = p.getToken()
	}
	if p.getToken().Value != ";" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ';' is missing.")
	}
	return s, nil
}

func (p *Parser) parseDeclarationOrExpression(tbl *SymbolTable, exprList bool) (*ast.Statement, error) {
	t := p.lookAHeadToken().ID
	for t == ast.Finally {
		p.getToken()
		t = p.lookAHeadToken().ID
	}
	if t.IsBasicType() {
		//基本类型声明
		token := p.getToken()
		i, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		declarators, err := p.parseDeclarators(tbl, ast.NewDeclaratorTypeDim(token.ID, i))
		return declarators, err
	} else {
		if t == ast.Identifier {
			i := p.nextIsClassType(1)
			if i > 0 && p.lookHeadNToken(i).ID == ast.Identifier {
				classType, err := p.parseClassType(tbl)
				if err != nil {
					return nil, err
				}
				if classType == nil {
					return nil, fmt.Errorf(p.syanxError() + "unkwon class.")
				}
				dimension, err := p.parseArrayDimension()
				if err != nil {
					return nil, err
				}
				declarators, err := p.parseDeclarators(tbl, ast.NewDeclaratorWithClassName(classType, dimension))
				return declarators, err
			}
		}

		var expr *ast.Statement
		var err error
		if exprList {
			expr, err = p.parseExprList(tbl)
		} else {
			node, err := p.parseExpression(tbl)
			if err != nil {
				return nil, err
			}
			expr = ast.NewStatementSingle(ast.Expr, node)
		}

		if p.getToken().Value != ";" {
			return nil, fmt.Errorf(p.syanxError() + "cause: ';' is missing.")
		} else {
			return expr, err
		}
	}
}

func (p *Parser) parseDeclarators(tbl *SymbolTable, d *ast.Declarator) (*ast.Statement, error) {
	var decl *ast.Statement

	for {
		node, err := p.ParseDeclarator(tbl, d)
		if err != nil {
			return nil, err
		}

		decl = ast.ConcatASTList(decl.GetASTList(), ast.NewStatementSingle(ast.Decl, node)).(*ast.Statement)

		t := p.getToken()
		if t.Value == ";" { // ';' 结束符
			return decl, nil
		} else if t.Value == "null" {
			p.getToken()
			return decl, nil
		} else if t.Value != "," { // 不是 ',' 逗号
			return nil, errors.New(p.syanxError() + "cause: ';' is missing.")
		}
	}
}

func (p *Parser) ParseDeclarator(tbl *SymbolTable, d *ast.Declarator) (*ast.Declarator, error) {
	if p.getToken().ID != ast.Identifier || d.GetType() == ast.Void {
		return nil, fmt.Errorf(p.syanxError())
	}
	name := p.getString()
	symbol := ast.NewSymbol(name)
	dim, err := p.parseArrayDimension()
	if err != nil {
		return nil, err
	}
	var init ast.Node
	if p.lookAHeadToken().Value == "=" {
		p.getToken()
		init, err = p.parseInitializer(tbl)
		if err != nil {
			return nil, err
		}
	}
	decl := d.Make(symbol, dim, init)
	tbl.Append(name, decl)
	return decl, nil

}

func (p *Parser) parseInitializer(tbl *SymbolTable) (ast.Node, error) {
	if p.lookAHeadToken().Value == "{" {
		return p.parseArrayInitializer(tbl)
	}
	return p.parseExpression(tbl)
}

/* array.initializer :
 *  '{' (( array.initializer | expression ) ',')* '}'
 */
func (p *Parser) parseArrayInitializer(tbl *SymbolTable) (*ast.ArrayInit, error) {
	p.getToken() // '{'
	if p.lookAHeadToken().Value == "}" {
		p.getToken()
		return ast.NewArrayInit(nil), nil
	}
	expr, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	init := ast.NewArrayInit(expr)
	for p.lookAHeadToken().Value == "," {
		p.getToken()
		expr, err = p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		ast.AppendASTList(init.ASTList, expr)
	}
	if p.getToken().Value != "}" {
		return nil, fmt.Errorf(p.syanxError() + "cause: '}' is missing.")
	}
	return init, nil
}

/* par.expression : '(' expression ')'
 */
func (p *Parser) parseExpression(tbl *SymbolTable) (ast.Node, error) {
	left, err := p.parseConditionalExpr(tbl)
	if err != nil {
		return nil, err
	}
	//如果不是赋值符
	if !p.lookAHeadToken().ID.IsAssignOp() {
		return left, nil
	}
	t := p.getToken()
	right, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	return ast.MakeAssign(t.ID, left, right), nil
}

// 获取全类名
func (p *Parser) nextIsClassType(i int) int {
	for {
		i++
		if p.lookAHeadToken().Value == "." {
			i++
			if p.lookAHeadToken().ID == ast.Identifier {
				continue
			}
			return 0
		}

		for {
			if p.lookHeadNToken(i).Value != "[" {
				return i
			}
			i++

			if p.lookHeadNToken(i).Value == "]" {
				break
			}
			i++
		}
		return 0
	}
}

func (p *Parser) parseExprList(tbl *SymbolTable) (*ast.Statement, error) {
	var expr *ast.Statement
	for {
		expression, err := p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		e := ast.NewStatementSingle(ast.Expr, expression)
		expr = ast.ConcatASTList(expr, ast.NewStatementSingle(ast.Block, e)).(*ast.Statement)

		if p.lookAHeadToken().Value != "," {
			return expr, nil
		}
		p.getToken()
	}
}

// 条件表达式解析
/* conditional.expr                 (right-to-left)
 *     : logical.or.expr [ '?' expression ':' conditional.expr ]
 */
func (p *Parser) parseConditionalExpr(tbl *SymbolTable) (ast.Node, error) {
	cond, err := p.parseBinaryExpr(tbl)
	if err != nil {
		return nil, err
	}
	if p.lookAHeadToken().Value == "?" {
		p.getToken()
		thenExpr, err := p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		if p.getToken().Value == ":" {
			return nil, fmt.Errorf(p.syanxError() + "cause: ':' is missing.")
		}
		elseExpr, err := p.parseExpression(tbl)
		if err != nil {
			return nil, err
		}
		return ast.NewCondExpr(cond, thenExpr, elseExpr), nil
	}
	return cond, nil
}

// 解析二元运算表达式
/* logical.or.expr          10 (operator precedence)
 * : logical.and.expr
 * | logical.or.expr OROR logical.and.expr          left-to-right
 *
 * logical.and.expr         9
 * : inclusive.or.expr
 * | logical.and.expr ANDAND inclusive.or.expr
 *
 * inclusive.or.expr        8
 * : exclusive.or.expr
 * | inclusive.or.expr "|" exclusive.or.expr
 *
 * exclusive.or.expr        7
 *  : and.expr
 * | exclusive.or.expr "^" and.expr
 *
 * and.expr                 6
 * : equality.expr
 * | and.expr "&" equality.expr
 *
 * equality.expr            5
 * : relational.expr
 * | equality.expr (EQ | NEQ) relational.expr
 *
 * relational.expr          4
 * : shift.expr
 * | relational.expr (LE | GE | "<" | ">") shift.expr
 * | relational.expr INSTANCEOF class.type ("[" "]")*
 *
 * shift.expr               3
 * : additive.expr
 * | shift.expr (LSHIFT | RSHIFT | ARSHIFT) additive.expr
 *
 * additive.expr            2
 * : multiply.expr
 * | additive.expr ("+" | "-") multiply.expr
 *
 * multiply.expr            1
 * : unary.expr
 * | multiply.expr ("*" | "/" | "%") unary.expr
 */

func (p *Parser) parseBinaryExpr(tbl *SymbolTable) (ast.Node, error) {
	expr, err := p.parseUnaryExpr(tbl)
	if err != nil {
		return nil, err
	}

	for {
		t := p.lookAHeadToken()
		prec := ast.GetOpPrecedence(t.Value)
		if prec == 0 {
			return expr, nil
		}
		expr, err = p.binaryExpr2(tbl, expr, prec)
		break
	}

	return expr, nil
}

/* unary.expr : "++"|"--" unary.expr
              | "+"|"-" unary.expr
              | "!"|"~" unary.expr
              | cast.expr
              | postfix.expr

   unary.expr.not.plus.minus is a unary expression starting without
   "+", "-", "++", or "--".
*/

func (p *Parser) parseUnaryExpr(tbl *SymbolTable) (ast.Node, error) {
	var t Token
	switch p.lookAHeadToken().Value {
	case "+", "-", "++", "--", "!", "~":
		t = p.getToken()
		if t.Value == "~" {
			t2 := p.lookAHeadToken().ID
			switch t2 {
			case ast.LongConstant, ast.IntConstant, ast.CharConstant:
				p.getToken()
				value, err := strconv.ParseInt(p.getString(), 10, 64)
				if err != nil {
					return nil, fmt.Errorf(p.syanxError() + "cause: not a Interget: " + p.getString())
				}
				return ast.NewIntConst(-value, t2), nil
			case ast.DoubleConstant, ast.FloatConstant:
				p.getToken()
				value, err := strconv.ParseFloat(p.getString(), 64)
				if err != nil {
					return nil, fmt.Errorf(p.syanxError() + "cause: not a Float or Double: " + p.getString())
				}
				return ast.NewDoubleConst(-value, t2), nil
			default:
				break
			}
		}
		expr, err := p.parseUnaryExpr(tbl)
		if err != nil {
			return nil, err
		}
		return ast.MakeExpressionSingle(t.ID, expr), nil
	case "(":
		return p.parseCast(tbl)
	default:
		return p.parsePostfix(tbl)
	}
}

func (p *Parser) parseParExpression(tbl *SymbolTable) (ast.Node, error) {
	if p.getToken().Value != "(" {
		return nil, fmt.Errorf(p.syanxError() + "cause: '(' is missing.")
	}
	expr, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	if p.getToken().Value != ")" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
	}
	return expr, nil
}

func (p *Parser) binaryExpr2(tbl *SymbolTable, expr ast.Node, prec int) (ast.Node, error) {
	t := p.getToken()
	if t.ID == ast.Instanceof {
		return p.parseInstanceOf(tbl, expr)
	}
	expr2, err := p.parseUnaryExpr(tbl)
	var newErr error
	if err != nil {
		return nil, err
	}
	for {
		t2 := p.lookAHeadToken()
		p2 := ast.GetOpPrecedence(t2.Value)
		if p2 != 0 && prec > p2 {
			expr2, newErr = p.binaryExpr2(tbl, expr2, p2)
			if newErr != nil {
				return nil, newErr
			}
		} else {
			return ast.MakeBinExpr(t.ID, expr, expr2), nil
		}
	}
}

func (p *Parser) parseInstanceOf(tbl *SymbolTable, expr ast.Node) (ast.Node, error) {
	t := p.lookAHeadToken()
	if t.ID.IsBasicType() {
		p.getToken()
		dimension, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		return ast.NewInstanceOfExprWithType(t.ID, dimension, expr), nil
	}
	className, err := p.parseClassType(tbl)
	if err != nil {
		return nil, err
	}
	dim, err := p.parseArrayDimension()
	if err != nil {
		return nil, err
	}
	return ast.NewInstanceOfExprWithClass(className, dim, expr), nil
}

func (p *Parser) parseCast(tbl *SymbolTable) (ast.Node, error) {
	t := p.lookHeadNToken(2)
	if t.ID.IsBasicType() && p.nextIsBuiltinCast() {
		p.getToken() // '('
		p.getToken() // primitive type
		dim, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		if p.getToken().Value != ")" {
			return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
		}
		expr, err := p.parseUnaryExpr(tbl)
		if err != nil {
			return nil, err
		}
		return ast.NewCastExprWithType(t.ID, dim, expr), nil
	} else if t.ID == ast.Identifier && p.nextIsClassCast() {
		p.getToken() // '('
		className, err := p.parseClassType(tbl)
		if err != nil {
			return nil, err
		}
		dim, err := p.parseArrayDimension()
		if err != nil {
			return nil, err
		}
		if p.getToken().Value != ")" {
			return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
		}
		expr, err := p.parseUnaryExpr(tbl)
		if err != nil {
			return nil, err
		}
		return ast.NewCastExprWithClass(className, dim, expr), nil
	} else {
		return p.parsePostfix(tbl)
	}
}

func (p *Parser) nextIsBuiltinCast() bool {
	var t Token
	i := 3
	for {
		t = p.lookHeadNToken(i)
		i++
		if t.Value != "[" {
			break
		}
		if p.lookHeadNToken(i).Value != "]" {
			return false
		}
		i++
	}
	return p.lookHeadNToken(i-1).Value == ")"
}

func (p *Parser) nextIsClassCast() bool {
	i := p.nextIsClassType(2)
	if i < 0 {
		return false
	}
	t := p.lookHeadNToken(i)
	if t.Value == ")" {
		return false
	}
	token := p.lookHeadNToken(i + 1)
	return token.Value == "(" || token.ID == ast.Null || token.ID == ast.StringL || token.ID == ast.Identifier || token.ID == ast.This || token.ID == ast.Super || token.ID == ast.New || token.ID == ast.True || token.ID == ast.False || token.ID == ast.LongConstant || token.ID == ast.IntConstant || token.ID == ast.CharConstant || token.ID == ast.DoubleConstant || token.ID == ast.FloatConstant
}

func (p *Parser) parsePostfix(tbl *SymbolTable) (ast.Node, error) {
	token := p.lookAHeadToken()
	switch token.ID {
	case ast.LongConstant, ast.IntConstant, ast.CharConstant:
		p.getToken()
		value, err := strconv.ParseInt(p.getString(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf(p.syanxError() + "cause: not a Interget: " + p.getString())
		}
		return ast.NewIntConst(value, token.ID), nil
	case ast.DoubleConstant, ast.FloatConstant:
		p.getToken()
		value, err := strconv.ParseFloat(p.getString(), 64)
		if err != nil {
			return nil, fmt.Errorf(p.syanxError() + "cause: not a Float or Double: " + p.getString())
		}
		return ast.NewDoubleConst(value, token.ID), nil
	default:
		break
	}
	var index ast.Node
	expr, err := p.parsePrimaryExpr(tbl)
	if err != nil {
		return nil, err
	}
	var t Token
	for {
		switch p.lookAHeadToken().Value {
		case "(":
			expr, err = p.parseMethodCall(tbl, expr)
			if err != nil {
				return nil, err
			}
			break
		case "[":
			if p.lookHeadNToken(2).Value == "]" {
				dim, err := p.parseArrayDimension()
				if err != nil {
					return nil, err
				}
				if p.getToken().Value != "." || p.getToken().ID != ast.Class {
					return nil, fmt.Errorf(p.syanxError())
				}
				expr, err = p.parseDotClass(expr, dim)
				if err != nil {
					return nil, err
				}
			} else {
				index, err = p.parseArrayIndex(tbl)
				if err != nil || index == nil {
					return nil, fmt.Errorf(p.syanxError())
				}
				expr = ast.MakeExpression(ast.Array, expr, index)
			}
			break
		case "++", "--":
			t = p.getToken()
			expr = ast.MakeExpression(t.ID, nil, expr)
			break
		case ".":
			p.getToken()
			t = p.getToken()
			if t.ID == ast.Class {
				expr, err = p.parseDotClass(expr, 0)
				if err != nil {
					return nil, err
				}
			} else if t.ID == ast.Super {
				name, err := toClassName(expr)
				if err != nil {
					return nil, fmt.Errorf(p.syanxError() + "cause: " + err.Error())
				}
				expr = ast.MakeExpression(ast.Dot, ast.NewSymbol(name), ast.NewKeyword(t.Value, t.ID))
			} else if t.ID == ast.Identifier {
				expr = ast.MakeExpression(ast.Dot, expr, ast.NewMemberSymbol(p.getString()))
			} else {
				return nil, fmt.Errorf(p.syanxError() + "cause: missing member name")
			}
			break
		case "#":
			p.getToken()
			t = p.getToken()
			if t.ID != ast.Identifier {
				return nil, fmt.Errorf(p.syanxError() + "cause: missing static member name")
			}
			name, err := toClassName(expr)
			if err != nil {
				return nil, fmt.Errorf(p.syanxError() + "cause: " + err.Error())
			}
			expr = ast.MakeExpression(ast.Member, ast.NewSymbol(name), ast.NewMemberSymbol(p.getString()))
			break
		default:
			return expr, nil
		}
	}
}

/* primary.expr : THIS | SUPER | TRUE | FALSE | NULL
 *              | StringL
 *              | Identifier
 *              | NEW new.expr
 *              | "(" expression ")"
 *              | builtin.type ( "[" "]" )* "." CLASS
 *
 * Identifier represents either a local variable name, a member name,
 * or a class name.
 */
func (p *Parser) parsePrimaryExpr(tbl *SymbolTable) (ast.Node, error) {
	t := p.getToken()
	var decl *ast.Declarator
	switch t.ID {
	case ast.Null, ast.This, ast.Super, ast.True, ast.False:
		return ast.NewKeyword(t.Value, t.ID), nil
	case ast.Identifier:
		name := p.getString()
		decl = tbl.Lookup(name)
		if decl == nil {
			// this or static member
			return ast.NewMemberSymbol(name), nil
		}
		// local variable
		return ast.NewVariable(name, decl), nil
	case ast.StringL:
		return ast.NewStringL(p.getString()), nil
	case ast.New:
		return p.parseNew(tbl)
	default:
		if t.Value == "(" {
			expr, err := p.parseExpression(tbl)
			if err != nil {
				return nil, err
			}
			if p.getToken().Value != ")" {
				return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
			}
			return expr, nil
		}
		if t.ID.IsBasicType() || t.ID == ast.Void {
			dim, err := p.parseArrayDimension()
			if err != nil {
				return nil, err
			}
			if p.getToken().Value == "." && p.getToken().ID == ast.Class {
				return p.parseDotClassbuiltinType(t.ID, dim)
			}
		}
		return nil, fmt.Errorf(p.syanxError())
	}
}

func (p *Parser) parseNew(tbl *SymbolTable) (*ast.NewExpr, error) {
	var init *ast.ArrayInit
	t := p.lookAHeadToken()
	if t.ID.IsBasicType() {
		p.getToken()
		size, err := p.parseArraySize(tbl)
		if err != nil {
			return nil, err
		}
		if p.lookAHeadToken().Value == "{" {
			init, err = p.parseArrayInitializer(tbl)
			if err != nil {
				return nil, err
			}
		}
		return ast.NewNewExprWithType(t.ID, size, init), nil
	} else if t.ID == ast.Identifier {
		className, err := p.parseClassType(tbl)
		if err != nil {
			return nil, err
		}
		t = p.lookAHeadToken()
		if t.Value == "(" {
			args, err := p.parseArgumentList(tbl)
			if err != nil {
				return nil, err
			}
			return ast.NewNewExprWithClass(className, args), nil
		} else if t.Value == "[" {
			size, err := p.parseArraySize(tbl)
			if err != nil {
				return nil, err
			}
			if p.lookAHeadToken().Value == "{" {
				init, err = p.parseArrayInitializer(tbl)
				if err != nil {
					return nil, err
				}
			}
			return ast.MakeObjectArray(className, size, init), nil
		}
	}
	return nil, fmt.Errorf(p.syanxError())
}

func (p *Parser) parseArraySize(tbl *SymbolTable) (*ast.ASTList, error) {
	var list *ast.ASTList
	for p.lookAHeadToken().Value == "[" {
		item, err := p.parseArrayIndex(tbl)
		if err != nil {
			return nil, err
		}
		list = ast.AppendASTList(list, item).(*ast.ASTList)
	}
	return list, nil
}

func (p *Parser) parseArrayIndex(tbl *SymbolTable) (ast.Node, error) {
	p.getToken() // '['
	if p.lookAHeadToken().Value == "]" {
		p.getToken()
		return nil, nil
	}
	index, err := p.parseExpression(tbl)
	if err != nil {
		return nil, err
	}
	if p.getToken().Value != "]" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ']' is missing.")
	}
	return index, nil
}

func (p *Parser) parseArgumentList(tbl *SymbolTable) (*ast.ASTList, error) {
	if p.getToken().Value != "(" {
		return nil, fmt.Errorf(p.syanxError() + "cause: '(' is missing.")
	}
	var list *ast.ASTList
	if p.lookAHeadToken().Value != ")" {
		for {
			expr, err := p.parseExpression(tbl)
			if err != nil {
				return nil, err
			}
			list = ast.AppendASTList(list, expr).(*ast.ASTList)
			if p.lookAHeadToken().Value == "," {
				p.getToken()
			} else {
				break
			}
		}
	}
	if p.getToken().Value != ")" {
		return nil, fmt.Errorf(p.syanxError() + "cause: ')' is missing.")
	}
	return list, nil
}

func (p *Parser) parseDotClassbuiltinType(builtinType ast.TokenID, dim int) (ast.Node, error) {
	if dim > 0 {
		cname := toJvmTypeName(builtinType, dim)
		return ast.MakeExpression(ast.Dot, ast.NewSymbol(cname), ast.NewMemberSymbol("class")), nil
	}
	var cname string
	switch builtinType {
	case ast.Boolean:
		cname = "java.lang.Boolean"
		break
	case ast.Byte:
		cname = "java.lang.Byte"
		break
	case ast.Char:
		cname = "java.lang.Character"
		break
	case ast.Short:
		cname = "java.lang.Short"
		break
	case ast.Int:
		cname = "java.lang.Integer"
		break
	case ast.Long:
		cname = "java.lang.Long"
		break
	case ast.Float:
		cname = "java.lang.Float"
		break
	case ast.Double:
		cname = "java.lang.Double"
		break
	case ast.Void:
		cname = "java.lang.Void"
		break
	default:
		return nil, fmt.Errorf(p.syanxError() + "cause: invalid builtin type:" + builtinType.String())
	}
	return ast.MakeExpression(ast.Member, ast.NewSymbol(cname), ast.NewMemberSymbol("TYPE")), nil
}

/* method.call : method.expr "(" argument.list ")"
 * method.expr : THIS | SUPER | Identifier
 *             | postfix.expr "." Identifier
 *             | postfix.expr "#" Identifier
 */
func (p *Parser) parseMethodCall(tbl *SymbolTable, expr ast.Node) (ast.Node, error) {
	if key, ok := expr.(*ast.Keyword); ok {
		id := key.TokenID
		if id != ast.This && id != ast.Super {
			return nil, fmt.Errorf(p.syanxError())
		}
	}
	if expression, ok := expr.(*ast.Expression); ok {
		op := expression.GetOperator()
		if op != ast.Dot && op != ast.Member {
			return nil, fmt.Errorf(p.syanxError())
		}
	}
	args, err := p.parseArgumentList(tbl)
	if err != nil {
		return nil, err
	}
	return ast.MakeCall(expr, args), nil
}

/* Parse a .class expression on a class type.  For example,
 * String.class   => ('.' "String" "class")
 * String[].class => ('.' "[LString;" "class")
 */
func (p *Parser) parseDotClass(className ast.Node, dim int) (ast.Node, error) {
	cname, err := toClassName(className)
	if err != nil {
		return nil, fmt.Errorf(p.syanxError() + "cause: " + err.Error())
	}
	if dim > 0 {
		// 构建类名
		var sbuf strings.Builder
		for ; dim > 0; dim-- {
			sbuf.WriteString("[")
		}
		sbuf.WriteString("L")
		sbuf.WriteString(strings.ReplaceAll(cname, ".", "/"))
		sbuf.WriteString(";")
		cname = sbuf.String()
	}
	return ast.MakeExpression(ast.Dot, ast.NewSymbol(cname), ast.NewMemberSymbol("class")), nil
}

func (p *Parser) HasMore() bool {
	return p.lex.HasNextToken()
}
