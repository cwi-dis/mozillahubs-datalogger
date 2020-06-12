package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

// inputData represents the structure of the data decoded from incoming HTTP
// POST requests.
type inputData struct {
	Info []interface{}   `json:"info"`
	Data [][]interface{} `json:"data"`
}

// successResponse is the structure of the data that is returned as JSON upon
// a successfully completed request.
type successResponse struct {
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// errorResponse is the structure of the data that is returned as JSON when the
// request encountered an error.
type errorResponse struct {
	Status string `json:"status"`
}

// getTimestamp returns the current Unix timestamp in seconds as a float.
func getTimestamp() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / math.Pow10(9)
}

// writeToFile writes the given struct of type inputData to a file identified
// by the parameter path, which is given as a string. If the file does not
// exist, the function attempts to create it. Otherwise, the existing file is
// opened in append mode. All writes to the file are GZip compressed. If the
// function encounters an error at any point, an error is returned.
func writeToFile(path string, body *inputData) error {
	start := time.Now().UnixNano()
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

	log.Printf("writeToFile %d", time.Now().UnixNano()-start)
	return nil
}

// parseRequestBody extracts the body of the given HTTP request and tries to
// decode it as a JSON data structure conforming to the type inputData. If
// successful, a pointer to such a data structure is returned. The function
// also ensures that the keys 'info' and 'data' are present in the decoded data
// and are nonempty.
func parseRequestBody(req *http.Request) (*inputData, error) {
	start := time.Now().UnixNano()
	// Initialise new decoder using body stream
	jsonDecoder := json.NewDecoder(req.Body)

	var bodyData *inputData = &inputData{}

	// Attempt to decode the stream and store results in bodyData
	if err := jsonDecoder.Decode(bodyData); err != nil {
		return nil, err
	}

	// Make sure the required keys are present and nonempty
	if len(bodyData.Info) == 0 || len(bodyData.Data) == 0 {
		return nil, errors.New("Required fields are missing or invalid")
	}

	// Return pointer to data if successful
	log.Printf("parseRequestBody %d", time.Now().UnixNano()-start)
	return bodyData, nil
}

// createHandlerWithPath returns a HTTP handler function which reacts to POST
// requests, takes the body data, parses it as JSON and writes the resulting
// data to a file at the path which is given by a parameter to this function.
func createHandlerWithPath(saveDir string) func(http.ResponseWriter, *http.Request) {
	// Initialise mutex to guaranteee exclusive access to output file
	var mutex = &sync.Mutex{}

	// Return HTTP handler function
	return func(writer http.ResponseWriter, req *http.Request) {
		// Make sure request is a POST request
		if req.Method == http.MethodPost {
			log.Println("Processing request")
			start := time.Now().UnixNano()

			// Attempt to decode the request body as JSON
			bodyData, err := parseRequestBody(req)

			if err != nil {
				log.Println("Could not decode body data:", err)

				// Send error message with code 400 to client
				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			// Generate filename and save path
			fileName := time.Now().Format("datalog-2006-01-02.csv.gz")
			savePath := path.Join(saveDir, fileName)

			// Lock the mutex for file access
			mutex.Lock()

			// Attempt to write the data to a file
			if err := writeToFile(savePath, bodyData); err != nil {
				// Unlock mutex and handle error
				mutex.Unlock()
				log.Println("Could not save data to file:", err)

				// Send error message with code 400 to client
				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			// Unlock mutex
			mutex.Unlock()

			// Generate response JSON with current timestamp
			msg, _ := json.Marshal(&successResponse{
				Status: "ok",
				Time:   getTimestamp(),
			})

			log.Printf("createHandlerWithPath %d", time.Now().UnixNano()-start)
			// Write response to client
			fmt.Fprintf(writer, string(msg))
		} else {
			// Send 404 to client if request is not POST
			http.NotFound(writer, req)
		}
	}
}

// startServer starts a new HTTP server at the given port which saves data
// received from clients in the given directory under an automatically
// generated name. The server mounts a handler for the route POST /mozillahubs
func startServer(saveDir string, port int) {
	portSpec := fmt.Sprintf(":%d", port)

	log.Printf("Server listening on port %d", port)

	http.HandleFunc("/mozillahubs", createHandlerWithPath(saveDir))
	err := http.ListenAndServe(portSpec, nil)

	if err != nil {
		log.Println(err)
	}
}

// main parses the command line arguments which contain the output path as
// positional argument and optionally a port for the HTTP server specified by
// the flag -p
func main() {
	port := flag.Int("p", 6000, "Port to listen on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("USAGE:", os.Args[0], "[-p port] saveDir")
		os.Exit(1)
	}

	startServer(flag.Arg(0), *port)
}
