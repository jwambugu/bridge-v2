FROM golang:1.19-alpine3.17 AS builder

WORKDIR /app

COPY . .

RUN go build -o server cmd/api/*
RUN apk add make
RUN make build-goose

FROM alpine:latest

WORKDIR /app

ENV APP_ENV=local

COPY internal/config/.local.env internal/config/.local.env
COPY internal/db/migrations internal/db/migrations
COPY --from=builder /app/server .
COPY --from=builder /app/bin/tools/goose .
COPY scripts/start.sh .

EXPOSE 8000
EXPOSE 8001

CMD ["/app/server"]

ENTRYPOINT ["/app/start.sh"]