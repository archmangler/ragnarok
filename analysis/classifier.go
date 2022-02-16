package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func count_files(dir_path string) (c int) {
	c = 0
	files, _ := ioutil.ReadDir(dir_path)
	c = len(files)
	return c
}

//check for blank files (zero content)
func check_empty_files(dir_path string, files []fs.FileInfo) (fc int) {

	fCount := 0

	for i := range files {

		f := files[i]

		fPath := string(dir_path + "/" + f.Name())
		fSize := f.Size()

		if fSize == 0 {
			fmt.Println("blank -> ", fPath)
			fCount++
		}

	}

	return fCount

}

func check_duplicate_content(dir_path string, files []fs.FileInfo) (dupCount int) {

	//Modify this to map the filename to the duplicate results
	dupCount = 0
	md5map := make(map[string]string)
	results := make(map[string]int)

	//first pass to get the file md5 hashes
	for i := range files {
		f := files[i]
		fPath := string(dir_path + "/" + f.Name())
		theString := getFileMD5(fPath)

		md5map[fPath] = theString
	}

	//check for duplicates
	for k, v := range md5map {
		cnt := 0
		for _, j := range md5map {
			if v == j {
				cnt++
			}
		}
		results[k] = cnt
	}
	dupCount = printDuplicates(results)
	return dupCount
}

func printDuplicates(theMap map[string]int) (dCount int) {

	for k, v := range theMap {
		if v > 1 {
			fmt.Printf("DUP: %s => %s \n", k, strconv.Itoa(v))
			dCount++
		}
	}
	return dCount
}

func getFileMD5(fPath string) (sum string) {

	// Open file for reading
	file, err := os.Open(fPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	// Create new hasher, which is a writer interface
	hasher := md5.New()

	_, err = io.Copy(hasher, file)

	if err != nil {
		log.Fatal(err)
	}

	// Hash and print. Pass nil since
	// the data is not coming in as a slice argument
	// but is coming through the writer interface
	md5sum := hasher.Sum(nil)

	md5String := hex.EncodeToString(md5sum)

	return md5String

}

func get_file_sizes(dir_path string) (s map[int64]int64) {

	s = make(map[int64]int64)
	var cnt int64 = 0

	files, _ := ioutil.ReadDir(dir_path)
	//first pass to get the sizes
	for i := range files {
		f := files[i]
		s[f.Size()] = f.Size()
	}

	for k, _ := range s {

		cnt = 0
		for i := range files {
			f := files[i]
			if k == f.Size() {
				cnt++
				s[k] = cnt
			}
		}
	}

	return s
}

func main() {

	source_dir := os.Args[1]

	fmt.Println("**************** BEGIN analysing contents ********************")
	fmt.Println("total: ", count_files(source_dir))

	files, _ := ioutil.ReadDir(source_dir)

	bin := get_file_sizes(source_dir)

	for k, v := range bin {
		fmt.Printf("size %s: count: %s\n", strconv.FormatInt(k, 10), strconv.FormatInt(v, 10))
	}

	//check for blank files (zero content)
	fCount := check_empty_files(source_dir, files)
	fmt.Printf("empty files: %s\n", strconv.Itoa(fCount))

	//check for duplicate files
	dupCount := check_duplicate_content(source_dir, files)
	fmt.Println("duplicated files: ", dupCount)

	//TODO: check for content with empty json fields
	//check_missing_fields(files)

	fmt.Println("**************** DONE analysing contents ********************")

}
