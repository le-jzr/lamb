
// Define the define.
__extern "__define" 2 "define" !(__extern "__define" 2);

define "pair" (\x \y \f f x y);
define "fst" (\x \y x);
define "snd" (\x ());

define "true" !(fst);
define "false" !(snd);
define "ite" (\x \y \z x y z);

define "cons" (\x \y (pair true (pair x y)));
define "nil" !(pair false ());
define "is_nil" !(fst);
define "head" (\x __fst (__snd x));
define "tail" (\x __snd (__snd x));

define "load" !(__extern "__load" 1);

define "+" !(__extern "__int_plus" 2);
define "-" !(__extern "__int_minus" 2);
define "*" !(__extern "__int_multiply" 2);
define "/" !(__extern "__int_divide" 2);
define "print" !(__extern "__print" 1);

define "true" !(__extern "__true" 0);
define "false" !(__extern "__false" 0);
define "_ite" !(__extern "__ite" 3);
define "ite" (\!x _ite x);

define "fact" (\x ite (<= x 0) 1 (* x (fact (- x 1))));

define "infinilist" (\x \y cons x (infinilist (y x) y));

define "seq" (infinilist 0 (+ 1));

define "first" (\x \y ite (or (<= x 0) (is_nil y)) nil (cons (head y) (first (- x 1) (tail y))));


define "lambda" !(__extern "__lambda" 1);
define "n2str" !(__extern "__n2str" 1);
define "str2n" !(__extern "__str2n" 1);

define "if" (\cond \_ \tt \_ \ff ite cond tt ff);

define "church-combinator" (\f (\x f (x x)) (\x f (x x)));
define "let"    (\x \_ \y \_ \z lambda (n2str x) z y);
define "letrec" (\x \_ \y \_ \z lambda (n2str x) z (church-combinator (lambda (n2str x) y)));

letrec fact2 = (\!x \!y if (<= y 0) then x else (fact2 (* x y) (- y 1))) in (define "fact" (\x fact2 1 x));

print (first 10 seq);


type Expression interface {
	Substitute(name string, value Expression) Expression
}

type BuiltinFunction int
const (
	_FORCE_EVAL = BuiltinFunction(iota)
)

type ExternalFuncCall = struct {
	args_given int
	args_remaining int
	
	func_name string
	args []Expression
}

// Lambda is represented as a pseudo-function: it is applied to the inner expression using function application.
// This must be taken into account when substituting.
type Lambda struct {
	name string
	strict bool
}

/*
func substitute(tmpl Expression, name string, val Expression) Expression {
	// TODO: this can be optimized by using string table and free variable bitmap.
	
	switch tt := tmpl.(type) {
	case int64, string, BuiltinFunction, *Bind:
		return tmpl
	
	case *ExternalFuncCall:
		nargs := make([]Expression, len(tt.args))
		for i, arg := range tt.args {
			nargs[i] = substitute(arg, name, val)
		}
		return &ExternalFuncCall{tt.args_given, tt.args_remaining, tt.func_name, nargs}
	
	case Name:
		if string(tt) == name {
			return val
		} else {
			return tmpl
		}
	
	case *FunctionApplication:
		ttt, ok := tt.fnc.(*Bind)
		if ok && ttt.name == name {
			return tmpl
		}

		// TODO: Avoid free variable capture. Depends on constructing free variable set.
		
		return &FunctionApplication{substitute(tt.fnc, name, val), substitute(tt.val, name, val)}
	}
	
	panic("missing case")
}*/

func reduceFunction(fnc, val) (nex Expression, changed bool) {
	switch fnc := e.fnc.(type) {
	case int64:
		// TODO: proper error
		panic("Function call on int.")
	case *ExternalFuncCall:
		// This is a special case, we must append arguments until the expected number is reached.
	
		switch val := e.val.(type) {
		case int64:
			// compute
		
		}
	case Name:
	case *Bind:
	
	
	case BuiltinFunction:
		switch fnc {
		case _FORCE_EVAL: 
			return fullReduce(val), true
		}
	}
	e.val
}

func reduce(ex Expression) (nex Expression, changed bool) {
	if ex == nil {
		return ex, false
	}
	
	switch e := ex.(type) {
	case BuiltinFunction, *Bind, int64, *PartialBinaryOp:
		// Irreducible
		return ex, false;
	case *FunctionApplication:
		return reduceFunction(e.fnc, e.val)
	case Name:
		val, ok := _namespace[string(e)]
		if !ok {
			return e, false
		}
		return val, true
	}
}


func readNext(r io.Reader) string {
	sbldr := new(StringBuilder)
	
	Char()
	
	sbldr.append()
}

func interpret(r io.Reader) {

	var current

	for {
		next := readNext(r)
		if r == "" {
			break
		}
		
		switch next {
		case ";":
			seq := applyFunction(&Lambda{"", true}, nil)
			current = applyFunction(seq, current)
			
		case "(":
			push(current)
			current = nil
			
		case ")":
			fn := pop()
				
			current = applyFunction(fn, current)

			for isForceEval(fn) || isBind(fn) {
				fn = pop()
				current = applyFunction(fn, current)
			}
			
		case "!(":
			push(current)
			push(_FORCE_EVAL)
			current = nil
		
		case "__extern":
			// TODO
		
		default:
			if next[0] == '\' {
				push(current)
				if next[1] == '!' {
					push(bind(next[2:]))
				} else {
					push(bind(next[1:]))
				}
				current = nil
			}
		
			if isInt(next) {
				current = applyFunction(current, parseInt(next))
			}
			
			current = applyFunction(current, name(next))
		}
	 
		if emptyStack() {
			current = fullReduce(current)
		}
	}
}
