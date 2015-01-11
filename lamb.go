package lamb

type ExternalFunc func(ctx *Context, self_name string, args []Expression) Expression

type Context struct {
	Externals map[string]ExternalFunc

	ns_lock   sync.Mutex
	namespace map[string]Expression
}

func NewContext() *Context {
	ctx := new(Context)
	ctx.Externals = make(map[string]ExternalFunc)
	ctx.namespace = make(map[string]Expression)
}

func (ctx *Context) Get(name string) Expression {
	ns_lock.Lock()
	defer ns_lock.Unlock()
	return ctx.namespace[name]
}

func (ctx *Context) Set(name string, exp Expression) {
	ns_lock.Lock()
	defer ns_lock.Unlock()
	ctx.namespace[name] = exp
}

func readNext(r *bufio.Reader) string {
	sbldr := new(StringBuilder)

	Char()

	sbldr.append()
}

func ParseExpression(r io.Reader, interactive_ctx *Context, debugTrace bool) Expression {
	buffered, ok := r.(*bufio.Reader)
	if !ok {
		buffered = bufio.NewReader(r)
	}

	var current Expression

	for {
		next := readNext(buffered)
		if next == "" {
			break
		}

		switch next {
		case next == ";":
			current = NewApplication(lambda("", true, false, nil), current)

		case next == "(":
			current = NewApplication(current, parseExpression(buffered, nil, false))

		case next == ")":
			return current

		case next == "!(":
			current = NewApplication(current, ForceEval{parseExpression(buffered, nil, false)})

		case next == "__extern":
			current = NewApplication(current, NewExternalFuncCall())

		case next[0] == '\\':
			// Lambda abstraction.

			name := next[1:]
			strict := name[0] == '!'
			if strict {
				name = name[1:]
			}
			return NewApplication(current, NewLambda(name, strict, parseExpression(buffered, nil, false)))

		case isInt(next):
			current = NewApplication(current, parseInt(next))

		case isString(next):
			// Indicated by leaving the initial quote in place.
			// Other than that, the string is already unescaped.

			current = NewApplication(current, String(next[1:]))

		default:
			// Probably a name, then.

			current = NewApplication(current, Name(next))
		}

		if interactive_ctx != nil {
			// In interactive mode, we reduce the partial application immediately.
			// This does not hamper the lazy evaluation,
			// since we assume that an expression to which
			// we can apply an argument is irreducible.

			if !debugTrace {
				current = fullReduce(ctx, current)
			} else {
				// Debug trace!
				//
				// We reduce by a single step and print
				// the result, until the expression is
				// completely reduced.

				writeTo(current, os.Stderr)
				os.Stderr.Write([]byte{'\n'})

				for {
					var ok bool
					current, ok = reduce(ctx, current)
					if !ok {
						break
					}

					writeTo(current, os.Stderr)
					os.Stderr.Write([]byte{'\n'})
				}
			}
		}
	}

	return current
}
