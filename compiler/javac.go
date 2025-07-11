package compiler

import (
	"github.com/Yyjccc/GoJavassist/compiler/ast"
	"github.com/Yyjccc/GoJavassist/compiler/reflect"
)

const (
	param0Name    = "$0"
	resultVarName = "$_"
	proceedName   = "$proceed"
)

// Reference: Javassist
type Javac struct {
	table *SymbolTable
	gen   *JvtCodeGenerator
}

func NewJavac(class *reflect.CtClass) *Javac {
	return &Javac{
		table: NewSymbolTable(nil),
		gen:   NewJvtCodeGenerator(class),
	}
}

func (j *Javac) GetGen() *JvtCodeGenerator {
	return j.gen
}

func CompileMethod(class *reflect.CtClass, src string) (*reflect.CtMethod, error) {
	compiler := NewJavac(class)
	err := compiler.compile(src)
	if err != nil {
		return nil, err
	}
	return compiler.gen.thisMethod, nil
}

func MakeCtMethod(src string, class *reflect.CtClass) (*reflect.CtMethod, error) {
	return CompileMethod(class, src)
}

func (compiler *Javac) compile(src string) error {
	p := NewParser(src)
	mem, isMethod, err := p.parseMember1(compiler.table)
	if err != nil {
		return err
	}
	if isMethod {
		err := compiler.compileMethod(p, &ast.MethodDecl{ASTList: mem})
		if err != nil {
			return err
		}
	} else {
		compiler.compileField(&ast.FieldDecl{ASTList: mem})
	}
	return nil
}

func (compiler *Javac) compileField(a *ast.FieldDecl) {

}

func (compiler *Javac) compileMethod(p *Parser, md *ast.MethodDecl) error {
	mod := getModifiers(md.GetModifiers())
	acc := reflect.NewAccDec(uint16(mod))
	plist, err := compiler.gen.makeParamList(md)
	if err != nil {
		return err
	}
	tlist, err := compiler.gen.makeThrowsList(md)
	if err != nil {
		return err
	}

	compiler.gen.recordParams(plist, acc.Static, "$", "$args", "$$", !acc.Static, 0, JavaToJvmName(compiler.gen.thisClass.QualifiedName), compiler.table)
	md, err = p.parseMethod2(compiler.table, md)
	if err != nil {
		return err
	}

	if md.IsConstructor() {
		//CtConstructor cons = new CtConstructor(plist,
		//	gen.getThisClass());
		//cons.setModifiers(mod);
		//md.accept(gen);
		//cons.getMethodInfo().setCodeAttribute(
		//	bytecode.toCodeAttribute());
		//cons.setExceptionTypes(tlist);
		//return cons;
	} else {
		r := md.GetReturn()
		rType, err := compiler.gen.resolver.lookupClass(r)
		if err != nil {
			return err
		}
		compiler.recordReturnType(rType, false)
		method := reflect.NewCtMethodInfo(rType, r.GetVariable().Identifier, plist, compiler.gen.thisClass)
		method.SetModifiers(acc)
		compiler.gen.thisMethod = method
		if err = md.Accept(compiler.gen); err != nil {
			return err
		}
		if md.GetBody() != nil {
			//设置code属性
			method.AddAttribute(compiler.gen.bytecodes.ToCodeAttribute())
		} else {
			acc.Abstract = true
			method.SetModifiers(acc)
		}

		method.SetExceptionTypes(tlist)
		return nil
	}

	return nil
}

func (compiler *Javac) CompileBody(method *reflect.CtMethod, src string) error {
	mod := method.Acc
	compiler.recordParams(method.GetParameterTypes(), mod.Static)
	compiler.gen.SetThisMethod(method)
	rType := method.GetReturnType()
	compiler.recordReturnType(rType, false)
	isVoid := rType == reflect.VoidType
	if src == "" {

	} else {
		p := NewParser(src)
		stb := NewSymbolTable(compiler.table)
		s, err := p.parseStatement(stb)
		if err != nil {
			return err
		}
		if p.HasMore() {
			return NewCompileError("the method/constructor body must be surrounded by {}")
		}
		//boolean callSuper = false;
		//if (method instanceof CtConstructor) {
		//	callSuper = !((CtConstructor)method).isClassInitializer();
		//}
		if err = compiler.gen.atMethodBody(s, false, isVoid); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Javac) recordReturnType(rType *reflect.CtClass, useResultVar bool) int {
	compiler.gen.recordType(rType)
	name := ""
	if useResultVar {
		name = resultVarName
	}
	return compiler.gen.recordReturnType(rType, "$r", name, compiler.table)
}

func (compiler *Javac) recordParams(params []*reflect.CtClass, isStatic bool) int {
	return compiler.gen.recordParams(params, isStatic, "$", "$args", "$$", !isStatic, 0, compiler.gen.GetThisName(), compiler.table)
}
