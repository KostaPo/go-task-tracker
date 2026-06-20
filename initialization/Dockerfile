# --- бинарник atlas, готовый, ничего собирать не нужно ---
FROM arigaio/atlas:latest-alpine AS atlas

# --- сборка Go-приложения ---
# Контекст сборки — КОРЕНЬ репозитория (там, где go.work),
# иначе go build не увидит workspace.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/initialization/cmd/app
RUN CGO_ENABLED=0 go build -o /migrate main.go

# --- финальный самодостаточный образ ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates

# бинарник atlas — кладём в $PATH, чтобы его можно было находить просто по имени "atlas"
COPY --from=atlas /atlas /usr/local/bin/atlas

# конфиг и шаблоны миграций — реальный путь: initialization/ (не cmd/app/)
COPY --from=build /src/initialization/atlas.hcl /migrations/atlas.hcl
COPY --from=build /src/initialization/migrations-template /migrations/migrations-template

WORKDIR /migrations

# наше Go-приложение
COPY --from=build /migrate /migrate

# контейнер стартует, выполняет миграции и завершается
ENTRYPOINT ["/migrate"]