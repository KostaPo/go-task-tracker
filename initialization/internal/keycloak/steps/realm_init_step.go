package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

// RealmInitStep создаёт целевой realm, если он ещё не существует.
type RealmInitStep struct{}

func (RealmInitStep) Order() int   { return 1 }
func (RealmInitStep) Name() string { return "RealmInitStep" }

func (RealmInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	realms, err := rt.Admin.GoCloak.GetRealms(ctx, token)
	if err != nil {
		return fmt.Errorf("не удалось получить список realm'ов: %w", err)
	}

	for _, r := range realms {
		if r.Realm != nil && *r.Realm == realmName {
			slog.Info("Realm уже существует", "realm", realmName)
			return nil
		}
	}

	realm := gocloak.RealmRepresentation{
		Realm:                       gocloak.StringP(realmName),
		Enabled:                     gocloak.BoolP(true),
		RememberMe:                  gocloak.BoolP(true),
		RegistrationAllowed:         gocloak.BoolP(true),
		LoginWithEmailAllowed:       gocloak.BoolP(false),
		RegistrationEmailAsUsername: gocloak.BoolP(false),
		DuplicateEmailsAllowed:      gocloak.BoolP(true),
	}

	if _, err := rt.Admin.GoCloak.CreateRealm(ctx, token, realm); err != nil {
		return fmt.Errorf("не удалось создать realm %q: %w", realmName, err)
	}

	slog.Info("Realm создан", "realm", realmName)
	return nil
}
