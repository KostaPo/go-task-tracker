package keycloakinit

import "context"

// Runtime агрегирует всё, что нужно шагу инициализации для работы:
// admin-клиент Keycloak и контекст целевого realm'а/клиента.
type Runtime struct {
	Admin *AdminClient
	Realm *RealmContext
}

// Step описывает единичный, идемпотентный шаг инициализации Keycloak.
// Аналог интерфейса KeycloakInitStep из Java-версии.
type Step interface {
	// Order определяет порядок выполнения шагов: меньшее значение — раньше.
	// Должен быть уникальным в пределах одного запуска.
	Order() int

	// Name используется только для логирования.
	Name() string

	// Execute выполняет шаг. Реализация обязана быть идемпотентной:
	// повторный запуск при уже примененных изменениях не должен ничего ломать.
	Execute(ctx context.Context, rt *Runtime) error
}
