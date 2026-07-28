BINARY := redfish_api_mock
CONFIG := config.json
DEFAULT_CONFIG := config.json.default
DIST_DIR := dist
PACKAGE_DIR := $(DIST_DIR)/$(BINARY)
PACKAGE := $(DIST_DIR)/$(BINARY).tar.gz

.PHONY: all build config run package test clean

all: build

build:
	go build -o $(BINARY) .

config:
	@if [ ! -f $(CONFIG) ]; then \
		cp $(DEFAULT_CONFIG) $(CONFIG); \
		echo "Created $(CONFIG) from $(DEFAULT_CONFIG)"; \
	fi

run: config build
	./$(BINARY) -config $(CONFIG) $(ARGS)

package: config build
	rm -rf $(PACKAGE_DIR) $(PACKAGE)
	mkdir -p $(PACKAGE_DIR)
	cp $(BINARY) $(CONFIG) $(PACKAGE_DIR)/
	tar -C $(DIST_DIR) -czf $(PACKAGE) $(BINARY)
	@echo "Created $(PACKAGE)"

test:
	go test ./...

clean:
	rm -rf $(BINARY) $(DIST_DIR)
