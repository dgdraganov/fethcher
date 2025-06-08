# Fethcher - Ethereum Transaction Fetcher

## Overview

Fethcher is a Go-based service that provides an API for fetching Ethereum transaction details from both a local database and the Ethereum network. It serves as a caching layer between clients and the Ethereum network, offering these key capabilities:

- Retrieves transaction details from both local database cache and Ethereum node
- Automatically caches transactions fetched from Ethereum network
- Maintains per-user history of all transactions queried
- Supports JWT-based authentication
- Handles both individual hashes and RLP-encoded transaction bundles

## Features

- **Authentication**: Secure JWT-based user authentication
- **Transaction Lookup**: Fetch transaction details by hash
- **RLP Support**: Parse RLP-encoded transaction hashes
- **User History**: Track and retrieve user transaction query history
- **Caching**: Automatically caches Ethereum network transactions in local database
- **Concurrent Fetching**: Efficiently fetches multiple transactions in parallel

## API Endpoints

### Authentication
- `POST /lime/authenticate` - Authenticate user and get JWT token

### Transaction Operations
- `GET /lime/eth` - Get transactions by hash (query parameter: transactionHashes)
- `GET /lime/eth/{rlpHash}` - Get transactions by RLP-encoded hash
- `GET /lime/all` - Get all transactions from database
- `GET /lime/my` - Get current user's transaction history

## Prerequisites

- Go 1.20+
- PostgreSQL
- Ethereum node access (Infura or similar)

## Environment variables

The service requires the below environment variables:

DB_CONNECTION_URL=host=db user=postgres password=postgres dbname=fethcher sslmode=disable
ETH_NODE_URL=https://mainnet.infura.io/v3/user_infura_token
API_PORT=9205
JWT_SECRET=4a03e22b74d3fc8edeff82390dc72f27c0a0bbf4c4e824a9ed15f1612a1c5cef

## How to run it? 

There is a `docker-compose.yaml` file that will use default values for the environment variables (including my own infura API key!). Included is a `Dockerfile` that will build a docker image for this service:

```bash
    make up
```

and

```bash
    make build-image
```

To bring the compose services down use the `make` target:

```bash
    make down
```

## Tests

Here is how to run the tests: 

```bash
    make test
```

## Seeded data

Upon start there will be several users with respective passwords seeded into the database. These can be used for authentication as there is no functionality for registering new ones:

- `alice`/ `alice`
- `bob`/ `bob`
- `carol` / `carol`
- `dave`/ `dave`
 

