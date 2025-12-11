package zinc

import "time"

type IndexShard struct {
	ShardNum int64               `json:"shard_num"`
	ID       string              `json:"id"`
	NodeID   string              `json:"node_id"` // remote instance ID
	Shards   []*IndexSecondShard `json:"shards"`
	Stats    IndexStat           `json:"stats"`
}

type IndexSecondShard struct {
	ID    int64     `json:"id"`
	Stats IndexStat `json:"stats"`
}

type IndexStat struct {
	DocTimeMin  int64  `json:"doc_time_min"`
	DocTimeMax  int64  `json:"doc_time_max"`
	DocNum      uint64 `json:"doc_num"`
	StorageSize uint64 `json:"storage_size"`
	WALSize     uint64 `json:"wal_size"`
}

type IndexSimple struct {
	Name        string                 `json:"name"`
	StorageType string                 `json:"storage_type"`
	ShardNum    int64                  `json:"shard_num"`
	Settings    *IndexSettings         `json:"settings,omitempty"`
	Mappings    map[string]interface{} `json:"mappings,omitempty"`
}

type IndexSettings struct {
	NumberOfShards   int64          `json:"number_of_shards,omitempty"`
	NumberOfReplicas int64          `json:"number_of_replicas,omitempty"`
	Analysis         *IndexAnalysis `json:"analysis,omitempty"`
}

type IndexAnalysis struct {
	Analyzer    map[string]*Analyzer   `json:"analyzer,omitempty"`
	CharFilter  map[string]interface{} `json:"char_filter,omitempty"`
	Tokenizer   map[string]interface{} `json:"tokenizer,omitempty"`
	TokenFilter map[string]interface{} `json:"token_filter,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"` // compatibility with es, alias for TokenFilter
}

type Analyzer struct {
	CharFilter  []string `json:"char_filter,omitempty"`
	Tokenizer   string   `json:"tokenizer,omitempty"`
	TokenFilter []string `json:"token_filter,omitempty"`
	Filter      []string `json:"filter,omitempty"` // compatibility with es, alias for TokenFilter

	// options for compatible
	Type      string   `json:"type,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`   // for type=pattern
	Lowercase bool     `json:"lowercase,omitempty"` // for type=pattern
	Stopwords []string `json:"stopwords,omitempty"` // for type=pattern,standard,stop
}

type Tokenizer struct {
	Type string `json:"type"`
}

type TokenFilter struct {
	Type string `json:"type"`
}

type User struct {
	ID        string    `json:"_id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Salt      string    `json:"salt,omitempty"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID         string    `json:"_id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Permission []string  `json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var RoleUser = &Role{
	ID:   "user",
	Name: "user",
	Permission: []string{
		"document.Bulk",
		"document.Multi",
		"document.Create",
		"document.Get",
		"document.Update",
		"document.Delete",
		"search.SearchDSL",
		"index.List",
		"index.IndexNameList",
		"index.Get",
		"index.Exists",
		"index.Refresh",
		"index.GetMapping",
		"index.GetSettings",
		"index.Analyze",
		"search.MultipleSearch",
		"search.DeleteByQuery",
		"index.ListTemplate",
		"index.GetTemplate",
		"elastic.PutDataStream",
		"elastic.GetDataStream",
		"index.GetESMapping",
		"index.GetESAliases",
		"document.ESBulk",
		"document.CreateUpdate",
	},
}
