# Build stage
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

COPY go.* ./

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o verse-extractor .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S -g appgroup appuser

# Create app directory with correct ownership
WORKDIR /app
COPY --from=builder /app/verse-extractor .
RUN chown -R appuser:appgroup /app

ENV PORT=8081

EXPOSE 8081

# Switch to non-root user
USER appuser

# Command to run
CMD ["./verse-extractor"]