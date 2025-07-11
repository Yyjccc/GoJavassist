package ast

import "strings"

// decl.go : include class decl , method decl ,field decl ,var decl

type Declarator struct {
	*ASTList
	varType        TokenID
	arrayDim       int
	localVar       int
	qualifiedClass string
}

// NewDeclaratorTypeDim 对应 Java 构造函数：Declarator(int type, int dim)
func NewDeclaratorTypeDim(t TokenID, dim int) *Declarator {
	return &Declarator{
		ASTList:        NewASTListSingle(nil), // 对应 super((Node)null)
		varType:        t,
		arrayDim:       dim,
		localVar:       -1,
		qualifiedClass: "",
	}
}

// NewDeclaratorWithClassName 对应 Java 构造函数：Declarator(ASTList className, int dim)
func NewDeclaratorWithClassName(className *ASTList, dim int) *Declarator {
	return &Declarator{
		ASTList:        NewASTListSingle(nil),
		varType:        Class,
		arrayDim:       dim,
		localVar:       -1,
		qualifiedClass: AstToClassName(className, '/'),
	}
}

// NewDeclaratorFull 对应 Java 构造函数：Declarator(int type, String jvmClassName, int dim, int var, Symbol sym)
func NewDeclaratorFull(t TokenID, jvmClassName string, dim int, localVar int, sym Node) *Declarator {
	d := &Declarator{
		ASTList:        NewASTListSingle(nil),
		varType:        t,
		arrayDim:       dim,
		localVar:       localVar,
		qualifiedClass: jvmClassName,
	}
	d.SetLeft(sym)
	// Java 代码中调用 append(this, (Node)null)；此处保持尾部为 nil
	d.SetTail(nil)
	return d
}

// Make 创建一个新的 Declarator，其 arrayDim 为当前值加上 dim，并设置新符号和初始化表达式。
// 对应 Java 中：public Declarator make(Symbol sym, int dim, Node init)
func (d *Declarator) Make(sym Node, dim int, init Node) *Declarator {
	newD := NewDeclaratorTypeDim(d.varType, d.arrayDim+dim)
	newD.qualifiedClass = d.qualifiedClass
	newD.SetLeft(sym)
	// AppendASTList 将 init 追加到 newD 链表的末尾
	AppendASTList(newD.ASTList, init)
	return newD
}

// GetType 返回变量类型，对应 Java 的 getType()
func (d *Declarator) GetType() TokenID {
	return d.varType
}

// GetArrayDim 返回数组维度，对应 Java 的 getArrayDim()
func (d *Declarator) GetArrayDim() int {
	return d.arrayDim
}

// AddArrayDim 增加数组维度，对应 Java 的 addArrayDim(int d)
func (d *Declarator) AddArrayDim(delta int) {
	d.arrayDim += delta
}

// GetClassName 返回完整类名，对应 Java 的 getClassName()
func (d *Declarator) GetClassName() string {
	return d.qualifiedClass
}

// SetClassName 设置完整类名，对应 Java 的 setClassName(String s)
func (d *Declarator) SetClassName(s string) {
	d.qualifiedClass = s
}

// GetVariable 返回变量符号，对应 Java 的 getVariable()
// 注意：这里假定 d.GetLeft() 返回的节点可断言为 *Symbol
func (d *Declarator) GetVariable() *Symbol {
	if sym, ok := d.GetLeft().(*Symbol); ok {
		return sym
	}
	return nil
}

// SetVariable 设置变量符号，对应 Java 的 setVariable(Symbol sym)
func (d *Declarator) SetVariable(sym Node) {
	d.SetLeft(sym)
}

// GetInitializer 返回初始化表达式，对应 Java 的 getInitializer()
func (d *Declarator) GetInitializer() Node {
	t := d.Tail()
	if t != nil {
		return t.Head()
	}
	return nil
}

// SetLocalVar 设置局部变量索引，对应 Java 的 setLocalVar(int n)
func (d *Declarator) SetLocalVar(n int) {
	d.localVar = n
}

// GetLocalVar 返回局部变量索引，对应 Java 的 getLocalVar()
func (d *Declarator) GetLocalVar() int {
	return d.localVar
}

// GetTag 返回标签，对应 Java 的 getTag()，这里固定返回 "decl"
func (d *Declarator) GetTag() string {
	return "decl"
}

// Accept 实现 Visitor 模式，
func (d *Declarator) Accept(v Visitor) error {
	return v.AtDeclarator(d)
}

// AstToClassName 将 ASTList 表示的类名转换为字符串，使用 sep 作为分隔符。
// 对应 Java 的 static String AstToClassName(ASTList name, char sep)
func AstToClassName(name *ASTList, sep rune) string {
	if name == nil {
		return ""
	}
	var sb strings.Builder
	AstToClassNameHelper(&sb, name, sep)
	return sb.String()
}

// AstToClassNameHelper 辅助函数，递归拼接类名部分。
func AstToClassNameHelper(sb *strings.Builder, name *ASTList, sep rune) {
	for name != nil {
		h := name.Head()
		switch v := h.(type) {
		case *Symbol:
			sb.WriteString(v.get())
		case *ASTList:
			AstToClassNameHelper(sb, v, sep)
		}
		namePtr := name.Tail()
		if namePtr == nil {
			return
		}
		name = namePtr
		sb.WriteRune(sep)
	}
}

// ClassDecl class declaration of ast node
type ClassDecl struct {
	BaseNode
	ClassName string
}

func (d *ClassDecl) setClassName(name string) {
	d.ClassName = name
}

// MethodDecl method declaration. ast tree
type MethodDecl struct {
	*ASTList
}

func NewMethodDecl(_head Node, _tail *ASTList) *MethodDecl {
	return &MethodDecl{
		ASTList: NewASTList(_head, _tail),
	}
}

func (d *MethodDecl) IsConstructor() bool {
	return false
}

func (d *MethodDecl) GetModifiers() *ASTList {
	return d.GetLeft().(*ASTList)
}

func (d *MethodDecl) GetReturn() *Declarator {
	return d.Tail().Head().(*Declarator)
}

func (d *MethodDecl) GetParams() *ASTList {
	return d.Sublist(2).(*ASTList).Head().(*ASTList)
}

func (d *MethodDecl) GetThrows() *ASTList {
	return d.Sublist(3).(*ASTList).Head().(*ASTList)
}

func (d *MethodDecl) GetBody() *Statement {
	return d.Sublist(4).(*ASTList).Head().(*Statement)
}

func (d *MethodDecl) Accept(v Visitor) error {
	return v.AtMethodDecl(d)
}

type FieldDecl struct {
	*ASTList
}

func NewFieldDecl(_head Node, _tail *ASTList) *FieldDecl {
	return &FieldDecl{
		ASTList: NewASTList(_head, _tail),
	}

}

func (d *FieldDecl) GetModifiers() *ASTList {
	return d.GetLeft().(*ASTList)
}

func (d *FieldDecl) GetDeclarator() *Declarator {
	return d.Tail().Head().(*Declarator)
}

func (d *FieldDecl) GetInit() Node {
	return d.Sublist(2).(*ASTList).Head()
}

func (d *FieldDecl) Accept(v Visitor) error {
	return v.AtFieldDecl(d)
}
