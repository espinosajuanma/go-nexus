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
