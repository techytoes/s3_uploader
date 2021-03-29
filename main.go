package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"encoding/json"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Replace this config
const (
	S3_REGION = ""
	S3_BUCKET = ""
)

// Size constants
const (
	MB = 1 << 20
)

type Sizer interface {
	Size() int64
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 * MB); err != nil {
		panic(err)
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, 10*MB) // 10 Mb

	// Creates S3 Client
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}

	// Fetches file from multipart request
	file, _, err := r.FormFile("file")
	fileName := r.FormValue("file_name")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a buffer to store the header of the file in
	size := file.(Sizer).Size()
	fileHeader := make([]byte, size)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		panic(err)
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		panic(err)
	}

	contentType := http.DetectContentType(fileHeader)

	log.Printf("Size: %#v\n", size)
	log.Printf("MIME: %#v\n", contentType)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(fileHeader),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "OK",
		"message": "🚀🌟🌈 Uploaded Successfully",
	})
}

func Home(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "OK",
		"message": "🚀🌟🌈",
	})
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Home).Methods("GET")
	router.HandleFunc("/file", UploadFile).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
