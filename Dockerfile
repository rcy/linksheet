FROM golang:1.24-alpine AS builder
WORKDIR /work
COPY go.mod go.sum main.go .
ADD linkmap linkmap
ADD db db
RUN go build -o /app

FROM alpine:latest AS production
COPY --from=builder /app .
EXPOSE 3333
CMD ["./app"]

