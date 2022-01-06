package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
)

//get running parameters from container environment
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))                        //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS"))                  //20
var broker1Address = os.Getenv("KAFKA_BROKER1_ADDRESS")                     // "192.168.65.2:9092"
var broker2Address = os.Getenv("KAFKA_BROKER2_ADDRESS")                     // "192.168.65.2:9092"
var broker3Address = os.Getenv("KAFKA_BROKER3_ADDRESS")                     // "192.168.65.2:9092"
var broker4Address = os.Getenv("KAFKA_BROKER4_ADDRESS")                     // "192.168.65.2:9092"
var broker5Address = os.Getenv("KAFKA_BROKER5_ADDRESS")                     // "192.168.65.2:9092"
var source_directory string = os.Getenv("DATA_SOURCE_DIRECTORY") + "/"      // "/datastore/"
var processed_directory string = os.Getenv("DATA_OUT_DIRECTORY") + "/"      //"/processed/"
var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/" + "producer.log" // "/applogs"
var topic0 string = os.Getenv("MESSAGE_TOPIC")                              // "messages"
var topic1 string = os.Getenv("DEADLETTER_TOPIC")                           // "deadLetter"
var topic2 string = os.Getenv("METRICS_TOPIC")                              // "metrics"
var hostname string = os.Getenv("HOSTNAME")                                 // "the pod hostname (in k8s) which ran this instance of go"

var port_specifier string = ":" + os.Getenv("METRICS_PORT_NUMBER") // port for metrics service to listen on

var taskCount int = 0

//produce a context for Kafka
var ctx = context.Background()
var mutex = &sync.Mutex{}

var taskMap = make(map[string][]string) //map of files to process
var purgeMap = make(map[string]string)  //map of files to 'purge' after processing

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

	logMessage = "[host=" + hostname + "]" + logMessage + " " + logFile
	fmt.Println(logMessage)

	/*
		mutex.Lock()

		f, e := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if e != nil {
			log.Fatalf("error opening log file: %v", e)
		}

		defer f.Close()

		log.SetOutput(f)

		//include the hostname on each log entry
		logMessage = "[host=" + hostname + "]" + logMessage
		log.Println(logMessage)

		mutex.Unlock()
	*/

}

func check_errors(e error, jobId int) {

	if e != nil {

		logMessage := "job error: " + strconv.Itoa(jobId) + e.Error()
		logger(logFile, logMessage)
	}

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

//pull an untouched file from the source directory and
func process_input_data(workerId int, jobNum int) {

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

			//check_errors(err, jobNum)

			payload := string(dat)

			//post the load message to kafka
			produce(payload, ctx, topic0)

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

				//logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to move " + input_file + "to  " + file_destination + " error code: " + err.Error()
				//logger(logFile, logMessage)

			} else {

				//record this as a metric
				recordLoadedMetrics()

				//logMessage := "OK: " + strconv.Itoa(workerId) + " moved " + input_file + " to  " + file_destination
				//logger(logFile, logMessage)

			}

		} else {
			//logMessage := "skipping file: " + input_file
			//logger(logFile, logMessage)
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

//Publish the message to kafka DeadLetter or other topics
//call: newOrderHandlers
func produce(message string, ctx context.Context, topic string) (err error) {

	//msglen := len(message)
	//logMessage := "wrote: " + message + " size:  " + strconv.Itoa(msglen) + " to topic " + topic
	//logger(logFile, logMessage)

	i := 0

	// intialize the writer with the broker addresses, and the topic

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker1Address, broker2Address, broker3Address, broker4Address, broker5Address},
		Topic:   topic,
	})

	// each kafka message has a key and value. The key is used
	// to decide which partition (and consequently, which broker)
	// the message gets published on

	err = w.WriteMessages(ctx, kafka.Message{
		Key: []byte(strconv.Itoa(i)),

		// create an arbitrary message payload for the value
		Value: []byte(message),
	})

	if err != nil {
		//No need to panic (these are just test messages)
		//panic("could not write message " + err.Error() + "to topic" + topic)
		//logMessage := "could not write message " + err.Error() + "to topic" + topic
		//logger(logFile, logMessage)

		//record this as a metric
		recordFailedMetrics()

	} else {
		//record this as a metric
		recordSubmittedMetrics()
	}

	// sleep for a second
	time.Sleep(time.Second)

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

func worker(id int, jobs <-chan int, results chan<- int) {

	for j := range jobs {

		//notify_job_start(id, j)

		process_input_data(id, j)

		//notify_job_finish(id, j)

		results <- numJobs //* numWorkers

	}

}

func read_input_sources(inputDir string) (inputs []string) {

	var inputQueue []string

	files, _ := ioutil.ReadDir(inputDir)

	//check_errors(err, 0)

	for _, f := range files {
		inputQueue = append(inputQueue, f.Name())
	}

	//To ensure all worker pods, in a kubernetes scenario, don't operate on the same batch of files at any given time:
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(inputQueue), func(i, j int) { inputQueue[i], inputQueue[j] = inputQueue[j], inputQueue[i] })

	return inputQueue
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

func divide_and_conquer(inputs []string, numWorkers int, numJobs int) (m map[string][]string) {

	tempMap := make(map[int][]string)

	size := len(inputs) / numWorkers

	var j int
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

	tMap := idAllocator(tempMap, numWorkers)

	return tMap
}

func main() {

	inputQueue := read_input_sources(source_directory)            //Get the total list of input files in the source dirs
	taskMap = divide_and_conquer(inputQueue, numWorkers, numJobs) //get the allocation of workers to task-sets

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
