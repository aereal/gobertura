package gobertura_test

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/aereal/gobertura"
	"github.com/google/go-cmp/cmp"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cov  *gobertura.Coverage
		name string
	}{
		{
			name: "single file, multiple lines, partially covered",
			cov: &gobertura.Coverage{
				Version:         "gobertura",
				Timestamp:       testTime.Unix(),
				LinesValid:      2,
				LinesCovered:    1,
				LineRate:        0.5,
				BranchRate:      0,
				BranchesValid:   0,
				BranchesCovered: 0,
				Sources:         gobertura.Sources{Sources: []string{"."}},
				Packages: []gobertura.Package{
					{
						Name:       "github.com/aereal/example",
						LineRate:   0.5,
						BranchRate: 0,
						Classes: []gobertura.Class{
							{
								Name:       "foo.go",
								Filename:   "github.com/aereal/example/foo.go",
								LineRate:   0.5,
								BranchRate: 0,
								Lines: []gobertura.Line{
									{Number: 1, Hits: 1, Branch: false},
									{Number: 2, Hits: 0, Branch: false},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple packages",
			cov: &gobertura.Coverage{
				Version:      "gobertura",
				Timestamp:    testTime.Unix(),
				LinesValid:   2,
				LinesCovered: 1,
				LineRate:     0.5,
				Sources:      gobertura.Sources{Sources: []string{"."}},
				Packages: []gobertura.Package{
					{
						Name:     "github.com/aereal/example",
						LineRate: 1.0,
						Classes: []gobertura.Class{
							{
								Name:     "foo.go",
								Filename: "github.com/aereal/example/foo.go",
								LineRate: 1.0,
								Lines: []gobertura.Line{
									{Number: 1, Hits: 1, Branch: false},
								},
							},
						},
					},
					{
						Name:     "github.com/aereal/example/sub",
						LineRate: 0.0,
						Classes: []gobertura.Class{
							{
								Name:     "bar.go",
								Filename: "github.com/aereal/example/sub/bar.go",
								LineRate: 0.0,
								Lines: []gobertura.Line{
									{Number: 1, Hits: 0, Branch: false},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty coverage",
			cov: &gobertura.Coverage{
				Version:   "gobertura",
				Timestamp: testTime.Unix(),
				Sources:   gobertura.Sources{Sources: []string{"."}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := gobertura.Write(&buf, tt.cov); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			output := buf.String()

			if !strings.HasPrefix(output, "<?xml") {
				t.Errorf("output does not start with XML declaration:\n%s", output)
			}

			const wantDoctype = `<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">`
			if !strings.Contains(output, wantDoctype) {
				t.Errorf("output does not contain DOCTYPE declaration:\n%s", output)
			}

			// Round-trip: re-parse the written XML and verify it matches the original Coverage.
			dec := xml.NewDecoder(strings.NewReader(output))
			var got gobertura.Coverage
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("xml.Decode() error = %v", err)
			}
			if diff := cmp.Diff(tt.cov, &got, ignoreXMLName); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
