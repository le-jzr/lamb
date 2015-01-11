package lamb

type Expression interface {
	Apply(arg Expression) (Expression, bool)
	Substitute(name string, value Expression) Expression
	Reduce(ctx *Context) (Expression, bool)
	FullReduce(ctx *Context) (Expression, bool)
	WriteTo(w Writer)
}

func apply(f Expression, arg Expression) (Expression, bool) {
	if e == nil {
		return arg
	}
	return f.Apply(arg)
}

func substitute(e Expression, name string, value Expression) Expression {
	if e == nil {
		return nil
	}
	return e.Substitute(name, value)
}

func reduce(ctx *Context, e Expression) (Expression, bool) {
	if e == nil {
		return nil, false
	}
	return e.Reduce(ctx)
}

func fullReduce(ctx *Context, e Expression) (Expression, bool) {
	if e == nil {
		return nil, false
	}
	return e.FullReduce(ctx)
}

func writeTo(e Expression, w Writer) {
	if e == nil {
		w.Write([]byte("nil"))
	} else {
		e.WriteTo(w)
	}
}
