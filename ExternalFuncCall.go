package lamb

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

	if ext.args_given == 0 {
		// Expecting fully reduced string.
		str := arg.(String).value
		return &ExternalFuncCall{e.args_given + 1, 0, str, nil}, true
	}

	if ext.args_given == 1 {
		// Expecting fully reduced int.
		i := arg.(Int).value
		return &ExternalFuncCall{e.args_given + 1, i, e.name, nil}, true
	}

	if ext.args_remaining > 0 {
		// Append argument.
		return &ExternalFuncCall{e.args_given + 1, e.args_remaining - 1, e.name, append(append([]Expression{}, e.args...), e.arg)}, true
	}

	// Not applicable, can be reduced.
	return nil, false
}

func (e *ExternalFuncCall) Substitute(name string, value Expression) Expression {
	// TODO: this can be optimized by using string table and free variable bitmap.

	nargs := make([]Expression, len(e.args))
	for i := range e.args {
		nargs[i] = substitute(e.args[i], name, val)
	}
	return &ExternalFuncCall{e.args_given, e.args_remaining, e.func_name, nargs}
}

func (e *ExternalFuncCall) Reduce(ctx *Context) (Expression, bool) {
	if args_given < 2 || args_remaining > 0 {
		// Incomplete call.
		return e, false
	}

	// Complete call.
	return ctx.Externals[func_name](ctx, func_name, args), true
}

func (e *ExternalFuncCall) FullReduce(ctx *Context) (Expression, bool) {
	result, ok := e.Reduce()
	if ok {
		reduced_result, _ := fullReduce(ctx, result)
		return reduced_result, true
	} else {
		return e, false
	}
}

func (e *ExternalFuncCall) WriteTo(w Writer) {
	if e.args_given == 0 {
		fmt.Fprint(w, "__external")
		return
	}

	fmt.Fprintf(w, "__external[%d, %d, %s", e.args_given, e.args_remaining, e.func_name)
	for _, arg := range e.args {
		fmt.Fprint(w, ", ")
		writeTo(arg, w)
	}
	fmt.Fprint(w, "]")
}
