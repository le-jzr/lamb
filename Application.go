package lamb

type application struct {
	fnc Expression
	arg Expression
}

func NewApplication(fnc, arg Expression) Expression {
	if fnc == nil {
		return arg
	}
	return &application{fnc, arg}
}

func (e *application) Apply(arg Expression) (Expression, bool) {
	// Unreduced application cannot take argument.
	return nil, false
}

func (e *application) Substitute(name string, value Expression) Expression {
	return &application{substitute(e.fnc, name, val), substitute(e.arg, name, val)}
}

func (e *application) Reduce(ctx *Context) (Expression, bool) {
	// This depends on the assumption that an applicable expression
	// is not reducible.

	nfnc, ok := reduce(ctx, e.fnc)
	if ok {
		return &application{nfnc, e.arg}, true
	}

	// Special case #1: function is a strict lambda.

	lmb, is_lambda := e.fnc.(*Lambda)
	strict := is_lambda && lmb.strict

	// Special case #2: argument is a forced evaluation bracket.

	_, forced := e.arg.(forceEval)
	strict = strict || forced

	if strict {
		narg, ok := reduce(ctx, e.arg())
		if ok {
			return &application{e.fnc, narg}, true
		}
	}

	// Common case: apply the argument.

	result, ok := apply(e.fnc, e.arg)
	if !ok {
		// TODO: print error or something
		return e, false
	}
	return result, ok
}

func (e *application) FullReduce(ctx *Context) (Expression, bool) {
	// This depends on the assumption that an applicable expression
	// is not reducible.

	nfnc, _ := fullReduce(ctx, e.fnc)

	// Special case #1: function is a strict lambda.

	lmb, is_lambda := nfnc.(*lambda)
	strict := is_lambda && lmb.strict

	// Special case #2: argument is a forced evaluation bracket.

	_, forced := e.arg.(forceEval)
	strict = strict || forced

	// Reduce the argument if necessary.

	narg := e.arg

	if strict {
		narg, _ := fullReduce(ctx, e.arg)
	}

	// Apply the argument.

	result, ok := apply(nfnc, narg)
	if !ok {
		// TODO: print error or something
		return &application{nfnc, narg}, true
	}

	// Reduce the result.

	reduced, _ := fullReduce(ctx, result)

	return reduced, true
}

func (e *application) WriteTo(w Writer) {
	writeTo(e.fnc, w)
	w.Write([]byte{' '})

	// Special case to inject parentheses around application-as-argument.
	_, is_app := e.arg.(*application)

	if is_app {
		w.Write([]byte{'('})
		writeTo(e.arg, w)
		w.Write([]byte{')'})
	} else {
		writeTo(e.arg, w)
	}
}
