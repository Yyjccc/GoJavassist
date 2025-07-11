package compiler

import (
	"GoJavassist/compiler/ast"
	"strings"
	"unicode"
)

type TokenType string

const errDisplayLength = 6

const (
	// KEYWORD 关键字
	KEYWORD TokenType = "KEYWORD"
	// IDENTIFIER 标识符
	IDENTIFIER TokenType = "IDENTIFIER"
	// NUMBER 数字
	NUMBER TokenType = "NUMBER"
	// STRING 字符串
	STRING TokenType = "STRING"
	// CHAR 字符
	CHAR TokenType = "CHAR"
	// OPERATOR 操作符
	OPERATOR TokenType = "OPERATOR"
	// SEPARATOR 界符
	SEPARATOR TokenType = "SEPARATOR"
	// COMMENT 注释
	COMMENT TokenType = "COMMENT"

	EOF TokenType = "EOF"
)

// 关键字列表
var keywords = ast.KeywordsMap

type Token struct {
	Type   TokenType
	Value  string
	ID     ast.TokenID
	Line   int
	Column int
}

var EOFToken = Token{
	Type:   EOF,
	Value:  "",
	ID:     -1,
	Line:   -1,
	Column: -1,
}

// Lexer 词法解析器
type Lexer struct {
	input  string //输入缓冲区
	ch     rune   //当前字符
	Max    int    //最终长度
	pos    int    //指针
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		Max:   len(input),
		ch:    0xA0,
		pos:   -1,
	}
	//if len(input) == 0 {
	//	l.ch = 0
	//}
	//l.ch = rune(input[0])
	//l.readChar()

	return l

}

// peekChar 查看下一个字符,但不消耗
func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0 // 文件结尾
	}
	return rune(l.input[l.pos])
}

// readChar 读取下一个字符
func (l *Lexer) readChar() {
	if len(l.input) == 0 {
		l.ch = 0
		return
	}
	if l.pos == l.Max-1 {
		l.ch = 0
		return
	}
	l.ch = rune(l.input[l.pos+1])
	l.pos++
	l.column++

	// 处理换行
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) NextToken() Token {
	//跳过空白字符
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
	startLine, startColumn := l.line, l.column
	switch {
	// 标识符字母开头
	case unicode.IsLetter(l.ch) || l.ch == '_':
		ident := l.readIdentifier()
		if id, ok := keywords[ident]; ok {
			return Token{Type: KEYWORD, ID: id, Value: ident, Line: startLine, Column: startColumn}
		}
		return Token{Type: IDENTIFIER, ID: ast.Identifier, Value: ident, Line: startLine, Column: startColumn}
	case unicode.IsDigit(l.ch):
		return Token{Type: NUMBER, ID: ast.IntConstant, Value: l.readNumber(), Line: startLine, Column: startColumn}
	case l.ch == '"':
		return Token{Type: STRING, ID: ast.StringL, Value: l.readString(), Line: startLine, Column: startColumn}
	case l.ch == '\'':
		l.readChar()
		return Token{Type: CHAR, ID: ast.CharConstant, Value: string(l.ch), Line: startLine, Column: startColumn}
	case l.ch == '/':
		return l.readCommentOrOperator(startLine, startColumn)
	case strings.ContainsRune("+-*/=<>!&|?:", l.ch):
		return l.readOperator(startLine, startColumn)
	case strings.ContainsRune("(){};,[].", l.ch):
		value := string(l.ch)
		l.readChar()
		return Token{Type: SEPARATOR, ID: ast.Separator, Value: value, Line: startLine, Column: startColumn}
	case l.ch == 0:
		l.pos = l.Max
		return EOFToken
	default:
		return Token{Type: EOF, Value: "not recognized : " + string(l.ch), Line: startLine, Column: startColumn}
	}
}

// readIdentifier 读取标识符
func (l *Lexer) readIdentifier() string {
	start := l.pos
	if start < 0 {
		start = 0
	}
	// 标识符 正规式
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' || l.ch == '$' {
		l.readChar()
	}
	return l.input[start:l.pos]
}

// 读取数字
func (l *Lexer) readNumber() string {
	start := l.pos
	for unicode.IsDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[start+1 : l.pos-1]
}

// 读取字符串
func (l *Lexer) readString() string {
	// 跳过开头的 "
	l.readChar()
	var sb strings.Builder

	// 当未遇到结束引号且未到文件尾时
	for l.ch != '"' && l.ch != 0 {
		// 处理转义字符
		if l.ch == '\\' {
			l.readChar() // 跳过反斜杠
			switch l.ch {
			case '"':
				sb.WriteRune('"')
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			// 根据需要可继续添加其它转义序列
			default:
				// 未识别的转义字符直接加入
				sb.WriteRune(l.ch)
			}
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}
	// 跳过结束的 "
	l.readChar()
	return sb.String()
}

// readCommentOrOperator 处理 `//` 单行注释 和 `/* ... */` 多行注释
func (l *Lexer) readCommentOrOperator(line, column int) Token {
	if l.peekChar() == '/' { // 处理 `//`
		l.readChar() // 跳过 `/`
		start := l.pos
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		return Token{Type: COMMENT, Value: l.input[start:l.pos], Line: line, Column: column}
	}

	if l.peekChar() == '*' { // 处理 `/* ... */`
		l.readChar() // 跳过 `*`
		start := l.pos
		for !(l.ch == '*' && l.peekChar() == '/') {
			if l.ch == 0 {
				return Token{Type: EOF, Value: "Unterminated block comment", Line: line, Column: column}
			}
			l.readChar()
		}
		l.readChar() // 跳过 `*`
		l.readChar() // 跳过 `/`
		return Token{Type: COMMENT, Value: l.input[start : l.pos-2], Line: line, Column: column}
	}
	l.readChar()
	// 如果不是注释，返回 `/` 作为运算符
	return Token{Type: OPERATOR, Value: "/", Line: line, Column: column}
}

// HasNextToken 是否还有 Token
func (l *Lexer) HasNextToken() bool {
	if l.pos == -1 {
		return len(l.input) > 0
	}
	for l.pos < l.Max-1 && unicode.IsSpace(l.ch) {
		l.readChar()
	}

	return l.pos < l.Max
}

// operatorMapping 定义所有操作符与对应 TokenID 的映射
var operatorMapping = map[string]ast.TokenID{
	// 两个字符或三个字符的操作符
	"!=":   ast.NEQ,
	"%=":   ast.MOD_E,
	"&=":   ast.AND_E,
	"*=":   ast.MUL_E,
	"+=":   ast.PLUS_E,
	"-=":   ast.MINUS_E,
	"/=":   ast.DIV_E,
	"<=":   ast.LE,
	"==":   ast.EQ,
	">=":   ast.GE,
	"^=":   ast.EXOR_E,
	"|=":   ast.OR_E,
	"++":   ast.PLUSPLUS,
	"--":   ast.MINUSMINUS,
	"<<":   ast.LSHIFT,
	"<<=":  ast.LSHIFT_E,
	">>":   ast.RSHIFT,
	">>=":  ast.RSHIFT_E,
	"||":   ast.OROR,
	"&&":   ast.ANDAND,
	">>>":  ast.ARSHIFT,
	">>>=": ast.ARSHIFT_E,
	// 单字符操作符
	"+": ast.Plus,
	"-": ast.Minus,
	"*": ast.Multiply,
	"/": ast.Divide,
	"=": ast.Assign,
	"<": ast.LT,
	">": ast.GT,
	"!": ast.Not,
	"&": ast.And,
	"|": ast.Or,
	"?": ast.Question,
	":": ast.Colon,
}

// isOperatorPrefix 检查 candidate 是否为某个合法操作符的前缀
func isOperatorPrefix(candidate string) bool {
	for op := range operatorMapping {
		if len(candidate) <= len(op) && op[:len(candidate)] == candidate {
			return true
		}
	}
	return false
}

// readOperator 采用贪心匹配方式读取操作符，返回最长匹配的 Token
func (l *Lexer) readOperator(startLine, startColumn int) Token {
	// 初始读取第一个字符
	op := string(l.ch)
	l.readChar() // 消耗当前字符
	// 如果第一个字符本身构成一个完整的操作符，记录下来
	validOp := op
	if _, ok := operatorMapping[validOp]; !ok {
		// 如果映射表中未定义单字符操作符，可直接设为 BadToken
		validOp = ""
	}
	// 尝试扩展操作符，最多扩展 3 个字符以应对 ">>>=" 等情况
	for i := 0; i < 3; i++ {
		// 如果已经读到文件末尾，则退出循环
		if l.ch == 0 {
			break
		}
		candidate := op + string(l.ch)
		if isOperatorPrefix(candidate) {
			op = candidate
			// 如果 candidate 刚好是一个完整的操作符，则记录下来
			if _, exists := operatorMapping[op]; exists {
				validOp = op
			}
			l.readChar()
		} else {
			break
		}
	}
	// 若 validOp 为空，则说明当前操作符未识别
	tokenID, ok := operatorMapping[validOp]
	if !ok {
		tokenID = ast.BadToken
	}
	return Token{
		Type:   OPERATOR,
		ID:     tokenID,
		Value:  validOp,
		Line:   startLine,
		Column: startColumn,
	}
}

func (l *Lexer) getTextAround() string {
	begin := l.pos - errDisplayLength
	if begin < 0 {
		begin = 0
	}

	end := l.pos + errDisplayLength
	if end > len(l.input) {
		end = len(l.input)
	}
	return l.input[begin:end]
}
