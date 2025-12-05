package workflowinstaller

// KnowledgeAPIResp a struct represents the response of knowledge api response.
type KnowledgeAPIResp struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

// KnowledgeFeedDelReq contains feed urls list.
type KnowledgeFeedDelReq struct {
	FeedUrls []string `json:"feed_urls"`
}
