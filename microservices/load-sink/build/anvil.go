package main

//Dummy API service to act as a load testing target

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*Troubleshooting note: Raw testing by hand is done as follows:

curl -X POST  http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"Name":"newOrder", "ID":"78912","Time":"223232113111","Data":"new order", "Eventname":"newOrder"}'

curl -s http://localhost:8080/orders| jq -r

*/

var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/load-sink.log" // /var/log
var output_dir string = os.Getenv("OUTPUT_DIR_PATH") + "/"              // e.g "/processed"

var port_specifier string = ":" + os.Getenv("PORT_NUMBER") // set the metrics endpoint port in the manifest
var orderCount int = 0                                     //total request count for current runtime
var errorCount int = 0                                     //total error count for current runtime

//Instrumentation
var (
	requestsSuccessful = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadsink_successul_requests_total",
		Help: "The total number of processed requests",
	})

	requestsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadsink_failed_requests_total",
		Help: "The total number of failed requests",
	})

	writesSuccessful = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadsink_successul_filewrites_total",
		Help: "The total number files successfully written",
	})

	writesFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadsink_failed_filewrites_total",
		Help: "The total number of file writes failed",
	})
)

func recordWriteSuccessMetrics() {
	go func() {
		writesSuccessful.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordWriteFailureMetrics() {
	go func() {
		writesFailed.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordSuccessMetrics() {
	go func() {
		requestsSuccessful.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordFailedMetrics() {
	go func() {
		requestsFailed.Inc()
		time.Sleep(2 * time.Second)

	}()
}

//Customised to the use case
type Order struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Time      string `json:"time"`
	Data      string `json:"data"`
	Eventname string `json:"eventname"`
}

type orderHandlers struct {
	sync.Mutex
	store map[string]Order
}

func logger(logFile string, logMessage string) {

	f, e := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if e != nil {
		log.Fatalf("error opening log file: %v", e)
	}

	defer f.Close()

	log.SetOutput(f)

	//include the hostname on each log entry
	log.Println(logMessage)

}

func (h *orderHandlers) orders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.publish(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *orderHandlers) get(w http.ResponseWriter, r *http.Request) {
	orders := make([]Order, len(h.store))

	fmt.Print("getting order")

	h.Lock()
	i := 0
	for _, order := range h.store {
		fmt.Println("debug> ", order)
		orders[i] = order
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *orderHandlers) getServiceHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Running health check ...")
	h.getRandomOrder(w, r)
}

func (h *orderHandlers) getRandomOrder(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Getting a random order")

	ids := make([]string, len(h.store))
	h.Lock()
	i := 0
	for id := range h.store {
		ids[i] = id
		i++
	}
	defer h.Unlock()

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))]
	}

	w.Header().Add("location", fmt.Sprintf("/orders/%s", target))
	w.WriteHeader(http.StatusFound)
}

func (h *orderHandlers) getOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if parts[2] == "random" {

		h.getRandomOrder(w, r)
		return
	}

	//Health service will return a random, but valid order
	if parts[2] == "health" {
		h.getServiceHealth(w, r)
		return
	}

	h.Lock()
	order, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func store_load_output_file(input_file string, msgPayload string) {

	logger(logFile, "prepared message payload: "+msgPayload)

	f, err := os.Create(input_file)

	if err != nil {

		logger(logFile, "error generating request message file: "+err.Error())

		//record as a failure metric
		recordWriteFailureMetrics()

	} else {
		//record as a success metric
		recordWriteSuccessMetrics()
	}
	_, err = f.WriteString(msgPayload)

	f.Close()
}

func check_empty_content(payload []byte) (empty bool) {
	empty = true
	if len(payload) > 0 {
		empty = false
	}
	return empty
}

func (h *orderHandlers) publish(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handling HTTP API (load) request:", orderCount+1)

	bodyBytes, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		errorCount++
		recordFailedMetrics()
		fmt.Println("ERROR in request: ", orderCount+1, " error reading body: ", err)

		return
	}

	//for debug
	empty := check_empty_content(bodyBytes)

	if empty {
		fmt.Println("[DEBUG] got empty request -> ", bodyBytes, " <- ")
	}

	ct := r.Header.Get("content-type")

	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("We don't speak '%s' around these parts ...", ct)))
		errorCount++
		fmt.Println("ERROR in request: ", orderCount+1, " error in content type: ", ct)

		return
	}

	orderCount++ //increment total incoming api requests

	//create an output file name to store this message
	input_file := output_dir + "/" + strconv.Itoa(orderCount)

	recordSuccessMetrics()

	msgPayload := string(bodyBytes)

	store_load_output_file(input_file, msgPayload)

	fmt.Println("processed request: ", orderCount, " data: ", msgPayload)

}

func idAllocator(taskMap map[int][]string, numWorkers int) (m map[string][]string) {

	tMap := make(map[string][]string)

	element := 0

	for i := 1; i <= numWorkers; i++ {

		taskID := strconv.Itoa(i)
		tMap[taskID] = taskMap[element]
		element++

	}

	return tMap
}

func main() {

	fmt.Println("Load sink service listening on ", port_specifier)

	//Administrative Web Interface
	admin := newAdminPortal()

	orderHandlers := newOrderHandlers()

	//POST URL for orders
	http.HandleFunc("/sink-orders", orderHandlers.orders)

	//GET URL for specific order information
	http.HandleFunc("/sink-order", orderHandlers.getOrder)

	//admin portal
	http.HandleFunc("/sink-admin", admin.handler)

	//metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(port_specifier, nil)

	if err != nil {
		panic(err)
	}

}
