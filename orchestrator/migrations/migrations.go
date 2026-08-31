// Package migrations embeds the SQL migration files applied by
// internal/db.Migrate. go:embed can only reach files within its own
// package directory, hence this thin wrapper package.
package migrations

import "embed"

//go:embed *.up.sql
var FS embed.FS
