package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const (
	uploadResponseContentType = "application/json"
	serverName                = "tomdawg"
)

//Config listen port and path to store files
type Config struct {
	ListenPort int
	AssetPath  string
}

//UploadResponse is returned as json response
type UploadResponse struct {
	Path, Status, Description string
	Time, Speed, Size, Recvd  int64
}

func (ur UploadResponse) out(dst io.Writer) {
	b, _ := json.Marshal(ur)
	dst.Write(b)
}

var config Config

func setupLogger() {
	logDir := path.Join(path.Dir(os.Args[0]), "logs")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	logw, err := os.Create(path.Join(logDir, fmt.Sprintf("%s.log", serverName)))
	if err != nil {
		log.Println(err)
		return
	}
	log.SetOutput(logw)
}

func headHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", serverName)
	w.Header().Set("Content-Type", uploadResponseContentType)

	log.Printf("RawQuery %s\n", r.URL.RawQuery)
	log.Printf("RawQuery %v\n", r.URL)

	saveTo := path.Join(config.AssetPath, r.URL.Path)
	fileInfo, err := os.Stat(saveTo)
	if err == nil {
		w.Header().Set("Offset", fmt.Sprintf("%d", fileInfo.Size()))
		w.WriteHeader(200)
	} else {
		w.WriteHeader(404)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", serverName)

	log.Printf("RawQuery %s\n", r.URL.RawQuery)
	log.Printf("RawQuery %v\n", r.URL)

	saveTo := path.Join(config.AssetPath, r.URL.Path)
	_, err := os.Stat(saveTo)
	if err == nil {
		http.ServeFile(w, r, saveTo)
	} else {
		w.WriteHeader(404)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {

	ur := UploadResponse{}
	reader, err := r.MultipartReader()
	if err != nil {
		log.Println(err)
		ur.Status = "multipart_reader_error"
		ur.Description = err.Error()
		ur.out(w)
		return
	}

	var totalBytes int64
	uploadCount := 0

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if part.FileName() == "" {
			continue
		}

		saveTo := path.Join(config.AssetPath, part.FormName(), part.FileName())
		saveToDir := filepath.Dir(saveTo)
		err = os.MkdirAll(saveToDir, 0755)
		if err != nil {
			log.Println(err)
			ur.Status = "multipart_mkdir_error"
			ur.Description = err.Error()
			ur.out(w)
			return
		}

		log.Printf("Save to %s\n", saveTo)
		saveTo, err = filepath.Abs(saveTo)
		if err != nil {
			log.Println(err)
			UploadResponse{
				Status:      "abs_file_error",
				Description: err.Error(),
				Path:        saveTo,
			}.out(w)
			return
		}

		ur.Path = saveTo

		out, err := os.Create(saveTo)
		if err != nil {
			log.Println(err)
			ur.Status = "multipart_create_error"
			ur.Description = err.Error()
			ur.out(w)
			return
		}

		defer out.Close()
		n, err := io.Copy(out, part)
		if err != nil {
			log.Println(err)
			ur.Status = "multipart_write_error"
			ur.Description = err.Error()
			ur.out(w)
			return
		}
		uploadCount += 1
		totalBytes += n
	}
	ur.Status = "success"
	//@todo
	ur.Size = totalBytes
	ur.Recvd = totalBytes
	ur.Description = fmt.Sprintf("Total Files: %d Total Bytes: %d", uploadCount, totalBytes)
	ur.out(w)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", serverName)
	w.Header().Set("Content-Type", uploadResponseContentType)

	log.Printf("RawQuery %s\n", r.URL.RawQuery)
	log.Printf("RawQuery %v\n", r.URL)

	saveTo := path.Join(config.AssetPath, r.URL.Path)

	saveToDir := filepath.Dir(saveTo)
	err := os.MkdirAll(saveToDir, 0755)
	if err != nil {
		log.Println(err)
		UploadResponse{
			Status:      "mkdir_path_error",
			Description: err.Error(),
			Path:        saveTo,
		}.out(w)
		return
	}

	log.Printf("Save to %s\n", saveTo)
	saveTo, err = filepath.Abs(saveTo)
	if err != nil {
		log.Println(err)
		UploadResponse{
			Status:      "abs_file_error",
			Description: err.Error(),
			Path:        saveTo,
		}.out(w)
		return
	}
	log.Printf("Absolute path Save to %s\n", saveTo)

	ur, _ := func() (interface{}, error) {

		ur := UploadResponse{
			Path: saveTo,
		}

		fileInfo, err := os.Stat(saveTo)
		if err == nil {
			ur.Status = "success"
			ur.Description = "Already Uploaded"
			ur.Size = fileInfo.Size()
			ur.Recvd = ur.Size
			log.Printf("File %s Exists %d bytes\n", saveTo, ur.Size)
			return ur, nil
		}

		contentLength, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			log.Println(err)
			ur.Status = "invalid_content_length"
			ur.Description = err.Error()
			return ur, err
		}
		ur.Size = contentLength

		partialSaveTo := fmt.Sprintf("%s.part", saveTo)
		log.Printf("Save to partial file %s\n", partialSaveTo)
		out, err := os.Create(partialSaveTo)
		if err != nil {
			log.Println(err)
			ur.Status = "file_create_error"
			ur.Description = err.Error()
			return ur, err
		}

		t0 := time.Now()
		defer out.Close()
		n, err := io.Copy(out, r.Body)
		if err != nil {
			log.Println(err)
			ur.Status = "upload_write_error"
			ur.Description = err.Error()
			return ur, err
		}
		ur.Recvd = n
		ur.Time = int64(time.Now().Sub(t0).Seconds())
		if ur.Time > 0 {
			ur.Speed = n / (1024 * ur.Time)
		}

		log.Printf("UPLOAD_STATS %s -> %d/%d bytes, %d KB/s total sec %d\n",
			partialSaveTo, n, contentLength, ur.Speed, ur.Time)
		if n != contentLength {
			err := fmt.Errorf("incomplete upload %d/%d", n, contentLength)
			log.Println(err)
			ur.Status = "incomplete_upload"
			ur.Description = err.Error()
			return ur, err
		}

		log.Printf("Renaming %s -> %s ", partialSaveTo, saveTo)
		err = os.Rename(partialSaveTo, saveTo)
		if err != nil {
			log.Println(err)
			ur.Status = "file_rename_error"
			ur.Description = err.Error()
			return ur, err
		}

		ur.Status = "success"
		ur.Description = "Uploaded successfully"
		return ur, nil
	}()
	log.Println("Upload Response ->", ur)
	ur.(UploadResponse).out(w)
}

func upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		putHandler(w, r)
	case "POST":
		postHandler(w, r)
	case "HEAD":
		headHandler(w, r)
	case "GET":
		getHandler(w, r)
	default:
		UploadResponse{
			Status:      "method_not_allowed",
			Description: r.Method,
		}.out(w)
	}
}

func main() {
	//Setup Logger
	setupLogger()

	configFile := flag.String("config", path.Join(path.Dir(os.Args[0]), "config.json"), "config file path")
	flag.Parse()

	log.Printf("Loading config file %s\n", *configFile)
	f, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Parsing config file %s\n", *configFile)
	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Asset path %s\n", config.AssetPath)
	err = os.MkdirAll(config.AssetPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", upload)
	log.Printf("Listening on port %d\n", config.ListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ListenPort), nil))
}
