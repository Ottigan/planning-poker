TAILWIND_CSS_INPUT := static/css/input.css
TAILWIND_CSS_OUTPUT := static/css/output.css
HTMX_GENERATE_CMD := templ generate
GO_BUILD_OUTPUT := bin/app

# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
dev/templ:
	$(HTMX_GENERATE_CMD) --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
dev/server:
	go run github.com/air-verse/air@v1.52.3 \
	--build.cmd "go build -o tmp/main" --build.bin "tmp/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle in watch mode.
dev/tailwind:
	./tailwindcss -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT) --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
dev/sync_assets:
	go run github.com/air-verse/air@v1.52.3 \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "static" \
	--build.include_ext "js,css"

# start all 4 watch processes in parallel.
dev: install 
	make -j4 dev/templ dev/server dev/tailwind dev/sync_assets

install:
	bun install && go mod tidy

build: install tailwind htmx
	@mkdir -p bin
	go build -o $(GO_BUILD_OUTPUT) .

tailwind:
	./tailwindcss -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT) --minify

htmx:
	$(HTMX_GENERATE_CMD)