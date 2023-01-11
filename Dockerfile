FROM golang:1.19-alpine as builder
WORKDIR /work
COPY go.mod go.sum main.go .
ADD linkmap linkmap
RUN go build -o /app

FROM alpine:latest as production
COPY --from=builder /app .
EXPOSE 3333
CMD ./app
