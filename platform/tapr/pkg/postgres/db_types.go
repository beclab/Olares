package postgres

const (
	PG_MASTER_HOST = "citus-0.citus-headless"
	PG_PORT        = 5432
)

type PGTable struct {
	Name string `db:"name"`
}

type PGPid struct {
	PID int `db:"pid"`
}
