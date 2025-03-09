wire:
	@echo "This command will add wire_gen.go in PATH={root}/internal/applications/{your-directory} make sure you already create {domain}_injector.go \nEnter directory: {your-directory} "; \
	read dir; \
	echo "Accessing directory and wire all DI $(WIRE_DIR)/$$dir"; \
	cd $(WIRE_DIR)/$$dir && wire