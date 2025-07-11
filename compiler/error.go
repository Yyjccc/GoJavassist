package compiler

import "github.com/Yyjccc/GoJavassist/compiler/ast"

type CompileError struct {
	error
	detail string
}

func (e CompileError) Error() string {
	return "[compile error] " + e.detail + "."
}

func NewCompileError(detail string) *CompileError {
	return &CompileError{detail: detail}
}

type NoFieldError struct {
	parent    *CompileError
	fieldName string
	expr      ast.Node
}

func (e *NoFieldError) Expr() ast.Node {
	return e.expr
}

func NewNoFieldError(name string, expr ast.Node) *NoFieldError {
	err := NewCompileError("no such field:" + name)
	return &NoFieldError{
		parent:    err,
		fieldName: name,
		expr:      expr,
	}
}

func (e *NoFieldError) Error() string {
	return e.parent.Error()
}
