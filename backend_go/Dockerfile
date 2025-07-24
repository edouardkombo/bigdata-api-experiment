FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o seeder ./cmd/seeder
RUN go build -o ingest ./cmd/ingest
RUN go build -o api-gateway ./cmd/api-gateway

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/seeder /usr/local/bin/seeder
COPY --from=builder /app/ingest /usr/local/bin/ingest
COPY --from=builder /app/api-gateway /usr/local/bin/api-gateway
CMD ["api-gateway"]
