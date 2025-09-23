COVERAGE_PROFILE ?= coverage.out
BIN_DIR  ?= ./bin
SRCS := $(shell find . -name '*.go' | grep -v bindata.go)

.PHONY: localdev_bootstrap
localdev_bootstrap:
	@echo "Starting local development environment"
	docker compose -f docker-compose.yml up -d


# DB COMMANDS

.PHONY: migration
migration:
	echo "---> Creating a new migration"
	echo "---> Input args... $(filename)"
	@migrate create -ext sql -dir ./db/migrations/ $(filename)

.PHONY: db-migrate
db-migrate:
	@echo "Running db migrations"
	go mod tidy
	@go run db/migrations/main.go up

.PHONY: db-migrate-down
db-migrate-down:
	@echo "Running db migrations down"
	go mod tidy
	@go run db/migrations/main.go down

.PHONY: db-rollback
db-rollback:
	@echo "Rolling back db migrations"
	go mod tidy
	@go run db/migrations/main.go down

# DEVELOPMENT COMMANDS

.PHONY: deps
deps:
	echo "Installing dependencies and making sure in sync"
	go mod download && \
	echo "succesfully downloaded dependencies"

.PHONY: lint
lint: ## Run linter
	@echo "---> Linting..."
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m || exit 1
	@echo "Running gofmt check..."
	@bash -c 'RES=$$(gofmt -l . | tee /dev/tty | wc -l); exit $$RES' && \
		echo "Successful!" || echo "Failed!"

.PHONY: lint-fix
lint-fix: ## Fix linting issues
	@echo "---> Linting"
	golangci-lint run --fix
	@echo "---> Fixing mockery..."
	$(BIN_DIR)/mockery --all --dir ./pkg --outpkg=mocks --output=./mocks

.PHONY: install
install: $(BIN_DIR)/golangci-lint $(BIN_DIR)/mockery

$(BIN_DIR)/golangci-lint:
	@echo "---> Installing golangci-lint..."
	@mkdir -p $(BIN_DIR)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) v1.55.2

$(BIN_DIR)/mockery:
	@echo "---> Installing mockery..."
	@mkdir -p $(BIN_DIR)
	@go install github.com/vektra/mockery/v2@latest
	@cp $(shell go env GOPATH)/bin/mockery $(BIN_DIR)/mockery

.PHONY: generate
generate: install
	@echo "---> Generating mocks"
	@$(BIN_DIR)/mockery
	
.PHONY: test
test:
	@echo "Running tests"
	mkdir -p build
	ENV=test go test -p 1 -cover -covermode=atomic -coverprofile=${COVERAGE_PROFILE} -timeout=30s ./...

.PHONY: test_ci
test_ci:
	@echo "Running CI tests"
	mkdir -p build
	ENV=test go test -p 1 -cover -covermode=atomic -coverprofile=${COVERAGE_PROFILE} -timeout=30s ./...

.PHONY: fmt
fmt:
	@echo "Formatting code"
	go fmt ./...


.PHONY: clean
clean: ## Clean up build artifacts
	@echo "---> Cleaning up build artifacts..."
	rm -rf build


.PHONY: build
build: ./build/service ## Build binary

./build/service: $(SRCS)
	@GOOS=darwin GOARCH=amd64 go generate ./...
	@echo "---> Building: (darwin, amd64)"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -race -o build/service cmd/server/main.go && \
		echo "Successful!"

	