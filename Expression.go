package lamb

type Expression interface {
	Apply(arg Expression) (Expression, bool)
	Substitute(name string, value Expression) Expression
	// Returns true if the expression changed.
	Reduce(ctx *Context) (Expression, bool)
	// Returns true if the reduction should be repeated.
	// This function therefore does not truly guarantee that
	// the expression is fully reduced on return,
	// as that allows us to efficiently handle tail recursion.
	FullReduce(ctx *Context) (Expression, bool)
	WriteTo(w Writer)
}

func Apply(fnc Expression, arg Expression) (Expression, bool) {
	if fnc == nil {
		return arg, true
	}
	return fnc.Apply(arg)
}

func Substitute(e Expression, name string, value Expression) Expression {
	if e == nil {
		return nil
	}
	return e.Substitute(name, value)
}

func Reduce(ctx *Context, e Expression) (Expression, bool) {
	if e == nil {
		return nil, false
	}
	return e.Reduce(ctx)
}

func FullReduce(ctx *Context, e Expression) (Expression, bool) {
	if e == nil {
		return nil, false
	}
	return e.FullReduce(ctx)
}

func WriteTo(e Expression, w Writer) {
	if e == nil {
		w.Write([]byte("nil"))
	} else {
		e.WriteTo(w)
	}
}
