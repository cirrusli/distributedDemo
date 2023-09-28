package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	//可能被多个线程并发地访问，因此为了保证线程安全，要加上互斥锁
	mutex *sync.RWMutex
}

// 添加服务注册
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	err := r.sendRequiredServices(reg)
	r.notify(patch{
		Added: []patchEntry{
			//待注册的服务的名称与URL
			{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})
	return err
}

// 当服务注册或被移除时进行通知
// todo receiver
func (r registry) notify(fullPatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	//遍历已经注册的服务
	for _, reg := range r.registrations {
		//使用goroutine并发地发送通知
		go func(reg Registration) {
			//针对每个注册的服务，对其所依赖的服务进行循环
			for _, reqService := range reg.RequiredServices {
				p := patch{
					Added:   []patchEntry{},
					Removed: []patchEntry{},
				}
				//发送更新的标志位
				sendUpdate := false
				//看看添加了哪些服务
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

// 请求所依赖的服务
// todo receiver
func (r registry) sendRequiredServices(reg Registration) error {
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

// 当一个服务出现时，想要通知依赖该服务的其他服务
// todo receiver
func (r registry) sendPatch(p patch, url string) error {
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

// 取消服务注册
func (r *registry) remove(url string) error {
	//check whether the url exist
	for i := range reg.registrations {
		if reg.registrations[i].ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("method remove of registry:service at URL %s not found", url)
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()
				successFlag := true
				for attempts := 0; attempts < 3; attempts++ {
					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println("In ./registry/server.go:Method heartbeat of registry:", err)
					} else if res.StatusCode == http.StatusOK {
						log.Println("Heartbeat check passed for", reg.ServiceName)
						if !successFlag {
							err := r.add(reg)
							if err != nil {
								return
							}
						}
						break
					}
					log.Println("Heartbeat check failed for", reg.ServiceName)
					if successFlag {
						successFlag = false
						err := r.remove(reg.ServiceURL)
						if err != nil {
							return
						}
					}
					time.Sleep(1 * time.Second)
				}
			}(reg)
			wg.Wait()
			time.Sleep(freq)
		}
	}
}

var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(10 * time.Second)
	})
}

// 包级变量
var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

// RegService 让如下结构体成为httpserver类型
type RegService struct{}

func (s RegService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Method ServeHTTP of RegService:Request received")
	switch r.Method {
	//注册服务
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println("Method ServeHTTP of RegService:Error decoding registration", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Method ServeHTTP of RegService:Adding service:%v with URL:%s\n", r.ServiceName, r.ServiceURL)
		err = reg.add(r)
		if err != nil {
			log.Println("Method ServeHTTP of RegService:add service failed", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	//取消服务
	case http.MethodDelete:
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Method ServeHTTP of RegService:Error decoding registration", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Method ServeHTTP of RegService:Removing service at URL:%s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println("Method ServeHTTP of RegService:add service failed", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	}
}
