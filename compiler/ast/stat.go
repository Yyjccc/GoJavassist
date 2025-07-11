package ast

type Statement struct {
	Node
	operatorId TokenID
}

func (s *Statement) GetASTList() *ASTList {
	if s == nil {
		return nil
	}
	return s.Node.(*ASTList)
}

func NewStatement(op TokenID, head, tail Node) *Statement {
	return &Statement{
		Node:       NewASTList(head, tail),
		operatorId: op,
	}
}

func ReNewNewStatement(op TokenID, list *ASTList) *Statement {
	return &Statement{
		Node:       list,
		operatorId: op,
	}
}

func NewStatementSingle(op TokenID, head Node) *Statement {
	return &Statement{
		Node:       NewASTList(head, nil),
		operatorId: op,
	}
}

func NewStatementEmpty(op TokenID) *Statement {
	return &Statement{
		Node:       NewASTList(nil, nil),
		operatorId: op,
	}
}

func MakeStatement(op TokenID, operand1, operand2 Node) *Statement {
	return NewStatement(op, operand1, NewASTList(operand2, nil))
}

func MakeStatementThree(op TokenID, op1, op2, op3 Node) *Statement {
	return NewStatement(op, op1, NewASTList(op2, NewASTList(op3, nil)))
}

func (s *Statement) Accept(v Visitor) error {
	return v.AtStatement(s)
}

func (s *Statement) GetOperator() TokenID {
	return s.operatorId
}

func (s *Statement) GetTag() string {
	if s.operatorId < 128 {
		return "stmnt:" + string(rune(s.operatorId))
	}
	return "stmnt:" + string(s.operatorId)
}
