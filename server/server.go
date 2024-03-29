package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/cwi-dis/mozillahubs-datalogger/util"
)

// SuccessResponse represents the structure of the data that is returned as
// JSON upon a successfully completed request.
type SuccessResponse struct {
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// LastSampleResponse represents the structure of the data that is returned as
// JSON when requesting the timestamp of the lastest sample
type LastSampleResponse struct {
	Status     string `json:"status"`
	LastSample string `json:"lastSample"`
}

// ErrorResponse represents the structure of the data that is returned as JSON
// when the request encountered an error.
type ErrorResponse struct {
	Status string `json:"status"`
}

// CreateLogHandler returns a HTTP handler function which reacts to POST
// requests, takes the body data, parses it as JSON and writes the resulting
// data to a file at the path which is given by a parameter to this function.
func CreateLogHandler(saveDir string) func(http.ResponseWriter, *http.Request) {
	// Initialise mutex to guaranteee exclusive access to output file
	var mutex = &sync.Mutex{}

	// Return HTTP handler function
	return func(writer http.ResponseWriter, req *http.Request) {
		// Make sure request is a POST request
		if req.Method != http.MethodPost {
			// Send error message with code 405 to client
			msg, _ := json.Marshal(&ErrorResponse{Status: "Method not allowed"})
			http.Error(writer, string(msg), http.StatusMethodNotAllowed)

			return
		}

		log.Println("Processing request")
		start := time.Now()

		// Attempt to decode the request body as JSON
		bodyData, err := util.ParseRequestBody(req)

		if err != nil {
			log.Println("Could not decode body data:", err)

			// Send error message with code 400 to client
			msg, _ := json.Marshal(&ErrorResponse{Status: "Input data invalid"})
			http.Error(writer, string(msg), http.StatusBadRequest)

			return
		}

		// Generate filename and save path
		fileName := time.Now().Format("datalog-2006-01-02.csv.gz")
		savePath := path.Join(saveDir, fileName)

		// Lock the mutex for file access
		mutex.Lock()
		defer mutex.Unlock()

		// Attempt to write the data to a file
		if err := util.WriteToFile(savePath, bodyData); err != nil {
			// Handle error
			log.Println("Could not save data to file:", err)

			// Send error message with code 500 to client
			msg, _ := json.Marshal(&ErrorResponse{Status: "Could not write to file"})
			http.Error(writer, string(msg), http.StatusInternalServerError)

			return
		}

		// Generate response JSON with current timestamp
		msg, _ := json.Marshal(&SuccessResponse{
			Status: "OK",
			Time:   util.GetTimestamp(),
		})

		log.Printf("createHandlerWithPath %.3f", util.ToMSec(time.Since(start)))
		// Write response to client
		fmt.Fprint(writer, string(msg))
	}
}

// CreateLastUpdateHandler returns a HTTP handler function that responds to GET
// requests to retrieve the timestamp of the last sample that was received.
func CreateLastUpdateHandler(saveDir string) func(http.ResponseWriter, *http.Request) {
	// Return HTTP handler function
	return func(writer http.ResponseWriter, req *http.Request) {
		// Make sure the request if a GET request
		if req.Method != http.MethodGet {
			msg, _ := json.Marshal(&ErrorResponse{Status: "Method not allowed"})
			http.Error(writer, string(msg), http.StatusMethodNotAllowed)

			return
		}

		// Get timestamp of last sample
		latestSampleTimestamp, err := util.LatestSampleTimestamp(saveDir)

		// Return HTTP 500 if there was an error retrieving the sample
		if err != nil {
			msg, _ := json.Marshal(&ErrorResponse{Status: "Could not retrieve timestamp"})
			http.Error(writer, string(msg), http.StatusInternalServerError)

			return
		}

		// Generate response JSON with timestamp
		msg, _ := json.Marshal(&LastSampleResponse{
			Status:     "OK",
			LastSample: latestSampleTimestamp.Format(time.RFC3339Nano),
		})

		// Write response to client
		fmt.Fprint(writer, string(msg))
	}
}

// StartServer starts a new HTTP server at the given port which saves data
// received from clients in the given directory under an automatically
// generated name. The server mounts a handler for the route POST /mozillahubs
// to store the data and a handler for retrieving the timestamp of the latest
// stored sample under GET /latest
func StartServer(saveDir string, port int) {
	portSpec := fmt.Sprintf(":%d", port)

	log.Printf("Server listening on port %d", port)

	// Install handlers
	http.HandleFunc("/mozillahubs", CreateLogHandler(saveDir))
	http.HandleFunc("/latest", CreateLastUpdateHandler(saveDir))

	// Start server
	if err := http.ListenAndServe(portSpec, nil); err != nil {
		log.Println(err)
	}
}
