FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o drift-detector ./cmd/drift-detector

FROM alpine:3.21

# Install ca-certificates for making HTTPS requests and tzdata for timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/drift-detector /app/drift-detector

# Copy default configuration if any
COPY config/config.yaml /app/config/config.yaml

RUN chmod +x /app/drift-detector

# Expose any necessary ports
# EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/drift-detector"]

# Set the default command
CMD ["--help"]
