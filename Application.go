package lamb

type Application struct {
	fnc Expression
	arg Expression
}

func NewApplication(fnc, arg Expression) Expression {
	if fnc == nil {
		return arg
	}
	return &Application{fnc, arg}
}

func (e *Application) Apply(arg Expression) (Expression, bool) {
	// Unreduced application cannot take argument.
	return nil, false
}

func (e *Application) Substitute(name string, value Expression) Expression {
	return &Application{Substitute(e.fnc, name, value), Substitute(e.arg, name, value)}
}

func (e *Application) Reduce(ctx *Context) (Expression, bool) {
	// This depends on the assumption that an applicable expression
	// is not reducible.

	nfnc, ok := Reduce(ctx, e.fnc)
	if ok {
		return &Application{nfnc, e.arg}, true
	}

	// Special case #1: function is a strict lambda.

	lmb, is_lambda := e.fnc.(*Lambda)
	strict := is_lambda && lmb.strict

	// Special case #2: argument is a forced evaluation bracket.

	_, forced := e.arg.(ForceEval)
	strict = strict || forced

	if strict {
		narg, ok := Reduce(ctx, e.arg)
		if ok {
			return &Application{e.fnc, narg}, true
		}
	}

	// Common case: apply the argument.

	result, ok := Apply(e.fnc, e.arg)
	if !ok {
		// TODO: print error or something
		return e, false
	}
	return result, ok
}

func (e *Application) FullReduce(ctx *Context) (Expression, bool) {
	// This depends on the assumption that an applicable expression
	// is not reducible further.

	nfnc := e.fnc
	repeat := true
	for repeat {
		nfnc, repeat = FullReduce(ctx, nfnc)
	}

	// Special case #1: function is a strict lambda.

	lmb, is_lambda := nfnc.(*Lambda)
	strict := is_lambda && lmb.strict

	// Special case #2: argument is a forced evaluation bracket.

	_, forced := e.arg.(ForceEval)
	strict = strict || forced

	// Reduce the argument if necessary.

	narg := e.arg

	if strict {
		repeat = true
		for repeat {
			narg, repeat = FullReduce(ctx, narg)
		}
	}

	// Apply the argument.

	result, ok := Apply(nfnc, narg)
	if !ok {
		// TODO: print error or something
		return &Application{nfnc, narg}, false
	}

	// DO NOT reduce the result.
	// Just tell caller to repeat fullReduce.
	// If we reduced here, recursive calls could consume
	// unbounded amount of stack.

	return result, true
}

func (e *Application) WriteTo(w Writer) {
	WriteTo(e.fnc, w)
	w.Write([]byte{' '})

	// Special case to inject parentheses around application-as-argument.
	_, is_app := e.arg.(*Application)

	if is_app {
		w.Write([]byte{'('})
		WriteTo(e.arg, w)
		w.Write([]byte{')'})
	} else {
		WriteTo(e.arg, w)
	}
}
