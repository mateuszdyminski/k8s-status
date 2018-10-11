package runner

import (
	"context"
)

// compositeChecker defines a health.Checker as a composite of
// several checkers run as a whole
type compositeChecker struct {
	name     string
	checkers []Checker
}

// Name returns the name of this checker
func (r *compositeChecker) Name() string { return r.name }

// Check runs an health check over the list of encapsulated checkers
// and reports errors to the specified Reporter
func (r *compositeChecker) Check(ctx context.Context, reporter Reporter) {
	for _, checker := range r.checkers {
		checker.Check(ctx, reporter)
	}
}

// NewCompositeChecker makes checker out of array of checkers
func NewCompositeChecker(name string, checkers []Checker) Checker {
	return &compositeChecker{name, checkers}
}
