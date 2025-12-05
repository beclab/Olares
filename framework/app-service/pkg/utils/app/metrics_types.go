package app

import (
	"bytetrade.io/web3os/app-service/pkg/kubesphere"
)

type Metrics struct {
	Results     []kubesphere.Metric `json:"results" description:"actual array of results"`
	CurrentPage int                 `json:"page,omitempty" description:"current page returned"`
	TotalPages  int                 `json:"total_page,omitempty" description:"total number of pages"`
	TotalItems  int                 `json:"total_item,omitempty" description:"page size"`
}

type Metadata struct {
	Data []kubesphere.Metadata `json:"data" description:"actual array of results"`
}

type MetricLabelSet struct {
	Data []map[string]string `json:"data" description:"actual array of results"`
}
