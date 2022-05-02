FROM golang:1.17-alpine AS builder

WORKDIR /usr/src/shrek

# Pre-copy/cache go.mod for pre-downloading dependencies and only re-downloading
# them in subsequent builds if they change.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy all other project files.
COPY . .

# Build app.
RUN go build -v -o /usr/local/bin/shrek ./cmd/shrek

FROM alpine:latest AS final

WORKDIR /app

# Copy compiled binary into final image.
COPY --from=builder /usr/local/bin/shrek .

# Define entry point.
ENTRYPOINT ["/app/shrek", "-d", "/app/generated/"]
