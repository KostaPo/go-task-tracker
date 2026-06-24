package steps

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
	"github.com/kostapo/tasktracker/initialization/internal/keycloak/userprofile"
)

// UserProfileInitStep приводит конфигурацию User Profile realm'а
// к желаемому набору атрибутов: username, email, custom.
type UserProfileInitStep struct {
	client *userprofile.Client
}

func NewUserProfileInitStep(keycloakBaseURL string) *UserProfileInitStep {
	return &UserProfileInitStep{client: userprofile.NewClient(keycloakBaseURL)}
}

func (*UserProfileInitStep) Order() int   { return 5 }
func (*UserProfileInitStep) Name() string { return "UserProfileInitStep" }

func (s *UserProfileInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	current, err := s.client.Get(ctx, token, realmName)
	if err != nil {
		return fmt.Errorf("не удалось получить конфигурацию user profile: %w", err)
	}

	desired := map[string]struct{}{"username": {}, "email": {}, "custom": {}}
	existing := current.AttributeNames()

	if reflect.DeepEqual(existing, desired) {
		slog.Info("User profile уже актуален")
		return nil
	}

	newConfig := &userprofile.Config{
		Attributes: []userprofile.Attribute{
			buildUsernameAttribute(),
			buildEmailAttribute(),
			buildCustomAttribute(),
		},
	}

	if err := s.client.Update(ctx, token, realmName, newConfig); err != nil {
		return fmt.Errorf("не удалось обновить user profile: %w", err)
	}

	slog.Info("User profile обновлён")
	return nil
}

func buildUsernameAttribute() userprofile.Attribute {
	return userprofile.Attribute{
		Name:        "username",
		DisplayName: "${username}",
		Permissions: userprofile.Permissions{
			View: []string{"admin"},
			Edit: []string{"admin"},
		},
		Required:    &userprofile.Required{Roles: []string{"user"}},
		Multivalued: false,
		Validations: map[string]any{
			"length":                         map[string]any{"min": 3, "max": 128},
			"username-prohibited-characters": map[string]any{},
			"up-username-not-idn-homograph":  map[string]any{},
		},
	}
}

func buildEmailAttribute() userprofile.Attribute {
	return userprofile.Attribute{
		Name:        "email",
		DisplayName: "${email}",
		Permissions: userprofile.Permissions{
			View: []string{"user", "admin"},
			Edit: []string{"user", "admin"},
		},
		Multivalued: false,
	}
}

func buildCustomAttribute() userprofile.Attribute {
	return userprofile.Attribute{
		Name:        "custom",
		DisplayName: "${custom}",
		Permissions: userprofile.Permissions{
			View: []string{"user", "admin"},
			Edit: []string{"user", "admin"},
		},
		Multivalued: false,
	}
}
