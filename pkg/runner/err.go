package runner

const noErrorDetail = ""

// NewProbeFromErr creates a new Probe given an error and a checker name
func NewProbeFromErr(name, detail string, err error) *Probe {
	return &Probe{
		Checker: name,
		Detail:  detail,
		Error:   userMessage(err),
		Status:  ProbeFailed,
	}
}

// NewSuccessProbe returns a successful probe for the given checker
func NewSuccessProbe(name string) *Probe {
	return &Probe{
		Checker: name,
		Status:  ProbeRunning,
	}
}

// UserMessage returns user-friendly part of the error
func userMessage(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
