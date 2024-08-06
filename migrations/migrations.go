package migrations

import "embed"

//go:embed rates/*.sql
var RatesMigrationsFS embed.FS

//go:embed customers/*.sql
var CustomersMigrationsFS embed.FS
