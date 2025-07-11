package ast

type Visitor interface {
	AtASTList(l *ASTList) error
	AtDeclarator(n *Declarator) error
	AtStatement(s *Statement) error
	AtExpression(e *Expression) error
	AtCondExpr(e *CondExpr) error
	AtCastExpr(e *CastExpr) error
	AtArrayInit(a *ArrayInit) error
	AtNewExpr(e *NewExpr) error
	AtAssignExpr(a *AssignExpr) error
	AtBinExpr(a *BinExpr) error
	AtCallExpr(c *CallExpr) error
	AtPair(p *Pair) error
	AtIntConst(i *IntConst) error
	AtDoubleConst(d *DoubleConst) error
	AtInstanceOfExpr(i *InstanceOfExpr) error
	AtStringL(s *StringLiteral) error
	AtMember(m *MemberSymbol) error
	AtVariable(v *Variable) error
	AtKeyword(k *Keyword) error
	AtMethodDecl(d *MethodDecl) error
	AtFieldDecl(d *FieldDecl) error
	AtSymbol(s *Symbol) error
}
