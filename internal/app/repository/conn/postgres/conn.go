package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	"github.com/Dopesonic/catalog-service/internal/app/config/section"
	"github.com/Dopesonic/catalog-service/migration"
)

type (
	Client struct {
		_bunDB
		rawBunDB *bun.DB

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   cfg.Address,
		Path:   cfg.Name,
	}

	query := u.Query()
	query.Set("sslmode", "disable")
	u.RawQuery = query.Encode()

	dsn := u.String()

	log.Printf(
		"postgres read timeout: %s, write timeout: %s",
		cfg.ReadTimeout,
		cfg.WriteTimeout,
	)

	sqlDB := sql.OpenDB(
		pgdriver.NewConnector(
			pgdriver.WithDSN(dsn),
			pgdriver.WithReadTimeout(cfg.ReadTimeout),
			pgdriver.WithWriteTimeout(cfg.WriteTimeout),
		),
	)

	sqlDB.SetMaxOpenConns(10)

	db := bun.NewDB(
		sqlDB,
		pgdialect.New(),
		bun.WithDiscardUnknownColumns(),
	)

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	return &Client{
		_bunDB:   db,
		rawBunDB: db,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err = migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("failed to discover migrations: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	migrator := migrate.NewMigrator(
		c.rawBunDB,
		migrations,
		opts...,
	)

	if err = migrator.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to init migrations: %w", err)
	}

	applied, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, mg := range applied {
		v, err := strconv.ParseInt(mg.Name, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse migration version %q: %w", mg.Name, err)
		}

		if v > oldVer {
			oldVer = v
		}
	}

	if _, err = migrator.Migrate(ctx); err != nil {
		return oldVer, 0, fmt.Errorf("failed to apply migrations: %w", err)
	}

	applied, err = migrator.AppliedMigrations(ctx)
	if err != nil {
		return oldVer, 0, fmt.Errorf("failed to get new migration version: %w", err)
	}

	for _, mg := range applied {
		v, err := strconv.ParseInt(mg.Name, 10, 64)
		if err != nil {
			return oldVer, 0, fmt.Errorf("failed to parse migration version %q: %w", mg.Name, err)
		}

		if v > newVer {
			newVer = v
		}
	}

	return oldVer, newVer, nil
}
