package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
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
var namespace = "ragnarok"                                     //os.Getenv("POD_NAMESPACE")
var grafana_dashboard_url = os.Getenv("GRAFANA_DASHBOARD_URL") // e.g http://192.168.1.4:32000/d/AtqYwRA7k/transaction-matching-system-load-metrics?orgId=1&refresh=10s
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))           //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS"))     //20

//pulsar connection details
var brokerServiceAddress = os.Getenv("PULSAR_BROKER_SERVICE_ADDRESS") // e.g "pulsar://pulsar-mini-broker.pulsar.svc.cluster.local:6650"
var subscriptionName = os.Getenv("PULSAR_CONSUMER_SUBSCRIPTION_NAME") //e.g sub003
var primaryTopic string = os.Getenv("MESSAGE_TOPIC")                  // "messages"

var source_directory string = os.Getenv("DATA_SOURCE_DIRECTORY") + "/"    // "/datastore"
var processed_directory string = os.Getenv("DATA_OUT_DIRECTORY") + "/"    //"/processed"
var backup_directory string = os.Getenv("BACKUP_DIRECTORY") + "/"         // "/backups"
var logFile string = os.Getenv("LOCAL_LOGFILE_PATH") + "/" + "loader.log" // "/applogs"

var topic1 string = os.Getenv("DEADLETTER_TOPIC")                  // "deadLetter" - for failed message file generation
var topic2 string = os.Getenv("METRICS_TOPIC")                     // "metrics" - for metrics that should be streamed via a topic/queue
var hostname string = os.Getenv("HOSTNAME")                        // "the pod hostname (in k8s) which ran this instance of go"
var start_sequence = 1                                             // start of message range count
var end_sequence = 10000                                           // end of message range count
var port_specifier string = ":" + os.Getenv("METRICS_PORT_NUMBER") // port for metrics service to listen on
var taskCount int = 0

//function call count tracking counter
var streamAllCount int = 0

//Redis data storage details
var redisWriteConnectionAddress string = os.Getenv("REDIS_MASTER_ADDRESS") //address:port combination e.g  "my-release-redis-master.default.svc.cluster.local:6379"
var redisReadConnectionAddress string = os.Getenv("REDIS_REPLICA_ADDRESS") //address:port combination e.g  "my-release-redis-replicas.default.svc.cluster.local:6379"
var redisAuthPass string = os.Getenv("REDIS_PASS")

//sometimes we operate on global variables ...
var mutex = &sync.Mutex{}

//map of task allocation to concurrent workers
var taskMap = make(map[string][]string)

type Payload struct {
	Name      string
	ID        string
	Time      string
	Data      string
	Eventname string
}

//Management Portal Component
type adminPortal struct {
	password string
}

//Metrics Instrumentation
var (
	inputFilesGenerated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_files_generated_total",
		Help: "The total number of message files generated for input",
	})

	inputFileWriteErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_files_errors_total",
		Help: "The total number of input message files generation errors",
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

func recordSuccessMetrics() {
	go func() {
		inputFilesGenerated.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordFailureMetrics() {
	go func() {
		inputFileWriteErrors.Inc()
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

func newAdminPortal() *adminPortal {

	//initialise the management portal with a loose requirement for username:password

	password := os.Getenv("ADMIN_PASSWORD")

	if password == "" {
		panic("required env var ADMIN_PASSWORD not set")
	}

	return &adminPortal{password: password}
}

func logger(logFile string, logMessage string) {
	now := time.Now()
	msgTs := now.UnixNano()

	//when we're not logging to file  ...
	fmt.Println("[log]", logFile, strconv.FormatInt(msgTs, 10), logMessage)

	//When we're logging to file ...logFile
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

func clear_directory(Dir string) (int, error) {

	//delete all items in the directory
	fCounter := 0

	dir, err := ioutil.ReadDir(Dir)
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{Dir, d.Name()}...))
		fCounter++
	}

	if err != nil {
		fmt.Println("error trying to remove files in ", Dir, err)
	}

	return fCounter, err

}

//split the available tasks into equitable chunks
func divide_and_conquer(inputs []string, numWorkers int) (m map[string][]string) {

	tempMap := make(map[int][]string)

	size := len(inputs) / numWorkers

	logger(logFile, "dividing load into chunks of size: "+strconv.Itoa(size))

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

//assign task ranges to workers
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

func generate_input_sources(inputDir string, startSequence int, endSequence int) (inputs []string) {
	//Generate filenames and ensure the list is randomised
	var inputQueue []string

	logger(logFile, "generating new file names: "+strings.Join(inputQueue, ","))

	for f := startSequence; f <= endSequence; f++ {

		//IF USING filestorage:		inputQueue = append(inputQueue, inputDir+strconv.Itoa(f))
		inputQueue = append(inputQueue, strconv.Itoa(f))

	}

	//To ensure all worker pods, in a kubernetes scenario, don't operate on the same batch of files at any given time:
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(inputQueue), func(i, j int) { inputQueue[i], inputQueue[j] = inputQueue[j], inputQueue[i] })

	logger(logFile, "generated file names: "+strings.Join(inputQueue, ","))

	return inputQueue
}

func put_to_redis(input_file string, fIndex int, fileCount int, conn redis.Conn) (fc int) {

	// Send our command across the connection. The first parameter to
	// Do() is always the name of the Redis command (in this example
	// HMSET), optionally followed by any necessary arguments (in this
	// example the key, followed by the various hash fields and values).

	now := time.Now()
	msgTimestamp := now.UnixNano()

	//build the message body inputs for json
	_, err := conn.Do("HMSET", fIndex, "Name", "newOrder", "ID", strconv.Itoa(fIndex), "Time", strconv.FormatInt(msgTimestamp, 10), "Data", hostname, "Eventname", "transactionRequest")

	if err != nil {
		fmt.Println("failed to put file to redis: ", input_file, err.Error())
		//record as a failure metric
		recordFailureMetrics()
	} else {
		//record as a success metric
		recordSuccessMetrics()
		fileCount++
	}

	return fileCount

}

func create_load_file(input_file string, fIndex int, fileCount int) (fc int) {

	logger(logFile, "trying to create input file: "+input_file)

	//format the request message payload (milliseconds for now)
	now := time.Now()
	msgTimestamp := now.UnixNano()

	msgPayload := `[{ "Name": "newOrder","ID":"` + strconv.Itoa(fIndex) + `","Time":"` + strconv.FormatInt(msgTimestamp, 10) + `","Data":"` + hostname + `","Eventname":"transactionRequest"}]`

	logger(logFile, "prepared message payload: "+msgPayload)

	f, err := os.Create(input_file)

	if err != nil {
		logger(logFile, "error generating request message file: "+err.Error())

		//record as a failure metric
		recordFailureMetrics()
	} else {
		//record as a success metric
		recordSuccessMetrics()
		fileCount++
	}
	_, err = f.WriteString(msgPayload)
	f.Close()

	fmt.Println("[debug] data -> ", msgPayload, " Error: ", err)
	return fileCount
}

func process_input_data(tmpFileList []string) int {

	logger(logFile, "processing files: "+strings.Join(tmpFileList, ","))

	//where the actual task gets done
	//To Store in file:
	//fileCount = create_load_file(input_file, fIndex, fileCount)

	//Alternatively: to store in Redis
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.

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

	//Use defer to ensure the connection is always
	//properly closed before exiting the main() function.
	defer conn.Close()

	fileCount := 0
	for fIndex := range tmpFileList {

		input_file := tmpFileList[fIndex] + ".json"
		fileCount = put_to_redis(input_file, fIndex, fileCount, conn)

	}

	logger(logFile, "done creating input files.")

	return fileCount
}

func MoveFile(sourcePath, destPath string) error {
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

	return err
}

func moveAllfiles(a_directory string, b_directory string) int {

	someFiles, err := ioutil.ReadDir(a_directory)
	fcount := 0

	for _, file := range someFiles {

		input_file := a_directory + file.Name()
		destination_file := b_directory + file.Name()

		err = MoveFile(input_file, destination_file)

		if err != nil {

			logMessage := "ERROR: failed to move " + input_file + "to  " + destination_file + " error code: " + err.Error()
			logger(logFile, logMessage)

		} else {

			logMessage := "OK: " + input_file + " moved " + " to " + destination_file
			logger(logFile, logMessage)

			fcount++

		}
	}
	return fcount

}

func backupProcessedData(w http.ResponseWriter, r *http.Request) (int, error) {

	fcount := 0
	errCount := 0

	logger(logFile, "backing up already processed data from previous workload ...")

	files, _ := ioutil.ReadDir(processed_directory)
	w.Write([]byte("<html><h1>Backing up previous load data ...</h1><br>" + strconv.Itoa(len(files)) + " files </html>"))

	//get "/processed" directory filecount before
	theFiles, err := ioutil.ReadDir(processed_directory)

	if err != nil {
		logger(logFile, "can't get file list from: "+processed_directory)
		w.Write([]byte("<html><h1>can't get file list from: " + processed_directory + "</html>"))
	}

	for _, file := range theFiles {

		//Move data from /processed/ to /backup/
		input_file := processed_directory + file.Name()
		file_destination := backup_directory + file.Name()

		err = MoveFile(input_file, file_destination)

		if err != nil {

			errCount++

			logMessage := "ERROR: failed to move " + input_file + "to  " + file_destination + " error code: " + err.Error()
			logger(logFile, logMessage)

		} else {

			logMessage := "OK: " + input_file + " moved " + " to " + file_destination
			logger(logFile, logMessage)

			fcount++

		}
	}

	//return count of files backed up
	return fcount, err
}

func loadSyntheticData(w http.ResponseWriter, r *http.Request, start_sequence int, end_sequence int) (string, error) {

	logger(logFile, "Creating synthetic data for workload ...")

	status := "ok"
	var err error

	w.Write([]byte("<html><h1>Generating and Loading Synthetic Data</h1></html>"))

	//Generate input file list and distribute among workers
	inputQueue := generate_input_sources(source_directory, start_sequence, end_sequence) //Generate the total list of input files in the source dirs
	metadata := strings.Join(inputQueue, ",")
	logger(logFile, "[debug] generated new workload metadata: "+metadata)

	//Generate the files.
	fcnt := process_input_data(inputQueue)

	w.Write([]byte("<br><html>Generated new load test data: " + strconv.Itoa(fcnt) + " files." + "</html>"))

	return status, err
}

func loadHistoricalData(w http.ResponseWriter, r *http.Request) (string, error) {

	status := "ok"
	var err error
	fcnt := 0

	w.Write([]byte("<br><html>Using last backed up data in " + backup_directory + " for new load test: " + strconv.Itoa(fcnt) + " files." + "</html>"))
	logger(logFile, "Restoring load data from last backed up ...")

	//clear input dir
	fcnt, err = clear_directory(source_directory)
	logger(logFile, "cleared input directory: "+strconv.Itoa(fcnt)+" files.")

	//move files from /backups/ to /datastore/
	fcnt = moveAllfiles(backup_directory, source_directory)
	logger(logFile, " moved "+strconv.Itoa(fcnt)+" files from backup dir to input directory.")
	w.Write([]byte("<br><html>Moved last backed up data in " + backup_directory + " for new load test: " + strconv.Itoa(fcnt) + " files." + "</html>"))

	//report done
	return status, err

}

func (a adminPortal) selectionHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	/* //Can debug as follows:
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	*/

	selection := make(map[string]string)
	params := r.URL.RawQuery

	parts := strings.Split(params, "=")
	selection[parts[0]] = parts[1]

	if selection[parts[0]] == "backupProcessedData" {
		fmt.Println("running backups")
		status, err := backupProcessedData(w, r)

		if err != nil {
			w.Write([]byte("<html> Woops! ... " + err.Error() + "</html>"))
		} else {
			w.Write([]byte("<html> selection result: " + strconv.Itoa(status) + "</html>"))
		}
	}

	if selection[parts[0]] == "LoadSyntheticData" {

		//User can select the start sequence file name and end
		start_of_sequence := strings.Join(r.Form["start"], " ")
		end_of_sequence := strings.Join(r.Form["stop"], " ")

		w.Write([]byte("<html><h1>Creating input load data from topic sequence ... </h1></html>"))

		w.Write([]byte("<html style=\"font-family:verdana;\"><h1>Generate new input load data ... </h1></html>"))
		w.Write([]byte("<html style=\"font-family:verdana;\">start of sequence: " + start_of_sequence + "<br></html>"))
		w.Write([]byte("<html style=\"font-family:verdana;\">end of sequence: " + end_of_sequence + "<br></html>"))

		so_seq, _ := strconv.Atoi(start_of_sequence)
		eo_seq, _ := strconv.Atoi(end_of_sequence)

		status, err := loadSyntheticData(w, r, so_seq, eo_seq)

		if err != nil {
			w.Write([]byte("<html> Woops! ... " + err.Error() + "</html>"))
		} else {
			w.Write([]byte("<html> selection result: " + status + "</html>"))
		}
	}

	if selection[parts[0]] == "LoadHistoricalData" {
		status, err := loadHistoricalData(w, r)

		if err != nil {
			w.Write([]byte("<html> Woops! ... " + err.Error() + "</html>"))
		} else {
			w.Write([]byte("<html> selection result: " + status + "</html>"))
		}
	}

	if selection[parts[0]] == "RestartLoadTest" {
		status := "null"

		w.Write([]byte("<html><h1>Resetting and restarting load test </h1></html>"))

		//restart all services that need re-initialisation for a new load test
		status = restart_loading_services("producer", namespace, w, r)
		w.Write([]byte("<html> <br>Restarted producers - " + status + "</html>"))

		status = restart_loading_services("consumer", namespace, w, r)
		w.Write([]byte("<html> <br>Restarted consumers - " + status + "</html>"))

	}

	if selection[parts[0]] == "LoadRequestSequence" {

		start_of_sequence := strings.Join(r.Form["start"], " ")
		end_of_sequence := strings.Join(r.Form["stop"], " ")

		w.Write([]byte("<html><h1>Creating input load data from topic sequence ... </h1></html>"))

		w.Write([]byte("<html style=\"font-family:verdana;\"><h1 >Creating input load data from topic sequence ... </h1></html>"))
		w.Write([]byte("<html style=\"font-family:verdana;\">start of sequence: " + start_of_sequence + "<br></html>"))
		w.Write([]byte("<html style=\"font-family:verdana;\">end of sequence: " + end_of_sequence + "<br></html>"))

		SoS, err := strconv.ParseInt(start_of_sequence, 10, 64)

		if err != nil {
			fmt.Println("unable to convert string start_of_sequence to int64")
		}

		EoS, err := strconv.ParseInt(end_of_sequence, 10, 64)

		if err != nil {
			fmt.Println("unable to convert string end_of_sequence to int64")
		}

		//modify to use pulsar
		dataCount, errorCount := dump_pulsar_messages_to_input(primaryTopic, SoS, EoS)

		w.Write([]byte("<html> <br>Loaded " + strconv.Itoa(dataCount) + " requests in range from topic. With " + strconv.Itoa(errorCount) + " errors. </html>"))
		w.Write([]byte("<html> <br>Initiating run with new request load data ...</html>"))
		//user can now restart the producers and consumer

	}

	html_content := `
	<body>
	<br>
    <form action="/loader-admin" method="get">
           <input type="submit" name="back" value="back to main page">
	</form>
	</body>
	`
	w.Write([]byte(html_content))
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {

	//Very Crude web-based option selection menu defined directly in html, no templates, no styles ...
	//<Insert Authentication Code Here>

	html_content := `
	<html><h1 style="font-family:verdana;">Transaction Matching Load Test Workbench</h1><br></html>
	<body>
	<div style="padding:10px;">
	<h3 style="font-family:verdana;">Select Load Testing Operations:</h3>
	  <br>
	
	 <form action="/selected?param=backupProcessedData" method="post" style="font-family:verdana;">
			<input type="submit" name="backupProcessedData" value="backup processed data" style="padding:20px;">
			<br>
	 </form>  

	<form action="/selected?param=LoadSyntheticData" method="post">
           <input type="submit" name="LoadSyntheticData" value="load from synthetic data" style="padding:20px;">
		   
	<html style="font-family:verdana;">Start of sequence:</html><input type="text" name="start" >
	<html style="font-family:verdana;">End of Sequence:</html><input type="text" name="stop" >

	</form>

	<form action="/selected?param=LoadHistoricalData" method="post">
           <input type="submit" name="LoadHistoricalData" value="load from historical data" style="padding:20px;">
		   <br>
	</form>

	<form action="/selected?param=RestartLoadTest" method="post">
	<input type="submit" name="RestartLoadTest" value="restart / reset load test" style="padding:20px;">
	<br>
	</form>

	<form action="/selected?param=LoadRequestSequence" method="post"">
	<input type="submit" name="LoadRequestSequence" value="load request sequence" style="padding:20px;" style="font-family:verdana;"> 

	<html style="font-family:verdana;">Start of sequence:</html><input type="text" name="start" >
	<html style="font-family:verdana;">End of Sequence:</html><input type="text" name="stop" >

	</form>
</div>
	<div>
		<a href="` + grafana_dashboard_url + `">Load Testing Metrics Dashboard</a>
	</div>
	</body>
	`
	w.Write([]byte(html_content))
}

func restart_loading_services(service_name string, namespace string, w http.ResponseWriter, r *http.Request) string {

	arg1 := "kubectl"
	arg2 := "rollout"
	arg3 := "status"
	arg4 := "deployment"
	arg5 := service_name
	arg6 := "--namespace"
	arg7 := namespace
	status := "unknown"

	//restart
	arg3 = "restart"
	cmd := exec.Command(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	logger(logFile, "Running command: "+arg1+" "+arg2+" "+arg3+" "+arg4+" "+arg5+" "+arg6+" "+arg7)

	time.Sleep(5 * time.Second) //really should have a loop here waiting for returns ...

	out, err := cmd.Output()

	if err != nil {
		logger(logFile, "cannot restart component: "+service_name+" error. "+err.Error())
		return "failed"

	} else {
		logger(logFile, "restarted service - ok")
		status = "ok"
	}

	temp := strings.Split(string(out), "\n")
	theOutput := strings.Join(temp, `\n`)
	logger(logFile, "restart command result: "+theOutput)

	//for the user
	w.Write([]byte("<html> <br>service status: " + theOutput + "</html>"))

	//get status
	cmd = exec.Command(arg1, arg2, arg3, arg4, arg5, arg6, arg7)

	logger(logFile, "Running command: "+arg1+" "+arg2+" "+arg3+" "+arg4+" "+arg5+" "+arg6+" "+arg7)

	time.Sleep(5 * time.Second)

	out, err = cmd.Output()

	if err != nil {

		logger(logFile, "cannot get status for component: "+service_name+" error. "+err.Error())
		return "failed"

	} else {

		logger(logFile, "got service status - ok")
		status = "ok"

	}

	temp = strings.Split(string(out), "\n")
	theOutput = strings.Join(temp, `\n`)
	logger(logFile, "restart command result: "+theOutput)

	//for the user
	w.Write([]byte("<html> <br>service status: " + theOutput + "</html>"))

	logger(logFile, "done resetting system components for "+service_name)
	return status
}

//modify to use pulsar
func write_message_from_pulsar(msgCount int, errCount int, temp []string, line int) (int, int) {

	file_name := source_directory + strconv.Itoa(msgCount) + ".json"

	//write each line to an individual file
	f, err := os.Create(file_name)

	if err != nil {
		logger("write_message_from_pulsar", "error generating request message file: "+err.Error())
		//record as a failure metric
		errCount++

	} else {
		logger("write_message_from_pulsar", "wrote historical topic message to file: "+file_name)
		//record as a success metric
		msgCount++
	}

	logger("write_message_from_pulsar", "will write this topic message to file: "+temp[line])
	_, err = f.WriteString(temp[line])
	f.Close()
	return msgCount, errCount
}

func write_binary_message_to_redis(msgCount int, errCount int, msgIndex int64, d map[string]string, conn redis.Conn) (int, int) {

	logger("write_binary_message_to_redis", "[calling write_binary_message_to_redis]")

	Name := d["Name"]
	ID := d["ID"]
	Time := d["Time"]
	Data := d["Data"]
	Eventname := d["Eventname"]

	//REDIFY
	fmt.Println("inserting : ", Name, ID, Time, Data, Eventname)

	logger("write_binary_message_to_redis", "DEBUG> inserting to redis with Do - HMSET ...")
	_, err := conn.Do("HMSET", msgIndex, "Name", Name, "ID", ID, "Time", Time, "Data", Data, "Eventname", Eventname)

	if err != nil {
		logger("write_binary_message_to_redis", "error writing message to redis: "+err.Error())
		//record as a failure metric
		errCount++

	} else {
		logger("write_binary_message_to_redis", "wrote message to redis. count: "+strconv.Itoa(msgCount))
		//record as a success metric
		msgCount++
	}

	return msgCount, errCount
}

func write_message_to_redis(msgCount int, errCount int, temp []string, line int, conn redis.Conn) (int, int) {
	//write an array of messages to redis, each line being a single entry in the array

	s := temp[line]

	var t Payload

	s = strings.Replace(s, "[", "", -1)
	s = strings.Replace(s, "]", "", -1)

	logger(logFile, "will write to redis: "+s)

	b := []byte(s)

	err := json.Unmarshal(b, &t)

	if err == nil {
		fmt.Printf("%+v\n", t)
	} else {
		fmt.Println(err)
		fmt.Printf("%+v\n", t)
	}

	Name := t.Name
	ID := t.ID
	Time := t.Time
	Data := t.Data
	Eventname := t.Eventname

	//REDIFY
	_, err = conn.Do("HMSET", line, "Name", Name, "ID", ID, "Time", Time, "Data", Data, "Eventname", Eventname)

	if err != nil {
		logger(logFile, "error writing message to redis: "+err.Error())
		//record as a failure metric
		recordFailureMetrics()
		errCount++

	} else {
		logger(logFile, "wrote message to redis. count: "+strconv.Itoa(msgCount))
		//record as a success metric
		recordSuccessMetrics()
		msgCount++
	}

	return msgCount, errCount
}

func streamAll(reader pulsar.Reader, startMsgIndex int64, stopMsgIndex int64) (msgCount int, errCount int) {

	streamAllCount++

	logger(logFile, "DEBUG> (streamAll)[calling streamAll] Call Count: "+strconv.Itoa(streamAllCount))

	//dump the extracted pulsar messages to REDIS
	msgCount = 0
	errCount = 0
	readCount := 0

	conn, err := redis.Dial("tcp", redisWriteConnectionAddress)

	if err != nil {
		log.Fatal(err)
	}

	logger(logFile, "DEBUG> (streamAll) connecting to redis...")

	// Now authenticate to Redis
	response, err := conn.Do("AUTH", redisAuthPass)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("redis auth response: ", response)
	}

	//Use defer to ensure the connection is always
	//properly closed before exiting the main() function.
	defer conn.Close()

	logger(logFile, "DEBUG> (streamAll) looping through messages with reader.HasNext() ...")

	for reader.HasNext() {

		msg, err := reader.Next(context.Background())

		if err != nil {
			log.Fatal(err)
		}

		//can I access the details of the message?
		msgIndex := msg.ID().EntryID()

		logger(logFile, "DEBUG> (streamAll) Extracting message details ...")
		fmt.Printf("Read message id: -> %v <- ... => %#v\n", msgIndex, msg.ID())

		//get the metadata as string for parsing
		metadata := fmt.Sprintf("%#v", msg.ID())

		//get the actual message content
		content := string(msg.Payload())

		//Can i serialize into bytes? Yes
		//logger(logFile, "DEBUG> (streamAll) Serializing message data ...")
		//myBytes := msg.ID().Serialize()

		//Can I store it somewhere? Perhaps a map ?  REDIS or even on disk in a file ?
		//In other words: Can I write a byte[] slice to a file? Yes!

		logger(logFile, "DEBUG> (streamAll) got message index...")

		/*
			if msgIndex == startMsgIndex {

				fmt.Println("start read: ", msgIndex)
				logger("streamAll", "DEBUG> (streamAll) Begin reading message stream at start point ...")
				read = true
			}

			if msgIndex == stopMsgIndex+1 {
				fmt.Println("stop reading: ", msgIndex)
				logger("streamAll", "DEBUG> (streamAll) stop reading message stream at start point ...")
				read = false
			}
		*/

		//This is terrible. There has to be a better way to do this.Please replace once you've read:
		//https://pulsar.apache.org/docs/en/2.2.0/concepts-messaging/
		for i := startMsgIndex; i <= stopMsgIndex; i++ {

			if msgIndex == i {

				logger("streamAll", "DEBUG> (streamAll) got message content for processing: "+content)
				logger("streamAll", "DEBUG> (streamAll) got message metadata for processing: "+metadata)

				dataMap := parseJSONmessage(content)

				for f := range dataMap {
					fmt.Println("got struct field from map -> ", f, dataMap[f])
				}

				msgCount, errCount = write_binary_message_to_redis(msgCount, errCount, msgIndex, dataMap, conn)

				if errCount == 0 {
					readCount++
				}

				logger("streamAll", "number of topic messages processed: "+strconv.Itoa(msgCount)+" errors: "+strconv.Itoa(errCount))
				logger("streamAll", "processing count = "+strconv.Itoa(readCount))
			}
		}

	}

	return msgCount, errCount
}

//custom parsing of JSON struct
//Expected format as read from Pulsar topic: [{ "Name":"newOrder","ID":"14","Time":"1644469469070529888","Data":"loader-c7dc569f-8bkql","Eventname":"transactionRequest"}]
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

//modified to use pulsar
func dump_pulsar_messages_to_input(pulsarTopic string, msgStartSeq int64, msgStopSeq int64) (msgCount int, errCount int) {

	logger("dump_pulsar_messages_to_input", "DEBUG> preparing to stream messages from topic "+pulsarTopic+" in range")
	logger("dump_pulsar_messages_to_input", "DEBUG> preparing a pulsar client connection ...")
	client, err := pulsar.NewClient(
		pulsar.ClientOptions{
			URL:               brokerServiceAddress,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
		})

	if err != nil {
		log.Fatalf("Could not instantiate Pulsar client: %v", err)
	}

	logger("dump_pulsar_messages_to_input", "DEBUG> (dump_pulsar_messages_to_input) deferring pulsar client close ....var")
	defer client.Close()

	logger("dump_pulsar_messages_to_input", "DEBUG> (dump_pulsar_messages_to_input) creating a pulsar reader (CreateReader) ...")
	reader, err := client.CreateReader(pulsar.ReaderOptions{
		Topic:          pulsarTopic,
		StartMessageID: pulsar.EarliestMessageID(),
	})

	logger("dump_pulsar_messages_to_input", "DEBUG> (dump_pulsar_messages_to_input) deferring reader close")
	defer reader.Close()

	logger("dump_pulsar_messages_to_input", "DEBUG> (dump_pulsar_messages_to_input) checking reader error (CreateReader) ...")

	if err != nil {
		fmt.Println("got error creating reader: ", err) //Alternatively: log.Fatal(err)
	}

	logger("dump_pulsar_messages_to_input", "DEBUG> (dump_pulsar_messages_to_input) preparing to stream out all messages ...")

	msgCount, errCount = streamAll(reader, msgStartSeq, msgStopSeq)

	return msgCount, errCount
}

func main() {

	//set up the metrics and management endpoint
	//Prometheus metrics UI
	http.Handle("/metrics", promhttp.Handler())

	//Management UI for Load Data Management
	//Administrative Web Interface
	admin := newAdminPortal()
	http.HandleFunc("/loader-admin", admin.handler)
	http.HandleFunc("/selected", admin.selectionHandler)

	//serve static content
	staticHandler := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", staticHandler))

	err := http.ListenAndServe(port_specifier, nil)

	if err != nil {
		fmt.Println("Could not start the metrics endpoint: ", err)
	}

}
