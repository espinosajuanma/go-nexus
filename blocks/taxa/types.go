package taxa

// SetFormat defines the allowed formatting options for Sets and Partitions.
type SetFormat string

const (
	StandardFormat SetFormat = "STANDARD"
	VectorFormat   SetFormat = "VECTOR"
)

// TaxSet represents a defined collection of taxa.
type TaxSet struct {
	Format   SetFormat
	TaxaList []string
}

// TaxPartition divides taxa into mutually exclusive subsets.
type TaxPartition struct {
	Format  SetFormat
	Subsets map[string][]string
}
