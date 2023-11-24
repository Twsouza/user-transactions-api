
FROM golang:1.21

RUN mkdir -p /src/user-transactions
WORKDIR /src/user-transactions

ADD go.mod go.sum ./

RUN go install go.uber.org/mock/mockgen@latest && \
  go mod download

ENV ADDR=0.0.0.0

EXPOSE 3000
