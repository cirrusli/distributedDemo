package registry

import (
	"bytes"
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
	mutex *sync.RWMutex
}

//添加服务注册
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	err := r.sendRequiredServices(reg)
	return err
}

//请求所依赖的服务
func (r *registry) sendRequiredServices(reg Registration) error {
	//仅需要一个读的锁
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	//查看要添加的服务是否存在
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				//存在则添加到待注册服务列表中
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}

		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}
func (r *registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	//使用NewBuffer将变量d变为ioReader类型
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
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
	mutex:         new(sync.RWMutex),
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
