package gobertura

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"path"
	"slices"
	"time"

	"golang.org/x/tools/cover"
)

// Coverage represents the root <coverage> element of a Cobertura XML document.
type Coverage struct {
	XMLName         xml.Name  `xml:"coverage"`
	Version         string    `xml:"version,attr"`
	Sources         Sources   `xml:"sources"`
	Packages        []Package `xml:"packages>package"`
	LineRate        float64   `xml:"line-rate,attr"`
	BranchRate      float64   `xml:"branch-rate,attr"`
	LinesValid      int       `xml:"lines-valid,attr"`
	LinesCovered    int       `xml:"lines-covered,attr"`
	BranchesValid   int       `xml:"branches-valid,attr"`
	BranchesCovered int       `xml:"branches-covered,attr"`
	Complexity      float64   `xml:"complexity,attr"`
	Timestamp       int64     `xml:"timestamp,attr"`
}

// Sources represents the <sources> element.
type Sources struct {
	Sources []string `xml:"source"`
}

// Package represents a <package> element.
type Package struct {
	Name       string  `xml:"name,attr"`
	Classes    []Class `xml:"classes>class"`
	LineRate   float64 `xml:"line-rate,attr"`
	BranchRate float64 `xml:"branch-rate,attr"`
	Complexity float64 `xml:"complexity,attr"`
}

// Class represents a <class> element, corresponding to a single Go source file.
type Class struct {
	Methods    Methods `xml:"methods"`
	Name       string  `xml:"name,attr"`
	Filename   string  `xml:"filename,attr"`
	Lines      []Line  `xml:"lines>line"`
	LineRate   float64 `xml:"line-rate,attr"`
	BranchRate float64 `xml:"branch-rate,attr"`
	Complexity float64 `xml:"complexity,attr"`
}

// Methods represents the <methods/> element. Always empty because Go coverage profiles
// do not include method-level information.
type Methods struct{}

// Line represents a <line> element.
type Line struct {
	Number int  `xml:"number,attr"`
	Hits   int  `xml:"hits,attr"`
	Branch bool `xml:"branch,attr"`
}

// Parse parses go test coverage output into a Coverage value for Cobertura XML output.
// timestamp is used for the <coverage timestamp="..."> attribute.
func Parse(r io.Reader, timestamp time.Time) (*Coverage, error) {
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("cover.ParseProfilesFromReader: %w", err)
	}
	pkgMap := map[string][]*cover.Profile{}
	for _, p := range profiles {
		pkg := path.Dir(p.FileName)
		pkgMap[pkg] = append(pkgMap[pkg], p)
	}

	// Sort package names for deterministic output.
	pkgNames := make([]string, 0, len(pkgMap))
	for name := range pkgMap {
		pkgNames = append(pkgNames, name)
	}
	slices.Sort(pkgNames)

	var (
		totalValid   int
		totalCovered int
		packages     []Package
	)
	for _, pkgName := range pkgNames {
		pkg, valid, covered := convertPackage(pkgName, pkgMap[pkgName])
		totalValid += valid
		totalCovered += covered
		packages = append(packages, pkg)
	}

	return &Coverage{
		Version:         "gobertura",
		Timestamp:       timestamp.Unix(),
		LinesValid:      totalValid,
		LinesCovered:    totalCovered,
		LineRate:        calcLineRate(totalCovered, totalValid),
		BranchRate:      0,
		BranchesValid:   0,
		BranchesCovered: 0,
		Sources:         Sources{Sources: []string{"."}},
		Packages:        packages,
	}, nil
}

// convertPackage converts a group of profiles for a single package into a Package.
func convertPackage(pkgName string, profiles []*cover.Profile) (Package, int, int) {
	sorted := make([]*cover.Profile, len(profiles))
	copy(sorted, profiles)
	slices.SortFunc(sorted, func(a, b *cover.Profile) int {
		return cmp.Compare(a.FileName, b.FileName)
	})

	var (
		totalValid   int
		totalCovered int
		classes      []Class
	)
	for _, p := range sorted {
		cls, valid, covered := convertClass(p)
		totalValid += valid
		totalCovered += covered
		classes = append(classes, cls)
	}

	return Package{
		Name:       pkgName,
		LineRate:   calcLineRate(totalCovered, totalValid),
		BranchRate: 0,
		Classes:    classes,
	}, totalValid, totalCovered
}

// convertClass converts a single file's coverage profile into a Class.
// Each ProfileBlock is expanded line by line; when multiple blocks overlap on the same line,
// the maximum hit count is used.
func convertClass(p *cover.Profile) (Class, int, int) {
	lineHits := map[int]int{}
	for _, block := range p.Blocks {
		for ln := block.StartLine; ln <= block.EndLine; ln++ {
			if existing, ok := lineHits[ln]; !ok || block.Count > existing {
				lineHits[ln] = block.Count
			}
		}
	}

	lineNums := make([]int, 0, len(lineHits))
	for ln := range lineHits {
		lineNums = append(lineNums, ln)
	}
	slices.Sort(lineNums)

	var (
		lines   []Line
		valid   int
		covered int
	)
	for _, ln := range lineNums {
		hits := lineHits[ln]
		lines = append(lines, Line{Number: ln, Hits: hits, Branch: false})
		valid++
		if hits > 0 {
			covered++
		}
	}

	return Class{
		Name:       path.Base(p.FileName),
		Filename:   p.FileName,
		LineRate:   calcLineRate(covered, valid),
		BranchRate: 0,
		Lines:      lines,
	}, valid, covered
}

// calcLineRate returns the ratio of covered to valid lines, rounded to 4 decimal places.
// Returns 0 if valid is 0.
func calcLineRate(covered, valid int) float64 {
	if valid == 0 {
		return 0
	}
	return math.Round(float64(covered)/float64(valid)*10000) / 10000
}
