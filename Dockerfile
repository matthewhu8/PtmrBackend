# Use the official Golang image to create a build artifact.
FROM golang:1.22 as builder
WORKDIR /app
COPY . .

ARG GH_TOKEN
RUN git config --global url."https://${GH_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

# Build UserService
WORKDIR /app/cmd/UserService
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/UserService/main .

# Build ApplicationService
WORKDIR /app/cmd/ApplicationService
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/ApplicationService/main .

# Build MatchingService
WORKDIR /app/cmd/MatchingService
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/MatchingService/main .

# Build JobWriter
WORKDIR /app/cmd/JobWriter
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/JobWriter/main .

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates bash

# Create directories for each service
WORKDIR /app
RUN mkdir -p /app/UserService /app/ApplicationService /app/MatchingService /app/JobWriter

# Copy the binaries and env files into their respective directories
COPY --from=builder /app/UserService /app/UserService
COPY --from=builder /app/ApplicationService /app/ApplicationService
COPY --from=builder /app/MatchingService /app/MatchingService
COPY --from=builder /app/JobWriter /app/JobWriter

# Set execute permissions for all binaries
RUN chmod +x /app/UserService/main
RUN chmod +x /app/ApplicationService/main
RUN chmod +x /app/MatchingService/main
RUN chmod +x /app/JobWriter/main

# Copy environment files into each service's directory
COPY --from=builder /app/cmd/UserService/app.env /app/UserService/app.env
COPY --from=builder /app/cmd/ApplicationService/app.env /app/ApplicationService/app.env
COPY --from=builder /app/cmd/MatchingService/app.env /app/MatchingService/app.env
COPY --from=builder /app/cmd/JobWriter/app.env /app/JobWriter/app.env

# Copy shared scripts and migrations
COPY start.sh /app/start.sh
COPY wait-for.sh /app/wait-for.sh
COPY ./pkg/db/migration /app/db/migration

RUN chmod +x /app/start.sh /app/wait-for.sh

# Expose the ports for the services
EXPOSE 8080 8081 8082 8083

ENTRYPOINT ["/app/start.sh"]
