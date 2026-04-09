# Variables
TAILWIND_CSS_INPUT := static/css/input.css
TAILWIND_CSS_OUTPUT := static/css/output.css
TEMPL_GENERATE_CMD := templ generate
GO_BUILD_OUTPUT := bin/app

install:
	bun install && go mod tidy

clean:
	rm -rf node_modules
	rm -rf tmp
	rm -rf bin
	rm -f $(TAILWIND_CSS_OUTPUT)
	clear

html:
	$(TEMPL_GENERATE_CMD)

tailwind:
	bunx @tailwindcss/cli -i $(TAILWIND_CSS_INPUT) -o $(TAILWIND_CSS_OUTPUT) --minify

build: install tailwind html
	@mkdir -p bin
	go build -o $(GO_BUILD_OUTPUT) .

dev: build
	make -j2 dev/templ dev/server

# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# default url: http://localhost:7331
dev/templ:
	$(TEMPL_GENERATE_CMD) --watch --proxy="http://localhost:8080" --open-browser=true -v

# run air to detect any go file changes to re-build and re-run the server.
dev/server:
	go run github.com/air-verse/air@v1.63.0 \
	--build.cmd "make tailwind && go build -o tmp/bin/main" \
	--build.bin "tmp/bin/main" \
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true
