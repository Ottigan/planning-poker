set shell := ["bash", "-euo", "pipefail", "-c"]

TAILWIND_CSS_INPUT := "static/css/input.css"
TAILWIND_CSS_OUTPUT := "static/css/output.css"
GO_BUILD_OUTPUT := "bin/app"

default:
    @just --list

install:
    bun install
    go mod download

build: templ tailwind
    mkdir -p bin
    go build -o {{ GO_BUILD_OUTPUT }} .

test:
    go test ./...

clean:
    rm -rf node_modules tmp bin
    rm -f {{ TAILWIND_CSS_OUTPUT }}

[parallel]
dev: templ-watch tailwind-watch server-watch

[private]
templ:
    templ generate

[private]
tailwind:
    bunx @tailwindcss/cli -i {{ TAILWIND_CSS_INPUT }} -o {{ TAILWIND_CSS_OUTPUT }} --minify

[private]
templ-watch:
    templ generate --watch --proxy="http://localhost:8080" --open-browser=true -v

[private]
tailwind-watch:
    bunx @tailwindcss/cli -i {{ TAILWIND_CSS_INPUT }} -o {{ TAILWIND_CSS_OUTPUT }} --watch

[private]
server-watch:
    go run github.com/air-verse/air@v1.63.0 \
    --build.cmd "go build -o tmp/bin/main" \
    --build.bin "tmp/bin/main" \
    --build.delay "100" \
    --build.exclude_dir "node_modules" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit true
