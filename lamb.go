package main

type Expression interface {
	Substitute(name string, value Expression) Expression
	Reduce() (Expression, bool)
	FullReduce() (Expression, bool)
}

// Pseudo function that is used to interpret '!' for strict evaluation.
type ForceEval struct {
}

func (e ForceEval) Substitute(name string, value Expression) Expression {
	return e
}

func (e ForceEval) Reduce() (Expression, bool) {
	return e, false
}

func (e ForceEval) FullReduce() (Expression, bool) {
	return e, false
}

type ExternalFuncCall struct {
	args_given     int
	args_remaining int

	func_name string
	args      []Expression
}

func (e *ExternalFuncCall) Substitute(name string, value Expression) Expression {
	// TODO: this can be optimized by using string table and free variable bitmap.

	nargs := make([]Expression, len(e.args))
	for i := range e.args {
		nargs[i] = substitute(e.args[i], name, val)
	}
	return &ExternalFuncCall{e.args_given, e.args_remaining, e.func_name, nargs}
}

func (e *ExternalFuncCall) Reduce() (Expression, bool) {
	if args_given < 2 || args_remaining > 0 {
		// Incomplete call.
		return e, false
	}

	// Complete call.
	return callExternal(func_name, args...), true
}

func (e *ExternalFuncCall) FullReduce() (Expression, bool) {
	result, ok := e.Reduce()
	if ok {
		reduced_result, _ := fullReduce(result)
		return reduced_result, true
	} else {
		return e, false
	}
}

type Int struct {
	value uint64
}

func (e Int) Substitute(name string, value Expression) Expression {
	return e
}

func (e Int) Reduce() (Expression, bool) {
	return e, false
}

func (e Int) FullReduce() (Expression, bool) {
	return e, false
}

type String struct {
	value string
}

func (e String) Substitute(name string, value Expression) Expression {
	return e
}

func (e String) Reduce() (Expression, bool) {
	return e, false
}

func (e String) FullReduce() (Expression, bool) {
	return e, false
}

type Name struct {
	name string
}

func (e Name) Substitute(name string, value Expression) Expression {
	if e.name == name {
		return value
	}
	return e
}

func (e Name) Reduce() (Expression, bool) {
	return _namespace[e.name], true
}

func (e Name) FullReduce() (Expression, bool) {
	expr := _namespace[e.name]
	expr_reduced, _ := fullReduce(expr)
	return expr_reduced, true
}

// Lambda is represented as a pseudo-function: it is applied to the inner expression using function application.
// This must be taken into account when substituting.
type Lambda struct {
	name   string
	strict bool

	// Created due to \x syntax. Stack must be popped again when it is applied due to a closing parenthesis.
	synthetic bool
}

func (e *Lambda) Substitute(name string, value Expression) Expression {
	return e
}

func (e *Lambda) Reduce() (Expression, bool) {
	return e, false
}

func (e *Lambda) FullReduce() (Expression, bool) {
	return e, false
}

type Application struct {
	fnc Expression
	arg Expression
}

func (e *Application) Substitute(name string, value Expression) Expression {
	lambda, ok := e.fnc.(*Lambda)
	if ok {
		// This is actually a lambda abstraction.

		if lambda.name == name {
			return e
		}
	}

	// TODO: Avoid free variable capture.
	//   - Construct the set of free variables in the value.
	//   - Rename all lambdas along the way that match any of the free vars.

	return apply(substitute(e.fnc, name, val), substitute(e.arg, name, val))
}

func (e *Application) Reduce() (Expression, bool) {
	// The Application can be reduced in four ways.
	// Way 1: fnc is an applied lambda (i.e. an application where fnc is Lambda and arg is lambda's body)

	app, ok := e.fnc.(*Application)
	if ok {
		lamb, ok := app.fnc.(*Lambda)
		if ok {
			if lamb.strict {
				// Must reduce argument first.
				narg, ok := reduce(e.arg)
				if ok {
					return apply(e.fnc, narg)
				}
			} else {
				a2, ok := e.arg.(*Application)
				if ok {
					_, ok := a2.fnc.(ForceEval)
					if ok {
						narg, ok := reduce(a2.arg)
						if ok {
							return apply(e.fnc, apply(ForceEval{}, narg)), true
						} else {
							return apply(e.fnc, narg), true
						}
					}
				}
			}

			return substitute(app.arg, lamb.name, e.arg), true
		}
	}

	// Way 2: fnc is a partial external.

	ext, ok := e.fnc.(*ExternalFuncCall)
	if ok {
		if ext.args_given == 0 {
			// Expecting fully reduced string.
			narg, ok := reduce(e.arg)
			if ok {
				return apply(e.fnc, narg)
			}

			str := narg.(String).value
			return &ExternalFuncCall{ext.args_given + 1, 0, str, nil}, true
		}

		if ext.args_given == 1 {
			// Expecting fully reduced int.
			narg, ok := reduce(e.arg)
			if ok {
				return apply(e.fnc, narg)
			}

			i := narg.(Int).value
			return &ExternalFuncCall{ext.args_given + 1, i, ext.name, nil}, true
		}

		if ext.args_remaining > 0 {
			// Process forced evaluation.
			a2, ok := e.arg.(*Application)
			if ok {
				_, ok := a2.fnc.(ForceEval)
				if ok {
					narg, ok := reduce(a2.arg)
					if ok {
						return apply(e.fnc, apply(ForceEval{}, narg)), true
					} else {
						return apply(e.fnc, narg), true
					}
				}
			}

			// Append argument.
			return &ExternalFuncCall{ext.args_given + 1, ext.args_remaining - 1, ext.name, append(append([]Expression{}, ext.args...), e.arg)}, true
		}
	}

	// Way 3: fnc in none of above, but can itself be reduced.

	nfnc, ok := reduce(e.fnc)
	if ok {
		return apply(nfnc, e.arg), true
	}

	// Way 4: fnc is ForceEval or nil, in which case it is treated as identity
	_, ok = e.fnc.(ForceEval)
	ok = ok || e.fnc == nil
	if ok {
		return e.arg, true
	}

	// If none of the above hold, and fnc is not Lambda, then this is a bug in the interpreted program.

	_, ok := e.fnc.(*Lambda)
	if ok {
		return e, false
	}

	panic("Cannot reduce function call.")
}

func (e *Application) FullReduce() (Expression, bool) {

	nfnc, ok := fullReduce(e.fnc)

	_, lambda := e.fnc.(*Lambda)
	if lambda {
		return apply(e.fnc, nfnc), ok
	}

	narg := e.arg

	// Process the strict evaluation bracket.
	a2, ok := narg.(*Application)
	if ok {
		_, ok = a2.fnc.(ForceEval)
	}
	if ok {
		narg, _ = fullReduce(a2.arg)
	}

	// fnc is ForceEval or nil, in which case it is treated as identity
	_, ok = nfnc.(ForceEval)
	ok = ok || nfnc == nil
	if ok {
		return narg, true
	}

	// fnc is an applied lambda (i.e. an application where fnc is Lambda and arg is lambda's body)

	app, ok := nfnc.(*Application)
	if ok2 {
		lamb, ok := app.fnc.(*Lambda)
		if ok {
			if lamb.strict {
				narg, _ = fullReduce(narg)
			}

			red, _ := fullReduce(substitute(app.arg, lamb.name, narg))
			return red, true
		}
	}

	// fnc is a partial external.

	ext, ok := e.fnc.(*ExternalFuncCall)
	if ok {
		if ext.args_given == 0 {
			// Expecting fully reduced string.
			narg, _ = fullReduce(narg)
			str := narg.(String).value
			return &ExternalFuncCall{ext.args_given + 1, 0, str, nil}, true
		}

		if ext.args_given == 1 {
			// Expecting fully reduced int.
			narg, _ := fullReduce(narg)
			i := narg.(Int).value
			ext := &ExternalFuncCall{ext.args_given + 1, i, ext.name, nil}
			red, _ := ext.FullReduce()
			return red, true
		}

		if ext.args_remaining > 0 {
			// Append argument.
			ext := &ExternalFuncCall{ext.args_given + 1, ext.args_remaining - 1, ext.name, append(append([]Expression{}, ext.args...), e.arg)}
			red, _ := ext.FullReduce()
			return red, true
		}
	}

	panic("Cannot reduce function call.")
}

func readNext(r io.Reader) string {
	sbldr := new(StringBuilder)

	Char()

	sbldr.append()
}

func interpret(r io.Reader) {

	var current Expression

	for {
		next := readNext(r)
		if r == "" {
			break
		}

		switch next {
		case ";":
			current = apply(apply(lambda("", true, false), nil), current)

		case "(":
			stk.push(current)
			current = nil

		case ")":
			fn := stk.pop()

			current = applyFunction(fn, current)

			for isForceEval(fn) || isSyntheticLambda(fn) {
				fn = stk.pop()
				current = applyFunction(fn, current)
			}

		case "!(":
			stk.push(current)
			stk.push(ForceEval{})
			current = nil

		case "__extern":
			current = applyFunction(current, new(ExternalFuncCall))

		default:
			switch {
			case next[0] == '\\':
				// Emulates parentheses.
				push(current)
				if next[1] == '!' {
					push(lambda(next[2:], true, true))
				} else {
					push(lambda(next[1:], false, true))
				}
				current = nil

			case isInt(next):
				current = applyFunction(current, parseInt(next))

			default:
				current = applyFunction(current, name(next))
			}
		}

		if stk.empty() {
			current = fullReduce(current)
		}
	}
}
