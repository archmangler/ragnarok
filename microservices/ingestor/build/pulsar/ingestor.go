package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//get running parameters from container environment
var sourceDirectory string = os.Getenv("DATA_SOURCE_DIRECTORY")         // "/datastore/inputs"
var s3bucketAddress string = os.Getenv("S3_BUCKET_ADDRESS")             //"s3://tradedatasource/inputs"
var remoteSourceDir string = os.Getenv("REMOTE_SOURCE_DIRECTORY") + "/" //path under s3/efs/afs share: e.g "/inputs" (leading slash, no trailing slash)

//Redis data storage details
var redisAuthPass string = os.Getenv("REDIS_PASS")
var redisWriteConnectionAddress string = os.Getenv("REDIS_MASTER_ADDRESS") //address:port combination e.g  "my-release-redis-master.default.svc.cluster.local:6379"
var port_specifier string = ":" + os.Getenv("METRICS_PORT_NUMBER")         // port for metrics service to listen on
var loadStatus string = "pending"                                          //pending|working|done

//{"instrumentId":128,"symbol":"BTC/USDC[Infi]","userId":25097,"side":2,"ordType":2,"price":5173054,"price_scale":2,"quantity":61300,"quantity_scale":6,"nonce":1645495020701,"blockWaitAck":0,"clOrdId":""}
type Order struct {
	InstrumentId   int    `json:"instrumentId"`
	Symbol         string `json:"symbol"`
	UserId         int    `json:"userId"`
	Side           int    `json:"side"`
	OrdType        int    `json:"ordType"`
	Price          int    `json:"price"`
	Price_scale    int    `json:"price_scale"`
	Quantity       int    `json:"quantity"`
	Quantity_scale int    `json:"quantity_scale"`
	Nonce          int    `json:"nonce"`
	BlockWaitAck   int    `json:"blockWaitAck "`
	ClOrdId        string `json:"clOrdId"`
}

//Management Portal Component
type adminPortal struct {
	password string
}

//Metrics Instrumentation
var (
	inputOrdersLoadTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_orders_ingested_total",
		Help: "The total number of orders ingested to redis",
	})

	inputOrdersLoadErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_orders_ingesterrors_total",
		Help: "The total number of orders failed to ingest to redis",
	})
)

func recordSuccessMetrics() {
	go func() {
		inputOrdersLoadTotal.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func recordFailureMetrics() {
	go func() {
		inputOrdersLoadErrors.Inc()
		time.Sleep(2 * time.Second)
	}()
}

func get_s3source_files(s3bucketAddress string) {

	//This is a very nasty workaround with potentially negative performance implications
	//aws s3 cp s3://tradedatasource/inputs ./source --recursive

	fmt.Println("(get_s3source_files): copying down source files from remote")

	arg1 := "aws"
	arg2 := "s3"
	arg3 := "sync"          //"sync" instead of "cp" for max parallelism
	arg4 := s3bucketAddress //"s3://tradedatasource"
	arg5 := sourceDirectory //arg6 := "--recursive"

	out, err := exec.Command(arg1, arg2, arg3, arg4, arg5).Output()

	fmt.Println("got output length: ", len(out))

	if err != nil {
		fmt.Println("(get_s3source_files) failed: ", err)
	} else {
		fmt.Println("(get_s3source_files) ok: done copying files down from remote")
	}

}

func list_s3source_files(s3bucketAddress string) (s []byte) {

	//This is a very nasty workaround with potentially negative performance implications
	fmt.Println("(list_s3source_files) listing source files in remote storage ...")

	arg1 := "aws"
	arg2 := "s3"
	arg3 := "ls"
	arg4 := s3bucketAddress //"s3://tradedatasource"
	arg5 := "--recursive"
	arg6 := "--human-readable"
	arg7 := "--summarize"

	out, err := exec.Command(arg1, arg2, arg3, arg4, arg5, arg6, arg7).Output()

	if err != nil {
		fmt.Println("sign_api_request error!", err)
	} else {
		fmt.Println("Done getting list of input files ...")
	}

	return out
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

}

func backupProcessedData(w http.ResponseWriter, r *http.Request) (fcount int, err error) {
	//TBI
	fcount = 0

	//return count of files backed up
	return fcount, err
}

//get a list of all files in the directory for loading
func getInputFiles(sourceDir string) (fileList []string) {

	files, err := ioutil.ReadDir(sourceDir + "/" + remoteSourceDir)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		fileList = append(fileList, file.Name())
	}

	return fileList
}

func getFileContent(fPath string) (content []string, err error) {

	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func jsonToMap(theString string) map[string]string {

	dMap := make(map[string]string)
	data := Order{}

	json.Unmarshal([]byte(theString), &data)

	dMap["instrumentId"] = fmt.Sprint(data.InstrumentId)
	dMap["symbol"] = string(data.Symbol)
	dMap["userId"] = fmt.Sprint(data.UserId)
	dMap["side"] = fmt.Sprint(data.Side)
	dMap["ordType"] = fmt.Sprint(data.OrdType)
	dMap["price"] = fmt.Sprint(data.Price)
	dMap["price_scale"] = fmt.Sprint(data.Price_scale)
	dMap["quantity"] = fmt.Sprint(data.Quantity)
	dMap["quantity_scale"] = fmt.Sprint(data.Quantity_scale)
	dMap["nonce"] = fmt.Sprint(data.Nonce)
	dMap["blockWaitAck"] = fmt.Sprint(data.BlockWaitAck)
	dMap["clOrdId"] = string(data.ClOrdId)

	fmt.Printf("debug %s\n", dMap)

	return dMap
}

func loadHistoricalData(sourceDir string, w http.ResponseWriter, r *http.Request) (string, error) {

	status := "ok"
	loadStatus = "started"
	var err error
	fCnt := 0
	errCnt := 0

	w.Write([]byte("<br><html>Using data in " + sourceDir + " for new load test: " + strconv.Itoa(fCnt) + " files." + "</html>"))
	logger("loadHistoricalData", "Using data in "+sourceDir+" for new load test")

	//fileList := getInputFiles(sourceDir)

	//get available files from source bucket
	files := list_s3source_files(s3bucketAddress)
	temp := strings.Split(string(files), "\n")
	var source []string

	//extract the file names
	for i := range temp {
		fields := strings.Fields(temp[i])
		if strings.Contains(temp[i], "inputs") {
			source = append(source, fields[4])
		}
	}

	//copy the files down to local
	get_s3source_files(s3bucketAddress)

	//read what local files we got
	fileList := getInputFiles(sourceDirectory)

	fmt.Println("debug -> ", fileList)

	//Connect to redis store to dump ingested order data
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

	for f := range fileList {

		fName := sourceDir + remoteSourceDir + string(fileList[f]) //hope this works out ...
		fmt.Println("read local file: " + fName)

		content, err := getFileContent(fName)

		if err != nil {

			fmt.Println("couldn't get file contents for ", f)

		} else {

			fData := strings.Join(content, "")

			order := jsonToMap(fData)

			fmt.Println("debug (order data): ", order)

			//process to redis ...
			fCnt, errCnt = write_binary_message_to_redis(fCnt, errCnt, f, order, conn)

			loadStatus = "working"

		}
	}

	loadStatus = "done"

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

	if selection[parts[0]] == "LoadHistoricalData" {
		status, err := loadHistoricalData(sourceDirectory, w, r)

		if err != nil {
			w.Write([]byte("<html> Woops! ... " + err.Error() + "</html>"))
		} else {
			w.Write([]byte("<html> selection result: " + status + "</html>"))
		}
	}

	html_content := `
	<body>
	<br>
    <form action="/ingestor-admin" method="get">
		   <input type="submit" name="back" value="back to main page">
		   
	</form>
	</body>
	`
	w.Write([]byte(html_content))
}

func (a adminPortal) initialiseHandler(w http.ResponseWriter, r *http.Request) {

	//Basic API Auth Example
	//Disabled for Testing

	/*user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}
	*/

	//don't block on this (the user can poll the status url for updates ...)
	go func() {

		status, err := loadHistoricalData(sourceDirectory, w, r)

		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(status))
		}
	}()

}

func (a adminPortal) statusHandler(w http.ResponseWriter, r *http.Request) {

	//Basic API Auth Example
	//Disabled for Testing

	/*user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}
	*/

	w.Write([]byte(loadStatus))

}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {

	//Very Crude web-based option selection menu defined directly in html, no templates, no styles ...
	//<Insert Authentication Code Here>

	html_content := `
	<html><h1 style="font-family:verdana;">Trade Matching Load Data Ingestor</h1><br></html>
	<body>
	<div style="padding:10px;">
	<h3 style="font-family:verdana;">Select Load Ingestion Operations:</h3>
	  <br>

	  <form action="/ingestorselected?param=LoadHistoricalData" method="post">
	      <input type="submit" name="LoadHistoricalData" value="load from historical data" style="padding:20px;">
	      <br>
     </form>
	 
	 <form action="/ingestorselected?param=getLoadStatus" method="post" style="font-family:verdana;">
			<input type="submit" name="getLoadStatus" value="get download status" style="padding:20px;">
			<br>
	 </form>  

	 <form action="/ingestorselected?param=backupProcessedData" method="post" style="font-family:verdana;">
			<input type="submit" name="backupProcessedData" value="backup processed data" style="padding:20px;">
			<br>
	 </form>  



	</form>
</div>
	</body>
	`
	w.Write([]byte(html_content))
}

func write_binary_message_to_redis(msgCount int, errCount int, msgIndex int, d map[string]string, conn redis.Conn) (int, int) {

	logger("write_binary_message_to_redis", "#debug (write_binary_message_to_redis): "+strconv.Itoa(msgIndex))

	InstrumentId := d["instrumentId"]
	Symbol := d["symbol"]
	UserId := d["userId"]
	Side := d["side"]
	OrdType := d["ordType"]
	Price := d["price"]
	Price_scale := d["price_scale"]
	Quantity := d["quantity"]
	Quantity_scale := d["quantity_scale"]
	Nonce := d["nonce"]
	BlockWaitAck := d["blockWaitAck"]
	ClOrdId := d["clOrdId"]

	//select correct DB (0)
	conn.Do("SELECT", 3)

	//REDIFY
	fmt.Println("inserting : ", InstrumentId, Symbol, UserId, Side, OrdType, Price, Price_scale, Quantity, Quantity_scale, Nonce, BlockWaitAck, ClOrdId)
	logger("write_binary_message_to_redis", "#debug (write_binary_message_to_redis): inserting to redis with Do - HMSET ...")

	_, err := conn.Do("HMSET", msgIndex, "instrumentId", InstrumentId, "symbol", Symbol, "userId", UserId, "side", Side, "ordType", OrdType, "price", Price, "price_scale", Price_scale, "quantity", Quantity, "quantity_scale", Quantity_scale, "nonce", Nonce, "blockWaitAck", BlockWaitAck, "clOrdId", ClOrdId)

	if err != nil {

		logger("write_binary_message_to_redis", "#debug #debug (write_binary_message_to_redis): error writing message to redis: "+err.Error())
		//record as a failure metric
		recordFailureMetrics()
		errCount++

	} else {

		logger("write_binary_message_to_redis", "#debug (write_binary_message_to_redis): wrote message to redis. count: "+strconv.Itoa(msgCount))
		//record as a success metric

		recordSuccessMetrics()
		msgCount++
	}

	return msgCount, errCount
}

func main() {

	/*
		[x] curl endpoint to trigger data load
		[x] put load status to readable endpoint
	*/

	//set up the metrics and management endpoint
	//Prometheus metrics UI
	http.Handle("/metrics", promhttp.Handler())

	//Management UI for Load Data Management
	//Administrative Web Interface
	admin := newAdminPortal()
	http.HandleFunc("/ingestor-admin", admin.handler)
	http.HandleFunc("/ingestorselected", admin.selectionHandler)

	//read the load status data (statusHandler)
	http.HandleFunc("/ingest-status", admin.statusHandler)
	//initiate data download
	http.HandleFunc("/ingest-start", admin.initialiseHandler)

	//serve static content
	staticHandler := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", staticHandler))

	err := http.ListenAndServe(port_specifier, nil)

	if err != nil {
		fmt.Println("Could not start http service endpoint: ", err)
	}

}
