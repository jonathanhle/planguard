.PHONY: build install clean test

# Build the converter
build:
	go build -o convert-terrascan .

# Install to $GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	rm -f convert-terrascan

# Run with example
test: build
	./convert-terrascan -file ../example-s3-versioning.rego --dry-run
