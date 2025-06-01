package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath" 
	"strings"      
	"blobstore/store" 
)

var myStore store.MySqlStore

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func sendJSONError(w http.ResponseWriter, message string, details string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message, Details: details})
}

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"

	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg": 
		return "audio/ogg"
	case ".aac":
		return "audio/aac"

	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogv": // Ogg video
		return "video/ogg"

	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"

	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"

	default:
		return "application/octet-stream" // idk what files you uploading
	}
}

func CreateBucketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method Not Allowed", "Only POST method is supported.", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/buckets/create" {
		sendJSONError(w, "Not Found", "The requested resource was not found.", http.StatusNotFound)
		return
	}

	var requestBody struct {
		BucketName string `json:"bucketName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendJSONError(w, "Bad Request", "Invalid JSON request body.", http.StatusBadRequest)
		return
	}

	if requestBody.BucketName == "" {
		sendJSONError(w, "Bad Request", "Bucket name cannot be empty.", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to create bucket: %s", requestBody.BucketName)
	err := myStore.CreateBucket(requestBody.BucketName)
	if err != nil {
		sendJSONError(w, "Internal Server Error", fmt.Sprintf("Failed to create bucket: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"message": fmt.Sprintf("Bucket '%s' created successfully.", requestBody.BucketName)}, http.StatusCreated)
}

// GetBucketsHandler handles GET /buckets requests to retrieve all bucket names.
func GetBucketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method Not Allowed", "Only GET method is supported.", http.StatusMethodNotAllowed)
		return
	}
	// Ensure that requests to /buckets/ are redirected to /buckets for GET
	if r.URL.Path != "/buckets" && r.URL.Path != "/buckets/" { // Allow both /buckets and /buckets/ for GET
		sendJSONError(w, "Not Found", "The requested resource was not found.", http.StatusNotFound)
		return
	}

	log.Println("Attempting to get all buckets.")
	bucketNames, err := myStore.GetBuckets()
	if err != nil {
		sendJSONError(w, "Internal Server Error", fmt.Sprintf("Failed to retrieve buckets: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string][]string{"buckets": bucketNames}, http.StatusOK)
}

func InsertFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method Not Allowed", "Only POST method is supported.", http.StatusMethodNotAllowed)
		return
	}

	// Extract bucketName from URL path /buckets/{bucketName}/files
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[2] == "" || parts[3] != "files" {
		sendJSONError(w, "Bad Request", "Invalid URL path. Expected /buckets/{bucketName}/files", http.StatusBadRequest)
		return
	}
	bucketName := parts[2]

	var requestBody struct {
		FileName string `json:"fileName"`
		FileData string `json:"fileData"` // Base64 encoded byte data
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendJSONError(w, "Bad Request", "Invalid JSON request body.", http.StatusBadRequest)
		return
	}

	if requestBody.FileName == "" || requestBody.FileData == "" {
		sendJSONError(w, "Bad Request", "File name and file data cannot be empty.", http.StatusBadRequest)
		return
	}

	// Decode Base64 file data
	fileDataBytes, err := base64.StdEncoding.DecodeString(requestBody.FileData)
	if err != nil {
		sendJSONError(w, "Bad Request", "Invalid Base64 encoding for fileData.", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to insert file '%s' into bucket '%s'", requestBody.FileName, bucketName)
	err = myStore.InsertFile(bucketName, requestBody.FileName, fileDataBytes)
	if err != nil {
		sendJSONError(w, "Internal Server Error", fmt.Sprintf("Failed to insert file: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string]string{"message": fmt.Sprintf("File '%s' inserted successfully into bucket '%s'.", requestBody.FileName, bucketName)}, http.StatusCreated)
}

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method Not Allowed", "Only GET method is supported.", http.StatusMethodNotAllowed)
		return
	}

	// Extract bucketName from URL path /buckets/{bucketName}/files
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[2] == "" || parts[3] != "files" {
		sendJSONError(w, "Bad Request", "Invalid URL path. Expected /buckets/{bucketName}/files", http.StatusBadRequest)
		return
	}
	bucketName := parts[2]

	log.Printf("Attempting to get files from bucket: %s", bucketName)
	fileNames, err := myStore.GetFiles(bucketName)
	if err != nil {
		sendJSONError(w, "Internal Server Error", fmt.Sprintf("Failed to retrieve files from bucket '%s': %v", bucketName, err), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, map[string][]string{"fileNames": fileNames}, http.StatusOK)
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method Not Allowed", "Only GET method is supported.", http.StatusMethodNotAllowed)
		return
	}

	// Extracting bucketName and fileName from URL path /buckets/{bucketName}/files/{fileName}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 || parts[2] == "" || parts[3] != "files" || parts[4] == "" {
		sendJSONError(w, "Bad Request", "Invalid URL path. Expected /buckets/{bucketName}/files/{fileName}", http.StatusBadRequest)
		return
	}
	bucketName := parts[2]
	fileName := parts[4]

	log.Printf("Attempting to get file '%s' from bucket '%s' for direct serving", fileName, bucketName)
	fileData, err := myStore.GetFile(bucketName, fileName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") { // Check for "not found" error message
			sendJSONError(w, "Not Found", fmt.Sprintf("File '%s' not found in bucket '%s'.", fileName, bucketName), http.StatusNotFound)
		} else {
			sendJSONError(w, "Internal Server Error", fmt.Sprintf("Failed to retrieve file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	contentType := getContentType(fileName)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// Writing the raw file data to the response body
	_, err = w.Write(fileData)
	if err != nil {
		log.Printf("Error writing file data to response for '%s': %v", fileName, err)
	}
	log.Printf("Successfully served file '%s' from bucket '%s' with Content-Type: %s", fileName, bucketName, contentType)
}

func main() {
	config := store.DBConfig{
		Host:     "localhost", 
		User:     "root",
		Password: "root",
		Database: "mysqlrest",
	}

	var err error
	myStore, err = store.ConnectToStore(config)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer myStore.Close() // Ensure the database connection is closed when main exits

	log.Println("MySQL Store connected successfully.")

	mux := http.NewServeMux()

	// GET /buckets (list all buckets)
	mux.HandleFunc("/buckets", GetBucketsHandler) 

	// POST /buckets (create a bucket)
	mux.HandleFunc("/buckets/create", CreateBucketHandler)

	// Handlers for /buckets/{bucketName}/files and /buckets/{bucketName}/files/{fileName}
	// Thse will use a single handler that dispatches based on URL structure.
	mux.HandleFunc("/buckets/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/") // Normalize path by removing trailing slash for easier parsing
		parts := strings.Split(path, "/")

		// /buckets/{bucketName}/files
		if len(parts) == 4 && parts[3] == "files" {
			if r.Method == http.MethodPost {
				InsertFileHandler(w, r)
				return
			} else if r.Method == http.MethodGet {
				GetFilesHandler(w, r)
				return
			}
		}

		// /buckets/{bucketName}/files/{fileName}
		if len(parts) == 5 && parts[3] == "files" {
			if r.Method == http.MethodGet {
				GetFileHandler(w, r) // This handler now serves the file directly
				return
			}
		}

		sendJSONError(w, "Not Found", "The requested resource was not found or method not supported.", http.StatusNotFound)
	})


	port := ":8080"
	log.Printf("Starting HTTP server on port %s...", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
