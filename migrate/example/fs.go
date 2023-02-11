package main

import "embed"

//go:embed migrations
var migrationsPath embed.FS
