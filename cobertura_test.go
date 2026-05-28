package gobertura_test

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/aereal/gobertura"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// ignoreXMLName excludes Coverage.XMLName from comparison, as it is populated automatically during encoding.
var ignoreXMLName = cmpopts.IgnoreFields(gobertura.Coverage{}, "XMLName")

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want  *gobertura.Coverage
		name  string
		input string
	}{
		{
			name:  "single file, single block, fully covered",
			input: "mode: set\ngithub.com/aereal/example/foo.go:1.1,3.2 1 1\n",
			want: &gobertura.Coverage{
				XMLName:         xml.Name{Local: "coverage"},
				Version:         "gobertura",
				Timestamp:       testTime.Unix(),
				LinesValid:      3,
				LinesCovered:    3,
				LineRate:        1.0,
				BranchRate:      0,
				BranchesValid:   0,
				BranchesCovered: 0,
				Sources:         gobertura.Sources{Sources: []string{"."}},
				Packages: []gobertura.Package{
					{
						Name:       "github.com/aereal/example",
						LineRate:   1.0,
						BranchRate: 0,
						Classes: []gobertura.Class{
							{
								Name:       "foo.go",
								Filename:   "github.com/aereal/example/foo.go",
								LineRate:   1.0,
								BranchRate: 0,
								Lines: []gobertura.Line{
									{Number: 1, Hits: 1, Branch: false},
									{Number: 2, Hits: 1, Branch: false},
									{Number: 3, Hits: 1, Branch: false},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "single file, multiple blocks, partially covered",
			input: "mode: set\n" +
				"github.com/aereal/example/foo.go:1.1,2.2 1 1\n" +
				"github.com/aereal/example/foo.go:3.1,4.2 1 0\n",
			want: &gobertura.Coverage{
				Version:      "gobertura",
				Timestamp:    testTime.Unix(),
				LinesValid:   4,
				LinesCovered: 2,
				LineRate:     0.5,
				Sources:      gobertura.Sources{Sources: []string{"."}},
				Packages: []gobertura.Package{
					{
						Name:     "github.com/aereal/example",
						LineRate: 0.5,
						Classes: []gobertura.Class{
							{
								Name:     "foo.go",
								Filename: "github.com/aereal/example/foo.go",
								LineRate: 0.5,
								Lines: []gobertura.Line{
									{Number: 1, Hits: 1},
									{Number: 2, Hits: 1},
									{Number: 3, Hits: 0},
									{Number: 4, Hits: 0},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple files in the same package yield a single <package>",
			input: "mode: set\n" +
				"github.com/aereal/example/a.go:1.1,1.10 1 1\n" +
				"github.com/aereal/example/b.go:1.1,1.10 1 0\n",
			want: &gobertura.Coverage{
				Version:      "gobertura",
				Timestamp:    testTime.Unix(),
				LinesValid:   2,
				LinesCovered: 1,
				LineRate:     0.5,
				Sources:      gobertura.Sources{Sources: []string{"."}},
				Packages: []gobertura.Package{
					{
						Name:     "github.com/aereal/example",
						LineRate: 0.5,
						Classes: []gobertura.Class{
							{
								Name:     "a.go",
								Filename: "github.com/aereal/example/a.go",
								LineRate: 1.0,
								Lines:    []gobertura.Line{{Number: 1, Hits: 1}},
							},
							{
								Name:     "b.go",
								Filename: "github.com/aereal/example/b.go",
								LineRate: 0.0,
								Lines:    []gobertura.Line{{Number: 1, Hits: 0}},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple packages yield multiple <package> elements",
			input: "mode: set\n" +
				"github.com/aereal/example/foo.go:1.1,1.10 1 1\n" +
				"github.com/aereal/example/sub/bar.go:1.1,1.10 1 0\n",
			want: &gobertura.Coverage{
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
								Lines:    []gobertura.Line{{Number: 1, Hits: 1}},
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
								Lines:    []gobertura.Line{{Number: 1, Hits: 0}},
							},
						},
					},
				},
			},
		},
		{
			name:  "mode:count reflects execution count in hits",
			input: "mode: count\ngithub.com/aereal/example/foo.go:1.1,1.10 1 42\n",
			want: &gobertura.Coverage{
				Version:      "gobertura",
				Timestamp:    testTime.Unix(),
				LinesValid:   1,
				LinesCovered: 1,
				LineRate:     1.0,
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
								Lines:    []gobertura.Line{{Number: 1, Hits: 42}},
							},
						},
					},
				},
			},
		},
		{
			name:  "empty profile",
			input: "mode: set\n",
			want: &gobertura.Coverage{
				Version:      "gobertura",
				Timestamp:    testTime.Unix(),
				LinesValid:   0,
				LinesCovered: 0,
				LineRate:     0,
				Sources:      gobertura.Sources{Sources: []string{"."}},
				Packages:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gobertura.Parse(strings.NewReader(tt.input), testTime)
			if err != nil {
				t.Fatalf("Parse2() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, ignoreXMLName); diff != "" {
				t.Errorf("Parse2() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
