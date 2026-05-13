package opencost

// AllocationRequest represents parameters for the OpenCost allocation API
type AllocationRequest struct {
	Window    string `json:"window"`
	Aggregate string `json:"aggregate,omitempty"`
	Step      string `json:"step,omitempty"`
}

// AllocationResponse wraps the OpenCost allocation API response
type AllocationResponse struct {
	Code int              `json:"code"`
	Data []AllocationData `json:"data"`
}

// AllocationData represents cost allocation for a single entity
type AllocationData struct {
	Name        string                `json:"name"`
	Properties  *AllocationProperties `json:"properties,omitempty"`
	CPUCost     float64               `json:"cpuCost"`
	RAMCost     float64               `json:"ramCost"`
	PVCost      float64               `json:"pvCost"`
	NetworkCost float64               `json:"networkCost"`
	TotalCost   float64               `json:"totalCost"`
	Start       string                `json:"start"`
	End         string                `json:"end"`
	Raw         map[string]any        `json:"-"`
}

// AllocationProperties contains labels and metadata for cost allocation
type AllocationProperties struct {
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
}