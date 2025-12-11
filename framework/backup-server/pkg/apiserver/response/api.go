package response

type ListResult struct {
	Items  []any `json:"items"`
	Totals int   `json:"totals"`
}

func NewListResult[T any](items []T) ListResult {
	vs := make([]any, 0)
	if items != nil && len(items) > 0 {
		for _, item := range items {
			vs = append(vs, item)
		}
	}
	return ListResult{Items: vs, Totals: len(items)}
}
