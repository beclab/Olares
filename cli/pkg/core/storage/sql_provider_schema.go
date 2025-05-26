package storage

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/storage/model"
	"github.com/jmoiron/sqlx"
)

func (p *SQLProvider) SchemaMigrate(ctx context.Context, up bool) (err error) {
	var (
		tx   *sqlx.Tx
		conn SQLXConnection
	)

	if tx, err = p.db.BeginTxx(ctx, nil); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	conn = tx

	if err = p.schemaMigrate(ctx, conn); err != nil {
		if tx != nil && err == ErrNoMigrationsFound {
			_ = tx.Rollback()
		}

		return err
	}

	if tx != nil {
		if err = tx.Commit(); err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("failed to commit the transaction with: commit error: %w, rollback error: %+v", err, rerr)
			}

			return fmt.Errorf("failed to commit the transaction but it has been rolled back: commit error: %w", err)
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrate(ctx context.Context, conn SQLXConnection) (err error) {
	migrations, err := loadMigrations(p.name) // schemaMigrate
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return ErrNoMigrationsFound
	}

	for _, migration := range migrations {
		if !migration.Up {
			continue
		}
		if err = p.schemaMigrateApply(ctx, conn, migration); err != nil {
			return p.schemaMigrateRollback(ctx, conn, err)
		}
	}

	return nil
}

func (p *SQLProvider) schemaMigrateApply(ctx context.Context, conn SQLXConnection, migration model.SchemaMigration) (err error) {
	if migration.NotEmpty() && migration.Up {
		if _, err = conn.ExecContext(ctx, migration.Query); err != nil {
			return fmt.Errorf(errFmtFailedMigration, migration.Version, migration.Name, err)
		}
	}

	if err = p.schemaMigrateFinalize(ctx, conn, migration); err != nil {
		return err
	}

	return nil
}

func (p *SQLProvider) schemaMigrateFinalize(ctx context.Context, conn SQLXConnection, migration model.SchemaMigration) (err error) {
	if migration.Version == 1 && !migration.Up {
		return nil
	}

	if _, err = conn.ExecContext(ctx, p.sqlInsertMigration, time.Now(), migration.Before(), migration.After()); err != nil {
		return fmt.Errorf("failed inserting migration record: %w", err)
	}

	// p.log.Debugf("Storage schema migrated from version %d to %d", migration.Before(), migration.After())

	return nil
}

func (p *SQLProvider) schemaMigrateRollback(ctx context.Context, conn SQLXConnection, merr error) (err error) {
	switch tx := conn.(type) {
	case *sqlx.Tx:
		return p.schemaMigrateRollbackWithTx(ctx, tx, merr)
	default:
		return p.schemaMigrateRollbackWithoutTx(ctx, merr)
	}
}

func (p *SQLProvider) schemaMigrateRollbackWithTx(_ context.Context, tx *sqlx.Tx, merr error) (err error) {
	if err = tx.Rollback(); err != nil {
		return fmt.Errorf("error applying rollback %+v. rollback caused by: %w", err, merr)
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %w", merr)
}

func (p *SQLProvider) schemaMigrateRollbackWithoutTx(ctx context.Context, merr error) (err error) {
	migrations, err := loadMigrations(p.name) // schemaMigrateRollbackWithoutTx
	if err != nil {
		return fmt.Errorf("error loading migrations for rollback: %+v. rollback caused by: %w", err, merr)
	}

	for _, migration := range migrations {
		if err = p.schemaMigrateApply(ctx, p.db, migration); err != nil {
			return fmt.Errorf("error applying migration version %d to version %d for rollback: %+v. rollback caused by: %w", migration.Before(), migration.After(), err, merr)
		}
	}

	return fmt.Errorf("migration rollback complete. rollback caused by: %w", merr)
}
