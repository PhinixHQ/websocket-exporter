package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	websocket_successful = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_successful",
			Help: "( 0 = false , 1 = true )",
		})

	websocket_status_code = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_status_code",
			Help: "( 101 is normal status code )",
		})

	websocket_response_time = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_response_time",
			Help: "( response time in second )",
		})
	transport string
	resp_code float64
	err_read  error
)

func probeHandler(w http.ResponseWriter, r *http.Request, TimeOutHandshake int) {

	websocket.DefaultDialer.HandshakeTimeout = time.Duration(TimeOutHandshake) * time.Second

	targets, tr_ok := r.URL.Query()["target"]
	transports, tp_ok := r.URL.Query()["transport"]

	if !tr_ok || len(targets[0]) < 1 {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	target := targets[0]

	if tp_ok == true {
		if len(transports[0]) > 1 {
			transport = transports[0]
			target = target + "&transport=" + transport
		}
	}

	ur, _ := url.Parse(target)

	u := url.URL{Scheme: ur.Scheme, Host: ur.Host, Path: ur.Path, RawQuery: ur.RawQuery}

	start := time.Now()

	c, resp, err_con := websocket.DefaultDialer.Dial(u.String(), nil)

	if err_con == nil {
		c.UnderlyingConn().SetDeadline(time.Now().Add(time.Duration(0) * time.Second))
	}

	if err_con == nil && resp != nil {
		resp_code = float64(resp.StatusCode)
	} else {
		resp_code = 0
	}

	if err_con != nil {
		websocket_successful.Set(0)
		websocket_status_code.Set(resp_code)

	} else {
		_, _, err_read := c.ReadMessage()

		if err_read != nil {
			websocket_successful.Set(0)
			websocket_status_code.Set(resp_code)
			c.Close()
		}
	}

	if err_con == nil && err_read == nil {
		websocket_successful.Set(1)
		websocket_status_code.Set(resp_code)
		c.Close()
	}

	elapsed := time.Since(start).Seconds()
	websocket_response_time.Set(elapsed)

	reg := prometheus.NewRegistry()
	reg.MustRegister(websocket_successful)
	reg.MustRegister(websocket_status_code)
	reg.MustRegister(websocket_response_time)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {

	Port := flag.Int("port", 9143, "Port Number to listen")
	TimeOutHandshake := flag.Int("timeout", 5, "HandshakeTimeout specifies the duration for the handshake to complete")

	flag.Parse()
	var port = ":" + strconv.Itoa(*Port)

	fmt.Println("exporter working on port", port)

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, *TimeOutHandshake)
	})

	http.ListenAndServe(port, nil)

}
