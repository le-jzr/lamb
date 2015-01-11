package lamb

import (
	"os"
)

func externalLoad(ctx *Context, self_name string, args []Expression) Expression {
	if len(args) != 1 {
		panic("Bad number of arguments to __load.")
	}

	file, err := os.Open(string(args[0].(String)))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	return ParseExpression(file, nil, false)
}

func externalDefine(ctx *Context, self_name string, args []Expression) Expression {
	if len(args) != 2 {
		panic("Bad number of arguments to __define.")
	}

	ctx.Set(string(args[0].(String)), args[1])
	return nil
}

func externalPrint(ctx *Context, self_name string, args []Expression) Expression {
	if len(args) != 1 {
		panic("Bad number of arguments to __define.")
	}

	WriteTo(args[0], os.Stdout)
	return &ExternalFuncCall{2, 1, self_name, nil}
}
