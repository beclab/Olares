package storage

import "regexp"

const (
	tableInstallConfig = "install_config"
	tableInstallLogs   = "install_logs"
)

const (
	providerDriverName     = "sqlite3" // sqlite3e
	providerDataSourceName = "storage.db"
	providerSQLite         = "sqlite"
)

type ctxKey int

const (
	pathMigrations = "migrations"
)

const (
	ctxKeyTransaction ctxKey = iota
)

var (
	reMigration = regexp.MustCompile(`^V(?P<Version>\d{4})\.(?P<Name>[^.]+)\.(?P<Direction>(up|down))\.sql$`)
)
