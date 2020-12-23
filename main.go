package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type config struct {
	Listen    string
	UploadDir string
	PrefixURL string
	POSTField string
}

var conf config

func init() {
	conf.Listen = "0.0.0.0:8080"
	conf.UploadDir = "./upload"
	conf.PrefixURL = "http://192.168.1.77:8080/files/"
	conf.POSTField = "file"
}

func noIndex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/upload", upload)

	fileServer := http.FileServer(http.Dir(conf.UploadDir))
	http.Handle("/files/", http.StripPrefix("/files", noIndex(fileServer)))

	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}

func upload(w http.ResponseWriter, req *http.Request) {

	//reqInfo, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	log.Println("Failed to dump request")
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}
	//log.Println(string(reqInfo))

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	formFile, header, err := req.FormFile(conf.POSTField)
	if err != nil {
		log.Printf("Error getting file from field %s: %s\n", conf.POSTField, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	outputName := conf.UploadDir + "/" + header.Filename
	outputFile, err := os.Create(outputName)
	if err != nil {
		log.Printf("Failed to create output file %s: %s\n", outputName, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	if _, err := io.Copy(outputFile, formFile); err != nil {
		log.Printf("Failed to write to output file %s: %s", outputName, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(conf.PrefixURL + "/" + header.Filename))

}
