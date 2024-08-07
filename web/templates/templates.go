package templates

import (
	_ "embed"
)

//go:embed "rate_update.tmpl"
var MessageTemplate string
