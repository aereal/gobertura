package gobertura

import (
	"encoding/xml"
	"io"
)

const coberturaDoctype = `<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">` + "\n"

// Write encodes cov as Cobertura XML to w.
// The output includes the XML declaration and the Cobertura DOCTYPE declaration.
func Write(w io.Writer, cov *Coverage) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	if _, err := io.WriteString(w, coberturaDoctype); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if err := enc.Encode(cov); err != nil {
		return err
	}
	return enc.Close()
}
