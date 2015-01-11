package lamb

import (
	"fmt"
)

type ExternalFuncCall struct {
	args_given     int
	args_remaining int

	func_name string
	args      []Expression
}

func NewExternalFuncCall() Expression {
	return new(ExternalFuncCall)
}

func (e *ExternalFuncCall) Apply(arg Expression) (Expression, bool) {

	if e.args_given == 0 {
		// Expecting fully reduced string.
		str := string(arg.(String))
		return &ExternalFuncCall{e.args_given + 1, 0, str, nil}, true
	}

	if e.args_given == 1 {
		// Expecting fully reduced int.
		i := arg.(Int).value
		return &ExternalFuncCall{e.args_given + 1, int(i), e.func_name, nil}, true
	}

	if e.args_remaining > 0 {
		// Append argument.
		return &ExternalFuncCall{e.args_given + 1, e.args_remaining - 1, e.func_name, append(append([]Expression{}, e.args...), arg)}, true
	}

	// Not applicable, can be reduced.
	return nil, false
}

func (e *ExternalFuncCall) Substitute(name string, value Expression) Expression {
	nargs := make([]Expression, len(e.args))
	for i := range e.args {
		nargs[i] = Substitute(e.args[i], name, value)
	}
	return &ExternalFuncCall{e.args_given, e.args_remaining, e.func_name, nargs}
}

func (e *ExternalFuncCall) Reduce(ctx *Context) (Expression, bool) {
	if e.args_given < 2 || e.args_remaining > 0 {
		// Incomplete call.
		return e, false
	}

	// Complete call.
	return ctx.Externals[e.func_name](ctx, e.func_name, e.args), true
}

func (e *ExternalFuncCall) FullReduce(ctx *Context) (Expression, bool) {
	// Coincidentally, here Reduce and FullReduce are the same.
	return e.Reduce(ctx)
}

func (e *ExternalFuncCall) WriteTo(w Writer) {
	if e.args_given == 0 {
		fmt.Fprint(w, "__external")
		return
	}

	fmt.Fprintf(w, "__external[%d, %d, %s", e.args_given, e.args_remaining, e.func_name)
	for _, arg := range e.args {
		fmt.Fprint(w, ", ")
		WriteTo(arg, w)
	}
	fmt.Fprint(w, "]")
}
