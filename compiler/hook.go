package compiler

import (
	"github.com/Yyjccc/GoJavassist/classfile"
	"github.com/Yyjccc/GoJavassist/compiler/ast"
)

type ProceedHandler interface {
	DoIt(gen *JvtCodeGenerator, args *ast.ASTList) error
	SetReturnType(check *JvstTypeChecker, args *ast.ASTList) error
}

type ReturnHook struct {
	next *ReturnHook
}

func NewReturnHook(gen *CCodeGenerator) *ReturnHook {
	hook := &ReturnHook{}
	hook.next = gen.returnHooks
	gen.returnHooks = hook
	return hook
}

func (h *ReturnHook) DoIt(gen *JvtCodeGenerator, opcode classfile.OpCode) bool {
	return false
}

func (h *ReturnHook) Remove(gen *CCodeGenerator) {
	gen.returnHooks = h.next
}

type JsrHook struct {
	*ReturnHook
	jsrList  [][]int
	cgen     *CCodeGenerator
	varIndex int
}

func NewJsrHook(gen *CCodeGenerator) *JsrHook {
	return &JsrHook{
		ReturnHook: NewReturnHook(gen),
		jsrList:    [][]int{},
		cgen:       gen,
		varIndex:   -1,
	}
}

func (j *JsrHook) GetVar(size int) int {
	if j.varIndex < 0 {
		j.varIndex = j.cgen.bytecodes.MaxLocals
		j.cgen.bytecodes.incMaxLocals(size)
	}
	return j.varIndex
}

func (j *JsrHook) jsrJmp(b *ByteCodes) {
	b.AddOpcode(classfile.OpGoto)
	j.jsrList = append(j.jsrList, []int{b.currentPc(), j.varIndex})
	b.WriteIndex(0)
}

func (j *JsrHook) DoIt(b *ByteCodes, opcode classfile.OpCode) bool {
	switch opcode {
	case classfile.OpReturn:
		j.jsrJmp(b)
		break
	case classfile.OpAReturn:
		b.addAStore(j.GetVar(1))
		j.jsrJmp(b)
		b.AddLload(j.varIndex)
		break
	case classfile.OpIReturn:
		b.addIStore(j.GetVar(1))
		j.jsrJmp(b)
		b.addIload(j.varIndex)
		break
	case classfile.OpLReturn:
		b.addLstore(j.GetVar(2))
		j.jsrJmp(b)
		b.AddLload(j.varIndex)
		break
	case classfile.OpDReturn:
		b.addDstore(j.GetVar(2))
		j.jsrJmp(b)
		b.addDload(j.varIndex)
		break
	case classfile.OpFReturn:
		b.addFstore(j.GetVar(1))
		j.jsrJmp(b)
		b.addFload(j.varIndex)
		break
	default:
		panic("fatal JsrHook")
	}
	return false
}
