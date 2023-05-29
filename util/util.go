package util

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// InputData represents the structure of the data decoded from incoming HTTP
// POST requests.
type InputData struct {
	Info []interface{}   `json:"info"`
	Data [][]interface{} `json:"data"`
}

// ToMSec converts the given Duration value into milliseconds represented as
// float64.
func ToMSec(d time.Duration) float64 {
	return float64(d) / math.Pow10(6)
}

// GetTimestamp returns the current Unix timestamp in seconds as a float.
func GetTimestamp() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / math.Pow10(9)
}

// WriteToFile writes the given struct of type inputData to a file identified
// by the parameter path, which is given as a string. If the file does not
// exist, the function attempts to create it. Otherwise, the existing file is
// opened in append mode. All writes to the file are GZip compressed. If the
// function encounters an error at any point, an error is returned.
func WriteToFile(path string, body *InputData) error {
	start := time.Now()
	// Open file in append mode or create it otherwise
	outFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)

	if err != nil {
		return err
	}

	defer outFile.Close()
	// Initialise new gzip writer with maximum compression
	zipWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)

	if err != nil {
		return err
	}

	defer zipWriter.Close()
	receivedAt := GetTimestamp()
	infoPart := fmt.Sprintf("%f,", receivedAt)

	// Collect contents of key 'info' into a single string delimited by commas
	for _, item := range body.Info {
		// Skip nil values
		if item == nil {
			infoPart += ","
		} else {
			infoPart += fmt.Sprintf("%v,", item)
		}
	}

	// Iterate over key 'data'
	for _, entry := range body.Data {
		dataPart := ""

		// Collect contents into a single string delimited by commas
		for _, item := range entry {
			// Skip nil values
			if item == nil {
				dataPart += ","
			} else {
				dataPart += fmt.Sprintf("%v,", item)
			}
		}

		// Combine keys 'info' and 'data' and write it to zip writer
		zipWriter.Write([]byte(infoPart + dataPart[:len(dataPart)-1] + "\n"))
	}

	log.Printf("writeToFile %.3f", ToMSec(time.Since(start)))
	return nil
}

// ParseRequestBody extracts the body of the given HTTP request and tries to
// decode it as a JSON data structure conforming to the type inputData. If
// successful, a pointer to such a data structure is returned. The function
// also ensures that the keys 'info' and 'data' are present in the decoded data
// and are nonempty.
func ParseRequestBody(req *http.Request) (*InputData, error) {
	start := time.Now()
	// Initialise new decoder using body stream
	jsonDecoder := json.NewDecoder(req.Body)

	var bodyData *InputData = &InputData{}

	// Attempt to decode the stream and store results in bodyData
	if err := jsonDecoder.Decode(bodyData); err != nil {
		return nil, err
	}

	// Make sure the required keys are present and nonempty
	if len(bodyData.Info) == 0 || len(bodyData.Data) == 0 {
		return nil, errors.New("required fields are missing or invalid")
	}

	log.Printf("parseRequestBody %.3f", ToMSec(time.Since(start)))
	// Return pointer to data if successful
	return bodyData, nil
}

// IgnoreSighup attaches a handler to the SIGHUP signal so the program doesn't
// quit when the terminal session ends.
func IgnoreSighup() {
	// Launching goroutine to catch SIGHUP signals
	go func() {
		// Listend for SIGHUP on channel
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Signal(syscall.SIGHUP))

		// Listen on channel and discard signals
		for {
			<-ch
		}
	}()
}

// CheckAndCreateFolder checks if the folder given as parameter exists and
// creates it if not. Returns an error if directory could not be created
func CheckAndCreateFolder(dir string) error {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		log.Println("Storing data in existing directory", dir)

		// Return if directory exists
		return nil
	}

	log.Println("Creating output directory", dir)

	// Attempt to create directory hierarchy
	return os.MkdirAll(dir, 0755)
}

// LatestSampleTimestamp returns the modification timestamp of the latest file
// whose name starts with 'datalog-' in the given directory.
func LatestSampleTimestamp(dir string) (time.Time, error) {
	// Read directory
	files, err := os.ReadDir(dir)

	// Return error if directory could not be read
	if err != nil {
		return time.Time{}, err
	}

	datalogFiles := make([]os.DirEntry, 0, len(files))

	// Collect entries whose name starts with 'datalog-' into list
	for _, entry := range files {
		if strings.HasPrefix(entry.Name(), "datalog-") {
			datalogFiles = append(datalogFiles, entry)
		}
	}

	// Return error if there are no 'datalog-' files in the directory
	if len(datalogFiles) == 0 {
		return time.Time{}, errors.New("no logs captured yet")
	}

	// Get last entry and retrieve file info
	latestEntryInfo, err := datalogFiles[len(datalogFiles)-1].Info()

	if err != nil {
		return time.Time{}, err
	}

	// Return nodification time of last entry
	return latestEntryInfo.ModTime(), nil
}
