.PHONY: build run

build:
	CGO_ENABLED=1 go build -o vn_index_systray

run: build
	./vn_index_systray
