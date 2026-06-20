# --- бинарник atlas, готовый, ничего собирать не нужно ---
FROM arigaio/atlas:latest-alpine AS atlas

# --- сборка Go-приложения ---
# Контекст сборки — КОРЕНЬ репозитория (там, где go.work),
# иначе go build не увидит workspace.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN echo "=== /src/initialization ===" && find /src/initialization -maxdepth 4
WORKDIR /src/initialization/cmd/app
RUN CGO_ENABLED=0 go build -o /migrate main.go

# --- финальный самодостаточный образ ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates

# бинарник atlas
COPY --from=atlas /atlas /atlas

# конфиг и шаблоны миграций — берём уже из build-стадии,
# так пути не зависят от того, откуда докер взял контекст
COPY --from=build /src/initialization/cmd/app/atlas.hcl /migrations/atlas.hcl
COPY --from=build /src/initialization/cmd/app/migrations-template /migrations/migrations-template

# наше Go-приложение
COPY --from=build /migrate /migrate

# контейнер стартует, выполняет миграции и завершается
ENTRYPOINT ["/migrate"]