package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
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

func main() {
	http.HandleFunc(
		"/",
		logRequest(http.NotFoundHandler().ServeHTTP),
	)

	http.HandleFunc("/upload", logRequest(basicAuth(upload)))

	fileServer := http.FileServer(http.Dir(conf.UploadDir))
	http.Handle(
		"/files/",
		logRequest(noIndex(http.StripPrefix("/files", fileServer).ServeHTTP)),
	)

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
		log.WithFields(makeRequestLog(req)).WithField("POSTField", POSTField).WithError(err).Error("Failed to get file from POST")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	outputName := conf.UploadDir + "/" + path.Base(header.Filename)
	outputFile, err := os.Create(outputName)
	if err != nil {
		log.WithFields(makeRequestLog(req)).WithField("outputName", outputName).WithError(err).Error("Failed to create output file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	if _, err := io.Copy(outputFile, formFile); err != nil {
		log.WithFields(makeRequestLog(req)).WithField("outputName", outputName).WithError(err).Error("Failed to write to output file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(conf.PrefixURL + "/files/" + header.Filename))

}
