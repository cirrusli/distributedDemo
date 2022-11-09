package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServiceURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	//可能被多个线程并发地访问，因此为了保证线程安全，要加上互斥锁
	mutex *sync.Mutex
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

//包级变量
var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
}

// RegService 让如下结构体成为httpserver类型
type RegService struct{}

func (s RegService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request Received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println("Error decoding registration", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service:%v with URL:%s\n", r.ServiceName, r.ServiceURL)
		err = reg.add(r)
		if err != nil {
			log.Println("Error decoding registration", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	}
}
