package nats

type Config struct {
	HTTPPort   int       `json:"http_port" mapstructure:"http_port"`
	Jetstream  Jetstream `json:"jetstream" mapstructure:"jetstream"`
	Accounts   Accounts  `json:"accounts" mapstructure:"accounts"`
	Port       int       `json:"port" mapstructure:"port"`
	PidFile    string    `json:"pid_file" mapstructure:"pid_file"`
	ServerName string    `json:"server_name" mapstructure:"server_name"`
}

type Jetstream struct {
	MaxFileStore   int64  `json:"max_file_store" mapstructure:"max_file_store"`
	MaxMemoryStore int    `json:"max_memory_store" mapstructure:"max_memory_store"`
	StoreDir       string `json:"store_dir" mapstructure:"store_dir"`
}

type Publish struct {
	Allow []string `json:"allow" mapstructure:"allow"`
}

type Subscribe struct {
	Allow []string `json:"allow" mapstructure:"allow"`
}

type Permissions struct {
	Publish   Publish   `json:"publish" mapstructure:"publish"`
	Subscribe Subscribe `json:"subscribe" mapstructure:"subscribe"`
}

type User struct {
	Username    string      `json:"user" mapstructure:"user"`
	Password    string      `json:"password" mapstructure:"password"`
	Permissions Permissions `json:"permissions" mapstructure:"permissions"`
}

type Terminus struct {
	Jetstream string `json:"jetstream" mapstructure:"jetstream"`
	Users     []User `json:"users" mapstructure:"users"`
}

type Accounts struct {
	Terminus Terminus `json:"terminus" mapstructure:"terminus"`
}
