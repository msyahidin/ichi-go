wire:
	@echo "This command will add wire_gen.go in PATH={root}/internal/applications/{your-directory} make sure you already create {domain}_injector.go \nEnter directory: {your-directory} "; \
	read dir; \
	echo "Accessing directory and wire all DI $(WIRE_DIR)/$$dir"; \
	cd $(WIRE_DIR)/$$dir && wire


ENT_DIR := $(shell pwd)/internal/infrastructure/database/ent
# Generate ent models
schema:
	go generate ./internal/infrastructure/database/ent

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
	go run ariga.io/entimport/cmd/entimport -dsn "mysql://root:password@tcp(localhost:3306)/oauth" -tables $$schema -schema-path $$ENT_DIR