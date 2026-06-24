// Package config отвечает за чтение и валидацию конфигурации приложения
// из переменных окружения (с опциональной подгрузкой .env файла).
package config

import (
	"fmt"
	"os"
	"strings"
)

// KeycloakAdminConfig — параметры подключения к Keycloak под админской учёткой.
// Аналог @Value("${keycloak...}") полей KeycloakInitializer в Java-версии.
type KeycloakAdminConfig struct {
	URL           string // KEYCLOAK_URI
	AdminRealm    string // KEYCLOAK_ADMIN_REALM (обычно "master")
	AdminClientID string // KEYCLOAK_ADMIN_CLIENT_ID (обычно "admin-cli")
	AdminUsername string // KEYCLOAK_ADMIN_USERNAME
	AdminPassword string // KEYCLOAK_ADMIN_PASSWORD
}

// AppClientConfig — параметры realm'а и клиента приложения.
// Аналог @Value("${app...}") полей KeycloakContext и ClientInitStep.
type AppClientConfig struct {
	Realm        string   // KEYCLOAK_REALM
	ClientID     string   // KEYCLOAK_CLIENT_ID
	RedirectURIs []string // KEYCLOAK_REDIRECT_URIS (через запятую)
	WebOrigins   []string // KEYCLOAK_WEB_ORIGINS (через запятую)

	// PostLogoutRedirectURIs хранится как есть, без разбиения:
	// Keycloak ожидает значение атрибута "post.logout.redirect.uris"
	// в виде строк, объединённых через "##" — это формат самого Keycloak,
	// а не наш собственный разделитель, поэтому мы его не трогаем.
	PostLogoutRedirectURIs string // KEYCLOAK_LOGOUT_REDIRECT_URIS
}

// GoogleConfig — параметры Google OAuth для настройки Identity Provider.
type GoogleConfig struct {
	ClientID     string // GOOGLE_CLIENT_ID
	ClientSecret string // GOOGLE_CLIENT_SECRET
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
	loadDotEnvIfPresent(".env")

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
			AdminPassword: req("KEYCLOAK_ADMIN_PASSWORD"),
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
			ClientSecret: req("GOOGLE_CLIENT_SECRET"),
		},
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("не заданы обязательные переменные окружения: %s", strings.Join(missing, ", "))
	}

	cfg.Keycloak.URL = strings.TrimRight(cfg.Keycloak.URL, "/")

	return cfg, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// loadDotEnvIfPresent — минимальный загрузчик .env без внешних зависимостей:
// разбирает строки вида KEY=VALUE и проставляет их в окружение процесса,
// только если переменная ещё не задана снаружи.
func loadDotEnvIfPresent(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // .env необязателен — например, в проде переменные задаются оркестратором
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
