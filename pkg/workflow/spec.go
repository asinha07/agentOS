package workflow

type StepSpec struct {
    Type   string         `json:"type"`
    Query  string         `json:"query,omitempty"`
    Output string         `json:"output,omitempty"`
}

type WorkflowSpec struct {
    Steps []StepSpec `json:"steps"`
}

