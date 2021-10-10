FROM golang:1.16-alpine

WORKDIR /

COPY *.go ./
COPY go.* ./

RUN go get
RUN go build -o /websocket-exporter

CMD [ "/websocket-exporter" ]
