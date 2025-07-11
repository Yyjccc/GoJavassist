package ast

import (
	"fmt"
	"reflect"
	"strings"
)

/*
		ast tree node abstract
		struct:
							   [Node]
	                           /    \
	                          /      \
						 Left/        \Right
*/
type Node interface {
	GetLeft() Node // 获取左子树
	SetLeft(n Node)
	GetRight() Node // 获取右子树
	SetRight(n Node)
	ToString() string
	// GetTag 返回类型标识（默认为类型名称）
	GetTag() string
	Accept(visitor Visitor) error
}

func GetIdentifier(node Node) string {
	if m, ok := node.(*MemberSymbol); ok {
		return m.Identifier
	}
	if s, ok := node.(*Symbol); ok {
		return s.Identifier
	}
	if v, ok := node.(*Variable); ok {
		return v.Identifier
	}
	return ""
}

// BaseNode : Node of empty implement
type BaseNode struct{}

func (b *BaseNode) ToString() string { return fmt.Sprintf("<%s>", b.GetTag()) }
func (b *BaseNode) GetTag() string {
	// 利用反射返回类型名称
	t := reflect.TypeOf(b)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (b *BaseNode) SetLeft(n Node) {}

func (b *BaseNode) GetLeft() Node { return nil }

func (b *BaseNode) GetRight() Node { return nil }

func (b *BaseNode) SetRight(n Node) {}

/*
	ASTList AST Node List 的Node 接口实现
	Left: current Node
	Right : next list ,must a ASTList

struct:

					   +-------------------------+
	                   |        ASTList          |
	                   +-------------------------+
	                   | Left: 当前节点的值 (Node) |
	                   | Right: 后继节点 (*ASTList)|
	                   +-------------------------+
	                            |
	                            v
	                   +-------------------------+
	                   |        ASTList          |
	                   +-------------------------+
	                   | Left: 当前节点的值 (Node)|
	                   | Right: 后继节点 (Node)   |
	                   +-------------------------+
	                            |
	                            v
	                           ...
*/
type ASTList struct {
	//BaseNode
	Left  Node // 链表当前节点的值
	Right Node // 链表后继节点
}

// NewASTList 创建一个 ASTList 对象，接收 head 和 tail 两个参数
func NewASTList(head Node, tail Node) *ASTList {
	return &ASTList{
		Left:  head,
		Right: tail,
	}
}

// NewASTListSingle 创建仅包含一个节点的 ASTList
func NewASTListSingle(head Node) *ASTList {
	return &ASTList{
		Left:  head,
		Right: nil,
	}
}

// MakeASTList 构造一个由三个 Node 组成的链表
func MakeASTList(e1, e2, e3 Node) *ASTList {
	return NewASTList(e1, NewASTList(e2, NewASTListSingle(e3)))
}

// 实现 Node 接口的方法

// ToString 返回 ASTList 的字符串表示，格式类似于：(<ASTList> elem1 elem2 …)
func (list *ASTList) ToString() string {
	var sb strings.Builder
	sb.WriteString("(<")
	sb.WriteString(list.GetTag())
	sb.WriteString(">")
	// 遍历链表中的每个节点

	for cur := Node(list); cur != nil; cur = cur.GetRight() {
		sb.WriteString(" ")
		if cur.GetLeft() == nil {
			sb.WriteString("<null>")
		} else {
			sb.WriteString(cur.GetLeft().ToString())
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (list *ASTList) Accept(visitor Visitor) error {
	return visitor.AtASTList(list)
}

// GetTag 返回类型标识
func (list *ASTList) GetTag() string {
	return "ASTList"
}

// GetLeft 返回当前节点的左子树
func (list *ASTList) GetLeft() Node {
	if list == nil {
		return nil
	}
	return list.Left
}

// GetRight 返回当前节点的右子树（链表中下一个元素）
func (list *ASTList) GetRight() Node {
	if list == nil {
		return nil
	}
	if list.Right != nil {
		return list.Right
	}
	return nil
}

// SetLeft 设置当前节点的左子树
func (list *ASTList) SetLeft(t Node) {
	list.Left = t
}

// SetRight 要求传入的 Node 必须是 *ASTList 类型，否则会 panic
func (list *ASTList) SetRight(t Node) {
	list.Right = t
	//if tail, ok := t.(*ASTList); ok {
	//	list.Right = tail
	//} else {
	//	panic("SetRight: argument is not *ASTList")
	//}
}

// 以下为 ASTList 独有的方法

// Head 返回当前链表节点的值（等同于 GetLeft）
func (list *ASTList) Head() Node {
	return list.Left
}

// SetHead 设置当前节点的值（等同于 SetLeft）
func (list *ASTList) SetHead(t Node) {
	list.Left = t
}

// Tail 返回当前节点的后继链表
func (list *ASTList) Tail() *ASTList {
	if list.Right == nil {
		return nil
	}
	return list.Right.(*ASTList)
}

// SetTail 设置当前节点的后继链表
func (list *ASTList) SetTail(tail Node) {
	list.Right = tail
}

// Length 返回链表的长度
func (list *ASTList) Length() int {
	if list == nil {
		return 0
	}
	return ASTListLength(list)
}

// ASTListLength 计算从给定链表起始的长度
func ASTListLength(list *ASTList) int {
	n := 0
	for cur := Node(list); cur != nil; cur = cur.GetRight() {
		n++
	}
	return n
}

// Sublist 返回从当前节点开始，跳过 nth 个节点后的子链表
func (list *ASTList) Sublist(nth int) Node {
	cur := Node(list)
	for nth > 0 && cur != nil {
		cur = cur.GetRight()
		nth--
	}
	return cur
}

// Subst 在链表中查找第一个等于 oldObj 的节点，将其替换为 newObj，成功则返回 true
func (list *ASTList) Subst(newObj, oldObj Node) bool {
	for cur := Node(list); cur != nil; cur = cur.GetRight() {
		// 这里采用指针相等或者根据 ToString() 判断（实际比较方式依据需求调整）
		if cur.GetLeft() == oldObj {
			cur.SetLeft(newObj)
			return true
		}
	}
	return false
}

// AppendASTList 将一个 Node 元素追加到链表 a 的末尾
func AppendASTList(a Node, b Node) Node {
	return ConcatASTList(a, NewASTListSingle(b))
}

// ConcatASTList 连接两个 ASTList 链表
func ConcatASTList(a, b Node) Node {
	if isNil(a) {
		return b
	}
	cur := a
	for !isNil(cur.GetRight()) {
		cur = cur.GetRight()
	}
	cur.SetRight(b)
	return a
}

func isNil(x interface{}) bool {
	return x == nil || (reflect.ValueOf(x).Kind() == reflect.Ptr && reflect.ValueOf(x).IsNil())
}
