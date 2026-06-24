package keycloakinit

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
)

// Run выполняет переданные шаги в порядке возрастания Step.Order().
// При ошибке любого шага выполнение прекращается и ошибка оборачивается
// с указанием шага, на котором всё остановилось — это соответствует
// поведению Java-варианта, где необработанное исключение в forEach
// прерывает дальнейшую инициализацию.
func Run(ctx context.Context, rt *Runtime, steps []Step) error {
	ordered := make([]Step, len(steps))
	copy(ordered, steps)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Order() < ordered[j].Order() })

	slog.Info("Запуск инициализации Keycloak...")

	for _, step := range ordered {
		slog.Info("Выполняется шаг", "order", step.Order(), "step", step.Name())
		if err := step.Execute(ctx, rt); err != nil {
			return fmt.Errorf("шаг %q (order=%d) завершился с ошибкой: %w", step.Name(), step.Order(), err)
		}
	}

	slog.Info("Инициализация Keycloak завершена! ✅")
	return nil
}
