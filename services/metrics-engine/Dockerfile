FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY *.go .
RUN go build -o metrics-engine .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/metrics-engine .
EXPOSE 5002
CMD ["./metrics-engine"]
