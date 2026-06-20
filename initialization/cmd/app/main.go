package main

import (
	"context"
	"log"
	"os"

	"ariga.io/atlas/atlasexec"
)

func main() {
	if err := runMigrations(); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
}

func runMigrations() error {
	// Пытаемся получить пути из переменных окружения.
	// Если их нет, используем значения по умолчанию для локального запуска:
	// "atlas" - ищет в system PATH, "." - текущая директория
	atlasPath := os.Getenv("ATLAS_PATH")
	if atlasPath == "" {
		atlasPath = "atlas"
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "."
	}

	client, err := atlasexec.NewClient(migrationsDir, atlasPath)
	if err != nil {
		return err
	}

	res, err := client.MigrateApply(context.Background(), &atlasexec.MigrateApplyParams{
		Env:        "local",
		AllowDirty: true,
	})
	if err != nil {
		return err
	}

	log.Printf("applied %d migrations", len(res.Applied))
	return nil
}
