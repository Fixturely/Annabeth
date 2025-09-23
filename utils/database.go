package utils

import (
	"crypto/tls"
	"database/sql"
	"fixture-ingestor/config"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDatabase() (*bun.DB, error) {
	databaseCfg := config.GetConfig().Database
	opts := []pgdriver.Option{
		pgdriver.WithAddr(fmt.Sprintf("%s:%d", databaseCfg.Host, databaseCfg.Port)),
		pgdriver.WithDatabase(databaseCfg.Name),
		pgdriver.WithUser(databaseCfg.User),
		pgdriver.WithPassword(databaseCfg.Password),
	}

	if databaseCfg.SSLMode {
		opts = append(opts, pgdriver.WithTLSConfig(&tls.Config{ServerName: databaseCfg.Host}))
	} else {
		opts = append(opts, pgdriver.WithInsecure(true))
	}

	pgconn := pgdriver.NewConnector(opts...)
	sqldb := sql.OpenDB(pgconn)

	db := bun.NewDB(sqldb, pgdialect.New())

	// Ensure the database can connect.
	_, err := db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return db, nil
}
