package keycloakinit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
)

// AdminClient — обёртка над gocloak.GoCloak, которая умеет лениво
// логиниться под admin-пользователем и обновлять токен по истечении срока
// его жизни. В Java-версии за это бесплатно отвечал keycloak-admin-client
// (resteasy-прокси), в Go аналогичного автообновления из коробки нет,
// поэтому делаем это явно.
type AdminClient struct {
	GoCloak *gocloak.GoCloak
	BaseURL string

	adminRealm    string
	adminUsername string
	adminPassword string

	mu        sync.Mutex
	token     *gocloak.JWT
	expiresAt time.Time
}

// NewAdminClient создаёт клиент Keycloak Admin REST API.
//
// Примечание: gocloak.LoginAdmin всегда использует служебный клиент
// "admin-cli" — это поведение зашито в самой библиотеке и не настраивается,
// поэтому adminClientID из конфигурации здесь не используется напрямую,
// но сохраняется в конфигурации для наглядности и на случай ручных вызовов.
func NewAdminClient(baseURL, adminRealm, adminUsername, adminPassword string) *AdminClient {
	return &AdminClient{
		GoCloak:       gocloak.NewClient(baseURL),
		BaseURL:       baseURL,
		adminRealm:    adminRealm,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}
}

// Token возвращает действующий access-token администратора, при
// необходимости заново авторизуясь в Keycloak.
func (ac *AdminClient) Token(ctx context.Context) (string, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.token != nil && time.Now().Before(ac.expiresAt) {
		return ac.token.AccessToken, nil
	}

	token, err := ac.GoCloak.LoginAdmin(ctx, ac.adminUsername, ac.adminPassword, ac.adminRealm)
	if err != nil {
		return "", fmt.Errorf("не удалось авторизоваться в Keycloak под админ-пользователем: %w", err)
	}

	ac.token = token
	// небольшой запас на сетевые задержки между проверкой токена и его использованием
	const safetyMargin = 10 * time.Second
	ac.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn)*time.Second - safetyMargin)

	return token.AccessToken, nil
}
