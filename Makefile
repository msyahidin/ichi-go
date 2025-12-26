.PHONY: gen-build gen-install gen-clean gen-test gen-example

BINARY_NAME=ichigen
INSTALL_PATH=./

ichigen-build:
	@go build -o $(BINARY_NAME) pkg/ichigen/cmd/main.go
	@echo "✓ Built $(BINARY_NAME)"

ichigen-install: ichigen-build
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "✓ Installed to $(INSTALL_PATH)"

ichigen-clean:
	@rm -f $(BINARY_NAME)
	@echo "✓ Cleaned"

ichigen-test:
	@go test ./internal/generator/...

# Examples
ichigen-example-full:
	@go run ./pkg/ichigen/cmd/main.go g full product --domain=catalog --crud

ichigen-example-controller:
	@go run ./pkg/ichigen/cmd/main.go g controller notification --domain=alert