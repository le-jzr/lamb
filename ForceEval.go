package lamb

// Used to interpret '!' for strict evaluation.
type ForceEval struct {
	body Expression
}

func (e ForceEval) Apply(arg Expression) (Expression, bool) {
	// ForceEval is reducible.
	return nil, false
}

func (e ForceEval) Substitute(name string, value Expression) Expression {
	return ForceEval{Substitute(e.body, name, value)}
}

func (e ForceEval) Reduce(ctx *Context) (Expression, bool) {
	reduced, ok := Reduce(ctx, e.body)
	if ok {
		return ForceEval{reduced}, true
	}

	return reduced, true
}

func (e ForceEval) FullReduce(ctx *Context) (Expression, bool) {
	return FullReduce(ctx, e.body)
}

func (e ForceEval) WriteTo(w Writer) {
	w.Write([]byte{'!', '('})
	WriteTo(e.body, w)
	w.Write([]byte{')'})
}
