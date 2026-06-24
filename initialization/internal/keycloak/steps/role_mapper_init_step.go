package steps

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nerzal/gocloak/v13"

	keycloakinit "github.com/kostapo/tasktracker/initialization/internal/keycloak"
)

const roleMapperName = "my-app-roles-mapper"

// RoleMapperInitStep добавляет protocol mapper, который кладёт client-роли
// пользователя в claim "roles" с префиксом "ROLE_".
type RoleMapperInitStep struct{}

func (RoleMapperInitStep) Order() int   { return 4 }
func (RoleMapperInitStep) Name() string { return "RoleMapperInitStep" }

func (RoleMapperInitStep) Execute(ctx context.Context, rt *keycloakinit.Runtime) error {
	token, err := rt.Admin.Token(ctx)
	if err != nil {
		return err
	}

	realmName := rt.Realm.RealmName

	clientUUID, err := rt.Realm.ResolveClientUUID(ctx, rt.Admin)
	if err != nil {
		return err
	}

	client, err := rt.Admin.GoCloak.GetClient(ctx, token, realmName, clientUUID)
	if err != nil {
		return fmt.Errorf("не удалось получить клиента: %w", err)
	}

	if client.ProtocolMappers != nil {
		for _, m := range *client.ProtocolMappers {
			if m.Name != nil && *m.Name == roleMapperName {
				slog.Info("Role mapper уже существует")
				return nil
			}
		}
	}

	mapper := gocloak.ProtocolMapperRepresentation{
		Name:           gocloak.StringP(roleMapperName),
		Protocol:       gocloak.StringP("openid-connect"),
		ProtocolMapper: gocloak.StringP("oidc-usermodel-client-role-mapper"),
		Config: &map[string]string{
			"multivalued":                            "true",
			"claim.name":                             "roles",
			"jsonType.label":                         "String",
			"access.token.claim":                     "true",
			"id.token.claim":                         "true",
			"userinfo.token.claim":                   "true",
			"usermodel.clientRoleMapping.rolePrefix": "ROLE_",
		},
	}

	if _, err := rt.Admin.GoCloak.CreateClientProtocolMapper(ctx, token, realmName, clientUUID, mapper); err != nil {
		return fmt.Errorf("не удалось создать role mapper: %w", err)
	}

	slog.Info("Role mapper создан")
	return nil
}
