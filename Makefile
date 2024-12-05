build:
	go build -o ./tmp/main -v

start:
	air

css-minify:
	./tailwindcss -i css/input.css -o css/output.css --minify

css-watch:
	./tailwindcss -i css/input.css -o css/output.css --watch