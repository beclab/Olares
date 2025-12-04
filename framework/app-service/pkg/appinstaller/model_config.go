package appinstaller

// ModelConfig defines config for llm model.
type ModelConfig struct {
	SourceURL   string `json:"source_url"`
	ID          string `json:"id"`
	Object      string `json:"object"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Format      string `json:"format"`
	Settings    struct {
		CtxLen         int    `json:"ctx_len"`
		PromptTemplate string `json:"prompt_template"`
	} `json:"settings"`
	Parameters struct {
		Temperature      float64       `json:"temperature"`
		TopP             float64       `json:"top_p"`
		Stream           bool          `json:"stream"`
		MaxTokens        int           `json:"max_tokens"`
		Stop             []interface{} `json:"stop"`
		FrequencyPenalty int           `json:"frequency_penalty"`
		PresencePenalty  int           `json:"presence_penalty"`
	} `json:"parameters"`
	Metadata struct {
		Author string   `json:"author"`
		Tags   []string `json:"tags"`
		Size   int64    `json:"size"`
	} `json:"metadata"`
	Engine string `json:"engine"`
}
