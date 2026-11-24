package dto

// Document is the source submitted to Object's read-only preparation use case.
type Document struct {
	Filename string
	Data     []byte
}

type Candidate struct {
	CategoryPath []string
	Name         string
	Quantity     *string
	Unit         *string
	Location     *string
	Quote        string
}

type CandidateResult struct {
	Candidates  []Candidate
	RawResponse []byte
}
