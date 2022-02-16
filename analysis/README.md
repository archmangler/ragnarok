# Scripts to analyse and process data from the load testing engine

* This directory contains a set of script for processing load data after the load test in order to check for errors in processing.

# Analysing processed data

* The containerised target application can be used to capture data processed through the load generation engine for error analysis
* Load testing data processed into the sample target application can be copied out of the application container and processed to understand errors as follows:

```

- Make a target directory:

```
mkdir -p  16022022007
```

- Copy down from the targt application container:

```
(base) welcome@Traianos-MacBook-Pro analysis % kubectl cp ragnarok/load-sink-6948f578bc-djghs:/processed/ 16022022007
tar: removing leading '/' from member names
```

- Run the processing script:

```
(base) welcome@Traianos-MacBook-Pro analysis % ./do.sh 16022022007| head
**************** BEGIN analysing contents ********************
total:  9909
size 127: count: 902
size 128: count: 8914
size 126: count: 93
empty files: 0

DUP: 16022022007/22916 => 10 
DUP: 16022022007/24361 => 12 
DUP: 16022022007/27661 => 14 
DUP: 16022022007/26666 => 6 
.
.
.
```

- Interpretation: From the above output we can tell the latest processed data captured by the dummy target application container 0 empty files , multiple messages were duplicated and there were messages of sizes ranging from 126 bytes to 127 ...
