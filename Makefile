
run:
	API_PORT=9205 \
	ETH_NODE_URL=fake-node-url \
	DB_CONNECTION_URL=fake_db_conn_string \
	JWT_SECRET=5up3r_53cr3t \
	go run main.go

test:
	go test -v ./...

compose:
	docker compose up -d --build

decompose:
	docker compose down
