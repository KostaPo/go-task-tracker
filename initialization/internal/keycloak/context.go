package keycloakinit

import (
	"context"
	"fmt"
	"sync"

	"github.com/Nerzal/gocloak/v13"
)

// RealmContext хранит сведения о целевом realm'е и клиенте приложения,
// общие для всех шагов инициализации. Аналог Java-класса KeycloakContext.
//
// UUID клиента (внутренний id записи в Keycloak, отличный от публичного
// clientId) резолвится лениво при первом обращении и кешируется —
// это безопасно делать после того, как отработал ClientInitStep.
type RealmContext struct {
	RealmName string
	ClientID  string

	mu               sync.Mutex
	cachedClientUUID string
}

// NewRealmContext создаёт контекст realm'а приложения.
func NewRealmContext(realmName, clientID string) *RealmContext {
	return &RealmContext{RealmName: realmName, ClientID: clientID}
}

// ResolveClientUUID возвращает внутренний UUID клиента приложения в Keycloak.
// Безопасно вызывать после того, как ClientInitStep уже отработал.
func (rc *RealmContext) ResolveClientUUID(ctx context.Context, admin *AdminClient) (string, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.cachedClientUUID != "" {
		return rc.cachedClientUUID, nil
	}

	token, err := admin.Token(ctx)
	if err != nil {
		return "", err
	}

	clients, err := admin.GoCloak.GetClients(ctx, token, rc.RealmName, gocloak.GetClientsParams{
		ClientID: gocloak.StringP(rc.ClientID),
	})
	if err != nil {
		return "", fmt.Errorf("не удалось получить список клиентов realm'а %q: %w", rc.RealmName, err)
	}

	for _, c := range clients {
		if c.ID != nil {
			rc.cachedClientUUID = *c.ID
			return rc.cachedClientUUID, nil
		}
	}

	return "", fmt.Errorf("клиент %q не найден в realm'е %q: убедитесь, что ClientInitStep выполнился раньше", rc.ClientID, rc.RealmName)
}
