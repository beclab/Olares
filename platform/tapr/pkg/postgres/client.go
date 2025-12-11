package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"bytetrade.io/web3os/tapr/pkg/utils"

	"github.com/jmoiron/sqlx"
	"k8s.io/klog/v2"

	// import driver
	_ "github.com/lib/pq"
)

var extensionMap = map[string]string{
	"pgvector":   "vector",
	"vector":     "vector",
	"vectors":    "vectors",
	"pgvecto.rs": "vectors",
}

type DBLogger struct {
	*sqlx.DB
	debug bool
}

func (d *DBLogger) Debug() { d.debug = true }
func (d *DBLogger) log(sql string) {
	if d.debug {
		klog.Info("query: ", sql)
	}
}

func (d *DBLogger) ExecContext(ctx context.Context, query string) (sql.Result, error) {
	d.log(query)
	return d.DB.ExecContext(ctx, query)
}

func (d *DBLogger) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	d.log(query)
	return d.DB.NamedExecContext(ctx, query, arg)
}

func (d *DBLogger) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	d.log(query)
	return d.DB.NamedQueryContext(ctx, query, arg)
}

func (d *DBLogger) QueryContext(ctx context.Context, query string) (*sql.Rows, error) {
	d.log(query)
	return d.DB.QueryContext(ctx, query)
}

func (d *DBLogger) QueryxContext(ctx context.Context, query string) (*sqlx.Rows, error) {
	d.log(query)
	return d.DB.QueryxContext(ctx, query)
}

type client struct {
	DB      *DBLogger
	builder *clientBuilder
}

type clientBuilder struct {
	user     string
	password string
	host     string
	port     int
	database string
}

func NewClientBuidler(user, password, host string, port int) *clientBuilder {
	return &clientBuilder{
		user:     user,
		password: password,
		host:     host,
		port:     port,
		database: "postgres",
	}
}

func (cb *clientBuilder) WithDatabase(db string) *clientBuilder {
	cb.database = db
	return cb
}

func (cb *clientBuilder) Build() (*client, error) {
	return newClient(DSN(cb.user, cb.password, cb.host, cb.port, cb.database), cb)
}

func DSN(user, password, host string, port int, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
}

func newClient(dsn string, builder *clientBuilder) (*client, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	dbProxy := DBLogger{DB: db}

	dbProxy.Debug()

	return &client{DB: &dbProxy, builder: builder}, nil
}

func (c *client) Close() {
	err := c.DB.Close()
	if err != nil {
		klog.Error("close db error, ", err)
	}
}

func (c *client) SwitchDatabase(db string) error {
	c.Close()
	newClient, err := c.builder.WithDatabase(db).Build()
	if err != nil {
		return err
	}

	c.DB = newClient.DB

	return nil
}

func (c *client) CreateDatabaseIfNotExists(ctx context.Context, db, owner string) error {
	database, err := c.findDatabase(ctx, db)
	if err != nil {
		return err
	}

	if database != nil && database.Name == db {
		return nil
	}

	// FIXME: sql inject
	sql := fmt.Sprintf("create database %s", db)
	if owner != "" {
		sql += " owner "
		sql += owner
	}

	_, err = c.DB.ExecContext(ctx, sql)

	return err
}

func (c *client) findDatabase(ctx context.Context, databaseName string) (*PGTable, error) {
	if databaseName == "" {
		return nil, errors.New("databaseName is empty")
	}
	sql := "select datname as name from pg_catalog.pg_database where datname=:name ORDER BY 1"
	database := new(PGTable)
	rows, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{"name": databaseName})
	if err != nil {
		klog.Error("select database error, sql: ", sql, ", error, ", err)
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	err = rows.StructScan(database)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (c *client) CreateCitus(ctx context.Context) error {
	_, err := c.DB.ExecContext(ctx, "create extension if not exists citus")

	return err
}

func (c *client) CreateExtensions(ctx context.Context, extensions []string) error {
	errs := make([]error, 0)
	for _, e := range extensions {
		extension, ok := extensionMap[e]
		if !ok {
			extension = e
		}
		_, err := c.DB.ExecContext(ctx, fmt.Sprintf("create extension if not exists %s cascade;", extension))
		if err != nil {
			errs = append(errs, err)
		}
	}
	return utils.AggregateErrs(errs)
}

func (c *client) ExecuteScript(ctx context.Context, databaseName, dbUsername string, scripts []string) error {
	var sb strings.Builder
	for _, cmd := range scripts {
		// replace databasename, dbusername with real database name and username in postgres
		cmd = strings.ReplaceAll(cmd, "$databasename", databaseName)
		cmd = strings.ReplaceAll(cmd, "$dbusername", dbUsername)
		sb.WriteString(cmd)
		if !strings.HasSuffix(cmd, ";") {
			sb.WriteString(";")
		}
	}
	_, err := c.DB.ExecContext(ctx, sb.String())
	return err
}

func (c *client) AddWorkerNode(ctx context.Context, nodeAddr string, port int) error {
	sql := "SELECT * from citus_add_node(:node_addr, :node_port);"

	_, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"node_addr": nodeAddr,
		"node_port": port,
	})

	if err != nil {
		return err
	}

	// TODO: debug info

	return nil
}

func (c *client) SetMasterNode(ctx context.Context, nodeAddr string, port int) error {
	sql := "SELECT citus_set_coordinator_host(:node_addr, :node_port);"

	_, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"node_addr": nodeAddr,
		"node_port": port,
	})

	if err != nil {
		return err
	}

	// TODO: debug info

	return nil

}

func (c *client) Rebalance(ctx context.Context) error {
	_, err := c.DB.ExecContext(ctx, "select citus_rebalance_start()")

	return err
}

func (c *client) ChangeAdminUser(ctx context.Context, olduser, newuser, newpwd string) error {
	sql := "select rolname from pg_catalog.pg_roles where rolname=:role"
	roles, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"role": newuser,
	})

	if err != nil {
		return err
	}

	defer roles.Close()
	found := roles.Next()
	if !found {
		klog.Info("create a new pg cluster admin user")
		sql = fmt.Sprintf("create role %s with superuser login password '%s'", newuser, newpwd)
	} else {
		klog.Info("update pg cluster admin user password")
		sql = fmt.Sprintf("alter role %s with superuser password '%s'", newuser, newpwd)
	}
	_, err = c.DB.ExecContext(ctx, sql)

	if err != nil {
		return err
	}

	if olduser != newuser {
		sql = fmt.Sprintf("alter role %s with nosuperuser", olduser)
		_, err = c.DB.ExecContext(ctx, sql)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) CreateOrUpdateUser(ctx context.Context, user, pwd string) error {
	sql := "select usename from pg_catalog.pg_user where usename=:user"

	res, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"user": user,
	})

	if err != nil {
		return err
	}

	defer res.Close()
	exists := res.Next()
	if exists {
		sql = fmt.Sprintf("alter role %s with password '%s'", user, pwd)
	} else {
		sql = fmt.Sprintf("create role %s with login password '%s'", user, pwd)
	}

	_, err = c.DB.ExecContext(ctx, sql)

	if err != nil {
		klog.Error("create or update user error, ", err, ", ", user)
		return err
	}

	sql = fmt.Sprintf("alter role %s connection limit 30", user)

	_, err = c.DB.ExecContext(ctx, sql)

	if err != nil {
		klog.Error("update user connection limit error, ", err, ", ", user)
		return err
	}

	sql = fmt.Sprintf("grant pg_write_server_files, pg_read_server_files to %s", user)

	_, err = c.DB.ExecContext(ctx, sql)

	if err != nil {
		klog.Error("grant role to user error, ", err, ", ", user)
		return err
	}

	return nil
}

func (c *client) ListDatabaseByOwner(ctx context.Context, owner string) ([]string, error) {
	sql := "select datname as name from pg_catalog.pg_database where pg_catalog.pg_get_userbyid(datdba)=:owner"
	rows, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"owner": owner,
	})

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var databases []string
	for rows.Next() {
		row := new(PGTable)
		err = rows.StructScan(row)
		if err != nil {
			return nil, err
		}

		databases = append(databases, row.Name)
	}

	return databases, nil
}

func (c *client) DropDatabase(ctx context.Context, db string) error {

	// terminate backend process first
	sql := "SELECT pid from pg_stat_activity where datname=:db"
	rows, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"db": db,
	})

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		row := new(PGPid)
		err = rows.StructScan(row)
		if err != nil {
			return err
		}

		sql = fmt.Sprintf("SELECT pg_terminate_backend(%d)", row.PID)
		_, err = c.DB.ExecContext(ctx, sql)
		if err != nil {
			return err
		}

	}

	sql = fmt.Sprintf("drop database if exists %s", db)
	_, err = c.DB.ExecContext(ctx, sql)

	return err
}

func (c *client) DeleteUser(ctx context.Context, user string) error {
	sql := fmt.Sprintf("drop role if exists %s", user)

	_, err := c.DB.QueryContext(ctx, sql)

	if err != nil {
		return err
	}

	return nil
}
