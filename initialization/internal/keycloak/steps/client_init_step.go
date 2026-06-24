package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

// ClientInitStep создаёт публичный OIDC-клиент приложения, если он ещё не существует.
type ClientInitStep struct {
	ClientID               string
	RedirectURIs           []string
	WebOrigins             []string
	PostLogoutRedirectURIs string // формат Keycloak: значения через "##"
}

func NewClientInitStep(clientID string, redirectURIs, webOrigins []string, postLogoutRedirectURIs string) *ClientInitStep {
	return &ClientInitStep{
		ClientID:               clientID,
		RedirectURIs:           redirectURIs,
		WebOrigins:             webOrigins,
		PostLogoutRedirectURIs: postLogoutRedirectURIs,
	}
}

func (*ClientInitStep) Order() int   { return 2 }
func (*ClientInitStep) Name() string { return "ClientInitStep" }

func (s *ClientInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	existing, err := rt.Admin.GoCloak.GetClients(ctx, token, realmName, gocloak.GetClientsParams{
		ClientID: gocloak.StringP(s.ClientID),
	})
	if err != nil {
		return fmt.Errorf("не удалось получить список клиентов: %w", err)
	}
	if len(existing) > 0 {
		slog.Info("Клиент уже существует", "clientId", s.ClientID)
		return nil
	}

	client := gocloak.Client{
		ClientID:                  gocloak.StringP(s.ClientID),
		Name:                      gocloak.StringP(s.ClientID),
		PublicClient:              gocloak.BoolP(true),
		Protocol:                  gocloak.StringP("openid-connect"),
		StandardFlowEnabled:       gocloak.BoolP(true),
		DirectAccessGrantsEnabled: gocloak.BoolP(false),
		RedirectURIs:              &s.RedirectURIs,
		WebOrigins:                &s.WebOrigins,
		FrontChannelLogout:        gocloak.BoolP(true),
		Attributes: &map[string]string{
			"pkce.code.challenge.method": "S256",
			"post.logout.redirect.uris":  s.PostLogoutRedirectURIs,
			"use.refresh.tokens":         "true",
		},
	}

	if _, err := rt.Admin.GoCloak.CreateClient(ctx, token, realmName, client); err != nil {
		return fmt.Errorf("не удалось создать клиента %q: %w", s.ClientID, err)
	}

	slog.Info("Клиент создан", "clientId", s.ClientID)
	return nil
}
