package http

// HTTPFrameworkType HTTP引擎类型枚举
type HTTPFrameworkType string

const (
	GinFramework   HTTPFrameworkType = "gin"
	IrisFramework  HTTPFrameworkType = "iris"
	EchoFramework  HTTPFrameworkType = "echo"
	FiberFramework HTTPFrameworkType = "fiber"
	ChiFramework   HTTPFrameworkType = "chi"
	HertzFramework HTTPFrameworkType = "hertz"
)

// HTTPFactory HTTP框架抽象工厂
type HTTPFactory interface {
	CreateHTTPServer(config HTTPConfig) HTTPProtocol
}

// GetHTTPFramework 根据框架类型获取对应工厂方法
func GetHTTPFramework(framework HTTPFrameworkType) HTTPFactory {
	switch framework {
	case GinFramework:
		return &ginHTTPFactory{}
	case IrisFramework:
		return &irisHTTPFactory{}
	case EchoFramework:
		return &echoHTTPFactory{}
	case FiberFramework:
		return &fiberHTTPFactory{}
	case ChiFramework:
		return &chiHTTPFactory{}
	case HertzFramework:
		return &hertzHTTPFactory{}
	default:
		return &ginHTTPFactory{} // 默认使用Gin引擎
	}
}

// HTTPConfig 服务器配置结构
type HTTPConfig struct {
	Framework  HTTPFrameworkType `mapstructure:"framework"`
	Port       int               `mapstructure:"port"`
	Host       string            `mapstructure:"host"`
	Workers    int               `mapstructure:"workers"`
	MultiNodes []string          `mapstructure:"multi_nodes"` // 集群节点
}

// ginHTTPFactory Gin框架工厂实现
type ginHTTPFactory struct{}

func (g *ginHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewGinAdapter(config)
}

// irisHTTPFactory Iris框架工厂实现
type irisHTTPFactory struct{}

func (i *irisHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewIrisAdapter(config)
}

// echoHTTPFactory Echo框架工厂实现
type echoHTTPFactory struct{}

func (e *echoHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewEchoAdapter(config)
}

// fiberHTTPFactory Fiber框架工厂实现
type fiberHTTPFactory struct{}

func (f *fiberHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewFiberAdapter(config)
}

// chiHTTPFactory Chi框架工厂实现
type chiHTTPFactory struct{}

func (c *chiHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewChiAdapter(config)
}

// hertzHTTPFactory Hertz框架工厂实现
type hertzHTTPFactory struct{}

func (h *hertzHTTPFactory) CreateHTTPServer(config HTTPConfig) HTTPProtocol {
	return NewHertzAdapter(config)
}
