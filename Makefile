TAILWIND_CSS_INPUT := static/css/input.css
TAILWIND_CSS_OUTPUT := static/css/output.css
HTMX_GENERATE_CMD := templ generate
GO_BUILD_OUTPUT := bin/app

install:
	bun install && go mod tidy

start:
	@pkill -f "air" || true
	air

build: install css-minify htmx
	@mkdir -p bin
	go build -o $(GO_BUILD_OUTPUT) .

htmx:
	$(HTMX_GENERATE_CMD)

htmx-watch:
	$(HTMX_GENERATE_CMD) --watch --proxy=http://localhost:8080

css:
	./tailwindcss -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT)

css-minify:
	./tailwindcss -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT) --minify

css-watch:
	./tailwindcss -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT) --watch

clean:
	rm -rf bin
	rm -rf dist
