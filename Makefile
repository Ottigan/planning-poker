start:
	air

build: install css-minify htmx
	go build -o ./bin/app -v

htmx:
	templ generate

static-watch:
	templ generate --watch --proxy="http://localhost:8080" --cmd="make css"

install:
	bun install && go mod tidy

css:
	./tailwindcss -i static/css/input.css -o static/css/output.css

css-minify:
	./tailwindcss -i static/css/input.css -o static/css/output.css --minify

css-watch:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch