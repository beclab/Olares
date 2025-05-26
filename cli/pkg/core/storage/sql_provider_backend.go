package storage

import (
	"database/sql"
	"encoding/base64"

	"github.com/mattn/go-sqlite3"
)

var Db Provider

type SQLiteProvider struct {
	SQLProvider
}

func NewSQLiteProvider(dbPath string) (provider *SQLiteProvider) {
	provider = &SQLiteProvider{
		SQLProvider: NewSQLProvider(providerSQLite, dbPath),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = querySQLiteSelectExistingTables

	Db = provider
	return provider
}

func sqlite3BLOBToTEXTBase64(data []byte) (b64 string) {
	return base64.StdEncoding.EncodeToString(data)
}

func sqlite3TEXTBase64ToBLOB(b64 string) (data []byte, err error) {
	return base64.StdEncoding.DecodeString(b64)
}

func init() {
	sql.Register("sqlite3e", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) (err error) {
			if err = conn.RegisterFunc("BIN2B64", sqlite3BLOBToTEXTBase64, true); err != nil {
				return err
			}

			if err = conn.RegisterFunc("B642BIN", sqlite3TEXTBase64ToBLOB, true); err != nil {
				return err
			}

			return nil
		},
	})
}
