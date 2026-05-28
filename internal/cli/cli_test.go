package cli_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aereal/gobertura/internal/cli"
	"github.com/google/go-cmp/cmp"
)

const validProfile = "mode: set\n" +
	"github.com/aereal/example/foo.go:1.10,3.2 1 1\n" +
	"github.com/aereal/example/foo.go:5.10,7.2 1 0\n"

const invalidProfile = "github.com/aereal/example/foo.go:1.10,3.2 1 1\n"

type bufferCloser struct {
	*bytes.Buffer
}

func (*bufferCloser) Close() error { return nil }

func newCLI(stdin string) (*cli.CLI, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	outBuf := &bufferCloser{&stdout}
	c := cli.New(
		func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(stdin)), nil },
		func() (io.WriteCloser, error) { return outBuf, nil },
		&stderr,
	)
	return c, &stdout, &stderr
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "profile-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	return f.Name()
}

func assertCoberturaXML(t *testing.T, output string) {
	t.Helper()
	for _, marker := range []string{
		`<?xml`,
		`<!DOCTYPE coverage`,
		`<coverage `,
		`<package `,
		`<line `,
	} {
		if !strings.Contains(output, marker) {
			t.Errorf("output does not contain %q\n--- output ---\n%s", marker, output)
		}
	}
}

func TestCLI_Run(t *testing.T) {
	t.Parallel()

	t.Run("stdin to stdout, happy path", func(t *testing.T) {
		t.Parallel()
		c, stdout, _ := newCLI(validProfile)
		if code := c.Run([]string{}); code != 0 {
			t.Fatalf("Run() = %d, want 0", code)
		}
		assertCoberturaXML(t, stdout.String())
	})

	t.Run("-input flag reads from file", func(t *testing.T) {
		t.Parallel()
		inputFile := writeTempFile(t, validProfile)
		c, stdout, _ := newCLI("")
		if code := c.Run([]string{"-input", inputFile}); code != 0 {
			t.Fatalf("Run() = %d, want 0", code)
		}
		assertCoberturaXML(t, stdout.String())
	})

	t.Run("-output flag writes to file", func(t *testing.T) {
		t.Parallel()
		outputFile := filepath.Join(t.TempDir(), "coverage.xml")
		c, _, _ := newCLI(validProfile)
		if code := c.Run([]string{"-output", outputFile}); code != 0 {
			t.Fatalf("Run() = %d, want 0", code)
		}
		data, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		assertCoberturaXML(t, string(data))
	})

	t.Run("invalid profile: exit 1 with error on stderr", func(t *testing.T) {
		t.Parallel()
		c, _, stderr := newCLI(invalidProfile)
		if code := c.Run([]string{}); code != 1 {
			t.Fatalf("Run() = %d, want 1", code)
		}
		want := "gobertura.Parse: cover.ParseProfilesFromReader: bad mode line: github.com/aereal/example/foo.go:1.10,3.2 1 1\n"
		if diff := cmp.Diff(want, stderr.String()); diff != "" {
			t.Errorf("stderr (-want, +got):\n%s", diff)
		}
	})

	t.Run("-input with nonexistent file: exit 1 with error on stderr", func(t *testing.T) {
		t.Parallel()
		c, _, stderr := newCLI("")
		if code := c.Run([]string{"-input", "/nonexistent/coverage.out"}); code != 1 {
			t.Fatalf("Run() = %d, want 1", code)
		}
		want := "failed to open input: os.Open: open /nonexistent/coverage.out: no such file or directory\n"
		if diff := cmp.Diff(want, stderr.String()); diff != "" {
			t.Errorf("stderr (-want, +got):\n%s", diff)
		}
	})

	t.Run("-output with unwritable path: exit 1 with error on stderr", func(t *testing.T) {
		t.Parallel()
		// Use a path under a nonexistent directory to make os.Create fail.
		badOutput := filepath.Join(t.TempDir(), "nonexistent", "coverage.xml")
		c, _, stderr := newCLI(validProfile)
		if code := c.Run([]string{"-output", badOutput}); code != 1 {
			t.Fatalf("Run() = %d, want 1", code)
		}
		wantPrefix := "failed to open output: os.Create: open"
		got := stderr.String()
		if !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("stderr:\n\twant prefix: %s\n\t got: %s", wantPrefix, got)
		}
	})

	t.Run("-h flag: exit 0", func(t *testing.T) {
		t.Parallel()
		c, _, _ := newCLI("")
		if code := c.Run([]string{"-h"}); code != 0 {
			t.Fatalf("Run() = %d, want 0", code)
		}
	})
}
