package main

import (
	"context"
	"log"
	"os"

	"github.com/kostapo/tasktracker/shared/logger"

	"github.com/kostapo/tasktracker/initialization/internal/config"

	"ariga.io/atlas/atlasexec"
	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
	"github.com/kostapo/tasktracker/initialization/internal/keycloak/steps"
)

func main() {

	log := logger.New("INIT-SERVICE")
	if err := runMigrations(); err != nil {
		log.Error("migrations failed", "error", err)
	}

	if err := run(); err != nil {
		log.Error("Инициализация Keycloak завершилась ошибкой", "error", err)
		os.Exit(1)
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

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	admin := keycloakinit.NewAdminClient(
		cfg.Keycloak.URL,
		cfg.Keycloak.AdminRealm,
		cfg.Keycloak.AdminUsername,
		cfg.Keycloak.AdminPassword.Val(),
	)

	realmCtx := keycloakinit.NewRealmContext(cfg.App.Realm, cfg.App.ClientID)

	rt := &keycloakinit.Runtime{
		Admin: admin,
		Realm: realmCtx,
	}

	allSteps := []keycloakinit.Step{
		steps.RealmInitStep{},
		steps.NewClientInitStep(cfg.App.ClientID, cfg.App.RedirectURIs, cfg.App.WebOrigins, cfg.App.PostLogoutRedirectURIs),
		steps.RolesInitStep{},
		steps.RoleMapperInitStep{},
		steps.NewUserProfileInitStep(cfg.Keycloak.URL),
		steps.NewGoogleIdpInitStep(cfg.Google.ClientID, cfg.Google.ClientSecret.Val()),
		steps.UsersInitStep{},
	}

	return keycloakinit.Run(ctx, rt, allSteps)
}
