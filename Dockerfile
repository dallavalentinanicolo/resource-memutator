# Use Go image top build
FROM golang:1.23 AS builder

# Set the target architecture to ARM64
ENV GOARCH=arm64
ENV GOOS=linux

WORKDIR /app
COPY . .

# Compile the Go binary for ARM64 architecture
RUN go mod tidy
RUN go build -o /webhook-server

FROM debian:bullseye-slim
WORKDIR /
COPY --from=builder /webhook-server /webhook-server
EXPOSE 443

CMD ["/webhook-server"]
