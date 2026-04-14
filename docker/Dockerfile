# syntax=docker/dockerfile:1
FROM golang:1.26.2-alpine3.23 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/github-issue-cms .

FROM alpine:3.22

RUN adduser -D -u 10001 app && apk add --no-cache ca-certificates

COPY --from=builder /out/github-issue-cms /usr/local/bin/github-issue-cms

USER app

ENTRYPOINT ["/usr/local/bin/github-issue-cms"]
CMD ["--help"]
