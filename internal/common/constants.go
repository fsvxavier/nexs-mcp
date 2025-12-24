package common

// Status constants for execution results.
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusError   = "error"
)

// Element type constants.
const (
	ElementTypePersona  = "persona"
	ElementTypeSkill    = "skill"
	ElementTypeTemplate = "template"
	ElementTypeAgent    = "agent"
	ElementTypeMemory   = "memory"
	ElementTypeEnsemble = "ensemble"
)

// Method constants.
const (
	MethodMention  = "mention"
	MethodKeyword  = "keyword"
	MethodSemantic = "semantic"
	MethodPattern  = "pattern"
)

// Sort order constants.
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Common constants.
const (
	SelectorAll = "all"
	BranchMain  = "main"
)
