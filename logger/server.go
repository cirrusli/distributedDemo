package logger

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var logger *log.Logger

type fileLog string

// RegisterHandlers 注册路由
func RegisterHandlers() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

//将日志数据写入文件
func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println("open file failed:")
		return 0, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Failed to close")
		}
	}(f)
	return f.Write(data)
}

// Run 存储日志文件的路径
func Run(destination string) {
	logger = log.New(fileLog(destination), "[go] - ", log.LstdFlags)
}

func write(message string) {
	log.Println("func write:\n", message)
	//由此写入文件
	logger.Printf("%v\n", message)
}
