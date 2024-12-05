start:
	air
	
build: install css-minify htmx
	go build -o ./bin/app -v

htmx:
	templ generate

install:
	bun install && go mod tidy

css-minify:
	./tailwindcss -i static/css/input.css -o static/css/output.css --minify

css-watch:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch