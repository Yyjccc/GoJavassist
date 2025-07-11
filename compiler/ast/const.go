package ast

import "strconv"

// StringLiteral string Literal
type StringLiteral struct {
	Node
	Text string
}

func NewStringL(text string) *StringLiteral {
	return &StringLiteral{
		Text: text,
	}
}

func (s *StringLiteral) ToString() string {
	return "\"" + s.Text + "\""
}

func (s *StringLiteral) Accept(v Visitor) error {
	return v.AtStringL(s)
}

type IntConst struct {
	Node
	Value  int64
	TypeID TokenID
}

func NewIntConst(val int64, typeID TokenID) *IntConst {
	return &IntConst{
		Value:  val,
		TypeID: typeID,
	}
}

func (i *IntConst) ToString() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i *IntConst) Accept(v Visitor) error {
	return v.AtIntConst(i)
}

func (i *IntConst) Compute(op string, right Node) Node {
	if intConst, ok := right.(*IntConst); ok {
		return i.ComputeInt(op, intConst)
	}
	if doubleConst, ok := right.(*DoubleConst); ok {
		return i.ComputeDouble(op, doubleConst)
	}
	return nil
}

func (ic *IntConst) ComputeInt(op string, right *IntConst) *IntConst {
	if right == nil {
		return nil
	}

	type1 := ic.TypeID
	type2 := right.TypeID
	var newType TokenID

	if type1 == LongConstant || type2 == LongConstant {
		newType = LongConstant
	} else if type1 == CharConstant && type2 == CharConstant {
		newType = CharConstant
	} else {
		newType = IntConstant
	}

	value1 := ic.Value
	value2 := right.Value
	var newValue int64

	switch op {
	case "+":
		newValue = value1 + value2
	case "-":
		newValue = value1 - value2
	case "*":
		newValue = value1 * value2
	case "/":
		newValue = value1 / value2
	case "%":
		newValue = value1 % value2
	case "|":
		newValue = value1 | value2
	case "^":
		newValue = value1 ^ value2
	case "&":
		newValue = value1 & value2
	case "<<":
		newValue = value1 << uint(value2)
		newType = type1
	case ">>":
		newValue = value1 >> uint(value2)
		newType = type1
	default:
		return nil
	}

	return NewIntConst(newValue, newType)
}

func (ic *IntConst) ComputeDouble(op string, right *DoubleConst) *DoubleConst {
	return computeDoubleOp(op, float64(ic.Value), right.Value, right.TypeID)
}

type DoubleConst struct {
	Node
	Value  float64
	TypeID TokenID
}

func (d *DoubleConst) GetType() TokenID {
	return d.TypeID
}

func NewDoubleConst(val float64, typeID TokenID) *DoubleConst {
	return &DoubleConst{
		Value:  val,
		TypeID: typeID,
	}
}

func (d *DoubleConst) ToString() string {
	return strconv.FormatFloat(d.Value, 'f', -1, 64)
}

func (d *DoubleConst) Accept(v Visitor) error {
	return v.AtDoubleConst(d)
}

func (d *DoubleConst) ComputeDouble(op string, right *DoubleConst) *DoubleConst {
	var newType TokenID
	if d.TypeID == DoubleConstant || right.TypeID == DoubleConstant {
		newType = DoubleConstant
	} else {
		newType = FloatConstant
	}
	return computeDoubleOp(op, d.Value, right.Value, newType)

}

func (d *DoubleConst) ComputeInt(op string, right *IntConst) *DoubleConst {
	return computeDoubleOp(op, d.Value, float64(right.Value), d.TypeID)
}

func computeDoubleOp(op string, value1, value2 float64, newType TokenID) *DoubleConst {
	var newValue float64
	switch op {
	case "+":
		newValue = value1 + value2
		break
	case "-":
		newValue = value1 - value2
		break
	case "*":
		newValue = value1 * value2
		break
	case "/":
		newValue = value1 / value2
		break
	case "%":
		newValue = float64(int64(value1) % int64(value2))
		break
	default:
		return nil
	}
	return NewDoubleConst(newValue, newType)
}

func (d *DoubleConst) Compute(op string, right Node) Node {
	if intConst, ok := right.(*IntConst); ok {
		return d.ComputeInt(op, intConst)
	}
	if doubleConst, ok := right.(*DoubleConst); ok {
		return d.ComputeDouble(op, doubleConst)
	}
	return nil
}
