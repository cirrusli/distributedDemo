package registry

type Registration struct {
	ServiceName      ServiceName
	ServiceURL       string
	RequiredServices []ServiceName
	//如：当registry中有grade所依赖的logger服务时，由此返回
	ServiceUpdateURL string
	//“心跳”检测服务是否正常的URL
	HeartbeatURL string
}

// ServiceName 注册的服务名称
type ServiceName string

//注册的具体服务名称
const (
	LoggerService = ServiceName("LoggerService")
	GradeService  = ServiceName("GradeService")
	PortalService = ServiceName("Portal")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}
type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
