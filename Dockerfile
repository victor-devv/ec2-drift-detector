FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /ec2-drift-detector ./cmd/drift-detector

ENTRYPOINT ["/ec2-drift-detector"]
