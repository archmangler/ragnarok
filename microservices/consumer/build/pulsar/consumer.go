package main

/* Consumer Pool Management Service*/

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Get script configuration from the shell environment
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))       //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS")) //20
var port_specifier string = ":" + os.Getenv("PORT_NUMBER") // /var/log

var pulsarBrokerURL = os.Getenv("PULSAR_BROKER_SERVICE_ADDRESS")      // e.g "????"
var subscriptionName = os.Getenv("PULSAR_CONSUMER_SUBSCRIPTION_NAME") //e.g sub001

var topic0 string = os.Getenv("MESSAGE_TOPIC") // "messages" or  "ragnarok/requests/transactions"
var topic1 string = os.Getenv("ERROR_TOPIC")   // "api-failures"
var topic2 string = os.Getenv("METRICS_TOPIC") // "metrics"

var target_api_url string = os.Getenv("TARGET_API_URL") // e.g To use the dummy target api, provide: http://<some_ip_address>:<someport>/orders

var hostname string = os.Getenv("HOSTNAME")                            // "the pod hostname (in k8s) which ran this instance of go"
var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/consumer.log" // "/data/applogs/consumer.log"
//var consumer_group = os.Getenv("CONSUMER_GROUP")                       // we set the consumer group name to the podname / hostname
var consumer_group = os.Getenv("HOSTNAME") // we set the consumer group name to the podname / hostname

//produce a context for Kafka
//var ctx = context.Background()

//Global Error Counter during lifetime of this service run
var errorCount int = 0
var requestCount int = 0

/*Payload message format example
[
  {
    "Name": "newOrder",
    "ID": "8276",
    "Time": "8276",
    "Data": "new order",
    "Eventname": "newOrder"
  }
]
*/

type Payload struct {
	Name      string
	ID        string
	Time      string
	Data      string
	Eventname string
}

//Instrumentation
var (
	consumedRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_consumer_consumed_requests_total",
		Help: "The total number of requests taken from load queue",
	})

	requestsSuccessful = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_consumer_successul_requests_total",
		Help: "The total number of processed requests",
	})

	requestsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_consumer_failed_requests_total",
		Help: "The total number of failed requests",
	})

	goJobs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_consumer_concurrent_jobs",
		Help: "The total number of concurrent jobs per instance",
	})

	goWorkers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_consumer_concurrent_workers",
		Help: "The total number of concurrent workers per instance",
	})
)

func recordConsumedMetrics() {
	go func() {
		consumedRequests.Inc()
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

func recordConcurrentJobs() {
	go func() {
		goJobs.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordConcurrentWorkers() {
	go func() {
		goWorkers.Inc()
		time.Sleep(2 * time.Second)
	}()
}

//destination directory is used for now to simulate the remote API
//messages consumed from kafka are dumped into the output-api shared folder.
var output_directory string = os.Getenv("OUTPUT_DIRECTORY_PATH") + "/" // "/data/output-api"

func check_errors(e error, jobId int) {

	if e != nil {
		logMessage := "error " + e.Error() + "skipping over " + strconv.Itoa(jobId)
		logger(logFile, logMessage)
	}

}

//custom parsing of JSON struct
//Expected format as read from Pulsar topic:
//[{ "Name":"newOrder","ID":"14","Time":"1644469469070529888","Data":"loader-c7dc569f-8bkql","Eventname":"transactionRequest"}]
func parseJSONmessage(theString string) map[string]string {

	dMap := make(map[string]string)

	theString = strings.Trim(theString, "[")
	theString = strings.Trim(theString, "]")

	data := Payload{}

	json.Unmarshal([]byte(theString), &data)

	dMap["Name"] = string(data.Name)
	dMap["ID"] = string(data.ID)
	dMap["Time"] = string(data.Time)
	dMap["Data"] = string(data.Data)
	dMap["Eventname"] = string(data.Eventname)

	return dMap
}

//simple illustrative data check for message (this is optional, really)
func data_check(message string) (err error) {

	dMap := parseJSONmessage(message)

	for k := range dMap {
		if len(dMap[k]) > 0 {
			fmt.Println("checking payload message field (ok): ", k, " -> ", dMap[k])
		} else {
			return errors.New("empty field! ... " + k)
		}
	}

	return nil
}

func notify_job_start(workerId int, jobNum int) {

	logMessage := "worker" + strconv.Itoa(workerId) + "started  job" + strconv.Itoa(jobNum)
	logger(logFile, logMessage)

}

func notify_job_finish(workerId int, jobNum int) {

	logMessage := "worker" + strconv.Itoa(workerId) + "finished job" + strconv.Itoa(jobNum)
	logger(logFile, logMessage)
}

func logger(logFile string, logMessage string) {

	now := time.Now()
	msgTimestamp := now.UnixNano()

	logMessage = strconv.FormatInt(msgTimestamp, 10) + " [host=" + hostname + "]" + logMessage + " " + logFile

	fmt.Println(logMessage)
}

func consume_payload_data(client pulsar.Client, topic string, id int) {

	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages

	logMessage := "worker " + strconv.Itoa(id) + "consuming from topic " + topic
	logger(logFile, logMessage)

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscriptionName,
		Type:             pulsar.Shared,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer consumer.Close()

	for {

		// the `Receive` method blocks until we receive the next event
		msg, err := consumer.Receive(context.Background())

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))

		//message acknowledgment
		consumer.Ack(msg)

		message := string(msg.Payload())

		err = data_check(message)

		if err != nil {
			//incremement error metric
			errorCount += 1
			logMessage := "Error Count: " + strconv.Itoa(errorCount)
			logger(logFile, logMessage)
		}
		//But carry  on regardless anyway ...

		//TODO: POST to the URL of the target API service ("The System Under Load Test") - target_api_url
		var jsonStr = []byte(message)
		req, err := http.NewRequest("POST", target_api_url, bytes.NewBuffer(jsonStr))
		req.Header.Set("X-Custom-Header", "loadtest")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}

		//Post to the target URL
		resp, err := client.Do(req)

		if err != nil {
			b, _ := httputil.DumpResponse(resp, true)
			logMessage := "Failed to post payload to target API" + string(b)
			logger(logFile, logMessage)

			if err != nil {

				//publish metric to the metrics topic
				errorCount += 1

				//produce(message, ctx, topic2)

				logMessage := "could not write message to API" + err.Error()
				logger(logFile, logMessage)

			} else {

				//If all went well ...
				logMessage = "wrote message to API: " + message
				logger(logFile, logMessage)

			}

		}

	}

	/*
		if err := consumer.Unsubscribe(); err != nil {
			log.Fatal(err)
		}
	*/

}

func dumb_worker(id int, client pulsar.Client) {
	for {
		consume_payload_data(client, topic0, id)
	}
}

func main() {

	//Connect to Pulsar
	client, err := pulsar.NewClient(
		pulsar.ClientOptions{
			URL:               pulsarBrokerURL,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
		})

	if err != nil {
		log.Fatalf("Could not instantiate Pulsar client: %v", err)
	}

	defer client.Close()

	go func() {
		//metrics endpoint
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(port_specifier, nil)

		if err != nil {
			fmt.Println("Could not start the metrics endpoint: ", err)
		}
	}()

	//using a simple, single threaded loop - sequential consumption
	dumb_worker(1, client)

}
