package main

/*调度器
to do : 修改hard code：主机信息，算法中的排序。
分离调度算法?
*/

import (
	"encoding/json"
	"etcd"
	"fmt"
	"functions"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"golang.org/x/net/netutil"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"monitorloop"
	"net"
	"net/http"
	"os"
	"strconv"
	//"strings"
	"time"
	"varpac"
)

// type host struct {
// 	totalMem    float64
// 	memload     float64
// 	probability float64
// }
// type probability struct {
// 	host1 host
// 	host2 host
// 	host3 host
// }

// var pro probability
var userid string
var logFile *os.File
var loger *log.Logger

func init() {
	//设置日志
	setLog()
	//初始化集群信息
	var sum float64
	for _, host := range varpac.Cluster {
		sum += host.TotalMem
	}
	for index, host := range varpac.Cluster {
		varpac.Section[index] = host.TotalMem / sum
		if index > 0 {
			varpac.Section[index] = varpac.Section[index] + varpac.Section[index-1]
		}
	}
}
func main() {
	//循环检测docker容器状态
	//go loopfunc()

	listener, err := net.Listen("tcp", varpac.Master.IP+":"+varpac.DispatPort)
	if err != nil {
		log.Fatal("Listen: %v", err)
	}
	defer listener.Close()

	criticalPoint := 15
	listener = netutil.LimitListener(listener, criticalPoint)

	dispatch := http.HandlerFunc(dispatch)
	http.Handle("/dispatch", dispatch)
	http.Serve(listener, dispatch)

}

func setLog() {
	logFile, err := os.Create("./dispatcher/logs.txt")
	if err != nil {
		fmt.Println(err)
	}
	loger = log.New(logFile, "cloudlab_dispatcher_server_", log.Ldate|log.Ltime|log.Lshortfile)
}

func loopfunc() {
	spec := "* */15 * * * *"
	var f = func() {
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err := client.NewClient("http://"+varpac.Master.IP+":3375", "v1.23", nil, defaultHeaders)
		if err != nil {
			log.Print(err)
		}
		options := types.ContainerListOptions{All: true}
		containers, err := cli.ContainerList(context.Background(), options)
		if err != nil {
			log.Print(err)
		}
		i := 0
		for _, c := range containers {
			//fmt.Println(c.ID)
			//var CPUusage types.CPUUsage
			//var CPUstats types.CPUStats
			var Stats types.Stats
			body, err := cli.ContainerStats(context.Background(), c.ID, false)
			if err != nil {
				log.Fatal(err)
			}
			dec := json.NewDecoder(body.Body)
			err = dec.Decode(&Stats)
			if err != nil {
				break
			}
			i++
			log.Println(i)
			log.Println(c.Names)
			log.Println(Stats.CPUStats.SystemUsage)
		}
	}
	monitorloop.Execute(spec, f)
}

func dispatch(w http.ResponseWriter, r *http.Request) {

	userid = r.FormValue("userid")
	// defer r.Body.Close()
	// data, _ := ioutil.ReadAll(r.Body)
	// var user map[string]interface{}
	// json.Unmarshal(data, &user)
	// userid := user["userid"]
	loger.Println(userid)
	if varpac.Concurrency == 0 {
		loger.Println("fast")
		fast(w)
	} else if varpac.Concurrency == 1 {
		loger.Println("accurate")
		accurate(w)
	} else if varpac.Concurrency == 2 {

		//typeBsed(w)
	}
}

func fast(w http.ResponseWriter) {
	loger.Println("in fast")
	section := varpac.Section
	seed := rand.NewSource(time.Now().Unix())
	newrand := rand.New(seed)
	randnum := newrand.Float64()

	var hostip string
	var portnum int64
	for index, sec := range section {
		if randnum < sec {
			hostip = varpac.Cluster[index].IP
			portnum = varpac.Cluster[index].PortNum
			varpac.Cluster[index].PortNum = (portnum + 1) % 500
			break
		}
	}

	var conid string
	loger.Println("before create")
	hostip = "11.0.0.172" //for test
	conid = createContainer(hostip, portnum)
	url := urlgenerater(hostip, portnum)
	// set lab session
	labSession := etcd.Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
		Url:    url,
	}
	loger.Println("before set session,userid:" + userid)
	labSession.Set(userid)
	varpac.FastVol()
	loger.Println("before write resp")
	loger.Println(url)
	w.Write([]byte(url))

}

//生成前端对应的容器访问url,因为需要做一个从前端路由器到服务器的端口映射
//比如 11.0.0.172:5901对应前端url是100.64.16.72:25901
func urlgenerater(ip string, portnum int64) (url string) {
	hostnum := varpac.IpMap[ip]
	port := 9500 + portnum
	port_str := strconv.Itoa(int(port))

	url = varpac.OutIp + ":" + hostnum + port_str
	return
}
func accurate(w http.ResponseWriter) {
	var hostloadMin float64
	var hostip string
	var portnum int64
	hostloadMin = 0
	for index, host := range varpac.Cluster {

		resp, err := http.Get("http://" + host.IP + ":" + varpac.AgentPort + "/bindMachine")
		if err != nil {
			log.Fatal("error in Agent")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("error in jiexi")
		}
		hostload, _ := strconv.ParseFloat(string(body), 32)
		if hostload < hostloadMin {
			hostloadMin = hostload
			hostip = host.IP
			portnum = host.PortNum
			varpac.Cluster[index].PortNum = (portnum + 1) % 500 //修改端口差值
		}
	}

	var conid string
	conid = createContainer(hostip, portnum)
	url := urlgenerater(hostip, portnum)
	// set lab session
	labSession := etcd.Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
	}
	labSession.Set(userid)
	varpac.AccurateVol()
	w.Write([]byte(url))

}

func typeBsed(w http.ResponseWriter, cpudemand int) {

	var maxpace float64
	var hostip string
	var portnum int64

	// var option types.ContainerCreateConfig
	maxpace = -10
	for index, host := range varpac.Cluster {
		ip := host.IP
		totalmem := host.TotalMem
		pace := getpace(ip, cpudemand, totalmem)
		if pace > maxpace {
			maxpace = pace
			hostip = ip
			portnum = host.PortNum
			varpac.Cluster[index].PortNum = (portnum + 1) % 500 //修改端口差值
		}
	}
	var conid string
	conid = createContainer(hostip, portnum)
	url := urlgenerater(hostip, portnum)
	// set lab session
	labSession := etcd.Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
	}
	labSession.Set(userid)
	varpac.AccurateVol()
	w.Write([]byte(url))
}

func getpace(ip string, cpudemand int, totalmem float64) float64 {
	var cpuratio, memratio, diff_before, diff_after, pace float64
	//conmemory := options.HostConfig.Resources.Memory
	var conMemory, conCpu float64
	memratio = functions.GetMemload(ip)
	cpuratio = functions.GetCpuload(ip)
	if memratio >= 0.8 || cpuratio >= 0.8 {
		return -1.0
	}
	//这里的逻辑仅仅是用来跑实验的，真正的算法需要获取容器真实的请求选项，
	//可以是全局变量。
	if cpudemand == 0 {
		conMemory = float64(0.0)
		conCpu = 0.0625
	} else {

		conMemory = float64(varpac.Option.HostConfig.Resources.Memory)
		conCpu = 0.0
	}
	diff_before = math.Abs(memratio - cpuratio)

	memratio_after := (memratio*totalmem*1024*1024*1024 + conMemory) / (totalmem * 1024 * 1024 * 1024)
	cpuratio_after := cpuratio + conCpu
	diff_after = math.Abs(float64(memratio_after) - float64(cpuratio_after))
	pace = diff_before - diff_after
	//fmt.Println("memratio",memratio)
	//fmt.Println("cpuratio",cpuratio)
	//fmt.Println("diff_before",diff_before)
	//
	//fmt.Println("memratio_after",memratio_after)
	//fmt.Println("cpuratio_after",cpuratio_after)
	//fmt.Println("diff_after",diff_after)
	result := pace + diff_before
	return result
}

//used by getpace

//used by getpace

func createContainer(hostip string, portnum int64) string {
	var option types.ContainerCreateConfig
	option = varpac.Option
	port := 9500 + portnum
	port_str := strconv.Itoa(int(port))
	option.HostConfig.PortBindings["6080/tcp"][0].HostPort = port_str
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://"+hostip+":2375", "v1.23", nil, defaultHeaders)
	if err != nil {
		log.Print(err)
	}
	response, err := cli.ContainerCreate(context.Background(), option.Config, option.HostConfig, option.NetworkingConfig, option.Name)
	if err != nil {
		log.Print("fail to create this container")
		log.Print(err)
	}

	err = cli.ContainerStart(context.Background(), response.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Print("fail to start this container")
		log.Print(err)
	}
	return response.ID
}
