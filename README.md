# [WIP] - Bridge

A simple backend service built using go and gRPC. This implementation uses
the [Hexagonal architecture](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)) which allows the business
logic to not rely on the data sources which can be easily swapped on demand.

The projects also uses [gRPC gateway](https://github.com/grpc-ecosystem/grpc-gateway) to proxy gRPC to JSON following
the gRPC HTTP spec.

The service can:

- Login an existing user.
- Register a new user.
- Get auth user details.
- Update auth user details.

For unit tests, we use [dockertest](https://github.com/ory/dockertest) to boot up containers used to make
integration tests easier and also [vault](https://www.vaultproject.io/) for managing secrets.

### Endpoints

> TODO: Add gRPC and grpc-gateway endpoints

### Roadmap

- [x] Store credentials on vault
- [] Add observability using OpenTelemetry
- [] Add a worker to run background tasks

## Running on Docker [Requires docker]

To run the project on docker, run the following command:

```bash
  make compose-up
```

## Run Locally

> ⚠️ Requires postgres

Clone the project

```bash
  git clone https://github.com/jwambugu/bridge-v2.git
```

Go to the project directory

```bash
  cd bridge-v2
```

Copy the config file `internal/config/.example.env` file to `internal/config/.test.env`

Update the config as per your credentials.

```bash
   cp internal/config/.example.env internal/config/example.env 
```

Run the migrations

```bash
   make build-goose && sh scripts/goose.sh up
```

Start the two webservers

- gRPC server
- grpc gateway server

```bash
  go run cmd/api/*.go
```

## Running Tests

> ⚠️ Requires postgres and updated config - see above.

To run tests, run the following command:

```bash
  make test
```

[//]: # (https://stackoverflow.com/questions/129693/is-duplicated-code-more-tolerable-in-unit-tests)
