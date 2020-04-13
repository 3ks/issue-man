FROM golang:latest

WORKDIR /go/src/issue-man

COPY . .

ENV CGO_ENABLED=0 GO111MODULE="on"

RUN go build  -o bin/issue-man

FROM alpine:latest

WORKDIR /app

COPY --from=0 /go/src/issue-man/bin/issue-man /app/

RUN  apk add tzdata && \
     cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
     echo "Asia/Shanghai" > /etc/timezone && \
     apk del tzdata

CMD ["./issue-man"]
