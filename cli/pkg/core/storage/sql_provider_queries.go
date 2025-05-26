package storage

// Table: install_config
const (
	queryFmtInsertInstallConfig = `
	  INSERT INTO %s (terminus_os_domainname, terminus_os_username, kube_type, vendor, gpu_enable, gpu_share, version)
		VALUES (?, ?, ?, ?, ?, ?, ?);`
	queryFmtInsertInstallLog = `
		INSERT INTO %s (message, state, percent, created_at)
		VALUES (?, ?, ?, ?);`
	queryFmtQueryInstallState = `
		SELECT message, state, percent, created_at FROM %s WHERE created_at >= ? ORDER BY id, created_at;`
)

const (
	querySQLiteSelectExistingTables = `
		SELECT name
		FROM sqlite_master
		WHERE type = 'table';`
)
