package characters

// DataType acts as an enum for standard NEXUS format data types.
type DataType string

const (
	Standard   DataType = "STANDARD"
	DNA        DataType = "DNA"
	RNA        DataType = "RNA"
	Nucleotide DataType = "NUCLEOTIDE"
	Protein    DataType = "PROTEIN"
	Continuous DataType = "CONTINUOUS"
)

// StateType tracks if a character is a single state, polymorphic, or uncertain.
type StateType string

const (
	StateSingle      StateType = "SINGLE"
	StatePolymorphic StateType = "POLYMORPHIC"
	StateUncertain   StateType = "UNCERTAIN"
)

// Dimensions captures the NCHAR value for the CHARACTERS block.
type Dimensions struct {
	NChar int
}

// Format specifies the format of the data MATRIX.
type Format struct {
	DataType DataType
	Missing  string
	Gap      string
	Symbols  string
}

// Character represents a single column in the NEXUS matrix.
type Character struct {
	Index       int
	Name        string
	StateLabels []string
}

// TaxonReference represents a single row in the NEXUS matrix.
type TaxonReference struct {
	Index  int
	Name   string
	parent *CharactersBlock // Back-pointer for fluent API
}

// CharacterState holds individualized information about a single character state.
type CharacterState struct {
	Index int
	Type  StateType
	Value []string
	Label string
}

// MatrixRow holds the sequence/data for a single taxon.
type MatrixRow struct {
	TaxonName string
	States    []CharacterState
}
