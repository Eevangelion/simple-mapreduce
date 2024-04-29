FROM golang:latest AS builder

ENV GOOS=linux \
    CGO_ENABLED=0 \
    GOARCH=amd64

WORKDIR /go/src/app

COPY . .

RUN go mod download && \
    go build -o /go/src/app/main .
RUN chmod +x /go/src/app/main

FROM alpine

COPY --from=builder /go/src/app/main /app

EXPOSE 8082

CMD ["./app"]