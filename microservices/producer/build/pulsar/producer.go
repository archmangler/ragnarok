package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//get running parameters from container environment
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))       //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS")) //20

var pulsarBrokerURL = os.Getenv("PULSAR_BROKER_SERVICE_ADDRESS")      // e.g "pulsar://pulsar-mini-broker.pulsar.svc.cluster.local:6650"
var subscriptionName = os.Getenv("PULSAR_CONSUMER_SUBSCRIPTION_NAME") //e.g sub002
var primaryTopic string = os.Getenv("MESSAGE_TOPIC")                  // "messages"

var source_directory string = os.Getenv("DATA_SOURCE_DIRECTORY") + "/"      // "/datastore/"
var processed_directory string = os.Getenv("DATA_OUT_DIRECTORY") + "/"      //"/processed/"
var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/" + "producer.log" // "/applogs"
var topic1 string = os.Getenv("DEADLETTER_TOPIC")                           // "deadLetter"
var topic2 string = os.Getenv("METRICS_TOPIC")                              // "metrics"
var hostname string = os.Getenv("HOSTNAME")                                 // "the pod hostname (in k8s) which ran this instance of go"

//Redis configuration for better storage performance ...
var dbIndex int = 11 // Separate namespace for management data. integer index > 0  e.g 11 (default 11)
var dataDbIndex int = 0

var redisWriteConnectionAddress string = os.Getenv("REDIS_MASTER_ADDRESS") //address:port combination e.g  "my-release-redis-master.default.svc.cluster.local:6379"
var redisReadConnectionAddress string = os.Getenv("REDIS_REPLICA_ADDRESS") //address:port combination e.g  "my-release-redis-replicas.default.svc.cluster.local:6379"
var redisAuthPass string = os.Getenv("REDIS_PASS")

var port_specifier string = ":" + os.Getenv("METRICS_PORT_NUMBER") // port for metrics service to listen on

var taskCount int = 0

var mutex = &sync.Mutex{}

var taskMap = make(map[string][]string) //map of files to process
var purgeMap = make(map[string]string)  //map of files to 'purge' after processing

//for job completion checking
var readFileList = make(map[int][]string)

//Instrumentation
var (
	inputRequestsLoaded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_input_requests_total",
		Help: "The total number of requests loaded from input source",
	})

	inputRequestsSubmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_submit_requests_total",
		Help: "The total number of loaded requests submitted to the load queue",
	})

	resultsRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_results_read_total",
		Help: "The total number of concurrent worker results read",
	})

	inputRequestsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_failed_requests_total",
		Help: "The total number of requests failed submission to load queue",
	})

	goJobs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_concurrent_jobs",
		Help: "The total number of concurrent jobs per instance",
	})

	goWorkers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_producer_concurrent_workers",
		Help: "The total number of concurrent workers per instance",
	})
)

func recordLoadedMetrics() {
	go func() {
		inputRequestsLoaded.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordSubmittedMetrics() {
	go func() {
		inputRequestsSubmitted.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordFailedMetrics() {
	go func() {
		inputRequestsFailed.Inc()
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

func recordConcurrentResults() {
	go func() {
		resultsRead.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func logger(logFile string, logMessage string) {

	now := time.Now()
	msgTimestamp := now.UnixNano()

	logMessage = strconv.FormatInt(msgTimestamp, 10) + " [host=" + hostname + "]" + logMessage + " " + logFile

	fmt.Println(logMessage)

}

func check_errors(e error, jobId int) {

	if e != nil {

		logMessage := "job error: " + strconv.Itoa(jobId) + e.Error()
		logger(logFile, logMessage)
	}

}

func purgeProcessedRedis(conn redis.Conn) {
	//purge this workers processed data from the input redis db (db 0)

	purgeCntr := 0
	var purged []string

	logger(logFile, "purging processed files ... "+strconv.Itoa(purgeCntr))

	//select correct DB
	conn.Do("SELECT", dataDbIndex)

	for input_id, _ := range purgeMap {

		result, err := conn.Do("DEL", input_id)

		if err != nil {
			fmt.Printf("Failed removing original input data from redis: %s\n", err)
		} else {
			fmt.Printf("deleted input %s -> %s\n", input_id, result)
			purged = append(purged, input_id)
		}

		purgeCntr++
	}

	logger(logFile, "purged files: "+strings.Join(purged, ",")+" total "+strconv.Itoa(purgeCntr))
}

func purgeProcessedMetadataRedis(conn redis.Conn) {
	//purge this workers entry from the work allocation metadata table (db 11)

	logger(logFile, "purging work allocation entry for "+hostname)

	//switch DB to metadata DB
	response, err := conn.Do("SELECT", dbIndex)

	if err != nil {
		fmt.Println("can't connect to redis metadata db: ", dbIndex)
		panic(err)
	} else {
		fmt.Println("redis select db response: ", response, " db index = ", dbIndex)
	}

	result, err := conn.Do("DEL", hostname)

	if err != nil {
		fmt.Printf("Failed removing work allocation entry for %s from redis: %s\n", hostname, err.Error())
	} else {
		fmt.Printf("deleted work allocation entry for %s -> %s\n", hostname, result)
	}

	//ensure we switch back to the default db background ...
	conn.Do("SELECT", dataDbIndex)
}

func purgeProcessed() {

	purgeCntr := 0

	for k, _ := range purgeMap {

		//	_ = MoveFile(k, v) ... this is very slow

		err := os.Remove(k)

		if err != nil {
			fmt.Printf("Failed removing original file: %s", err)
		}

		purgeCntr++
	}

	logger(logFile, "purged processed files: "+strconv.Itoa(purgeCntr))
}

func MoveFile(sourcePath string, destPath string) error {

	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}

	_, err = io.Copy(outputFile, inputFile)

	inputFile.Close()
	outputFile.Close()

	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}

	time.Sleep(2 * time.Second)

	return nil

}

func readFromRedis(input_id string, conn redis.Conn) (ds string, err error) {

	msgPayload := ""
	err = nil

	//select correct DB
	fmt.Println("select redis db: ", 0)

	conn.Do("SELECT", 0)

	//build the message body inputs for json
	//_, err := conn.Do("HMSET", fIndex, "Name", "newOrder", "ID", strconv.Itoa(fIndex), "Time", strconv.FormatInt(msgTimestamp, 10), "Data", hostname, "Eventname", "transactionRequest")
	msgID, err := redis.String(conn.Do("HGET", input_id, "ID"))
	if err != nil {
		fmt.Println("oops, got this: ", err, " skipping ", input_id)
		return msgID, err
	}

	msgName, err := redis.String(conn.Do("HGET", input_id, "Name"))
	if err != nil {
		fmt.Println("oops, got this: ", err, " skipping ", input_id)
		return msgID, err
	}

	msgTimestamp, err := redis.String(conn.Do("HGET", input_id, "Time"))
	if err != nil {
		fmt.Println("oops, got this: ", err, " skipping ", input_id)
		return msgID, err
	}

	msgData, err := redis.String(conn.Do("HGET", input_id, "Data"))
	if err != nil {
		fmt.Println("oops, got this: ", err, " skipping ", input_id)
		return msgID, err
	}

	msgEventname, err := redis.String(conn.Do("HGET", input_id, "Eventname"))
	if err != nil {
		fmt.Println("oops, got this: ", err, " skipping ", input_id)
		return msgID, err
	}

	//pack the fields into json structure, this has performance impact
	//but allows us to insert a tracing id into the data (try putting it into the metadata instead)
	//We should marshall this json using a well defined struct but lets
	//take the shortcut for now ...
	msgPayload = `[{"Name":"` + msgName + `","ID":"` + input_id + `","Time":"` + msgTimestamp + `","Data":"` + msgData + `","producerName":"` + hostname + `","Eventname":"` + msgEventname + `"}]`

	fmt.Println("got msg from redis ->", msgPayload, "<-")

	//get all the required data for the input id and return as json string
	return msgPayload, err

}

func check_payload_data(payload string) (e error) {

	//check for empty message payload data (and other issues in input data) before we
	//push into the messaging queue

	if len(payload) == 0 {
		return errors.New("skip empty message to message queue: " + payload)
	} else {
		fmt.Println("check message content nonzero: ", len(payload))
	}

	return nil
}

func process_input_data_redis_concurrent(workerId int, jobId int) {

	//Pulsar client connection ...
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarBrokerURL,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: primaryTopic,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer producer.Close()

	ctx := context.Background()

	//error counter
	errCount := 0
	var tmpFileList []string

	//Get the unique key for the set of input tasks for this worker-job combination
	taskID := strconv.Itoa(workerId) // can be keyed with: jobId + "-" + strconv.Itoa(jobNum)

	tmpFileList = taskMap[taskID] //get the messages assigned to this go worker

	logger("process_input_data_redis_concurrent", "[#debug-1]worker: "+taskID+" starting on workload: "+strings.Join(tmpFileList, ","))

	//Open Redis connection here again ...
	conn, err := redis.Dial("tcp", redisWriteConnectionAddress)

	if err != nil {
		log.Fatal(err)
	}

	// Now authenticate
	response, err := conn.Do("AUTH", redisAuthPass)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("redis auth response: ", response)
	}

	defer conn.Close()

	//select correct DB
	conn.Do("SELECT", 0)

	for fIndex := range tmpFileList {

		input_id := tmpFileList[fIndex]

		payload, err := readFromRedis(input_id, conn) //readFromRedis(input_id string, c conn.Redis) (ds string, err error)

		if err != nil {

			errCount++

			logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to read data for " + input_id + " error code: " + err.Error()

			logger(logFile, logMessage)

		} else {

			//record this as a metric
			recordLoadedMetrics()

			//keep a list of completed input jobs per worker process
			readFileList[workerId] = append(readFileList[workerId], input_id)

		}

		//check for data issues before input to pulsar
		err = check_payload_data(payload)

		if err != nil {

			fmt.Println("error: got empty message, not posting to message queue: ", payload)

		} else {

			//post the load message to pulsar
			produce(payload, producer, ctx, primaryTopic)
		}

		//keep a record of files that should be moved to /processed after the workers stop
		mutex.Lock()
		purgeMap[input_id] = input_id
		mutex.Unlock()

		fmt.Println("completed job: ", jobId)

	}

	logger(logFile, "completing task: "+strconv.Itoa(taskCount))

	//it's a global variable being updated concurrently, so mutex lock ...
	mutex.Lock()
	taskCount++
	mutex.Unlock()

	logger(logFile, "completed task: "+strconv.Itoa(taskCount))

	//Summarise the work completed by this worker:
	for k, v := range readFileList {
		logger(logFile, "[#debug-1] worker process: "+strconv.Itoa(k)+" completed tasks: "+strings.Join(v, ","))
	}

	//please fix this!!
	if taskCount == numWorkers-1 || taskCount == numWorkers {
		//delete (or move) all processed files from Redis to somewhere else
		purgeProcessedRedis(conn)

		//then remove the work allocation entry in Redis
		purgeProcessedMetadataRedis(conn)

	}

}

//pull unprocessed input data from Redis, divide data processing up
//among workers and process into kafka
func process_input_data_redis(workerId int) {

	//Pulsar client connection ...
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarBrokerURL,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: primaryTopic,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer producer.Close()

	ctx := context.Background()

	//error counter
	errCount := 0
	var tmpFileList []string

	//Get the unique key for the set of input tasks for this worker-job combination
	taskID := strconv.Itoa(workerId) // + "-" + strconv.Itoa(jobNum)
	tmpFileList = taskMap[taskID]

	//Open Redis connection here again ...
	conn, err := redis.Dial("tcp", redisWriteConnectionAddress)

	if err != nil {
		log.Fatal(err)
	}

	// Now authenticate
	response, err := conn.Do("AUTH", redisAuthPass)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("redis auth response: ", response)
	}

	defer conn.Close()

	//select correct DB
	conn.Do("SELECT", 0)

	for fIndex := range tmpFileList {

		input_id := tmpFileList[fIndex]

		payload, err := readFromRedis(input_id, conn) //readFromRedis(input_id string, c conn.Redis) (ds string, err error)

		if err != nil {

			errCount++

			logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to read data for " + input_id + " error code: " + err.Error()

			logger(logFile, logMessage)

		} else {

			//record this as a metric
			recordLoadedMetrics()

			logMessage := "OK: " + strconv.Itoa(workerId) + " read payload data for " + input_id

			logger(logFile, logMessage)

		}

		//post the load message to pulsar
		produce(payload, producer, ctx, primaryTopic)

		//keep a record of files that should be moved to /processed after the workers stop
		mutex.Lock()
		purgeMap[input_id] = input_id
		mutex.Unlock()

	}

	logger(logFile, "completing task"+strconv.Itoa(taskCount))

	//it's a global variable being updated concurrently, so mutex lock ...
	mutex.Lock()
	taskCount++
	mutex.Unlock()

	logger(logFile, "completed task"+strconv.Itoa(taskCount))

	if taskCount == numWorkers-1 {
		//delete (or move) all processed files from Redis to somewhere else
		purgeProcessedRedis(conn)
	}

}

//pull an untouched file from the source directory and process as file data to kafka
func process_input_data(workerId int, jobNum int, producer pulsar.Producer, ctx context.Context) {

	//error counter
	errCount := 0

	var input_file string
	var tmpFileList []string

	//Get the unique key for the set of input tasks for this worker-job combination
	taskID := strconv.Itoa(workerId) // + "-" + strconv.Itoa(jobNum)

	tmpFileList = taskMap[taskID]

	for fIndex := range tmpFileList {

		input_file = source_directory + tmpFileList[fIndex]

		if _, err := os.Stat(input_file); err == nil {

			dat, err := ioutil.ReadFile(input_file)

			payload := string(dat)

			//post the load message to pulsar
			produce(payload, producer, ctx, primaryTopic)

			//move processed file to "processed" (output) directory
			file_destination := processed_directory + tmpFileList[fIndex]

			//We need to use MoveFile instead of RenameFile because
			//this does not work across devices in Docker
			//err = MoveFile(input_file, file_destination)

			//keep a record of files that should be moved to /processed after the workers stop
			mutex.Lock()
			purgeMap[input_file] = file_destination
			mutex.Unlock()

			if err != nil {

				errCount++

				logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to move " + input_file + "to  " + file_destination + " error code: " + err.Error()
				logger(logFile, logMessage)

			} else {

				//record this as a metric
				recordLoadedMetrics()

				logMessage := "OK: " + strconv.Itoa(workerId) + " moved " + input_file + " to  " + file_destination
				logger(logFile, logMessage)

			}

		} else {
			logMessage := "skipping file: " + input_file

			logger(logFile, logMessage)
		}

	}

	//it's a global variable being updated concurrently, so mutex lock ...
	mutex.Lock()
	taskCount++
	mutex.Unlock()

	logger(logFile, "completed task"+strconv.Itoa(taskCount))

	if taskCount == numWorkers-1 {

		//move all processed files out of  "/datastore"
		purgeProcessed()

	}

}

//Publish the message to Pulsar topics
func produce(message string, producer pulsar.Producer, ctx context.Context, topic string) (err error) {

	msglen := len(message)
	logMessage := "writing: " + message + " size:  " + strconv.Itoa(msglen) + " to topic " + topic
	logger(logFile, logMessage)

	//write the payload to Pulsar
	msgId, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(fmt.Sprintf(message))})

	if err != nil {

		//record this as a metric
		recordFailedMetrics()

		log.Fatal(err)
	} else {

		log.Println("Published message: ", msgId)

		//record this as a success metric
		recordSubmittedMetrics()

		fmt.Println("wrote payload to topic: ", topic)

	}

	return err
}

func notify_job_start(workerId int, jobNum int) {
	logMessage := "worker " + strconv.Itoa(workerId) + " started job " + strconv.Itoa(jobNum)
	logger(logFile, logMessage)
}

func notify_job_finish(workerId int, jobNum int) {
	logMessage := "worker " + strconv.Itoa(workerId) + " finished job " + strconv.Itoa(jobNum)
	logger(logFile, logMessage)
}

func dumb_worker(id int) {

	//optional pre-tasks
	//notify_job_start(id, j)

	process_input_data_redis(id)

	//optional post-tasks
	//notify_job_finish(id, j)

}

func worker(id int, jobs <-chan int, results chan<- int) {

	for j := range jobs {

		//optional pre-tasks
		//notify_job_start(id, j)

		process_input_data_redis_concurrent(id, j)

		//optional post-tasks
		//notify_job_finish(id, j)

		results <- numJobs //* numWorkers

	}

}

//#debug
func idAllocator(taskMap map[int][]string, numWorkers int) (m map[string][]string) {

	tMap := make(map[string][]string)
	element := 0

	//print out incoming taskMap
	fmt.Println("#debug1 (idAllocator): received incoming tasks map", taskMap)

	for i := 1; i <= numWorkers; i++ {
		taskID := strconv.Itoa(i)
		tMap[taskID] = taskMap[element]
		element++
		logMessage := "worker process: " + taskID + " assigned tasks: " + strings.Join(tMap[taskID], ",")
		logger(logFile, logMessage)
	}

	fmt.Println("#debug1 (idAllocator): created assigned task map", tMap)

	return tMap
}

//debug this
func divide_and_conquer(inputs []string, numWorkers int, numJobs int) (m map[string][]string) {

	tempMap := make(map[int][]string)

	//+1 to ensure the allocation remains under the worker count
	size := (len(inputs) / numWorkers) + 1

	var j int = 0
	var lc int = 0

	for i := 0; i < len(inputs); i += size {

		j += size

		if j > len(inputs) {
			j = len(inputs)
		}

		//just populate the map for now
		//later assign worker-job pairs
		tempMap[lc] = inputs[i:j]

		lc++

	}

	//debugging
	fmt.Println("#debug (divide_and_conquer): index: " + strconv.Itoa(j) + " size: " + strconv.Itoa(size) + " assignment iterations: " + strconv.Itoa(lc) + " number of workers: " + strconv.Itoa(numWorkers))

	tMap := idAllocator(tempMap, numWorkers)

	return tMap
}

func getWorkAllocation(inputId string, conn redis.Conn) (ds []string, err error) {

	err = nil

	//build the message body inputs for json
	fmt.Println("getting data for key: ", inputId)

	//select correct DB
	conn.Do("SELECT", dbIndex)

	msgPayload, err := redis.Strings(conn.Do("LRANGE", inputId, 0, -1))

	if err != nil {

		fmt.Println("OOPS, got this for workload: ", err, " skipping ", inputId)

		return msgPayload, err

	} else {

		fmt.Println("OK - got this work list: ", msgPayload)

	}

	fmt.Println("OK - returning this work list: ", msgPayload)

	return msgPayload, err
}

func main() {

	//Connect to redis
	conn, err := redis.Dial("tcp", redisWriteConnectionAddress)

	if err != nil {
		log.Fatal(err)
	}

	// Now authenticate
	response, err := conn.Do("AUTH", redisAuthPass)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("redis auth response: ", response)
	}

	//connect to a separately assigned db/namespace in redis
	//reserved for the work allocation table
	response, err = conn.Do("SELECT", dbIndex)

	if err != nil {
		fmt.Println("can't connect to redis namespace: ", dbIndex)
		panic(err)
	} else {
		fmt.Println("redis select work allocation db response: ", response, dbIndex)
	}

	//Use defer to ensure the connection is always
	//properly closed before exiting the main() function.
	defer conn.Close()

	workAllocation, err := getWorkAllocation(hostname, conn) //get the assigned work for this worker

	fmt.Println("work allocation returned: ", workAllocation)

	if err != nil {
		fmt.Println("WARNING: could not get work allocation for pod: ", hostname)
	} else {
		fmt.Println("OK - got work allocation: ", workAllocation)
	}

	//inputQueue := read_input_sources_redis(workAllocation) //Alternatively, get the total list of input data files from redis instead

	taskMap = divide_and_conquer(workAllocation, numWorkers, numJobs) //get the allocation of workers to task-sets

	//Making use of Go Worker pools for concurrency within a pod here ...
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	go func() {
		//metrics endpoint
		http.Handle("/metrics", promhttp.Handler())

		err := http.ListenAndServe(port_specifier, nil)

		if err != nil {
			fmt.Println("Could not start the metrics endpoint: ", err)
		}
	}()

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

	for r := 0; r <= numJobs*numWorkers; r++ {

		<-results

		//record as metric
		recordConcurrentResults()

	}

}
