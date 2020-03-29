FROM golang:latest

WORKDIR /go/src/

COPY . .

ENV CGO_ENABLED=0 GO111MODULE="on"

RUN cd issue-man && go build . -o bin/issue-man

FROM alpine:latest

WORKDIR /app

COPY --form=0 /go/src/issue-man/bin/issue-man /app/

CMD ["./issue-man"]
