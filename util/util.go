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
	"time"
)

// InputData represents the structure of the data decoded from incoming HTTP
// POST requests.
type InputData struct {
	Info []interface{}   `json:"info"`
	Data [][]interface{} `json:"data"`
}

// ToMSec converts the given Duration value into seconds represented as float64.
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
	infoPart := ""

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

	log.Printf("writeToFile %.3f", ToMSec(time.Now().Sub(start)))
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
		return nil, errors.New("Required fields are missing or invalid")
	}

	// Return pointer to data if successful
	log.Printf("parseRequestBody %.3f", ToMSec(time.Now().Sub(start)))
	return bodyData, nil
}
