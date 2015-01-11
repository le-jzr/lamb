package lamb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"unicode"
)

type Writer interface {
	io.Writer
}

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

	ctx.Externals["__load"] = externalLoad
	ctx.Externals["__define"] = externalDefine
	ctx.Externals["__print"] = externalPrint

	ctx.namespace["__load"] = &ExternalFuncCall{2, 1, "__load", nil}
	ctx.namespace["__define"] = &ExternalFuncCall{2, 2, "__define", nil}
	ctx.namespace["__print"] = &ExternalFuncCall{2, 1, "__print", nil}
	return ctx
}

func (ctx *Context) Get(name string) Expression {
	ctx.ns_lock.Lock()
	defer ctx.ns_lock.Unlock()
	return ctx.namespace[name]
}

func (ctx *Context) Set(name string, exp Expression) {
	ctx.ns_lock.Lock()
	defer ctx.ns_lock.Unlock()
	ctx.namespace[name] = exp
}

// Used to split the stream on spaces, parentheses and semicolons.
func readNext(r *bufio.Reader) (string, error) {
	buf := new(bytes.Buffer)

	for {
		c, _, err := r.ReadRune()
		if err != nil {
			return buf.String(), err
		}

		if !unicode.IsSpace(c) {
			r.UnreadRune()
			break
		}
	}

	for {
		c, _, err := r.ReadRune()
		if err != nil || unicode.IsSpace(c) {
			return buf.String(), err
		}

		if c == ';' || c == '(' || c == ')' {
			str := buf.String()
			if c == '(' && str == "!" {
				return "!(", nil
			}

			if buf.Len() == 0 {
				return fmt.Sprintf("%c", c), nil
			}
			r.UnreadRune()
			return buf.String(), nil
		}

		buf.WriteRune(c)
	}
}

func ParseExpression(r io.Reader, interactive_ctx *Context, debugTrace bool) Expression {
	buffered, ok := r.(*bufio.Reader)
	if !ok {
		buffered = bufio.NewReader(r)
	}

	var current Expression

	for {
		next, err := readNext(buffered)
		if err != nil {
			// TODO: error handling
			break
		}

		if next == "" {
			break
		}

		switch {
		case next == ";":
			current = NewApplication(NewLambda("", true, nil), current)

		case next == "(":
			current = NewApplication(current, ParseExpression(buffered, nil, false))

		case next == ")":
			return current

		case next == "!(":
			current = NewApplication(current, ForceEval{ParseExpression(buffered, nil, false)})

		case next == "__extern":
			current = NewApplication(current, NewExternalFuncCall())

		case next[0] == '\\':
			// Lambda abstraction.

			name := next[1:]
			strict := name[0] == '!'
			if strict {
				name = name[1:]
			}
			return NewApplication(current, NewLambda(name, strict, ParseExpression(buffered, nil, false)))

		case unicode.IsDigit(rune(next[0])):
			i, err := strconv.ParseInt(next, 0, 64)
			if err != nil {
				// TODO: error handling
			}
			current = NewApplication(current, Int{i})

		case next[0] == '"':
			s, err := strconv.Unquote(next)
			if err != nil {
				// TODO: error handling
			}
			current = NewApplication(current, String(s))

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
				repeat := true
				for repeat {
					current, repeat = FullReduce(interactive_ctx, current)
				}
			} else {
				// Debug trace!
				//
				// We reduce by a single step and print
				// the result, until the expression is
				// completely reduced.

				WriteTo(current, os.Stderr)
				os.Stderr.Write([]byte{'\n'})

				for {
					var ok bool
					current, ok = Reduce(interactive_ctx, current)
					if !ok {
						break
					}

					WriteTo(current, os.Stderr)
					os.Stderr.Write([]byte{'\n'})
				}
			}
		}
	}

	return current
}
