package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/Dopesonic/catalog-service/internal/app/builder"
)

func Migrate() *cli.Command {
	return &cli.Command{
		Name:    "migrate",
		Aliases: []string{"m"},
		Usage:   "Apply pending database migrations",
		Description: strings.TrimSpace(`
Connects to PostgreSQL, checks current schema version,
and applies any pending migrations.
`),
		Action:          cmdMigrate,
		HideHelpCommand: true,
	}
}

func cmdMigrate(cCtx *cli.Context) error {
	Builder := builder.NewBuilder(cCtx)
	Builder.BuildConfig()
	Builder.BuildRepoConnPostgres()
	Builder.BuildRepoConnMigrator()
	Builder.Run()

	return nil
}
