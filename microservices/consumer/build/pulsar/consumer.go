package main

/* Consumer Pool Management Service*/

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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
var consumer_group = os.Getenv("HOSTNAME")                             // we set the consumer group name to the podname / hostname

//API login details
var base_url string = os.Getenv("API_BASE_URL")                //"trading-api.dexp-qa.com"
var password string = os.Getenv("TRADING_API_PASSWORD")        //"lolEx@20222@@"
var username string = os.Getenv("TRADING_API_USERNAME")        //"ngocdf1_qa_indi_7uxp@mailinator.com"
var userID int = 2661                                          // really arbitrary placehodler
var clOrdId string = os.Getenv("TRADING_API_CLORID")           //"test-1-traiano45"
var blockWaitAck, _ = strconv.Atoi(os.Getenv("BLOCKWAIT_ACK")) //blockwaitack
var account int = 0                                            //updated with the value of requestID for each new login to the API

//Global Error Counter during lifetime of this service run
var errorCount int = 0
var requestCount int = 0

type Payload struct {
	InstrumentId   string `json:"instrumentId"`
	Symbol         string `json:"symbol"`
	UserId         string `json:"userId"`
	Side           string `json:"side"`
	OrdType        string `json:"ordType"`
	Price          string `json:"price"`
	Price_scale    string `json:"price_scale"`
	Quantity       string `json:"quantity"`
	Quantity_scale string `json:"quantity_scale"`
	Nonce          string `json:"nonce"`
	BlockWaitAck   string `json:"blockWaitAck "`
	ClOrdId        string `json:"clOrdId"`
}

//Logon to the API
type authCredential struct {
	Id            int    `json:"id"`
	RequestToken  string `json:"requestToken"`
	RequestSecret string `json:"requestSecret"`
}

//We need to eventually replace this with a Golang native method as callouts to python
//are inefficient
func sign_api_request(apiSecret string, requestBody string) (s string) {
	//This is a very nasty workaround with potentially negative performance implications
	out, err := exec.Command("/usr/bin/python3", "use.py", apiSecret, requestBody).Output()

	if err != nil {
		fmt.Println("sign_api_request error!", err)
	}

	s = strings.TrimSuffix(string(out), "\n")
	return s
}

//4. Build up the request body
func create_order(secret_key string, api_key string, base_url string, orderParameters map[string]string, orderIndex int) {

	//Request body for POSTing a Trade
	params, err := json.Marshal(orderParameters)

	if err != nil {
		fmt.Println("(create_order) failed to jsonify: ", "(", orderIndex, ")", params)
	}

	requestString := string(params)

	//debug
	fmt.Println("(create_order) request parameters -> ", "(", orderIndex, ")", requestString)
	sig := sign_api_request(secret_key, requestString)

	//debug
	fmt.Println("(create_order) request signature -> ", "(", orderIndex, ")", sig)

	trade_request_url := "https://" + base_url + "/api/order"

	//Set the client connection custom properties
	fmt.Println("(create_order) setting client connection properties.", "(", orderIndex, ")")
	client := http.Client{}

	//POST body
	fmt.Println("(create_order) creating new POST request: ")
	request, err := http.NewRequest("POST", trade_request_url, bytes.NewBuffer(params))

	//set some headers
	fmt.Println("(create_order) setting request headers ...")
	request.Header.Set("Content-type", "application/json")
	request.Header.Set("requestToken", api_key)
	request.Header.Set("signature", sig)

	if err != nil {
		fmt.Println("(create_order) error after header addition: ", "(", orderIndex, ")")
		log.Fatalln(err)
	}

	fmt.Println("(create_order) executing the POST to ", "(", orderIndex, ")", trade_request_url)
	resp, err := client.Do(request)

	if err != nil {
		fmt.Println("(create_order) error after executing POST: ", "(", orderIndex, ")")
		log.Fatalln(err)

		//record as a failure metric
		recordFailedMetrics()
	}

	defer resp.Body.Close()
	fmt.Println("(create_order) reading response body ...", "(", orderIndex, ")")
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("(create_order) error reading response body: ", "(", orderIndex, ")")
		log.Fatalln(err)
	}

	sb := string(body)

	fmt.Println("(create_order) got response output: ", "(", orderIndex, ")", sb)

	//record this as a success metric
	recordSuccessMetrics()

}

//Instrumentation and metrics
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

//destination directory is used for now to simulate the remote API
//messages consumed from kafka are dumped into the output-api shared folder.
var output_directory string = os.Getenv("OUTPUT_DIRECTORY_PATH") + "/" // "/data/output-api"

//skeleton for an error logging and handling function
func check_errors(e error, jobId int) {

	if e != nil {
		logMessage := "error " + e.Error() + "skipping over " + strconv.Itoa(jobId)
		logger(logFile, logMessage)
	}

}

//3. Modify JSON document with newuser ID and any other details that's needed to update old order data
//updateOrder(order, account, blockWaitAck, userId, clOrdId)
func updateOrder(order map[string]string, account int, blockWaitAck int, userId int, clOrdId string) (Order map[string]string) {

	//replace the userId with the currently active user
	//Order = order
	fmt.Println("(updateOrder): updating this map: ", order)

	//debug
	fmt.Println("(updateOrder): updating these fields: ", userId, clOrdId, blockWaitAck, account)

	order["clOrdId"] = clOrdId
	order["userId"] = strconv.Itoa(userId)
	order["blockWaitAck"] = strconv.Itoa(blockWaitAck)
	order["account"] = strconv.Itoa(account)

	//debug
	fmt.Println("(updateOrder) after updating map: ", order)

	return order
}

//custom parsing of JSON struct
//Expected format as read from Pulsar topic:
//[{ "Name":"newOrder","ID":"14","Time":"1644469469070529888","Data":"loader-c7dc569f-8bkql","Eventname":"transactionRequest"}]
func parseJSONmessage(theString string) map[string]string {

	var dMap map[string]string

	fmt.Println("(parseJSONmessage) before stripping:   BEGIN->", theString, "<-END")

	theString = strings.Trim(theString, "[")
	theString = strings.Trim(theString, "]")

	fmt.Println("(parseJSONmessage) after stripping:    BEGIN->", theString, "<-END")
	fmt.Println("(parseJSONmessage) before marshalling: BEGIN->", theString, "<-END")

	//data := Payload{}

	//marshalling issue
	json.Unmarshal([]byte(theString), &dMap)

	//fmt.Println("(parseJSONmessage) after marshalling: ", dMap)

	/*
		dMap["instrumentId"] = data.InstrumentId
		dMap["symbol"] = data.Symbol
		dMap["userId"] = data.UserId
		dMap["side"] = data.Side
		dMap["ordType"] = data.OrdType
		dMap["price"] = data.Price
		dMap["price_scale"] = data.Price_scale
		dMap["quantity"] = data.Quantity
		dMap["quantity_scale"] = data.Quantity_scale
		dMap["nonce"] = data.Nonce
		dMap["blockWaitAck"] = data.BlockWaitAck
		dMap["clOrdId"] = data.ClOrdId
	*/

	fmt.Println("(parseJSONmessage) after marshaling: ", dMap)

	return dMap
}

func empty_msg_check(message string) (err error) {

	if len(message) == 0 {

		return errors.New("empty message from message queue: " + message)

	}

	return nil
}

//simple illustrative data check for message (this is optional, really)
//Add all your pre-POST data checking here!
func data_check(message string) (err error) {

	dMap := parseJSONmessage(message)

	for k := range dMap {
		if k != "clOrdId" {
			if len(dMap[k]) > 0 {
				fmt.Println("(data_check) checking payload message field (ok): ", k, " -> ", dMap[k])
			} else {
				return errors.New("(data_check) empty field in message! ... " + k)
			}
		}
	}

	return nil
}

//Example of a pretask run before the main work function
func notify_job_start(workerId int, jobNum int) {

	logMessage := "worker" + strconv.Itoa(workerId) + "started  job" + strconv.Itoa(jobNum)
	logger(logFile, logMessage)

}

//example of a post task run after the main work function
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

//2. Expand JSON into POST body
func jsonToMap(theString string) map[string]string {

	dMap := make(map[string]string)
	data := Payload{}

	fmt.Println("(jsonToMap) before stripping:   BEGIN->", theString, "<-END")
	theString = strings.Trim(theString, "[")
	theString = strings.Trim(theString, "]")
	fmt.Println("(jsonToMap) after stripping:    BEGIN->", theString, "<-END")

	fmt.Println("(jsonToMap) (parseJSONmessage) before marshalling: ", theString)

	fmt.Println("(jsonToMap) (parseJSONmessage) before marshalling - convert to byte: ", []byte(theString))

	//The Problem is Here
	json.Unmarshal([]byte(theString), &data)

	fmt.Println("(jsonToMap) (parseJSONmessage) after marshalling: ", data)

	dMap["instrumentId"] = data.InstrumentId
	dMap["symbol"] = data.Symbol
	dMap["userId"] = data.UserId
	dMap["side"] = data.Side
	dMap["ordType"] = data.OrdType
	dMap["price"] = data.Price
	dMap["price_scale"] = data.Price_scale
	dMap["quantity"] = data.Quantity
	dMap["quantity_scale"] = data.Quantity_scale
	dMap["nonce"] = data.Nonce
	dMap["blockWaitAck"] = data.BlockWaitAck
	dMap["clOrdId"] = data.ClOrdId

	fmt.Printf("(jsonToMap) returning map %s\n", dMap)

	return dMap
}

func consume_payload_data(client pulsar.Client, topic string, id int, credentials map[string]string) {

	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages

	var orderIndex int = 0

	logMessage := "(consume_payload_data) worker " + strconv.Itoa(id) + " consuming from topic " + topic
	logger("(consume_payload_data)", logMessage)

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

		fmt.Printf("(consume_payload_data) Received message msgId: %#v -- content: '%s' (%s)\n",
			msg.ID(), string(msg.Payload()), strconv.Itoa(orderIndex))

		//message acknowledgment
		//Do we need this ? Under what conditions ?
		consumer.Ack(msg)

		message := string(msg.Payload())

		orderIndex++

		err = empty_msg_check(message)

		if err != nil {

			logMessage := "(consume_payload_data) ERROR: empty message (skipping): " + "(" + strconv.Itoa(orderIndex) + ") " + message
			logger("(consume_payload_data)", logMessage)

		} else {

			//record this as a metric
			recordConsumedMetrics()

			err = data_check(message)

			if err != nil {

				fmt.Println("(consume_payload_data) data check error: ", "(", orderIndex, ")", err)
				//incremement error metric
				errorCount += 1
				logMessage := "(consume_payload_data) Error Count: " + "(" + strconv.Itoa(orderIndex) + ")" + strconv.Itoa(errorCount)
				logger("(consume_payload_data)", logMessage)

			} else {

				//sign the body and create an order (order map[string]string )
				fmt.Println("(consume_payload_data) converting this json string to map: ", "(", orderIndex, ")", message)
				order := jsonToMap(message) //convert the json string to a map[string]string to access the order elements

				fmt.Println("(consume_payload_data) json converted map: ", "(", orderIndex, ")", order)

				order = updateOrder(order, account, blockWaitAck, userID, clOrdId)

				fmt.Println("(consume_payload_data) updated order details: ", "(", orderIndex, ")", order)

				create_order(credentials["secret_key"], credentials["api_key"], base_url, order, orderIndex)

			}
		}
	}

}

func dumb_worker(id int, client pulsar.Client, credentials map[string]string) {

	for {
		consume_payload_data(client, topic0, id, credentials)
	}

}

func apiLogon(username string, password string, userID int, base_url string) (credentials map[string]string) {

	params := `{ "login":"` + username + `",  "password":"` + password + `",  "userId":"` + strconv.Itoa(userID) + `"}`
	responseBytes := []byte(params)
	responseBody := bytes.NewBuffer(responseBytes)

	credentials = make(map[string]string)

	fmt.Println(responseBody)
	resp, err := http.Post("https://"+base_url+"/api/logon", "application/json", responseBody)

	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}

	defer resp.Body.Close()

	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	fmt.Println("#debug: getting request response body: ", sb) //marshall into an authcredential struct
	loginData := authCredential{}
	json.Unmarshal([]byte(sb), &loginData)

	//extract tokens
	fmt.Println("#debug: getting request ID: ", strconv.Itoa(loginData.Id))
	credentials["request_id"] = strconv.Itoa(loginData.Id)

	fmt.Println("#debug: getting api_key: ", loginData.RequestToken)
	credentials["api_key"] = loginData.RequestToken

	fmt.Println("#debug: getting secret_key: ", loginData.RequestSecret)
	credentials["secret_key"] = loginData.RequestSecret

	return credentials

}

func main() {

	//login to the Trading API (assuming a single user for now)
	credentials := apiLogon(username, password, userID, base_url)

	//update old order data with unique, current information
	request_id, _ := strconv.Atoi(credentials["request_id"])
	userID = request_id
	account = request_id

	//Connect to Pulsar
	client, err := pulsar.NewClient(
		pulsar.ClientOptions{
			URL:               pulsarBrokerURL,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
		})

	if err != nil {
		log.Fatalf("(main) Could not instantiate Pulsar client: %v", err)
	}

	defer client.Close()

	go func() {
		//metrics endpoint
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(port_specifier, nil)

		if err != nil {
			fmt.Println("(main) Could not start the metrics endpoint: ", err)
		} else {
			fmt.Println("(main) done setting up metrics endpoint: ")
		}
	}()

	//using a simple, single threaded loop - sequential consumption
	fmt.Println("(main) contemplating running main worker loop ...")
	dumb_worker(1, client, credentials)
	fmt.Println("(main) done running main worker loop ...")

}
