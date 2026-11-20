package policy

func ValidState(state string) bool {
	return state == "open" || state == "in_progress" || state == "completed"
}

func CanTransition(prior, next string) bool {
	return next == prior || (prior == "open" && (next == "in_progress" || next == "completed")) || (prior == "in_progress" && next == "completed") || (prior == "completed" && next == "open")
}
