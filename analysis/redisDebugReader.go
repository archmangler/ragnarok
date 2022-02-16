package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var logFile string = "script"

//get running parameters from container environment
var numJobs, _ = strconv.Atoi(os.Getenv("NUM_JOBS"))       //20
var numWorkers, _ = strconv.Atoi(os.Getenv("NUM_WORKERS")) //20

var hostname string = os.Getenv("HOSTNAME") // "the pod hostname (in k8s) which ran this instance of go"

//Redis configuration for better storage performance ...
var redisWriteConnectionAddress string = os.Getenv("REDIS_MASTER_ADDRESS") //address:port combination e.g  "my-release-redis-master.default.svc.cluster.local:6379"
var redisReadConnectionAddress string = os.Getenv("REDIS_REPLICA_ADDRESS") //address:port combination e.g  "my-release-redis-replicas.default.svc.cluster.local:6379"
var redisAuthPass string = os.Getenv("REDIS_PASS")

var taskCount int = 0

var mutex = &sync.Mutex{}

var taskMap = make(map[string][]string) //map of files to process
var purgeMap = make(map[string]string)  //map of files to 'purge' after processing

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

	purgeCntr := 0

	logger(logFile, "purging processed files ... "+strconv.Itoa(purgeCntr))

	for input_id, _ := range purgeMap {

		fmt.Println("dummy purging redis item ", input_id)

		/*result, err := conn.Do("DEL", input_id)

		if err != nil {
			fmt.Printf("Failed removing original input data from redis: %s\n", err)
		} else {
			fmt.Printf("deleted input %s -> %s\n", input_id, result)
		}*/

		purgeCntr++
	}

	logger(logFile, "purged processed files: "+strconv.Itoa(purgeCntr))
}

func readFromRedis(input_id string, conn redis.Conn) (ds string, err error) {

	msgPayload := ""
	err = nil

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

	//We should marshall this json into a well defined struct but lets
	//take the shortcut for now ...
	msgPayload = `[{ "Name":"` + msgName + `","ID":"` + input_id + `","Time":"` + msgTimestamp + `","Data":"` + msgData + `","Eventname":"` + msgEventname + `"}]`

	//get all the required data for the input id and return as json string
	return msgPayload, err

}

func process_input_data_redis_concurrent(workerId int, jobId int) {

	//error counter
	errCount := 0
	var tmpFileList []string

	//Get the unique key for the set of input tasks for this worker-job combination
	taskID := strconv.Itoa(workerId) // can be keyed with: jobId + "-" + strconv.Itoa(jobNum)
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

	for fIndex := range tmpFileList {
		input_id := tmpFileList[fIndex]

		payload, err := readFromRedis(input_id, conn) //readFromRedis(input_id string, c conn.Redis) (ds string, err error)

		if err != nil {

			errCount++

			logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to read data for " + input_id + " error code: " + err.Error()

			logger(logFile, logMessage)

		} else {

			logMessage := "OK: " + strconv.Itoa(workerId) + " read payload data for " + input_id
			logger(logFile, logMessage)

		}
		fmt.Println("Read payload from Redsi: ", payload)

		//keep a record of files that should be moved to /processed after the workers stop
		mutex.Lock()
		purgeMap[input_id] = input_id
		mutex.Unlock()

		fmt.Println("completed job: ", jobId)

	}

	logger(logFile, "completing task"+strconv.Itoa(taskCount))

	//it's a global variable being updated concurrently, so mutex lock ...
	mutex.Lock()
	taskCount++
	mutex.Unlock()

	logger(logFile, "completed task"+strconv.Itoa(taskCount))

	//please fix this!!
	if taskCount == numWorkers-1 || taskCount == numWorkers {
		//delete (or move) all processed files from Redis to somewhere else
		purgeProcessedRedis(conn)
	}

}

//pull unprocessed input data from Redis, divide data processing up
//among workers and process into kafka
func process_input_data_redis(workerId int) {

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

	for fIndex := range tmpFileList {
		input_id := tmpFileList[fIndex]

		payload, err := readFromRedis(input_id, conn)

		if err != nil {

			errCount++

			logMessage := "FAILED: " + strconv.Itoa(workerId) + " failed to read data for " + input_id + " error code: " + err.Error()

			logger(logFile, logMessage)

		} else {

			logMessage := "OK: " + strconv.Itoa(workerId) + " read payload data for " + input_id + ": " + payload
			logger(logFile, logMessage)

		}

		//post the load message to pulsar
		//produce(payload, producer, ctx, primaryTopic)

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

func read_input_sources_redis() (inputs []string) {
	var inputQueue []string

	//Get from Redis
	conn, err := redis.Dial("tcp", redisReadConnectionAddress)

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

	//GET ALL VALUES FROM DISCOVERED KEYS ONLY
	data, err := redis.Strings(conn.Do("KEYS", "*"))

	if err != nil {
		log.Fatal(err)
	}

	for _, msgID := range data {

		inputQueue = append(inputQueue, msgID)

	}

	//return the list of messages that exist in redis ...
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

	//inputQueue := read_input_sources(source_directory)            //Get the total list of input files in the source dirs
	inputQueue := read_input_sources_redis()                      //Alternatively, get the total list of input data files from redis instead
	taskMap = divide_and_conquer(inputQueue, numWorkers, numJobs) //get the allocation of workers to task-sets

	//Making use of Go Worker pools for concurrency within a pod here ...
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numWorkers; w++ {

		go worker(w, jobs, results)

	}

	for j := 1; j <= numJobs; j++ {

		jobs <- j

	}

	close(jobs)

	for r := 0; r <= numJobs*numWorkers; r++ {

		<-results

	}

}
