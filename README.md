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

### Architecture

> Why I chose to use DDD and Clean Architecture?

DDD and Clean Architecture focus on the domain and have a clear separation of concerns, the business logic and the application complexity are in different levels, which contributes to the maintainability of the project, and also they are close to the SOLID principles making it more modular and easy to make changes.

Implementing the Bulk transaction was easier due to the architecture of the project, using the Open-Closed principle, I was able to add a channel to send the transactions without changing the existing code.

> Why I didn't add additional data or tables?

I focused on implementing the core functionality first, the provided fields were sufficient to capture all necessary details for a transaction. However, the project was structured in a way to easily add new features and data/tables.

> How I implemented the pagination and filtering?

I've used the postgres built-in Limit and Offset functions, as they let you easily paginate through the list. The filtering is done with Query method from GORM. The service layer handles filtering, ensuring that users can only filter by certain fields. To optimize the queries, I've added indexes to the fields that are used in the filters.

An alternative that can be implemented in the future is to use Keyset pagination, which is more efficient than Offset pagination.

> How I implemented bulk transactions? And why 100 transactions at a time or every second?

I did some tests on Postman with 100 virtual users, roughly the best results were achieved with 100 transactions. With more tests and varying numbers of users, this number could change. To ensure some consistency I chose to run at every second if the 100 transactions are not matched. The Bulk method is not perfect, but due to time constraints I implemented it in a simple way, if I had more time I'd add retry option, exponential backoff (with jitter), maybe send the transactions to a queue to be processed by another process. One thing that I missed was to configure the connection pool on GORM, that'd increase the total requests made and the response time.

> How would I implement notification?

I'd use the notification pattern, creating a transaction notification to which other components can subscribe, create a "queue" component that will subscribe to the transaction notification, when the transaction is created, it notifies the queue component which sends a message to the desired queue. This approach can be used to notify internal and external components.

Additional components could use the notification pattern to subscribe to the transaction notification and send to webhooks, if necessary.

> Why I used Postman to document the API?

It's a tool many developers use and are familiar with. I was already using it to make requests and test performance. It provides a simple and interactive way to use the documentation.

In larger projects and with more time, I'd use Swagger/OpenAPI to document the API, there are some tools that can generate the documentation from the code, which makes it easier to maintain (e.g. github.com/go-openapi/swag).

> What future improvements or additional features I'd add?

* Improve overall system resilience by implementing graceful shutdown, health checks, retry mechanisms, exponential backoff, circuit breakers, etc.
* Add a cache layer to improve the performance.
* Add E2E and performance tests.
* Add observability and monitoring.
