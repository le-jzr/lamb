package lamb

type String string

func (e String) Apply(arg Expression) (Expression, bool) {
	return nil, false
}

func (e String) Substitute(name string, value Expression) Expression {
	return e
}

func (e String) Reduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e String) FullReduce(ctx *Context) (Expression, bool) {
	return e, false
}

func (e String) WriteTo(w Writer) {
	// FIXME: Escaping.
	fmt.Fprintf(w, "\"%s\"", string(e))
}
