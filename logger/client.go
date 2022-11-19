package logger

import (
	"bytes"
	"distributedDemo/registry"
	"fmt"
	"log"
	"net/http"
)

func SetClientLogger(serviceURL string, clientService registry.ServiceName) {
	log.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	log.SetFlags(0)
	log.SetOutput(&clientLogger{url: serviceURL})
}

type clientLogger struct {
	url string
}

func (cl clientLogger) Write(data []byte) (int, error) {
	b := bytes.NewBuffer([]byte(data))
	res, err := http.Post(cl.url+"/logger", "text/plain", b)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send logger message. Service responded with %d - %s", res.StatusCode, res.Status)
	}
	return len(data), nil
}
