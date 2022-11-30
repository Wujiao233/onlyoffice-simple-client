FROM golang:1.19.3-alpine3.16 AS backend
WORKDIR /backend-build

RUN apk update
RUN apk --no-cache add gcc musl-dev

COPY . .
RUN go env -w GOPROXY=https://goproxy.cn
RUN go env -w GO111MODULE=on

RUN go build -o client ./cmd/main.go

# Make workspace with above generated files.
FROM alpine:3.16 AS monolithic
RUN apk add tzdata
ENV ZONEINFO /usr/share/timezone
WORKDIR /app

COPY --from=backend /backend-build/client /app/
COPY ./static/* /app/static/
COPY ./templates/* /app/templates/
RUN mkdir -p /app/configs && \
    touch /app/configs/config.json && \
    echo '{}' > /app/configs/config.json && \
    chmod +x /app/client
RUN mkdir -p /app/data/

ENTRYPOINT ["./client"]
