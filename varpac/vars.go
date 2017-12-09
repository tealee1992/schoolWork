package varpac

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

var Concurrency = 0
var sem1 = make(chan struct{}, 1)
var sem2 = make(chan struct{}, 1)
var Master = host{
	IP: "11.0.0.172",
}
var Cluster = []host{
	host{
		IP:       "11.0.0.171",
		TotalMem: 32,
		PortNum: 0,
	},
	host{
		IP:       "11.0.0.172",
		TotalMem: 32,
		PortNum: 0,
	},
	host{
		IP:       "11.0.0.176",
		TotalMem: 4,
		PortNum: 0,
	},
}
var Section [3]float64
var (
	TEMPLATE_DIR = "./views"
	AgentPort    = "9902"
	DispatPort   = "9903"
	Option       = types.ContainerCreateConfig{
		Config: &container.Config{
			AttachStdin: true,
			Tty:         true,
			Image:       "os-zh-cn-bochs2.4-2",
			Volumes: map[string]struct{}{
				"/tempfiles": {},
			},
		},
		HostConfig: &container.HostConfig{
			Binds:     []string{"/home/docker/GoWorkspace/tempfiles/:/tempfiles"},
			Resources: container.Resources{
			//CPUShares:1,
			//Memory:0,//Memory:314572800,//300M内存
			},
			PortBindings: {
				"6080/tcp": [{
					//HostIp: "",
					"HostPort": "9500"//端口范围是9500~9999
				}]
			},
		},
	}
)

//保存实验image的标志
var Title = "labimage"

//novnc密码
var Password = "novnc"

type host struct {
	IP       string
	TotalMem float64
	PortNum int64
}
type Prob struct { //概率法计算数据
	Host        host
	Memload     float64
	Probability float64 //内存在集群中的占比
}

func FastVol() {
	select {
	case sem1 <- struct{}{}:
	default:
		Concurrency = (Concurrency + 1) % 3
	}
}

func AccurateVol() {
	select {
	case sem2 <- struct{}{}:
	default:
		Concurrency = (Concurrency + 1) % 3
	}
}

func TypeBased() {
	select {
	case <-sem1:
		<-sem2
	default:
		Concurrency = (Concurrency + 1) % 3
	}

}
