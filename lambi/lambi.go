// lambi is a command-line interface for lamb interpreter.
//
// Usage: lambi [-d] [-i] [-b] [--] [filename1] [filename2] ...
//
//	-d
//		Debug mode.
//
//		When expression is being evaluated, lambi prints the
//		expression after every reduction step to standard error
//              output.
//
//	-i
//		Interactive mode (default).
//
//		After all the files provided on the command line
//		are processed, further input is read from standard
//		input.
//
//	-b
//		Batch mode.
//
//		No reading from standard input.
//		If multiple -i and -b switches are given,
//		the last one counts.
//
//
//	--hoard
//		Hoarder mode.
//
//		Input is fully read before it starts being evaluated.
//		Probably only useful with -d.
//		Can use more memory than the default, which is
//		to evaluate program fragments as soon as possible.
//
//
package main

import (
	lamb ".."
	"fmt"
	"os"
)

const EXIT_FAILURE = 1

func main() {

	debug := false
	hoard := false
	batch := false

	files := []string{}

L:
	for i, arg := range os.Args[1:] {
		switch arg {
		case "-d":
			debug = true

		case "-i":
			batch = false

		case "-b":
			batch = true

		case "--hoard":
			hoard = true

		case "--":
			files = append(files, os.Args[i+1:]...)
			break L

		default:
			files = append(files, arg)
		}
	}

	if !hoard {
		ctx := lamb.NewContext()

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot open file %s: %s\n", file, err.Error())
				if batch {
					os.Exit(EXIT_FAILURE)
				}
				break
			}
			fmt.Fprintf(os.Stderr, "Loading file %s\n", file)

			lamb.ParseExpression(f, ctx, debug)
			f.Close()
		}

		if !batch {
			lamb.ParseExpression(os.Stdin, ctx, debug)
		}

		return
	}

	// The rest of the function is Hoarder mode.

	current := lamb.Expression(nil)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot open file %s: %s\n", file, err.Error())
			if batch {
				os.Exit(EXIT_FAILURE)
			}
			break
		}
		fmt.Fprintf(os.Stderr, "Loading file %s\n", file)

		current = lamb.NewApplication(lamb.NewLambda("", true, nil), current)
		current = lamb.NewApplication(current, lamb.ParseExpression(f, nil, false))

		f.Close()
	}

	if !batch {
		current = lamb.NewApplication(lamb.NewLambda("", true, nil), current)
		current = lamb.NewApplication(current, lamb.ParseExpression(os.Stdin, nil, false))
	}

	ctx := lamb.NewContext()

	if debug {
		lamb.WriteTo(current, os.Stderr)
		os.Stderr.Write([]byte{'\n'})

		for {
			var ok bool
			current, ok = lamb.Reduce(ctx, current)
			if !ok {
				break
			}

			lamb.WriteTo(current, os.Stderr)
			os.Stderr.Write([]byte{'\n'})
		}
	} else {
		repeat := true
		for repeat {
			current, repeat = lamb.FullReduce(ctx, current)
		}
	}
}
