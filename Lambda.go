package lamb

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

var bound int64

type Lambda struct {
	// Name of the bound variable.
	name string

	// This is checked by a special case in Application.Reduce
	strict bool

	body Expression
}

// We exploit the fact that names in lamb are not allowed
// to begin with a number. We substitute the bound name
// with a numerical name which makes sure that later substitution
// cannot capture a free variable.
// This way, lambda abstraction only binds to names that
// are syntactically within it.
func NewLambda(name string, strict bool, body Expression) Expression {

	num := strconv.FormatInt(atomic.AddInt64(&bound, 1), 10)

	return &Lambda{num, strict, Substitute(body, name, Name(num))}
}

func (e *Lambda) Apply(arg Expression) (Expression, bool) {

	return Substitute(e.body, e.name, arg), true
}

func (e *Lambda) Substitute(name string, value Expression) Expression {
	if name == e.name {
		return e
	}

	return &Lambda{e.name, e.strict, Substitute(e.body, name, value)}
}

func (e *Lambda) Reduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e *Lambda) FullReduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e *Lambda) WriteTo(w Writer) {
	if e.strict {
		fmt.Fprint(w, "\\!")
	} else {
		fmt.Fprint(w, "\\")
	}
	fmt.Fprintf(w, "%s (", e.name)
	WriteTo(e.body, w)
	fmt.Fprintf(w, ")")
}
