package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

// RolesInitStep создаёт client-роли приложения и добавляет роль USER
// в композит дефолтной realm-роли (default-roles-{realm}).
type RolesInitStep struct{}

var rolesToCreate = []string{"USER", "SELLER", "MODERATOR", "ADMIN"}

func (RolesInitStep) Order() int   { return 3 }
func (RolesInitStep) Name() string { return "RolesInitStep" }

func (RolesInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	clientUUID, err := rt.Realm.ResolveClientUUID(ctx, rt.Admin)
	if err != nil {
		return err
	}

	existingRoles, err := rt.Admin.GoCloak.GetClientRoles(ctx, token, realmName, clientUUID, gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("не удалось получить список ролей клиента: %w", err)
	}
	existingNames := make(map[string]struct{}, len(existingRoles))
	for _, r := range existingRoles {
		if r.Name != nil {
			existingNames[*r.Name] = struct{}{}
		}
	}

	for _, roleName := range rolesToCreate {
		if _, ok := existingNames[roleName]; ok {
			continue
		}
		role := gocloak.Role{Name: gocloak.StringP(roleName)}
		if _, err := rt.Admin.GoCloak.CreateClientRole(ctx, token, realmName, clientUUID, role); err != nil {
			return fmt.Errorf("не удалось создать роль %q: %w", roleName, err)
		}
		slog.Info("Роль создана", "role", roleName)
	}

	return assignUserRoleAsDefault(ctx, rt, token, realmName, clientUUID)
}

func assignUserRoleAsDefault(ctx context.Context, rt *keycloakinit.Runtime, token, realmName, clientUUID string) error {
	defaultRoleName := "default-roles-" + realmName

	composites, err := rt.Admin.GoCloak.GetCompositeRealmRoles(ctx, token, realmName, defaultRoleName)
	if err != nil {
		return fmt.Errorf("не удалось получить композитные роли %q: %w", defaultRoleName, err)
	}

	for _, c := range composites {
		if c.Name != nil && *c.Name == "USER" && c.ContainerID != nil && *c.ContainerID == clientUUID {
			slog.Info("USER уже входит в роль по умолчанию realm'а")
			return nil
		}
	}

	userRole, err := rt.Admin.GoCloak.GetClientRole(ctx, token, realmName, clientUUID, "USER")
	if err != nil {
		return fmt.Errorf("не удалось получить роль USER: %w", err)
	}

	if err := rt.Admin.GoCloak.AddRealmRoleComposite(ctx, token, realmName, defaultRoleName, []gocloak.Role{*userRole}); err != nil {
		return fmt.Errorf("не удалось добавить USER в роль по умолчанию: %w", err)
	}

	slog.Info("USER добавлена в роль по умолчанию realm'а")
	return nil
}
