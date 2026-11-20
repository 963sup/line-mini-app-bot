package port

import "context"

// Investigation bounds a read-only request for additional source evidence.
type Investigation struct {
	Actor, Repository, SourceRevision string
	AllowedPages                      []int
	StepLimit                         int
}

type Evidence struct {
	Page  int
	Quote string
}

// Investigator returns supporting evidence; objectimport decides whether to accept it.
type Investigator interface {
	Investigate(context.Context, Investigation) ([]Evidence, error)
}
