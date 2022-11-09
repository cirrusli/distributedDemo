package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

//添加服务注册
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

//取消服务注册
func (r *registry) remove(url string) error {
	//check whether the url exist
	for i := range reg.registrations {
		if reg.registrations[i].ServiceURL == url {
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("method remove:service at URL %s not found", url)
}

//包级变量
var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.Mutex),
}

// RegService 让如下结构体成为httpserver类型
type RegService struct{}

func (s RegService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Method ServeHTTP:Request Received")
	switch r.Method {
	//注册服务
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println("Method ServeHTTP:Error decoding registration", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Method ServeHTTP:Adding service:%v with URL:%s\n", r.ServiceName, r.ServiceURL)
		err = reg.add(r)
		if err != nil {
			log.Println("Method ServeHTTP:add service failed", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	//取消服务
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Method ServeHTTP:Error decoding registration", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Method ServeHTTP:Removing service at URL:%s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println("Method ServeHTTP:add service failed", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	}
}
