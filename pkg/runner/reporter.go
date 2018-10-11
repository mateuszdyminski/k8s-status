package runner

import "context"

// Reporter defines an obligation to report structured errors.
type Reporter interface {
	// Add adds a health probe for a specific node.
	Add(probe *Probe)
	// Status retrieves the collected status after executing all checks.
	GetProbes() []*Probe
	// NumProbes returns the number of probes this reporter contains
	NumProbes() int
}

// Probes is a list of probes.
// It implements the Reporter interface.
type Probes []*Probe

// Add adds a health probe for a specific node.
// Implements Reporter
func (r *Probes) Add(probe *Probe) {
	*r = append(*r, probe)
}

// Status retrieves the collected status after executing all checks.
// Implements Reporter
func (r Probes) GetProbes() []*Probe {
	return []*Probe(r)
}

// NumProbes returns the number of probes this reporter contains
// Implements Reporter
func (r Probes) NumProbes() int {
	return len(r)
}

// GetFailed returns all probes that reported an error
func (r Probes) GetFailed() []*Probe {
	var failed []*Probe

	for _, probe := range r {
		if probe.Status == ProbeFailed {
			failed = append(failed, probe)
		}
	}

	return failed
}

// Status computes the node status based on collected probes.
func (r Probes) Status() NodeStatusType {
	result := NodeStatusRunning
	for _, probe := range r {
		if probe.Status == ProbeFailed {
			result = NodeStatusDegraded
			break
		}
	}
	return result
}

// Checker is an interface for executing a health check.
type Checker interface {
	Name() string
	// Check runs a health check and records any errors into the specified reporter.
	Check(context.Context, Reporter)
}

// Checkers is a collection of checkers.
// It implements CheckerRepository interface.
type Checkers []Checker

func (r *Checkers) AddChecker(checker Checker) {
	*r = append(*r, checker)
}

// CheckerRepository represents a collection of checkers.
type CheckerRepository interface {
	AddChecker(checker Checker)
}

// AddFrom copies probes from src to dst
func AddFrom(dst, src Reporter) {
	for _, probe := range src.GetProbes() {
		dst.Add(probe)
	}
}
