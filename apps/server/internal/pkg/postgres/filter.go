package postgres

import (
	"fmt"
	"strings"
)

type QueryFilter struct {
	conditions []string
	Args       []any
}

func NewFilter(initialArgs ...any) *QueryFilter {
	args := make([]any, 0, len(initialArgs)+4)
	args = append(args, initialArgs...)
	return &QueryFilter{
		conditions: make([]string, 0, 4),
		Args:       args,
	}
}

func (f *QueryFilter) NextIdx() int {
	return len(f.Args) + 1
}

// Add formats clauseFormat with the next $N placeholder and appends val to Args.
func (f *QueryFilter) Add(clauseFormat string, val any) *QueryFilter {
	idx := f.NextIdx()
	f.conditions = append(f.conditions, fmt.Sprintf(clauseFormat, idx))
	f.Args = append(f.Args, val)
	return f
}

// Add2 formats clauseFormat with the same $N used twice in one clause, appending val once.
func (f *QueryFilter) Add2(clauseFormat string, val any) *QueryFilter {
	idx := f.NextIdx()
	f.conditions = append(f.conditions, fmt.Sprintf(clauseFormat, idx, idx))
	f.Args = append(f.Args, val)
	return f
}

// Add3 formats clauseFormat with the same $N used thrice in one clause, appending val once.
func (f *QueryFilter) Add3(clauseFormat string, val any) *QueryFilter {
	idx := f.NextIdx()
	f.conditions = append(f.conditions, fmt.Sprintf(clauseFormat, idx, idx, idx))
	f.Args = append(f.Args, val)
	return f
}

// AddRaw appends a static clause with no bound argument.
func (f *QueryFilter) AddRaw(clause string) *QueryFilter {
	f.conditions = append(f.conditions, clause)
	return f
}

// AddArgs appends trailing args (e.g. limit, offset) without adding a WHERE condition.
func (f *QueryFilter) AddArgs(args ...any) *QueryFilter {
	f.Args = append(f.Args, args...)
	return f
}

// Join joins all conditions with " AND ". If no conditions exist, returns defaultClause (or "TRUE").
func (f *QueryFilter) Join(defaultClause string) string {
	if len(f.conditions) == 0 {
		if defaultClause != "" {
			return defaultClause
		}
		return "TRUE"
	}
	return strings.Join(f.conditions, " AND ")
}
