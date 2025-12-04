FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler .

FROM ubuntu:latest

WORKDIR /app

RUN mkdir /db

COPY --from=builder /app/scheduler /app/scheduler
COPY --from=builder /app/web /app/web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/db/scheduler.db
ENV TODO_PASSWORD=""

EXPOSE 7540

CMD ["/app/scheduler"]