
FROM golang:1.20 AS builder


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


RUN go build -o main .


FROM scratch


COPY --from=builder /app/main /app/main


ENTRYPOINT ["/app/main"]
