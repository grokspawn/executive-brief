.PHONY: build install clean test

# Binary name
BINARY=exec-brief

# Build the binary
build:
	go build -o $(BINARY) .

## Install to ~/bin
#install: build
#	mkdir -p ~/bin
#	cp $(BINARY) ~/bin/
#	@echo "Installed to ~/bin/$(BINARY)"
#	@echo "Make sure ~/bin is in your PATH"

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f exec-brief-*.md exec-brief-*.html

# Run tests
test:
	go test ./...

# Run the brief
run: build
	./$(BINARY) --daily

# Format code
fmt:
	go fmt ./...
