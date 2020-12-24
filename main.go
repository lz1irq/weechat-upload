package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const POSTField = "file"

type config struct {
	Listen       string
	UploadDir    string
	PrefixURL    string
	AuthUsername string
	AuthPassword string
}

var conf config

func init() {

	flag.StringVar(&conf.Listen, "listen", "0.0.0.0:8080", "IP and port to bind to")
	flag.StringVar(&conf.UploadDir, "upload.dir", "upload", "Directory for file uploads, must exist and be writeable")
	flag.StringVar(&conf.PrefixURL, "url.prefix", "http://127.0.0.1:8080", "Public URL to prefix upload file paths with")
	flag.StringVar(&conf.AuthUsername, "auth.user", "user", "HTTP basic auth username")
	flag.StringVar(&conf.AuthPassword, "auth.pass", "pass", "HTTP basic auth password")
	flag.Parse()
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

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			log.Println("Error parsing basic auth")
			w.WriteHeader(http.StatusForbidden)
		}
		if user != conf.AuthUsername || pass != conf.AuthPassword {
			log.Println("Wrong username/password")
			w.WriteHeader(http.StatusForbidden)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/upload", basicAuth(upload))

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

	formFile, header, err := req.FormFile(POSTField)
	if err != nil {
		log.Printf("Error getting file from field %s: %s\n", POSTField, err.Error())
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
