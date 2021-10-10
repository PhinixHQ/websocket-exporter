FROM golang:1.16-alpine

ENV port 9143
EXPOSE $port

WORKDIR /

COPY *.go ./
COPY go.* ./

RUN go get
RUN go build -o /websocket-exporter

CMD [ "/websocket-exporter" ]
