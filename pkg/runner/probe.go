package runner

type NodeStatusType string

const (
	NodeStatusUnknown  NodeStatusType = "unknown"
	NodeStatusRunning  NodeStatusType = "running"
	NodeStatusDegraded NodeStatusType = "degraded"
)

type ProbeType string

const (
	ProbeUnknown    ProbeType = "unknown"
	ProbeRunning    ProbeType = "running"
	ProbeFailed     ProbeType = "failed"
	ProbeTerminated ProbeType = "terminated"
)

// Severity defines the severity of the probe.
type ProbeSeverity string

const (
	// None severity denotes the severity of a running probe
	ProbeNone ProbeSeverity = "none"
	// Critical defines a serious error that requires immediate attention
	ProbeCritical ProbeSeverity = "critical"
	// Warning defines a (possibly transient) condition that requires attention
	// but is not critical
	ProbeWarning ProbeSeverity = "warning"
)

// Probe represents the outcome of a single check
type Probe struct {
	// Checker is the name of the checker that generated the probe
	Checker string `json:"checker"`
	// Detail is the optional detail specific to the checker
	Detail string `json:"detail"`
	// Code is the optional code specific to a checker (i.e. HTTP status code)
	Code string `json:"code"`
	// Status is the result of the probe
	Status ProbeType `json:"status"`
	// Error is the probe-specific error message
	Error string `json:"error"`
	// CheckerData is a free-form data specific to the checker
	CheckerData interface{} `json:"checkerData"`
	// Severity is the severity of the probe
	Severity ProbeSeverity `json:"severity"`
}

type FinalProbe struct {
	// Status is the result of the probe
	Status ProbeType `json:"status"`

	// Config is the result of check of configuration in ConfigMap.
	Config SingleFinalProbe `json:"config"`

	// Errors is the probe-specific error message
	Errors []SingleFinalProbe `json:"errors"`

	// Oks is the ok response
	Oks []SingleFinalProbe `json:"oks"`
}

type SingleFinalProbe struct {
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}

func (m *Probe) Reset() { *m = Probe{} }

func (m *Probe) GetChecker() string {
	if m != nil {
		return m.Checker
	}
	return ""
}

func (m *Probe) GetDetail() string {
	if m != nil {
		return m.Detail
	}
	return ""
}

func (m *Probe) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *Probe) GetStatus() ProbeType {
	if m != nil {
		return m.Status
	}
	return ProbeUnknown
}

func (m *Probe) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *Probe) GetCheckerData() interface{} {
	if m != nil {
		return m.CheckerData
	}
	return nil
}

func (m *Probe) GetSeverity() ProbeSeverity {
	if m != nil {
		return m.Severity
	}
	return ProbeNone
}
