FROM golang:1.24-alpine3.20 AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /go/src/github.com/rokuosan/github-issue-cms

ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go/bin/github-issue-cms .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/github-issue-cms /usr/local/bin/github-issue-cms

ENTRYPOINT ["/usr/local/bin/github-issue-cms"]
CMD ["--help"]
