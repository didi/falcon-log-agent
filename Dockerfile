FROM golang:latest as pre-build

WORKDIR /go/src/github.com/didi/falcon-log-agent
COPY . .
RUN go build -o falcon-log-agent


FROM alpine:latest

WORKDIR /app

ENV TZ=Asia/Shanghai

RUN echo "http://mirrors.aliyun.com/alpine/v3.4/main/" > /etc/apk/repositories && \
    apk update && apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=pre-build /go/src/github.com/didi/falcon-log-agent/falcon-log-agent .
COPY --from=pre-build /go/src/github.com/didi/falcon-log-agent/cfg/dev.cfg .
COPY --from=pre-build /go/src/github.com/didi/falcon-log-agent/cfg/strategy.dev.json .

EXPOSE 8003

CMD ['/app/falcon-log-agent', "-c", "/etc/conf/cfg.json", "-s", "/etc/conf/strategy.json"]
