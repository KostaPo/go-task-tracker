package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

const googleIdpAlias = "google"

// GoogleIdpInitStep подключает Google как Identity Provider, если он ещё не настроен.
type GoogleIdpInitStep struct {
	ClientID     string
	ClientSecret string
}

func NewGoogleIdpInitStep(clientID, clientSecret string) *GoogleIdpInitStep {
	return &GoogleIdpInitStep{ClientID: clientID, ClientSecret: clientSecret}
}

func (*GoogleIdpInitStep) Order() int   { return 6 }
func (*GoogleIdpInitStep) Name() string { return "GoogleIdpInitStep" }

func (s *GoogleIdpInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	idps, err := rt.Admin.GoCloak.GetIdentityProviders(ctx, token, realmName)
	if err != nil {
		return fmt.Errorf("не удалось получить список identity providers: %w", err)
	}
	for _, p := range idps {
		if p.Alias != nil && *p.Alias == googleIdpAlias {
			slog.Info("Google IdP уже существует")
			return nil
		}
	}

	google := gocloak.IdentityProviderRepresentation{
		Alias:                     gocloak.StringP(googleIdpAlias),
		ProviderID:                gocloak.StringP("google"),
		Enabled:                   gocloak.BoolP(true),
		FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
		Config: &map[string]string{
			"clientId":     s.ClientID,
			"clientSecret": s.ClientSecret,
			"defaultScope": "openid email profile",
		},
	}

	if _, err := rt.Admin.GoCloak.CreateIdentityProvider(ctx, token, realmName, google); err != nil {
		return fmt.Errorf("не удалось создать Google IdP: %w", err)
	}

	slog.Info("Google IdP создан")
	return nil
}
