# User transactions

An API service to save, retrieve and list user transactions.

## Development

### Requirements

- Docker
- Docker-Compose
- (Optional) VS Code with Remote Containers extension

### How to run

#### Using VS Code

1. Open the project in VS Code
2. Install the Remote Containers extension
3. Click on the button in the bottom left corner.
4. Select `Remote Containers: Reopen in Container`

#### Using Docker Compose

1. Run `docker-compose -f .devcontainer/docker-compose.yml up -d`

To enter inside of the container run:

```bash
docker-compose -f .devcontainer/docker-compose.yml exec transactions bash
```

#### How it works

It creates and run the containers with all dependencies installed. You can develop in the IDE of your choice and the code will be synced with the container.

#### How to run tests

If you are using VS Code, you can run the tests using the opening a terminal inside the container and running `go test ./...`.

If you are using Docker Compose, you can run the tests using `docker-compose -f .devcontainer/docker-compose.yml exec transactions go test ./...`.

#### How to run integration tests

Some of the tests requires a database to be running. To run the integration tests, run `docker-compose -f .devcontainer/docker-compose.yml exec transactions go test -tags=integration ./...`.

When creating an integration test, you should use the `integration` tag to make sure that the test will only run when the tag is provided.

```go
//go:build integration
// +build integration

package package_test
```

### How to use mock

The [gomock](https://github.com/uber-go/mock) library is used to generate mocks. To generate the mocks, run `mockgen -source=path/to/interface.go -destination=path/to/mock/interface_mock.go` inside the container.

## Production

To simulate the production environment, run `make run` and will spin up the containers (server application and database) and run the application.

The `Dockerfile` is already configured to build the application for production environment and run it (it's 28MB).

## Documentation

To provide a better understanding of the API, the documentation was created using Postman and is live on https://documenter.getpostman.com/view/2433332/2s9YeD8YrB. Also the Postman collection is available on the root of the project.
