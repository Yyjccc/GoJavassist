package ast

import (
	"GoJavassist/compiler/reflect"
)

type Symbol struct {
	Node
	Identifier string
}

func NewSymbol(name string) *Symbol {
	return &Symbol{
		Identifier: name,
	}
}

func (s *Symbol) get() string {
	return s.Identifier
}

func (s *Symbol) Accept(v Visitor) error {
	return v.AtSymbol(s)
}

type Keyword struct {
	BaseNode
	Name    string
	TokenID TokenID
}

func NewKeyword(name string, ID TokenID) *Keyword {
	return &Keyword{
		BaseNode: BaseNode{},
		Name:     name,
		TokenID:  ID,
	}
}

func (k *Keyword) Accept(visitor Visitor) error {
	return visitor.AtKeyword(k)
}

// field
type MemberSymbol struct {
	*Symbol
	Field *reflect.CtField
}

func NewMemberSymbol(name string) *MemberSymbol {
	return &MemberSymbol{
		Symbol: NewSymbol(name),
		Field:  nil,
	}
}

func (m *MemberSymbol) GetField() *reflect.CtField {
	if m == nil {
		return nil
	}
	return m.Field
}

func (m *MemberSymbol) Accept(v Visitor) error {
	return v.AtMember(m)
}

type Variable struct {
	*Symbol
	Declarator *Declarator
}

func NewVariable(name string, declarator *Declarator) *Variable {
	return &Variable{
		Symbol:     NewSymbol(name),
		Declarator: declarator,
	}
}

func (v *Variable) ToString() string {
	return v.Identifier + ":" + v.Declarator.GetType().String()
}

func (v *Variable) Accept(visitor Visitor) error {
	return visitor.AtVariable(v)
}
