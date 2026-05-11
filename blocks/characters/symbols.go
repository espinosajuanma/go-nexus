package characters

const (
	InternalMissing = "\x00"
	InternalGap     = "\x01"
)

// DefaultSymbols provides the standard allowed character states for NEXUS datatypes.
var DefaultSymbols = map[DataType][]string{
	DNA:        {"A", "C", "G", "T"},
	RNA:        {"A", "C", "G", "U"},
	Nucleotide: {"A", "C", "G", "T", "U"},
	Protein:    {"A", "R", "N", "D", "C", "Q", "E", "G", "H", "I", "L", "K", "M", "F", "P", "S", "T", "W", "Y", "V", "*"}, // * is often used for stop codons
	Standard:   {"0", "1"},
}

// DefaultEquates provides the standard IUPAC ambiguity mappings.
// When a parser sees 'R' in a DNA matrix, it knows it means '(A G)'.
var DefaultEquates = map[DataType]map[string]string{
	DNA: {
		"R": "(A G)", "Y": "(C T)", "M": "(A C)", "K": "(G T)",
		"S": "(C G)", "W": "(A T)", "H": "(A C T)", "B": "(C G T)",
		"V": "(A C G)", "D": "(A G T)", "N": "(A C G T)",
	},
	RNA: {
		"R": "(A G)", "Y": "(C U)", "M": "(A C)", "K": "(G U)",
		"S": "(C G)", "W": "(A U)", "H": "(A C U)", "B": "(C G U)",
		"V": "(A C G)", "D": "(A G U)", "N": "(A C G U)",
	},
	Nucleotide: {
		"R": "(A G)", "Y": "(C T U)", "M": "(A C)", "K": "(G T U)",
		"S": "(C G)", "W": "(A T U)", "H": "(A C T U)", "B": "(C G T U)",
		"V": "(A C G)", "D": "(A G T U)", "N": "(A C G T U)",
	},
	Protein: {
		"B": "(D N)",                                     // Aspartic Acid or Asparagine
		"Z": "(E Q)",                                     // Glutamic Acid or Glutamine
		"X": "(A R N D C Q E G H I L K M F P S T W Y V)", // Any amino acid
	},
}
