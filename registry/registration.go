package registry

type Registration struct {
	ServiceName     ServiceName
	ServiceURL      string
	RequiredService []ServiceName
	//如：当registry中有grade所依赖的logger服务时，由此返回
	ServiceUpdateURL string
}
type ServiceName string

//注册服务的名称
const (
	LoggerService = ServiceName("LoggerService")
	GradeService  = ServiceName("GradeService")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}
type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
