package v1alpha1

type ClusterMetrics struct {
	CPU    MetricV        `json:"cpu"`
	Memory MetricV        `json:"memory"`
	Disk   MetricV        `json:"disk"`
	Net    MetricNetValue `json:"net"`
}

type MetricV struct {
	Total float64 `json:"total"`
	Usage float64 `json:"usage"`
	Ratio float64 `json:"ratio"`
	Unit  string  `json:"unit"`
}

type MetricNetValue struct {
	Transmitted float64 `json:"transmitted"`
	Received    float64 `json:"received"`
}

type Metrics struct {
	Results     []Metric `json:"results" description:"actual array of results"`
	CurrentPage int      `json:"page,omitempty" description:"current page returned"`
	TotalPages  int      `json:"total_page,omitempty" description:"total number of pages"`
	TotalItems  int      `json:"total_item,omitempty" description:"page size"`
}

type MetadataWrap struct {
	Data []Metadata `json:"data" description:"actual array of results"`
}

type MetricLabelSet struct {
	Data []map[string]string `json:"data" description:"actual array of results"`
}
