package compiler

import "GoJavassist/compiler/ast"

// SymbolTable 表示符号表，内部保存一个 map 以及父符号表引用
type SymbolTable struct {
	parent *SymbolTable
	table  map[string]*ast.Declarator
}

// NewSymbolTable 创建一个新的符号表，传入父符号表可以为 nil
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		parent: parent,
		table:  make(map[string]*ast.Declarator),
	}
}

// GetParent 返回父符号表
func (s *SymbolTable) GetParent() *SymbolTable {
	return s.parent
}

// Lookup 在当前符号表中查找指定名称的 Declarator，如果未找到，则递归查找父符号表
func (s *SymbolTable) Lookup(name string) *ast.Declarator {
	if decl, ok := s.table[name]; ok {
		return decl
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil
}

// Append 添加一个符号到当前符号表
func (s *SymbolTable) Append(name string, value *ast.Declarator) {
	s.table[name] = value
}
