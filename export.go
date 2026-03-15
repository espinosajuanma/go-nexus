package nexus

import (
	"fmt"
	"io"
	"strings"
)

func (t *TaxaBlock) write(w io.Writer) error {
	fmt.Fprintln(w, "BEGIN TAXA;")
	fmt.Fprintf(w, "\tDIMENSIONS NTAX=%d;\n", t.Dimensions.Count)

	fmt.Fprint(w, "\tTAXLABELS\n")
	for _, label := range t.TaxLabels {
		fmt.Fprintf(w, "\t\t%s\n", label)
	}
	fmt.Fprintln(w, "\t;")
	fmt.Fprintln(w, "END;")
	return nil
}

func (c *CharactersBlock) write(w io.Writer) error {
	fmt.Fprintln(w, "BEGIN CHARACTERS;")
	fmt.Fprintf(w, "\tDIMENSIONS NCHAR=%d;\n", c.Dimensions.NChar)

	// Format command setup
	formatArgs := []string{fmt.Sprintf("DATATYPE=%s", c.Format.DataType)}
	if c.Format.Missing != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("MISSING=%s", c.Format.Missing))
	}
	if c.Format.Gap != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("GAP=%s", c.Format.Gap))
	}
	if c.Format.Symbols != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("SYMBOLS=\"%s\"", c.Format.Symbols))
	}
	fmt.Fprintf(w, "\tFORMAT %s;\n", strings.Join(formatArgs, " "))

	// Charstatelabels
	if len(c.CharStateLabels) > 0 {
		fmt.Fprintln(w, "\tCHARSTATELABELS")
		// Note: In a real implementation, you'd want to sort the keys to ensure deterministic output
		for idx, name := range c.CharStateLabels {
			fmt.Fprintf(w, "\t\t%d %s,\n", idx, name)
		}
		fmt.Fprintln(w, "\t;")
	}

	// Matrix
	fmt.Fprintln(w, "\tMATRIX")
	for _, row := range c.Matrix {
		fmt.Fprintf(w, "\t%s\t%s\n", row.TaxonName, row.Data)
	}
	fmt.Fprintln(w, "\t;")
	fmt.Fprintln(w, "END;")

	return nil
}
