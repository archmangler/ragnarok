package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
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

func main() {

	var missingIds []int

	lines, err := readLines("sequence.log")

	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	seq,_ := strconv.Atoi(lines[0])
	last,_ := strconv.Atoi(lines[0])

	missingCount := 0
	thing := "."

	for i := range lines {

		last = seq

		seq, _ = strconv.Atoi(lines[i])

		if seq == last+1 {

			fmt.Printf("%s", thing)

		} else {

			//fmt.Println("WARNING: sequence break: current = ", seq, "last = ", last, "delta = ", seq-last)

			missingCount = seq - last - 1 + missingCount

			for item := last; item < seq; item++ {

				missingIds = append(missingIds, item)

				fmt.Printf("%s ", strconv.Itoa(item))

			}

		}

	}

	fmt.Println("missing message total: ", missingCount)
	/*
		for id := range missingIds {
			fmt.Printf("%s ", strconv.Itoa(id))
		}
	*/

}
