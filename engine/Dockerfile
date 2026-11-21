FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mimic ./cmd/mimic

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/mimic /app/mimic

EXPOSE 50051 8080
ENTRYPOINT ["/app/mimic"]

