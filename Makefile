build:
	go build -o ./tmp/main -v

start:
	air

css-minify:
	./tailwindcss -i static/css/input.css -o static/css/output.css --minify

css-watch:
	./tailwindcss -i static/css/input.css -o static/css/output.css --watch