package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

// UsersInitStep создаёт двух демонстрационных пользователей с ролями,
// если они ещё не существуют: user/user (роль USER) и admin/admin (роль ADMIN).
type UsersInitStep struct{}

func (UsersInitStep) Order() int   { return 7 }
func (UsersInitStep) Name() string { return "UsersInitStep" }

func (UsersInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	clientUUID, err := rt.Realm.ResolveClientUUID(ctx, rt.Admin)
	if err != nil {
		return err
	}

	if err := createUserIfAbsent(ctx, rt, clientUUID, "user", "user", "USER"); err != nil {
		return err
	}
	return createUserIfAbsent(ctx, rt, clientUUID, "admin", "admin", "ADMIN")
}

func createUserIfAbsent(ctx context.Context, rt *keycloakinit.Runtime, clientUUID, username, password, roleName string) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	found, err := rt.Admin.GoCloak.GetUsers(ctx, token, realmName, gocloak.GetUsersParams{
		Username: gocloak.StringP(username),
		Exact:    gocloak.BoolP(true),
	})
	if err != nil {
		return fmt.Errorf("не удалось выполнить поиск пользователя %q: %w", username, err)
	}
	if len(found) > 0 {
		slog.Info("Пользователь уже существует", "username", username)
		return nil
	}

	user := gocloak.User{
		Username: gocloak.StringP(username),
		Enabled:  gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(password),
				Temporary: gocloak.BoolP(false),
			},
		},
	}

	userID, err := rt.Admin.GoCloak.CreateUser(ctx, token, realmName, user)
	if err != nil {
		return fmt.Errorf("не удалось создать пользователя %q: %w", username, err)
	}

	role, err := rt.Admin.GoCloak.GetClientRole(ctx, token, realmName, clientUUID, roleName)
	if err != nil {
		return fmt.Errorf("не удалось получить роль %q: %w", roleName, err)
	}

	if err := rt.Admin.GoCloak.AddClientRolesToUser(ctx, token, realmName, clientUUID, userID, []gocloak.Role{*role}); err != nil {
		return fmt.Errorf("не удалось назначить роль %q пользователю %q: %w", roleName, username, err)
	}

	slog.Info("Пользователь создан", "username", username, "role", roleName)
	return nil
}
