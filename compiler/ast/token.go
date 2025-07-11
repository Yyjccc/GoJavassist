package ast

import (
	"fmt"
	"strings"
)

type NodeType int

var TokenMap = make(map[TokenID]string)

var KeywordsMap = map[string]TokenID{

	"+":  Plus,
	"-":  Minus,
	"*":  Multiply,
	"/":  Divide,
	"%":  Mod,
	"|":  Or,
	"^":  Xor,
	"&":  And,
	"<<": LSHIFT,
	">>": RSHIFT,

	"class":     Class,
	"interface": Interface,
	"enum":      Enum,

	"import":     Import,
	"package":    Package,
	"extends":    Extends,
	"implements": Implements,

	"public":    Public,
	"private":   Private,
	"protected": Protected,
	"static":    Static,
	"abstract":  Abstract,
	"final":     Final,
	"transient": Transient,
	"volatile":  Volatile,
	"strict":    Strict,
	"native":    Native,

	"synchronized": Synchronized,

	"int":     Int,
	"long":    Long,
	"float":   Float,
	"double":  Double,
	"short":   Short,
	"byte":    Byte,
	"char":    Char,
	"boolean": Boolean,
	"void":    Void,

	"const": CONST,

	"if":       IF,
	"else":     Else,
	"do":       Do,
	"while":    While,
	"switch":   Switch,
	"case":     Case,
	"default":  Default,
	"break":    Break,
	"continue": Continue,
	"return":   Return,
	"goto":     Goto,

	"try":        Try,
	"catch":      Catch,
	"finally":    Finally,
	"throw":      Throw,
	"throws":     Throws,
	"instanceof": Instanceof,

	"new":   New,
	"super": Super,
	"this":  This,
	"true":  True,
	"false": False,
	"null":  Null,
}

func init() {
	for k, v := range KeywordsMap {
		TokenMap[v] = k
	}
}

type TokenID int

const (
	Empty TokenID = iota
	Class
	Method
	Interface
	Enum

	Import
	Package
	Extends
	Implements

	Public
	Private
	Protected
	Static
	Final
	Abstract
	Transient
	Volatile
	Strict
	Native

	Synchronized

	// 基本类型
	Int
	Long
	Float
	Double
	Short
	Byte
	Char
	Boolean

	Void

	CONST

	// 控制语句
	IF
	Else
	Do
	While
	For
	Switch
	Case
	Default
	Break
	Continue
	Goto
	Return

	Try
	Catch
	Finally
	Throw
	Throws
	Instanceof

	New
	Super
	This

	// 操作符
	Plus     // +
	Minus    // -
	Multiply // *
	Divide   // /
	Mod      // %
	Assign   // =
	LT       // <
	GT       // >
	Not      // !
	And      // &
	Or       // |
	BitNot   // ~
	Xor      // ^
	Question // ?
	Colon    // :

	NEQ        // !=
	MOD_E      // %=
	AND_E      // &=
	MUL_E      // *=
	PLUS_E     // +=
	MINUS_E    // -=
	DIV_E      // /=
	LE         // <=
	EQ         // ==
	GE         // >=
	EXOR_E     // ^=
	OR_E       // |=
	PLUSPLUS   // ++
	MINUSMINUS // --
	LSHIFT     // <<
	LSHIFT_E   // <<=
	RSHIFT     // >>
	RSHIFT_E   // >>=
	OROR       // ||
	ANDAND     // &&
	ARSHIFT    // >>>
	ARSHIFT_E  // >>>=

	Identifier
	CharConstant
	IntConstant
	LongConstant
	FloatConstant
	DoubleConstant
	StringL
	True
	False
	Null
	Call   // method call
	Array  // array access
	Member // static member access

	Expr  // expression statement
	Label // label statement
	Block // block statement
	Decl  // declaration statement

	Separator
	Dot // .
	BadToken
)

func (t TokenID) IsModifier() bool {
	return t >= Public && t <= Synchronized
}

func (t TokenID) IsOperator() bool {
	return t >= Plus && t <= ARSHIFT_E
}

func (t TokenID) IsBasicType() bool {
	return t >= Int && t <= Boolean
}

// 判断是否为赋值 操作符
func (t TokenID) IsAssignOp() bool {
	return t == Assign || t == MOD_E || t == AND_E || t == MUL_E || t == PLUS_E || t == MINUS_E || t == DIV_E || t == EXOR_E || t == OR_E || t == LSHIFT_E || t == RSHIFT_E || t == ARSHIFT_E
}

// 运算符优先级map
var binaryOpPrecedence = map[string]int{
	"%": 6, "&": 6, "*": 6, "+": 5, "-": 5, "/": 6, "<": 4, ">": 4,
	"<=": 4, ">=": 4, "==": 5, "!=": 5, "&&": 9, "||": 10, "^": 7, "|": 8,
	"<<": 3, ">>": 3, ">>>": 3,
}

// 获取运算符优先级
func GetOpPrecedence(op string) int {
	if precedence, exists := binaryOpPrecedence[op]; exists {
		return precedence
	}
	return 0 // 非二元运算符
}

func (t TokenID) String() string {
	return fmt.Sprintf("%d", int(t))
}

func (t TokenID) GetName() string {
	return strings.ToLower(TokenMap[t])
}

// 定义 AssignOps 数组，对应 Java 代码的 int AssignOps[]
var AssignOps = map[TokenID]TokenID{
	MOD_E:     '%',
	AND_E:     '&',
	MUL_E:     '*',
	PLUS_E:    '+',
	MINUS_E:   '-',
	DIV_E:     '/',
	EXOR_E:    '^',
	OR_E:      '|',
	LSHIFT_E:  LSHIFT,
	RSHIFT_E:  RSHIFT,
	ARSHIFT_E: ARSHIFT,
}

// binOp 对应 Java 代码中的 binOp 数组
var binOp = []TokenID{
	'+', '-', '*', '/', '%', '&', '|', '^', LSHIFT, RSHIFT, ARSHIFT,
}

// LookupBinOp 查找 binOp 数组中 `token` 的索引
func LookupBinOp(token TokenID) int {
	for k, v := range binOp {
		if v == token {
			return k
		}
	}
	return -1 // 没找到返回 -1
}

// 计算 token 位置
func GetBinOpIndex(op TokenID) int {
	// 先从 assignOps 取出对应的运算符
	token, exists := AssignOps[op]
	if !exists {
		return -1 // 若 op 不在 assignOps 里，直接返回 -1
	}

	// 在 binOp 中查找 token 的位置
	return LookupBinOp(token)
}
