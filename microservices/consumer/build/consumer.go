package main

/* Consumer Pool Management Service*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
)

//Get script configuration from the shell environment
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))       //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS")) //20
var port_specifier string = ":" + os.Getenv("PORT_NUMBER") // /var/log

var brokerServiceAddress = os.Getenv("KAFKA_BROKER_SERVICE_ADDRESS") // e.g "kafka.kafka.svc.cluster.local"

var broker1Address = os.Getenv("KAFKA_BROKER1_ADDRESS")   // "192.168.65.2:9092"
var broker2Address = os.Getenv("KAFKA_BROKER2_ADDRESS")   // "192.168.65.2:9092"
var broker3Address = os.Getenv("KAFKA_BROKER3_ADDRESS")   // "192.168.65.2:9092"
var broker4Address = os.Getenv("KAFKA_BROKER4_ADDRESS")   // "192.168.65.2:9092"
var broker5Address = os.Getenv("KAFKA_BROKER5_ADDRESS")   // "192.168.65.2:9092"
var broker6Address = os.Getenv("KAFKA_BROKER6_ADDRESS")   // "192.168.65.2:9092"
var broker7Address = os.Getenv("KAFKA_BROKER7_ADDRESS")   // "192.168.65.2:9092"
var broker8Address = os.Getenv("KAFKA_BROKER8_ADDRESS")   // "192.168.65.2:9092"
var broker9Address = os.Getenv("KAFKA_BROKER9_ADDRESS")   // "192.168.65.2:9092"
var broker10Address = os.Getenv("KAFKA_BROKER10_ADDRESS") // "192.168.65.2:9092"
var broker11Address = os.Getenv("KAFKA_BROKER11_ADDRESS") // "192.168.65.2:9092"
var broker12Address = os.Getenv("KAFKA_BROKER12_ADDRESS") // "192.168.65.2:9092"
var broker13Address = os.Getenv("KAFKA_BROKER13_ADDRESS") // "192.168.65.2:9092"
var broker14Address = os.Getenv("KAFKA_BROKER14_ADDRESS") // "192.168.65.2:9092"
var broker15Address = os.Getenv("KAFKA_BROKER15_ADDRESS") // "192.168.65.2:9092"
var broker16Address = os.Getenv("KAFKA_BROKER16_ADDRESS") // "192.168.65.2:9092"
var broker17Address = os.Getenv("KAFKA_BROKER17_ADDRESS") // "192.168.65.2:9092"
var broker18Address = os.Getenv("KAFKA_BROKER18_ADDRESS") // "192.168.65.2:9092"
var broker19Address = os.Getenv("KAFKA_BROKER19_ADDRESS") // "192.168.65.2:9092"
var broker20Address = os.Getenv("KAFKA_BROKER20_ADDRESS") // "192.168.65.2:9092"
var broker21Address = os.Getenv("KAFKA_BROKER21_ADDRESS") // "192.168.65.2:9092"
var broker22Address = os.Getenv("KAFKA_BROKER22_ADDRESS") // "192.168.65.2:9092"
var broker23Address = os.Getenv("KAFKA_BROKER23_ADDRESS") // "192.168.65.2:9092"
var broker24Address = os.Getenv("KAFKA_BROKER24_ADDRESS") // "192.168.65.2:9092"
var broker25Address = os.Getenv("KAFKA_BROKER25_ADDRESS") // "192.168.65.2:9092"
var broker26Address = os.Getenv("KAFKA_BROKER26_ADDRESS") // "192.168.65.2:9092"
var broker27Address = os.Getenv("KAFKA_BROKER27_ADDRESS") // "192.168.65.2:9092"
var broker28Address = os.Getenv("KAFKA_BROKER28_ADDRESS") // "192.168.65.2:9092"

var offSetCommitInterval, _ = strconv.Atoi(os.Getenv("CONSUMER_COMMIT_INTERVAL"))

var topic0 string = os.Getenv("MESSAGE_TOPIC") // "messages"
var topic1 string = os.Getenv("ERROR_TOPIC")   // "api-failures"
var topic2 string = os.Getenv("METRICS_TOPIC") // "metrics"

var target_api_url string = os.Getenv("TARGET_API_URL") // e.g To use the dummy target api, provide: http://<some_ip_address>:<someport>/orders

var hostname string = os.Getenv("HOSTNAME")                            // "the pod hostname (in k8s) which ran this instance of go"
var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/consumer.log" // "/data/applogs/consumer.log"
//var consumer_group = os.Getenv("CONSUMER_GROUP")                       // we set the consumer group name to the podname / hostname
var consumer_group = os.Getenv("HOSTNAME") // we set the consumer group name to the podname / hostname

//produce a context for Kafka
var ctx = context.Background()

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
	Name      string `json:"name"`
	ID        string `json:"id"`
	Time      string `json:"time"`
	Data      string `json:"data"`
	Eventname string `json:"eventname"`
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

//simple illustrative data check for message (this is optional, really)
func data_check(message string) (err error) {

	data := Payload{}

	err = json.Unmarshal([]byte(message), &data)

	if err != nil {

		logMessage := "incorrect message format (not readable json)" + err.Error()
		logger(logFile, logMessage)

		//publish metric to the metrics topic
		errorCount += 1

		logMessage = "Error Count: " + strconv.Itoa(errorCount)
		logger(logFile, logMessage)

		//TODO: Format as JSON
		//produce(message, ctx, topic2)
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

	/*
		f, e := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if e != nil {
			log.Fatalf("error opening log file: %v", e)
		}

		defer f.Close()

		//include the hostname on each log entry
		logMessage = "[host=" + hostname + "]" + logMessage
		log.Println(logMessage)

		log.SetOutput(f)
		log.Println(logMessage)
	*/

}

func produce(message string, ctx context.Context, topic string) {

	i := 0

	// intialize the writer with the broker addresses, and the topic
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerServiceAddress},
		Topic:   topic,
	})

	// each kafka message has a key and value. The key is used
	// to decide which partition (and consequently, which broker)
	// the message gets published on
	err := w.WriteMessages(ctx, kafka.Message{
		Key: []byte(strconv.Itoa(i)),

		// create an arbitrary message payload for the value
		Value: []byte(message),
	})

	if err != nil {

		//No need to panic
		//panic("could not write message " + err.Error() + "to topic" + topic)
		logMessage := "could not write message " + err.Error() + "to topic" + topic
		logger(logFile, logMessage)

		errorCount++
	}

	// log a confirmation once the message is written
	logMessage := "wrote:" + message + " to topic " + topic
	logger(logFile, logMessage)

	// sleep for a second
	time.Sleep(time.Second)
}

func consume_payload_data(ctx context.Context, topic string, id int) (message string) {
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages

	logMessage := "worker " + strconv.Itoa(id) + "consuming from topic " + topic
	logger(logFile, logMessage)

	dialer := &kafka.Dialer{
		Timeout: 600 * time.Second,
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerServiceAddress},
		Topic:   topic,
		GroupID: consumer_group,
		Dialer:  dialer,
	})

	for {

		// the `ReadMessage` method blocks until we receive the next event
		msg, err := r.ReadMessage(ctx)

		if err != nil {
			panic("could not read message " + err.Error())
		} else {
			recordConsumedMetrics()

			logMessage := "worker " + strconv.Itoa(id) + "OK - consumed a message from topic " + topic
			logger(logFile, logMessage)

		}

		message = string(msg.Value)

		err = data_check(message)

		if err != nil {

			//produce to errored messages topic
			//produce(message, ctx, topic1)

			//publish metric to the metrics topic
			errorCount += 1

			logMessage := "Error Count: " + strconv.Itoa(errorCount)
			logger(logFile, logMessage)

			//produce(message, ctx, topic2)

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

			//record the metric
			recordFailedMetrics()
		} else {

			recordSuccessMetrics()

		}

		defer resp.Body.Close()

		//resp.Body.Close()

		//log the response
		logMessage := "Posted message data: " + string(jsonStr) + " to topic: "
		logger(logFile, logMessage)

		logMessage = "Response Status from target API: " + resp.Status
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

func dumb_worker(id int) {

	for {

		payload := consume_payload_data(ctx, topic0, id)
		logMessage := "-> consumed payload message from" + topic0 + " : " + payload
		logger(logFile, logMessage)
	}

}

func worker(id int, jobs <-chan int, results chan<- int) {

	for j := range jobs {

		payload := consume_payload_data(ctx, topic0, id)

		//notify_job_start(id, j)

		logMessage := "-> consumed payload message from" + topic0 + " : " + payload
		logger(logFile, logMessage)

		time.Sleep(time.Second)

		//notify_job_finish(id, j)

		results <- j

	}

}

func main() {

	go func() {
		//metrics endpoint
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(port_specifier, nil)

		if err != nil {
			fmt.Println("Could not start the metrics endpoint: ", err)
		}
	}()

	//using a simple loop
	dumb_worker(1)

	//Using Go Worker Pools ..

	/*jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)

		//record as metric
		recordConcurrentWorkers()
	}

	for j := 1; j <= numJobs; j++ {

		jobs <- j

		//record as metric
		recordConcurrentJobs()

	}
	close(jobs)

	for r := 0; r <= numJobs; r++ {

		<-results
	}
	*/

}
