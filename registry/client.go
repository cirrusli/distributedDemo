package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RegisterService 给registryService服务发送一个POST请求
func RegisterService(r Registration) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(r)
	if err != nil {
		return err
	}
	res, err := http.Post(ServiceURL, "application/json", buf)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service."+
			"Registry service responded with %d", res.StatusCode)
	}
	return nil
}
