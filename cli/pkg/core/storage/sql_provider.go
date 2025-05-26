package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func NewSQLProvider(name, dataSourceName string) (provider SQLProvider) {
	dbName := path.Join(dataSourceName, providerDataSourceName)
	db, err := sqlx.Open(providerDriverName, fmt.Sprintf("file:%s?cache=shared", dbName))
	db.SetMaxIdleConns(1)
	provider = SQLProvider{
		db:         db,
		name:       name,
		driverName: providerDriverName,
		errOpen:    err,

		log: logger.GetLogger(),

		sqlInsertInstallConfig: fmt.Sprintf(queryFmtInsertInstallConfig, tableInstallConfig),
		sqlInsertInstallLog:    fmt.Sprintf(queryFmtInsertInstallLog, tableInstallLogs),
		sqlQueryInstallState:   fmt.Sprintf(queryFmtQueryInstallState, tableInstallLogs),
	}

	return provider
}

type SQLProvider struct {
	db *sqlx.DB

	name       string
	driverName string
	schema     string
	errOpen    error

	log *zap.SugaredLogger

	// Table: install_config
	sqlInsertInstallConfig string
	sqlInsertInstallLog    string
	sqlQueryInstallState   string

	// Utility.
	sqlSelectExistingTables string

	// Table: migrations.
	sqlInsertMigration       string
	sqlSelectMigrations      string
	sqlSelectLatestMigration string
}

func (p *SQLProvider) StartupCheck() (err error) {
	if p.errOpen != nil {
		return fmt.Errorf("error opening database: %w", p.errOpen)
	}

	for i := 0; i < 10; i++ {
		if err = p.db.Ping(); err == nil {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	ctx := context.Background()
	if err = p.SchemaMigrate(ctx, true); err != nil {
		return err
	}
	return nil
}

// BeginTX begins a transaction with the storage provider when applicable.
func (p *SQLProvider) BeginTX(ctx context.Context) (c context.Context, err error) {
	var tx *sql.Tx

	if tx, err = p.db.Begin(); err != nil {
		return nil, err
	}

	return context.WithValue(ctx, ctxKeyTransaction, tx), nil
}

// Commit performs a storage provider commit when applicable.
func (p *SQLProvider) Commit(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Commit()
}

// Rollback performs a storage provider rollback when applicable.
func (p *SQLProvider) Rollback(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Rollback()
}

func (p *SQLProvider) Ping() (err error) {
	return p.db.Ping()
}

// Close the underlying storage provider.
func (p *SQLProvider) Close() (err error) {
	return p.db.Close()
}

func (p *SQLProvider) SaveInstallConfig(config model.InstallModelReq) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err = p.db.ExecContext(ctx, p.sqlInsertInstallConfig,
		config.Config.DomainName, config.Config.UserName, config.Config.KubeType, config.Config.Vendor, config.Config.GpuEnable, config.Config.GpuShare, config.Config.Vendor); err != nil {
		return fmt.Errorf("error inserting install config for user '%s': %w", config.Config.UserName, err)
	}

	return nil
}

func (p *SQLProvider) SaveInstallLog(msg string, state string, percent int64) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var t = time.Now().UnixMilli()
	if _, err = p.db.ExecContext(ctx, p.sqlInsertInstallLog,
		msg, state, percent, t); err != nil {
		return fmt.Errorf("error inserting install log %s: %s, %w", state, msg, err)
	}
	return nil
}

func (p *SQLProvider) QueryInstallState(tspan int64) (data []model.InstallState, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data = make([]model.InstallState, 0, 10)
	if err = p.db.SelectContext(ctx, &data, p.sqlQueryInstallState, tspan); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return data, err
	}

	return data, nil
}
