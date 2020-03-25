FROM golang:latest

WORKDIR /go/src/

COPY . .

RUN CGO_ENABLED=0 go build . -o bin/issue-man

FROM alpine:latest

WORKDIR /app

COPY --form=0 /go/src/issue-man/bin/issue-man /app

CMD ["./issue-man"]
