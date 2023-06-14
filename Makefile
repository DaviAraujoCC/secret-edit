all: build

build:
	@echo "Building..."
	go mod tidy
	CGO_ENABLED=0  go build -o bin/secret-edit

install:
	@echo "Installing..."
	@sudo mv bin/secret-edit /usr/local/bin/secret-edit
