wire:
	@echo "This command will add wire_gen.go in PATH={root}/internal/applications/{your-directory} make sure you already create {domain}_injector.go \nEnter directory: {your-directory} "; \
	read dir; \
	echo "Accessing directory and wire all DI $(WIRE_DIR)/$$dir"; \
	cd $(WIRE_DIR)/$$dir && wire

# Build the project
build: test
	@echo "Building the project..."
	go build -o main cmd/main.go
	@echo "Build complete! Executable file: main"

# Run tests and generate coverage report
test:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

ENT_DIR := $(shell pwd)/internal/infra/database/ent
# Generate ent models
schema:
	go generate ./internal/infra/database/ent

schema-advance:
	@echo "*****\n Advance mode will granted you a super power, use it wisely\n [Generate with entgo feature sql/modifier,sql/execquery]\n*****";
	@cd $(ENT_DIR) && go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/modifier,sql/execquery ./schema

schema-import:
	@echo "This command will create ent schema and generate ent code";
	@read -p "Enter schema name: " schema; \
	if [ -z "$$schema" ]; then \
		echo "Schema name cannot be empty!"; \
		exit 1; \
	fi;
	@echo "Creating ent schema and generating ent code at $$ENT_DIR/$$schema";
	go run ariga.io/entimport/cmd/entimport -dsn "mysql://root:password@tcp(localhost:3306)/ichigo_db" -tables $$schema -schema-path $$ENT_DIR

migration-build:
	@echo "Warning this action will build unix executable file "
	go build -v -o migration db/cmd/main.go

migration-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: You must provide 'name' arguments."; \
		echo "Sample: make migration-create name=create_users_table type=sql"; \
		exit 1; \
	fi
	@if [ ! -f "./migration" ]; then \
		$(MAKE) migration-build; \
	fi
	@echo "🛠️  Creating migration: $(name) [$(type)]"
	@./migration mysql create $(name) $(type)

migration-up:
	@if [ ! -f "./migration" ]; then \
		$(MAKE) migration-build; \
    fi
	./migration mysql up

migration-down:
	@if [ ! -f "./migration" ]; then \
		$(MAKE) migration-build; \
    fi
	./migration mysql down

migration-down-to:
	@if [ ! -f "./migration" ]; then \
		$(MAKE) migration-build; \
    fi
	@if [ -z "$(version)" ]; then \
		echo "Error: Version cannot be empty."; \
		exit 1; \
	elif ! [[ "$(version)" =~ ^[0-9]+$$ ]]; then \
		echo "Error: Version must be a positive integer."; \
		exit 1; \
	elif [ "$(version)" = "0" ]; then \
		echo "Error: Version 0 is not allowed."; \
		exit 1; \
	elif ./migration mysql status | grep -q "Current: $(version)"; then \
		echo "Error: Version $(version) is the current version or higher."; \
		exit 1; \
	else \
		echo "Warning: This action will rollback to version $(version)"; \
		go build -v -o migration migrations/cmd/main.go; \
		./migration mysql down-to $(version); \
	fi

migration-status:
	@if [ ! -f "./migration" ]; then \
  		$(MAKE) migration-build; \
    fi
	./migration mysql status