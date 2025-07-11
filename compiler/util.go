package compiler

import (
	"fmt"
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
	r "reflect"
	"strings"
)

const (
	P_DOUBLE = 0
	P_FLOAT  = 1
	P_LONG   = 2
	P_INT    = 3
	P_OTHER  = -1
)

// toJvmTypeName tokenID convert jvm class type
func toJvmTypeName(typeId ast.TokenID, dim int) string {
	var c byte = 'I'
	switch typeId {
	case ast.Boolean:
		c = 'Z'
		break
	case ast.Byte:
		c = 'B'
		break
	case ast.Char:
		c = 'C'
	case ast.Short:
		c = 'S'
	case ast.Int:
		c = 'I'
	case ast.Long:
		c = 'J'
	case ast.Float:
		c = 'F'
	case ast.Double:
		c = 'D'
	case ast.Void:
		c = 'V'
	default:
	}

	var sbuf string
	for i := 0; i < dim; i++ {
		sbuf += "["
	}

	sbuf += string(c)
	return sbuf
}

func toJvmArrayName(name string, dim int) string {
	if name == "" {
		return ""
	}

	if dim == 0 {
		return name
	}

	var sbuf strings.Builder
	for i := 0; i < dim; i++ {
		sbuf.WriteByte('[')
	}

	sbuf.WriteByte('L')
	sbuf.WriteString(name)
	sbuf.WriteByte(';')

	return sbuf.String()
}

// className as a ast tree convert jvm class name
func toClassName(node ast.Node) (string, error) {
	var sbuf strings.Builder
	err := toClassNameRecursive(node, &sbuf)
	if err != nil {
		return "", err
	}
	return sbuf.String(), nil
}

func toClassNameRecursive(node ast.Node, sbuf *strings.Builder) error {
	if symbol, ok := node.(*ast.Symbol); ok {
		sbuf.WriteString(symbol.Identifier)
		return nil
	}
	//equal *ast.Symbol
	if mem, ok := node.(*ast.MemberSymbol); ok {
		sbuf.WriteString(mem.Identifier)
		return nil

	}

	if expr, ok := node.(*ast.Expression); ok {
		if expr.GetOperator() == ast.Dot {
			err := toClassNameRecursive(expr.Oprand1(), sbuf)
			if err != nil {
				return err
			}
			sbuf.WriteString(".")
			err = toClassNameRecursive(expr.Oprand2(), sbuf)
			return err
		}
	}
	return fmt.Errorf("bad static member access.")
}

func isNil(x interface{}) bool {
	return x == nil || (r.ValueOf(x).Kind() == r.Ptr && r.ValueOf(x).IsNil())
}

func needsSuperCall(body *ast.Statement) bool {
	if body.GetOperator() == ast.Block {
		body = body.GetASTList().Head().(*ast.Statement)
	}
	if body != nil && body.GetOperator() == ast.Expr {
		expr := body.GetASTList().Head()
		if !isNil(expr) {
			if express, ok := expr.(*ast.Expression); ok {
				if express.GetOperator() == ast.Call {
					target := express.Head()
					if k, ok := target.(*ast.Keyword); ok {
						token := k.TokenID
						return token != ast.This && token != ast.Super
					}
				}
			}
		}
	}

	return true
}

func is2word(token ast.TokenID, dim int) bool {
	return dim == 0 && (token == ast.Double || token == ast.Long)
}

func isRefType(id ast.TokenID) bool {
	return id == ast.Class || id == ast.Null
}

func typePrecedence(id ast.TokenID) int {
	if id == ast.Double {
		return P_DOUBLE
	} else if id == ast.Float {
		return P_FLOAT
	} else if id == ast.Long {
		return P_LONG
	} else if isRefType(id) {
		return P_OTHER
	} else {
		return P_INT
	}
}

func getArrayReadOp(id ast.TokenID, dim int) classfile.OpCode {
	if dim > 0 {
		return classfile.OpAALoad
	}
	switch id {
	case ast.Double:
		return classfile.OpDALoad
	case ast.Float:
		return classfile.OpFALoad
	case ast.Long:
		return classfile.OpLALoad
	case ast.Int:
		return classfile.OpIALoad
	case ast.Short:
		return classfile.OpSALoad
	case ast.CharConstant:
		return classfile.OpCALoad
	case ast.Boolean, ast.Byte:
		return classfile.OpBALoad
	default:
		return classfile.OpAALoad
	}
}

func invalidDim(srcType, destType ast.TokenID, srcDim, destDim int, srcClass, destClass string, isCast bool) bool {
	if srcType != destType {
		if srcType == ast.Null {
			return false
		}

		if destDim == 0 && destType == ast.Class && destClass == jvmJavaLangObject {
			return false
		}
		if isCast && srcDim == 0 && srcType == ast.Class && srcClass == jvmJavaLangObject {
			return false
		}
		return true
	}
	return false
}

func getArrayWriteOp(aType ast.TokenID, dim int) classfile.OpCode {
	if dim > 0 {
		return classfile.OpAAStore
	}
	switch aType {
	case ast.Double:
		return classfile.OpDAStore
	case ast.Float:
		return classfile.OpFAStore
	case ast.Long:
		return classfile.OpLAStore
	case ast.Int:
		return classfile.OpIAStore
	case ast.Short:
		return classfile.OpSAStore
	case ast.CharConstant:
		return classfile.OpCAStore
	case ast.Boolean, ast.Byte:
		return classfile.OpBAStore
	default:
		return classfile.OpAAStore
	}
}

// TODO
func isFromSameDeclaringClass(class *reflect.CtClass, class2 *reflect.CtClass) bool {
	return class2 == class
}

func descToType(c rune) ast.TokenID {
	switch c {
	case 'I':
		return ast.Int
	case 'J':
		return ast.Long
	case 'D':
		return ast.Double
	case 'F':
		return ast.Float
	case 'V':
		return ast.Void
	case 'Z':
		return ast.Boolean
	case 'B':
		return ast.Byte
	case 'C':
		return ast.Char
	case 'S':
		return ast.Short
	default:
		return ast.BadToken
	}
}

func makeMethod(class *reflect.CtClass, mname string) *reflect.CtMethod {
	method := reflect.CtMethod{
		Class: &reflect.CtClass{
			Fields:        nil,
			PackageName:   "java.lang",
			SimpleName:    "Class",
			QualifiedName: "java/lang/Class",
			Acc: &reflect.AccDesc{
				Public: true,
			},
		},
		Name:       mname,
		Descriptor: "(Ljava/lang/String;)Ljava/lang/Class;",
		Acc: &reflect.AccDesc{
			Public: true,
			Static: true,
		},
	}
	return &method
}

func getCompOperator(expr ast.Node) ast.TokenID {
	if bexpr, ok := expr.(*ast.Expression); ok {
		token := bexpr.GetOperator()
		if token == ast.Not {
			return ast.Not
		}
		return token
	}
	if binExpr, ok := expr.(*ast.BinExpr); ok {
		token := binExpr.GetOperator()
		if token == ast.Not {
			return ast.Not
		}
		if token != ast.OROR && token != ast.ANDAND && token != ast.And && token != ast.Or {
			return ast.EQ // ==, !=, ...
		}
		return token
	}
	return ast.Empty // others
}

func isAlwaysBranch(expr ast.Node, branchIf bool) bool {
	if isNil(expr) {
		return false
	}
	if e, ok := expr.(*ast.Keyword); ok {
		id := e.TokenID
		if branchIf {
			return id == ast.True
		}
		return id == ast.False
	}
	return false
}

// Hash string hash
func Hash(s string) int64 {
	var hash int64 = 0
	for _, char := range s {
		hash = 31*hash + int64(char)
	}
	return hash
}

func StripPlusExpr(expr ast.Node) ast.Node {
	if isNil(expr) {
		return nil
	}
	switch e := expr.(type) {
	case *ast.BinExpr:
		if e.GetOperator() == ast.Plus && isNil(e.Oprand2()) {
			return e.GetLeft()
		}
	case *ast.Expression:
		op := e.GetOperator()
		if op == ast.Member {
			cexpr := getConstantFieldValue(e.Oprand2().(*ast.MemberSymbol))
			if cexpr != nil {
				return cexpr
			}
		} else if op == '+' && e.GetRight() == nil {
			return e.GetLeft()
		}
	case *ast.MemberSymbol:
		cexpr := getConstantFieldValue(e)
		if cexpr != nil {
			return cexpr
		}
	}

	return expr
}

// GetConstantFieldValue for a Member
func getConstantFieldValue(mem *ast.MemberSymbol) ast.Node {
	return getConstantFieldValue0(mem.GetField())
}

// GetConstantFieldValue for a CtField
func getConstantFieldValue0(f *reflect.CtField) ast.Node {
	if f == nil {
		return nil
	}

	value := f.GetConstantValue()
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return ast.NewStringL(v)
	case float64:
		return ast.NewDoubleConst(v, ast.DoubleConstant)
	case float32:
		return ast.NewDoubleConst(float64(v), ast.FloatConstant)
	case int64:
		return ast.NewIntConst(v, ast.LongConstant)
	case int:
		return ast.NewIntConst(int64(v), ast.IntConstant)
	case bool:
		token := ast.False
		if v {
			token = ast.True
		}
		return ast.NewKeyword(token.GetName(), token)
	default:
		return nil
	}
}

func JvmToJavaName(cname string) string {
	return strings.ReplaceAll(cname, "/", ".")
}

func JavaToJvmName(cname string) string {
	return strings.ReplaceAll(cname, ".", "/")
}

// getModifiers 解析关键字列表并返回修饰符位掩码
func getModifiers(mods *ast.ASTList) int {
	m := 0
	node := ast.Node(mods)
	for !isNil(node) {
		key := node.GetLeft().(*ast.Keyword)
		node = node.GetRight()
		switch key.TokenID {
		case ast.Static:
			m |= classfile.AccStatic
			break
		case ast.Final:
			m |= classfile.AccFinal
			break
		case ast.Synchronized:
			m |= classfile.AccSynchronized
			break
		case ast.Abstract:
			m |= classfile.AccAbstract
			break
		case ast.Public:
			m |= classfile.AccPublic
			break
		case ast.Protected:
			m |= classfile.AccProtected
			break
		case ast.Private:
			m |= classfile.AccPrivate
			break
		case ast.Volatile:
			m |= classfile.AccVolatile
			break
		case ast.Transient:
			m |= classfile.AccTransient
			break
		case ast.Strict:
			m |= classfile.AccStrict
			break
		default:
			break
		}
	}
	return m
}

func getTypeName(typeId ast.TokenID) (string, error) {
	cname := ""
	switch typeId {
	case ast.Boolean:
		cname = "boolean"
		break
	case ast.Char:
		cname = "char"
		break
	case ast.Byte:
		cname = "byte"
		break
	case ast.Short:
		cname = "short"
		break
	case ast.Int:
		cname = "int"
		break
	case ast.Long:
		cname = "long"
		break
	case ast.Float:
		cname = "float"
		break
	case ast.Double:
		cname = "double"
		break
	case ast.Void:
		cname = "void"
		break
	default:
		return "", NewCompileError("fatal type")
	}

	return cname, nil
}

func NumOfParameters(desc string) (int, error) {
	n := 0
	i := 1

	for {
		if i >= len(desc) {
			return 0, fmt.Errorf("bad descriptor")
		}

		c := desc[i]

		if c == ')' {
			break
		}

		// 处理数组维度
		for c == '[' {
			i++
			if i >= len(desc) {
				return 0, fmt.Errorf("bad descriptor")
			}
			c = desc[i]
		}

		// 处理类类型
		if c == 'L' {
			i = strings.IndexByte(desc[i:], ';') + i + 1
			if i <= 0 {
				return 0, fmt.Errorf("bad descriptor")
			}
		} else {
			i++
		}

		n++
	}
	return n, nil
}

func dataSize(d string, withRet bool) int {
	n := 0
	c := d[0]
	if c == '(' {
		i := 1
		for {
			c = d[i]
			if c == ')' {
				c = d[i+1]
				break
			}

			array := false
			for c == '[' {
				array = true
				i++
				c = d[i]
			}

			if c == 'L' {
				i = strings.Index(d, ";") + 1
				if i <= 0 {
					panic("bad descriptor")
				}
			} else {
				i++
			}
			if !array && (c == 'J' || c == 'D') {
				n -= 2
			} else {
				n--
			}
		}
	}
	if withRet {
		if c == 'J' || c == 'D' {
			n += 2
		} else if c != 'V' {
			n++
		}
	}
	return n
}

func DescriptorDataSize(d string) int {
	return dataSize(d, true)
}
