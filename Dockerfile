FROM golang:1.16-alpine as builder

ENV port 9143
EXPOSE $port

WORKDIR /

COPY *.go ./
COPY go.* ./

RUN go mod download

RUN go build -o /websocket-exporter

CMD [ "/websocket-exporter" ]


FROM alpine:3.14 as production

ENV port 9143

COPY --from=builder websocket-exporter .

EXPOSE $port

CMD ./websocket-exporter
