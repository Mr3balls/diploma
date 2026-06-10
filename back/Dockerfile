FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
# Build migrate CLI for running DB migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/esports-api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app && apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/esports-api /app/esports-api
COPY --from=builder /go/bin/migrate /app/migrate
COPY migrations /app/migrations
# credentials/ is optional (Google Sheets). If absent, the folder is created empty.
COPY credentials/ /app/credentials/
USER app
EXPOSE 8080
CMD ["/app/esports-api"]
