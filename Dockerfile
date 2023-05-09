FROM golang:1.19-alpine3.17 AS builder

WORKDIR /app

RUN adduser -S appuser

COPY go.mod go.sum  ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN apk add make
RUN make build-goose

RUN apk --no-cache add gcc g++ git

RUN go build \
     -ldflags="-linkmode external -extldflags -static"\
     -tags netgo\
    -o server cmd/api/*

FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd
USER appuser

LABEL version="1.0.0"
LABEL author="jwambugu"

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
