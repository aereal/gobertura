// Package cli implements the core logic of the gobertura command.
// Input and output are abstracted via io.Reader and io.Writer for testability.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aereal/gobertura"
)

func NewStd() *CLI {
	return New(
		func() (io.ReadCloser, error) { return os.Stdin, nil },
		func() (io.WriteCloser, error) { return os.Stdout, nil },
		os.Stderr,
	)
}

func New(
	openInput func() (io.ReadCloser, error),
	openOutput func() (io.WriteCloser, error),
	errStream io.Writer,
) *CLI {
	return &CLI{
		openInput:  openInput,
		openOutput: openOutput,
		errStream:  errStream,
	}
}

// CLI implements the gobertura command for converting coverage profiles to Cobertura XML.
type CLI struct {
	openInput  func() (io.ReadCloser, error)
	openOutput func() (io.WriteCloser, error)
	errStream  io.Writer
}

// Run parses args and executes the coverage conversion, returning an exit code.
func (c *CLI) Run(args []string) int {
	fs := flag.NewFlagSet("gobertura", flag.ContinueOnError)
	fs.SetOutput(c.errStream)

	var inputPath, outputPath string
	fs.StringVar(&inputPath, "input", "", "input coverage profile `file` path (default: stdin)")
	fs.StringVar(&outputPath, "output", "", "output Cobertura XML `file` path (default: stdout)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if inputPath != "" {
		c.openInput = inputOpener(inputPath)
	}
	if outputPath != "" {
		c.openOutput = outputOpener(outputPath)
	}

	if err := c.run(); err != nil {
		fmt.Fprintln(c.errStream, err.Error())
		return 1
	}

	return 0
}

func inputOpener(path string) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("os.Open: %w", err)
		}
		return f, nil
	}
}

func outputOpener(path string) func() (io.WriteCloser, error) {
	return func() (io.WriteCloser, error) {
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("os.Create: %w", err)
		}
		return f, nil
	}
}

func (c *CLI) run() error {
	r, err := c.openInput()
	if err != nil {
		return fmt.Errorf("failed to open input: %w", err)
	}
	defer r.Close()

	w, err := c.openOutput()
	if err != nil {
		return fmt.Errorf("failed to open output: %w", err)
	}
	defer w.Close()

	cov, err := gobertura.Parse(r, time.Now())
	if err != nil {
		return fmt.Errorf("gobertura.Parse: %w", err)
	}

	if err := gobertura.Write(w, cov); err != nil {
		return fmt.Errorf("gobertura.Write: %w", err)
	}

	return nil
}
