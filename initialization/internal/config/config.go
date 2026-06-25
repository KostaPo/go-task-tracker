// Package config отвечает за чтение и валидацию конфигурации приложения
// из переменных окружения (с опциональной подгрузкой .env файла).
package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// KeycloakAdminConfig — параметры подключения к Keycloak под админской учёткой.
type KeycloakAdminConfig struct {
	URL           string    // KEYCLOAK_URI
	AdminRealm    string    // KEYCLOAK_ADMIN_REALM (обычно "master")
	AdminClientID string    // KEYCLOAK_ADMIN_CLIENT_ID (обычно "admin-cli")
	AdminUsername string    // KEYCLOAK_ADMIN_USERNAME
	AdminPassword secretStr // KEYCLOAK_ADMIN_PASSWORD
}

// AppClientConfig — параметры realm'а и клиента приложения.
type AppClientConfig struct {
	Realm        string   // KEYCLOAK_REALM
	ClientID     string   // KEYCLOAK_CLIENT_ID
	RedirectURIs []string // KEYCLOAK_REDIRECT_URIS (через запятую)
	WebOrigins   []string // KEYCLOAK_WEB_ORIGINS (через запятую)

	// PostLogoutRedirectURIs хранится как есть, без разбиения:
	// Keycloak ожидает значение атрибута "post.logout.redirect.uris"
	// в виде строк, объединённых через "##".
	PostLogoutRedirectURIs string // KEYCLOAK_LOGOUT_REDIRECT_URIS
}

// GoogleConfig — параметры Google OAuth для настройки Identity Provider.
type GoogleConfig struct {
	ClientID     string    // GOOGLE_CLIENT_ID
	ClientSecret secretStr // GOOGLE_CLIENT_SECRET
}

// Config — корневая конфигурация приложения.
type Config struct {
	Keycloak KeycloakAdminConfig
	App      AppClientConfig
	Google   GoogleConfig
}

// Load читает конфигурацию из переменных окружения. Если рядом есть .env —
// он подгружается заранее (значения из реального окружения имеют приоритет).
func Load() (*Config, error) {
	// godotenv не перезаписывает уже выставленные переменные окружения.
	if err := godotenv.Load(); err != nil {
		slog.Debug("файл .env не найден, продолжаем без него", "err", err)
	}

	var missing []string
	req := func(key string) string {
		v, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(v) == "" {
			missing = append(missing, key)
		}
		return v
	}
	opt := func(key, fallback string) string {
		if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
			return v
		}
		return fallback
	}

	cfg := &Config{
		Keycloak: KeycloakAdminConfig{
			URL:           req("KEYCLOAK_URI"),
			AdminRealm:    req("KEYCLOAK_ADMIN_REALM"),
			AdminClientID: opt("KEYCLOAK_ADMIN_CLIENT_ID", "admin-cli"),
			AdminUsername: req("KEYCLOAK_ADMIN_USERNAME"),
			AdminPassword: secretStr(req("KEYCLOAK_ADMIN_PASSWORD")),
		},
		App: AppClientConfig{
			Realm:                  req("KEYCLOAK_REALM"),
			ClientID:               req("KEYCLOAK_CLIENT_ID"),
			RedirectURIs:           splitCSV(req("KEYCLOAK_REDIRECT_URIS")),
			WebOrigins:             splitCSV(req("KEYCLOAK_WEB_ORIGINS")),
			PostLogoutRedirectURIs: req("KEYCLOAK_LOGOUT_REDIRECT_URIS"),
		},
		Google: GoogleConfig{
			ClientID:     req("GOOGLE_CLIENT_ID"),
			ClientSecret: secretStr(req("GOOGLE_CLIENT_SECRET")),
		},
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("не заданы обязательные переменные окружения: %s", strings.Join(missing, ", "))
	}

	cfg.Keycloak.URL = strings.TrimRight(cfg.Keycloak.URL, "/")

	if _, err := url.ParseRequestURI(cfg.Keycloak.URL); err != nil {
		return nil, fmt.Errorf("KEYCLOAK_URI невалидный URL: %w", err)
	}

	slog.Info("конфигурация загружена",
		"keycloak_url", cfg.Keycloak.URL,
		"admin_realm", cfg.Keycloak.AdminRealm,
		"app_realm", cfg.App.Realm,
		"client_id", cfg.App.ClientID,
	)

	return cfg, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// secretStr — обёртка над string, которая скрывает значение в логах и дампах.
// %v, %s, %+v и fmt.Sprint* вернут "***" вместо реального значения.
type secretStr string

func (s secretStr) String() string   { return "***" }
func (s secretStr) GoString() string { return `secretStr("***")` }

// Val возвращает реальное значение — используй явно только там, где оно нужно.
func (s secretStr) Val() string { return string(s) }
