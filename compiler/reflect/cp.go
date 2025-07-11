package reflect

import (
	"GoJavassist/classfile"
	"fmt"

	"strings"
)

// ConstPool 常量池 对[]classfile.ConstantInfo 进行封装
type ConstPool struct {
	entries       []classfile.ConstantInfo
	top           int //init -1
	This          *CtClass
	cache         map[string]int
	thisClassInfo int
}

func (p *ConstPool) GetPool() []classfile.ConstantInfo {
	return p.entries
}

func MakeConstPool() *ConstPool {
	c := &ConstPool{
		entries: make([]classfile.ConstantInfo, 0),
		cache:   make(map[string]int),
		top:     -1,
	}
	c.AddItem(nil)
	return c
}

func (p *ConstPool) NewLocalVariableAttribute() *classfile.LocalVariableTableAttribute {
	p.AddString(classfile.LocalVariableTable)
	return &classfile.LocalVariableTableAttribute{LocalVariableTable: make([]classfile.LocalVariableTableEntry, 0)}
}

func (p *ConstPool) AddItem(item classfile.ConstantInfo) int {
	p.entries = append(p.entries, item)
	p.top++
	return p.top
}

func NewExceptionAttribute() *classfile.ExceptionsAttribute {
	return &classfile.ExceptionsAttribute{
		ExceptionIndexTable: make([]uint16, 0),
	}
}

func (p *ConstPool) AddClassInfo(jvmType string) int {
	nameIdx := p.AddString(strings.ReplaceAll(jvmType, ".", "/"))
	info := classfile.ConstantClassInfo{NameIndex: uint16(nameIdx)}
	return p.AddConstantClassInfo(info)
}

func (p *ConstPool) AddMethodRefInfo(index int, name string, d string) int {
	nt := p.AddNameAndTypeInfo(name, d)
	return p.AddMethodRefInfo0(index, nt)
}

func (p *ConstPool) AddNameAndTypeInfo(name string, d string) int {
	// 组合 name 和 d 来生成唯一标识符
	key := fmt.Sprintf("%s:%s", name, d)
	if idx, exists := p.cache[key]; exists {
		return idx
	}
	info := classfile.ConstantNameAndTypeInfo{
		NameIndex:       uint16(p.AddString(name)),
		DescriptorIndex: uint16(p.AddString(d)),
	}
	idx := p.AddItem(info)
	p.cache[key] = idx
	return idx
}

func (p *ConstPool) AddString(s string) int {
	if idx, exists := p.cache[s]; exists {
		return idx
	}
	item := p.AddItem([]uint8(s))
	p.cache[s] = item
	return item
}

func (p *ConstPool) AddMethodRefInfo0(index int, nt int) int {
	key := fmt.Sprintf("@[MethodRefInfo-%d:%d", index, nt)
	if idx, exists := p.cache[key]; exists {
		return idx
	}

	info := classfile.ConstantMethodRefInfo{
		ClassIndex:       uint16(index),
		NameAndTypeIndex: uint16(nt),
	}
	return p.AddItem(info)
}

func (p *ConstPool) AddInterfaceMethodRefInfo(clazz int, name string, s string) int {
	nt := p.AddNameAndTypeInfo(name, s)
	return p.AddInterfaceMethodRefInfo0(clazz, nt)
}

func (p *ConstPool) AddInterfaceMethodRefInfo0(clazz int, nt int) int {
	key := fmt.Sprintf("@[InterfaceMethodRefInfo-%d:%d", clazz, nt)
	if idx, exists := p.cache[key]; exists {
		return idx
	}
	info := classfile.ConstantInterfaceMethodRefInfo{
		ClassIndex:       uint16(clazz),
		NameAndTypeIndex: uint16(nt),
	}
	item := p.AddItem(info)
	p.cache[key] = item
	return item
}

func (p *ConstPool) AddClassInfo0(class *CtClass) int {
	if class == p.This {
		return p.thisClassInfo
	}
	if !class.IsArray() {
		return p.AddClassInfo(class.QualifiedName)
	}
	return p.AddClassInfo(class.QualifiedName)
}

// ExceptionTable Exception表 对[]classfile.ExceptionTableEntry 进行封装
type ExceptionTable struct {
	pool    *ConstPool
	Entries []classfile.ExceptionTableEntry
}

func (t *ExceptionTable) Add(start int, end int, pc int, index int) {
	if start < end {
		entry := classfile.ExceptionTableEntry{
			StartPc:   uint16(start),
			EndPc:     uint16(end),
			HandlerPc: uint16(pc),
			CatchType: uint16(index),
		}
		t.Entries = append(t.Entries, entry)
	}
}

func (t *ExceptionTable) ToExceptionTable() []classfile.ExceptionTableEntry {
	return t.Entries
}

func NewExceptionTable(pool *ConstPool) *ExceptionTable {
	return &ExceptionTable{
		pool:    pool,
		Entries: make([]classfile.ExceptionTableEntry, 0),
	}
}

func (p *ConstPool) AddConstantClassInfo(info classfile.ConstantClassInfo) int {
	key := fmt.Sprintf("@ConstantClassInfo-%d", info.NameIndex)
	if idx, exists := p.cache[key]; exists {
		return idx
	}
	i := p.AddItem(info)
	p.cache[key] = i
	return i
}

func (p *ConstPool) AddStringRef(s string) int {
	strIndex := p.AddString(s)
	key := fmt.Sprintf("@StringRef-%d", strIndex)
	if idx, exists := p.cache[key]; exists {
		return idx
	}
	info := classfile.ConstantStringInfo{StringIndex: uint16(strIndex)}
	idx := p.AddItem(info)
	p.cache[key] = idx
	return idx
}

func (p *ConstPool) AddFieldrefInfo(classInfo int, name string, descriptor string) int {
	nt := p.AddNameAndTypeInfo(name, descriptor)
	return p.addFieldrefInfo(classInfo, nt)
}

func (p *ConstPool) addFieldrefInfo(info int, nt int) int {
	return p.AddItem(classfile.ConstantFieldRefInfo{
		ClassIndex:       uint16(info),
		NameAndTypeIndex: uint16(nt),
	})
}

func (p *ConstPool) AddInterfaceMethodrefInfo(clazz int, name string, t string) int {
	nt := p.AddNameAndTypeInfo(name, t)
	return p.addInterfaceMethodrefInfo0(clazz, nt)
}

func (p *ConstPool) addInterfaceMethodrefInfo0(clazz int, nt int) int {
	return p.AddItem(classfile.ConstantInterfaceMethodRefInfo{
		ClassIndex:       uint16(clazz),
		NameAndTypeIndex: uint16(nt),
	})
}

func (p *ConstPool) SetExceptions(ea *classfile.ExceptionsAttribute, names []string) {
	for _, name := range names {
		info := p.AddClassInfo(name)
		ea.ExceptionIndexTable = append(ea.ExceptionIndexTable, uint16(info))
	}
}
