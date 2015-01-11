package lamb

import (
	"fmt"
)

type Int struct {
	value int64
}

func (e Int) Apply(arg Expression) (Expression, bool) {
	return nil, false
}

func (e Int) Substitute(name string, value Expression) Expression {
	return e
}

func (e Int) Reduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e Int) FullReduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e Int) WriteTo(w Writer) {
	fmt.Fprintf(w, "%d", e.value)
}

/*
func externalAddInt(args ...[]Expression) Expression {
	// TODO
}

func externalSubInt(args ...[]Expression) Expression {
	// TODO
}

func externalMulInt(args ...[]Expression) Expression {
	// TODO
}

func externalDivInt(args ...[]Expression) Expression {
	// TODO
}
*/
