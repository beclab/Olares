package storage

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/storage/model"
)

//go:embed migrations/*
var migrationsFS embed.FS

func loadMigrations(providerName string) (migrations []model.SchemaMigration, err error) {
	var (
		migration model.SchemaMigration
		entries   []fs.DirEntry
	)

	if entries, err = migrationsFS.ReadDir(path.Join(pathMigrations)); err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var fileName = entry.Name()
		fileName = strings.TrimPrefix(fileName, "._")
		if migration, err = scanMigration(providerName, fileName); err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func skipMigration(up bool, target, prior int, migration *model.SchemaMigration) (skip bool) {
	if up {
		if !migration.Up {
			// Skip if we wanted an Up migration but it isn't an Up migration.
			return true
		}

		if migration.Version > target || migration.Version <= prior {
			// Skip if the migration version is greater than the target or less than or equal to the previous version.
			return true
		}
	} else {
		if migration.Up {
			// Skip if we didn't want an Up migration but it is an Up migration.
			return true
		}

		if migration.Version <= target || migration.Version > prior {
			// Skip the migration if we want to go down and the migration version is less than or equal to the target
			// or greater than the previous version.
			return true
		}
	}

	return false
}

func scanMigration(providerName, m string) (migration model.SchemaMigration, err error) {
	if !reMigration.MatchString(m) {
		return model.SchemaMigration{}, errors.New("invalid migration: could not parse the format")
	}

	result := reMigration.FindStringSubmatch(m)

	migration = model.SchemaMigration{
		Name:     strings.ReplaceAll(result[reMigration.SubexpIndex("Name")], "_", " "),
		Provider: providerName,
	}

	var data []byte

	if data, err = migrationsFS.ReadFile(path.Join(pathMigrations, m)); err != nil {
		return model.SchemaMigration{}, err
	}

	migration.Query = string(data)

	switch direction := result[reMigration.SubexpIndex("Direction")]; direction {
	case "up":
		migration.Up = true
	case "down":
		migration.Up = false
	default:
		return model.SchemaMigration{}, fmt.Errorf("invalid migration: value in Direction group '%s' must be up or down", direction)
	}

	migration.Version, _ = strconv.Atoi(result[reMigration.SubexpIndex("Version")])

	return migration, nil
}
