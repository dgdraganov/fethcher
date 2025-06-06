GOLANGCI_LINT_VERSION := v2.1.6

run:
	# API_PORT=9205 \
	# ETH_NODE_URL=fake-node-url \
	# DB_CONNECTION_URL=fake_db_conn_string \
	# JWT_SECRET=5up3r_53cr3t \
	go run main.go

test:
	go test -v ./...

up:
	docker compose up -d --build

down:
	docker compose down

install-deps:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint:
	golangci-lint run

gen-fakes:
	go get github.com/maxbrunsfeld/counterfeiter/v6
	go generate ./...

build-image:
	docker build -t limeapi .