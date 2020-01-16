package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const PATH = "./temp/"

func main() {
	fmt.Printf("[%s] Running.\n", time.Now().Local())
	// path check
	if _, err := os.Stat(PATH); err != nil {
		err := os.MkdirAll(PATH, 0711)
		if err != nil {
			fmt.Printf("[%s] Make dir error. %s\n", time.Now().Local(), err)
			os.Exit(1)
		}
	}

	// http server
	http.HandleFunc("/file_upload", fileUpload)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("[%s] Listen to port error. %s\n", time.Now().Local(), err)
		os.Exit(1)
	}
}

func fileUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] [%s] %s %s\n", time.Now().Local(), strings.Split(r.RemoteAddr, ":")[0], r.Method, r.URL)
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	}
	if "OPTIONS" == r.Method {
		return
	}
	if "POST" != r.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//get file
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("[%s] Get file error. %s\n", time.Now().Local(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	//save file
	name, err := save(file, header)
	if err != nil {
		fmt.Printf("[%s] Save file error. %s\n", time.Now().Local(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//execute command
	cmd := exec.Command("uncompyle6", PATH+name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[%s] Command execute error %s\n", time.Now().Local(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//result check
	if strings.Index(fmt.Sprintf("%s", out), "# uncompyle6") == -1 {
		fmt.Printf("[%s] Unable to parse file\n", time.Now().Local())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//return
	fmt.Printf("[%s] Success\n", time.Now().Local())
	w.Write(out)
}

//save file to local
func save(file multipart.File, header *multipart.FileHeader) (name string, err error) {
	name = strconv.FormatInt(time.Now().UnixNano(), 10) + "." + strings.Split(header.Filename, ".")[1]
	f, err := os.OpenFile(PATH+name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, file)
	return
}
