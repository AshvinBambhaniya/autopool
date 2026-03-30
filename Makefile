.PHONY: test bench lint fmt help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  test    Run all unit and integration tests"
	@echo "  bench   Run performance benchmarks"
	@echo "  lint    Run static analysis (go vet)"
	@echo "  fmt     Format all Go files"

test:
	go test -v ./...

bench:
	go test -v -bench=. -benchmem .

lint:
	go vet ./...

fmt:
	go fmt ./...
