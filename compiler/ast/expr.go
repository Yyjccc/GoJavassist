package ast

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
	"strconv"
)

// Expression 表达式
/*
  ----------------------------------------------------------
      Expression (操作符: OperatorId)
                  │
               [ASTList]
              ┌─────┴─────┐
     (Operand1)         [ASTList] ——> (Operand2)
          │                    │
    ASTList.Left           ASTList.Right.Left

  说明：
  - Expression 中嵌入的 ASTList 用于保存操作数：
      • ASTList.Left 存储第一个操作数（左子树）。
      • ASTList.Right 是一个 ASTList 节点，其 Left 存储第二个操作数。
  - OperatorId 表示当前表达式的运算符标识。
	Operand 可能是 某个具体的变量，操作数或者表达式（子树）
*/
type Expression struct {
	*ASTList
	OperatorId TokenID
}

// NewExpression 构造函数，创建 Expression 实例
func NewExpression(op TokenID, head Node, tail *ASTList) *Expression {
	return &Expression{
		ASTList:    NewASTList(head, tail),
		OperatorId: op,
	}
}

// NewExpressionSingle 创建单操作数的 Expression
func NewExpressionSingle(op TokenID, head Node) *Expression {
	return &Expression{
		ASTList:    NewASTListSingle(head),
		OperatorId: op,
	}
}

// MakeExpression 生成二元表达式
func MakeExpression(op TokenID, operand1, operand2 Node) *Expression {
	return NewExpression(op, operand1, NewASTListSingle(operand2))
}

// MakeExpressionSingle 生成一元表达式
func MakeExpressionSingle(op TokenID, operand1 Node) *Expression {
	return NewExpressionSingle(op, operand1)
}

// GetOperator 获取运算符 ID
func (e *Expression) GetOperator() TokenID {
	return e.OperatorId
}

// SetOperator 设置运算符 ID
func (e *Expression) SetOperator(op TokenID) {
	e.OperatorId = op
}

// Oprand1 获取第一个操作数（左子树）
func (e *Expression) Oprand1() Node {
	return e.GetLeft()
}

// SetOprand1 设置第一个操作数
func (e *Expression) SetOprand1(Expression Node) {
	e.SetLeft(Expression)
}

// Oprand2 获取第二个操作数（右子树的左节点）
func (e *Expression) Oprand2() Node {
	if right := e.GetRight(); right != nil {
		return right.GetLeft()
	}
	return nil
}

// SetOprand2 设置第二个操作数
func (e *Expression) SetOprand2(Expression Node) {
	if right := e.GetRight(); right != nil {
		right.SetLeft(Expression)
	}
}

// Accept 访问者模式（假设 Visitor 和 CompileError 在其他地方定义）
func (e *Expression) Accept(v Visitor) error {
	return v.AtExpression(e)
}

// GetName 返回运算符的名称
func (e *Expression) GetName() string {
	id := e.OperatorId
	return TokenMap[id]
}

// GetTag 返回标签
func (e *Expression) GetTag() string {
	return "op:" //+ e.GetName()
}

// CondExpr 条件表达式
/*
	2. CondExpr 条件表达式节点
	  ----------------------------------------------------------
				  CondExpr
					  │
				   [ASTList]
				  ┌─────┴─────┐
			 [Cond]         [ASTList]
							   ┌─────┴─────┐
						  [ThenExpr]    [ASTList]
											│
									   [ElseExpr]

	  说明：
	  - CondExpr 的 ASTList 链表构造了三部分：
		  • Head：条件表达式（Cond）。
		  • Tail.Head：then 分支表达式（ThenExpr）。
		  • Tail.Tail.Head：else 分支表达式（ElseExpr）。
	  - 这种链表结构使得条件表达式的三个组成部分顺序排列。
*/
type CondExpr struct {
	*ASTList
}

func NewCondExpr(cond Node, thenp Node, elsep Node) *CondExpr {
	return &CondExpr{
		ASTList: NewASTList(cond, NewASTList(thenp, NewASTListSingle(elsep))),
	}
}

func (e *CondExpr) CondExpr() Node {
	return e.Head()
}

func (e *CondExpr) SetCond(node Node) {
	e.SetHead(node)
}

func (e *CondExpr) ThenExpr() Node {
	return e.Tail().Head()
}

func (e *CondExpr) SetThenExpr(node Node) {
	e.Tail().SetHead(node)
}

func (e *CondExpr) ElseExpr() Node {
	return e.Tail().Tail().Head()
}
func (e *CondExpr) SetElseExpr(node Node) {
	e.Tail().Tail().SetHead(node)
}

func (e *CondExpr) Accept(v Visitor) error {
	return v.AtCondExpr(e)
}

// CastExpr 类型转换表达式
/*
	3. CastExpr 类型转换表达式节点
	  ----------------------------------------------------------
		arrayDim 数组类型维度 <------[CastExpr] ----> castType 转化类型
										  │
									   [ASTList]
							（转化类型）┌─────┴─────┐ （子表达式）
							   [ClassName?]      [ASTList]
													  │
												  [Operand]

	  说明：
	  - CastExpr 的 ASTList：
		  • Left 部分可能存储类名（ClassName），用于类转换；若为基本类型转换则为 nil。
		  • Right 部分的 Left 存储被转换的表达式（Operand）。
	  - 附加字段 castType 和 arrayDim 用于标识转换类型和数组维度。
*/
type CastExpr struct {
	*ASTList
	castType TokenID
	arrayDim int
}

func NewCastExprWithClass(className *ASTList, dim int, expr Node) *CastExpr {
	return &CastExpr{
		ASTList:  NewASTList(className, NewASTListSingle(expr)),
		castType: Class,
		arrayDim: dim,
	}
}

func NewCastExprWithType(types TokenID, dim int, expr Node) *CastExpr {
	return &CastExpr{
		ASTList:  NewASTList(nil, NewASTListSingle(expr)),
		castType: types,
		arrayDim: dim,
	}
}

func (c *CastExpr) GetType() TokenID {
	if c == nil {
		return Empty
	}
	return c.castType
}

func (c *CastExpr) GetArrayDim() int {
	if c == nil {
		return 0
	}
	return c.arrayDim
}

func (e *CastExpr) Accept(v Visitor) error {
	return v.AtCastExpr(e)
}

// GetClassName allow nil ,when nil ,is build-in type
func (e *CastExpr) GetClassName() *ASTList {
	if isNil(e.GetLeft()) {
		return nil
	}
	return e.GetLeft().(*ASTList)
}
func (e *CastExpr) GetOprand() Node {
	return e.GetRight().GetLeft()
}
func (e *CastExpr) SetOprand(t Node) {
	e.GetRight().SetLeft(t)
}

func (e *CastExpr) GetTag() string {
	return fmt.Sprintf("cast:%d:%d", e.castType, e.arrayDim)
}

// NewExpr new 表达式
/*
	  4. NewExpr new表达式节点（对象或数组的创建）
	  ----------------------------------------------------------
				   NewExpr (标记: new / new[])
					  │
				   [ASTList]
				  ┌─────┴─────┐
		 [ClassName 或 nil]   [ASTList]
								 ┌─────┴─────┐
							[Arguments / 数组大小]   (可选: 其他初始化信息)
									  │
							   (若存在 ASTList.Right.Tail, Head 为 Initializer)

	  说明：
	  - NewExpr.newArray 标识是否为数组创建。
	  - 当创建对象时，ASTList.Left 存储对象类型（ClassName），ASTList.Right.Left 存储构造参数。
	  - 当创建数组时，ASTList.Left 可能为 nil，ASTList.Right.Left 存储数组大小，
		且如果有初始化，则附加一个链表节点存储数组初始化器（Initializer）。
*/
type NewExpr struct {
	*ASTList
	newArray  bool
	arrayType TokenID
}

func NewNewExprWithClass(className *ASTList, args *ASTList) *NewExpr {
	return &NewExpr{
		ASTList:   NewASTList(className, NewASTListSingle(args)),
		newArray:  false,
		arrayType: Class,
	}
}

func NewNewExprWithType(types TokenID, arraySize *ASTList, init *ArrayInit) *NewExpr {
	expr := &NewExpr{
		ASTList:   NewASTList(nil, NewASTListSingle(arraySize)),
		newArray:  true,
		arrayType: types,
	}
	if init != nil {
		AppendASTList(expr.ASTList, init)
	}
	return expr
}

func MakeObjectArray(className, arraySize *ASTList, init *ArrayInit) *NewExpr {
	expr := NewNewExprWithClass(className, arraySize)
	expr.newArray = true
	if init != nil {
		AppendASTList(expr.ASTList, init)
	}
	return expr
}

func (e *NewExpr) IsArray() bool {
	return e.newArray
}

func (e *NewExpr) GetArrayType() TokenID {
	return e.arrayType
}

func (e *NewExpr) GetClassName() *ASTList {
	return e.GetLeft().(*ASTList)
}

func (e *NewExpr) GetArguments() *ASTList {
	return e.GetRight().GetLeft().(*ASTList)
}

func (e *NewExpr) GetArraySize() *ASTList {
	return e.GetArguments()
}

func (e *NewExpr) GetInitializer() *ArrayInit {
	node := e.GetRight().GetRight()
	if node == nil {
		return nil
	}
	return node.GetLeft().(*ArrayInit)
}

func (e *NewExpr) Accept(v Visitor) error {
	return v.AtNewExpr(e)
}

func (e *NewExpr) GetTag() string {
	if e.newArray {
		return "new[]"
	}
	return "new"
}

// ArrayInit 数组初始化节点
/*
	eg； new String[]{"a","b"} // ArrayInit 为{}中间的部分
	5. ArrayInit 数组初始化节点
	  ----------------------------------------------------------
				  ArrayInit
					  │
				   [ASTList]
				  ┌─────┴─────┐
			[第1个元素]    ──►  (后续元素通过链表结构依次连接)

	  说明：
	  - ArrayInit 通过 ASTList 以链表形式存储数组各个元素的初始化值，
		Head 存储当前元素，Tail 指向下一个初始化元素。
*/
type ArrayInit struct {
	*ASTList
}

func NewArrayInit(first Node) *ArrayInit {
	return &ArrayInit{
		ASTList: NewASTListSingle(first),
	}
}

func (a *ArrayInit) getTag() string {
	return "array"
}

func (a *ArrayInit) Size() int {
	s := a.ASTList.Length()
	if s == 1 && isNil(a.GetLeft()) {
		return 0
	}
	return s
}

func (a *ArrayInit) Accept(v Visitor) error {
	return v.AtArrayInit(a)
}

// AssignExpr 赋值表达式
/*
	----------------------------------------------------------------
	AssignExpr 赋值表达式结构示意图

				AssignExpr (运算符: OperatorId 赋值)
							 │
						 [ASTList]
						┌─────┴─────┐
			 ┌─────────┴─────────┐
			 │                   │
		(左操作数)         [ASTList 节点]
							   │
							   │  ┌───────────────────────┐
							   │  │ ASTList.Left: 右操作数 │
							   │  └───────────────────────┘
			 │
	  说明：
	  - AssignExpr 继承自 Expression，其内部嵌入的 ASTList 用于组织赋值表达式的操作数。
	  - ASTList.Left 存储赋值操作左侧的变量或目标。
	  - ASTList.Right（通过链表的形式）存储右侧的表达式。
	----------------------------------------------------------------
*/

type AssignExpr struct {
	*Expression
}

func NewAssignExpr(op TokenID, _head Node, _tail *ASTList) *AssignExpr {
	return &AssignExpr{
		Expression: NewExpression(op, _head, _tail),
	}
}

func MakeAssign(op TokenID, oprand1, oprand2 Node) *AssignExpr {
	return NewAssignExpr(op, oprand1, NewASTListSingle(oprand2))
}

func (a *AssignExpr) Accept(v Visitor) error {
	return v.AtAssignExpr(a)
}

// BinExpr 二元表达式
/*
	----------------------------------------------------------------
	BinExpr 二元运算表达式结构示意图

				BinExpr (运算符: OperatorId 二元运算)
							 │
						 [ASTList]
						┌─────┴─────┐
			 ┌─────────┴─────────┐
			 │                   │
		(左操作数)         [ASTList 节点]
							   │
							   │  ┌───────────────────────┐
							   │  │ ASTList.Left: 右操作数 │
							   │  └───────────────────────┘
			 │
	  说明：
	  - BinExpr 也是基于 Expression 的扩展，用于表示二元运算。
	  - 内部的 ASTList 采用链表形式：
		  • ASTList.Left 保存左操作数。
		  • ASTList.Right 的 Head（即 Left）保存右操作数。
	----------------------------------------------------------------
*/

type BinExpr struct {
	*Expression
}

func NewBinExpr(op TokenID, _head Node, _tail *ASTList) *BinExpr {
	return &BinExpr{
		Expression: NewExpression(op, _head, _tail),
	}
}

func MakeBinExpr(op TokenID, oprand1, oprand2 Node) *BinExpr {
	return NewBinExpr(op, oprand1, NewASTListSingle(oprand2))
}

func (a *BinExpr) Accept(v Visitor) error {
	return v.AtBinExpr(a)
}

// CallExpr 方法调用表达式
/*
	----------------------------------------------------------------
		CallExpr 方法调用表达式结构示意图

					  CallExpr (运算符: Call)
								 │
							 [ASTList]
							┌─────┴─────┐
				 ┌─────────┴─────────┐
				 │                   │
			(调用目标)         [ASTList 节点]
								   │
								   │  ┌────────────────────────────┐
								   │  │ ASTList.Left: 参数列表/实参 │
								   │  └────────────────────────────┘
				 │
		  附加说明：
		  - CallExpr 继承自 Expression，利用 ASTList 表示方法调用的目标和参数。
		  - ASTList.Left 存储调用的目标（如函数或对象）。
		  - ASTList.Right 的 Head 存储参数列表（可能为链表结构，逐一表示各参数）。
		  - 除 ASTList 外，还包含一个 method 字段，用于存放解析后的方法信息。
	----------------------------------------------------------------
*/
type CallExpr struct {
	*Expression
	method *reflect.CtMethod
}

func NewCallExpr(_head Node, _tail *ASTList) *CallExpr {
	return &CallExpr{
		Expression: NewExpression(Call, _head, _tail),
	}
}

func (c *CallExpr) SetMethod(method *reflect.CtMethod) {
	c.method = method
}
func (c *CallExpr) GetMethod() *reflect.CtMethod {
	return c.method
}

func MakeCall(target, args Node) *CallExpr {
	return NewCallExpr(target, NewASTListSingle(args))
}

func (c *CallExpr) Accept(v Visitor) error {
	return v.AtCallExpr(c)
}

// InstanceOfExpr InstanceOf 表达式
/*
	----------------------------------------------------------------
		InstanceOfExpr instanceof 表达式结构示意图

				   InstanceOfExpr (继承自 CastExpr)
						   castType, arrayDim
								 │
							[CastExpr 的 ASTList]
							┌─────┴─────┐
				 ┌─────────┴─────────┐
				 │                   │
		 (类型标识/类名)        [ASTList 节点]
								   │
								   │  ┌─────────────────────────┐
								   │  │ ASTList.Left: 待判断表达式 │
								   │  └─────────────────────────┘
				 │
		  说明：
		  - InstanceOfExpr 用于实现 instanceof 判断，继承自 CastExpr。
		  - CastExpr 的 ASTList：
			  • Left 部分存储用于类型转换的类名或类型标识。
			  • Right 的 Head 保存待判断的表达式。
		  - 同时，castType 和 arrayDim 参数用于记录转换的具体类型信息和数组维度。
	----------------------------------------------------------------
*/

type InstanceOfExpr struct {
	*CastExpr
}

func (i *InstanceOfExpr) GetTag() string {
	return "instanceof:" + strconv.Itoa(int(i.castType)) + ":" + strconv.Itoa(i.arrayDim)
}

func (i *InstanceOfExpr) Accept(v Visitor) error {
	return v.AtInstanceOfExpr(i)
}

func NewInstanceOfExprWithClass(className *ASTList, dim int, expr Node) *InstanceOfExpr {
	return &InstanceOfExpr{
		CastExpr: NewCastExprWithClass(className, dim, expr),
	}
}

func NewInstanceOfExprWithType(typeId TokenID, dim int, expr Node) *InstanceOfExpr {
	return &InstanceOfExpr{
		CastExpr: NewCastExprWithType(typeId, dim, expr),
	}
}
