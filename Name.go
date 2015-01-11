package lamb

type Name string

func (e Name) Apply(arg Expression) (Expression, bool) {
	return nil, false
}

func (e Name) Substitute(name string, value Expression) Expression {
	if string(e) == name {
		return value
	}
	return e
}

func (e Name) Reduce(ctx *Context) (Expression, bool) {
	return ctx.Get(string(e)), true
}

func (e Name) FullReduce(ctx *Context) (Expression, bool) {
	expr := ctx.Get(string(e))
	expr_reduced, _ := fullReduce(expr)
	return expr_reduced, true
}

func (e Name) WriteTo(w Writer) {
	w.Write([]byte(string(e)))
}
