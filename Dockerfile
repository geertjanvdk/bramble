ARG GOLANG=1.20
ARG ALPINE=3.17

FROM golang:${GOLANG}-alpine${ALPINE} AS buildGO
ARG VERSION=SNAPSHOT
WORKDIR /dist
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -ldflags="-s" -o bramble ./cmd/bramble
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'github.com/kelvin-green/bramble.Version=$VERSION'" -o bramble ./cmd/bramble

FROM alpine:${ALPINE} AS app

COPY --from=buildGO /dist/bramble /app/
RUN echo "{}" > /app/config.json
EXPOSE 8082
EXPOSE 8083
EXPOSE 8084

CMD [ "/app/bramble", "-config", "/app/config.json" ]