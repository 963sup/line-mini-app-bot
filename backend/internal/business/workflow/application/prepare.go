package application

import "context"

// Scope is set by the calling use case after it verifies the request identity.
type Scope struct {
	Actor, Target string
}

// Preparation connects a use case's read operation, optional knowledge retrieval,
// optional agent investigation and result builder. It does not execute commands.
type Preparation[Input, Data, Knowledge, Evidence, Output any] struct {
	Read          func(context.Context, Scope, Input) (Data, error)
	NeedKnowledge func(Data) bool
	Retrieve      func(context.Context, Scope, Input, Data) (Knowledge, error)
	NeedAgent     func(Data, *Knowledge) bool
	Investigate   func(context.Context, Scope, Input, Data, *Knowledge) (Evidence, error)
	Build         func(Data, *Knowledge, *Evidence) (Output, error)
}

// Prepare runs the required read first. Optional steps are selected by
// application predicates; none means that only the read and build run.
func Prepare[I, D, K, E, O any](ctx context.Context, scope Scope, input I, plan Preparation[I, D, K, E, O]) (O, error) {
	var result O
	if scope.Actor == "" || scope.Target == "" || plan.Read == nil || plan.Build == nil ||
		(plan.NeedKnowledge == nil) != (plan.Retrieve == nil) ||
		(plan.NeedAgent == nil) != (plan.Investigate == nil) {
		return result, ErrConflict
	}
	data, err := plan.Read(ctx, scope, input)
	if err != nil {
		return result, err
	}
	var knowledge *K
	if plan.NeedKnowledge != nil && plan.NeedKnowledge(data) {
		value, err := plan.Retrieve(ctx, scope, input, data)
		if err != nil {
			return result, err
		}
		knowledge = &value
	}
	var evidence *E
	if plan.NeedAgent != nil && plan.NeedAgent(data, knowledge) {
		value, err := plan.Investigate(ctx, scope, input, data, knowledge)
		if err != nil {
			return result, err
		}
		evidence = &value
	}
	return plan.Build(data, knowledge, evidence)
}

// Stage replaces a candidate before confirmation. The caller supplies the
// version it prepared against, so a late preparation cannot replace newer work.
func (s Service) Stage(ctx context.Context, scope Scope, runID string, version uint64, candidate Candidate) error {
	if scope.Actor == "" || runID == "" || scope.Target == "" || candidate.Target != scope.Target ||
		candidate.Version == "" || candidate.Command == "" {
		return ErrConflict
	}
	run, err := s.Runs.Load(ctx, scope.Actor, runID)
	if err != nil {
		return err
	}
	if run.Version != version || run.WriteKey != "" || run.Receipt != "" {
		return ErrConflict
	}
	run.Version++
	run.Candidate = candidate
	return s.Runs.CompareAndSave(ctx, scope.Actor, runID, version, run)
}
