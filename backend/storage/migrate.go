package storage

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs the database migrations once provided a db connection URL.
func Migrate(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("invalid database URL: %w", err)
	}

	db := dbmate.New(u)
	db.FS = migrationsFS
	db.MigrationsDir = []string{"migrations"}
	db.SchemaFile = "/dev/null"
	db.WaitBefore = true

	return db.CreateAndMigrate()
}
