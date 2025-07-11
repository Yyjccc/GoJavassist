package ast

type Pair struct {
	Left, Right Node
}

func NewPair(_left, _right Node) *Pair {
	return &Pair{Left: _left, Right: _right}
}

func (p *Pair) GetLeft() Node {
	return p.Left
}
func (p *Pair) GetRight() Node {
	return p.Right
}
func (p *Pair) SetLeft(n Node) {
	p.Left = n
}
func (p *Pair) SetRight(n Node) {
	p.Right = n
}

func (p *Pair) ToString() string {
	str := "(<Pair> "
	if p.Left != nil {
		str += p.Left.ToString()
	} else {
		str += "<null>"
	}
	str += " . "
	if p.Right != nil {
		str += p.Right.ToString()
	} else {
		str += "<null>"
	}
	str += ")"
	return str
}

func (p *Pair) GetTag() string {
	return "pair"
}

func (p *Pair) Accept(visitor Visitor) error {
	return visitor.AtPair(p)
}
