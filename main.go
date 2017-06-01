package main

import (

	//"strings"
	//"net/http"
	"encoding/json"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	//"github.com/docker/engine-api/types/events"
	"golang.org/x/net/context"
	//"io"
	//"syscall"
	//"fmt"
	"fmt"
	"github.com/docker/engine-api/types/container"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	//"github.com/docker/go-connections/nat"
	"math/rand"
	"net"
	"time"
	"varpac"
	//"netlimit"
	"monitorloop"
	"timestamp"

	"golang.org/x/net/netutil"
	"math"
	"strings"
)

//container config
var options1 = types.ContainerCreateConfig{
	Config: &container.Config{
		AttachStdin: true,
		Tty:         true,
		Image:       "stress-bigger1",
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
	},
}
var options2 = types.ContainerCreateConfig{
	Config: &container.Config{
		AttachStdin: true,
		Tty:         true,
		Image:       "stress-bigger2",
		Volumes: map[string]struct{}{
			"/tempfiles": {},
		},
	},
	HostConfig: &container.HostConfig{
		Binds: []string{"/home/docker/GoWorkspace/tempfiles/:/tempfiles"},
		Resources: container.Resources{
			//CPUShares:1,
			Memory: 2147484000, //Memory:157286400314572800,//150M内存
		},
	},
}

/*options := types.ContainerCreateConfig{
	Config : &container.Config{
		AttachStdin:true,
		Tty	:true,
		Image:"os-zh-cn-bochs2.4-2",
		Volumes:map[string]struct{}{
			"/tempfiles":{},
		},
	},
	HostConfig:&container.HostConfig{
		Binds:[]string{"/home/docker/GoWorkspace/tempfiles/"},
		PortBindings:nat.PortMap{
			nat.Port("6080/tcp") : []nat.PortBinding{
				{
					HostIP:"0.0.0.0",
					HostPort:"",
				},
			},
		},
		Resources:container.Resources{
			//CPUShares:1,
			Memory:524288000,
		},
	},
}*/
type host struct {
	totalMem    float64
	memload     float64
	probability float64
}
type probability struct {
	host1 host
	host2 host
	host3 host
}

var pro probability

func main() {

	//go monitorLoop()
	/*
		l, err := net.Listen("tcp", "127.0.0.1:10000")
		if err != nil {
		    log.Fatal("Listen: %v", err)
		}
		defer l.Close()
		l = netutil.LimitListener(l, 12)

		http.Serve(l, http.HandlerFunc(bindMaine))
	/**/
	//go loopfunc()
	pro.updateProb(options2)
	l, err := net.Listen("tcp", "11.0.57.2:9090")
	//l,err := net.Listen("tcp","127.0.0.1:9090")
	if err != nil {
		log.Fatal("Listen: %v", err)
	}
	defer l.Close()

	criticalPoint := 15
	//l = netlimit.LimitListener(l,criticalPoint)
	l = netutil.LimitListener(l, criticalPoint)

	handler := http.HandlerFunc(bindMachine)
	http.Handle("/containers/create", handler)
	http.Serve(l, handler)

	/*http.HandleFunc("/containers/create",bindMachine)
	err1 := http.ListenAndServe(":9090",nil)
	if err1 != nil{
		log.Fatal("ListenAndServe:",err)
	}*/

}
func loopfunc() {
	spec := "*/15 * * * * ?"
	var f = func() {
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err := client.NewClient("http://11.0.57.2:3375", "v1.23", nil, defaultHeaders)
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
			dec := json.NewDecoder(body)
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

//更新host的probability
func (p *probability) updateProb(options types.ContainerCreateConfig) {
	var mempercent1, mempercent2, mempercent3 float64

	//获取容器内存分配
	conmemory := options.HostConfig.Resources.Memory
	conMemory := float64(conmemory)
	//获取主机内存使用率
	resp, err := http.Get("http://11.0.57.1:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent1, err = strconv.ParseFloat(string(body), 32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	resp, err = http.Get("http://11.0.57.2:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent2, err = strconv.ParseFloat(string(body), 32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	resp, err = http.Get("http://11.0.57.3:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent3, err = strconv.ParseFloat(string(body), 32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	p.host1.memload = mempercent1
	p.host2.memload = mempercent2
	p.host3.memload = mempercent3
	//获取主机总内存
	p.host1.totalMem = 32
	p.host2.totalMem = 32
	p.host3.totalMem = 4
	totalMem := p.host1.totalMem + p.host2.totalMem
	//计算各主机分配概率
	//
	p.host1.probability = p.host1.totalMem / totalMem
	p.host2.probability = p.host2.totalMem / totalMem
	p.host3.probability = p.host3.totalMem / totalMem

	preMemper1 := (p.host1.memload*p.host1.totalMem*1024*1024*1024 + conMemory) / p.host1.totalMem * 1024 * 1024 * 1024
	if preMemper1 > 0.8 {
		p.host1.probability = p.host1.probability - 0.05
		p.host2.probability = p.host2.probability + 0.05
	}
	preMemper2 := (p.host2.memload*p.host2.totalMem*1024*1024*1024 + conMemory) / p.host2.totalMem * 1024 * 1024 * 1024
	if preMemper2 > 0.8 {
		p.host2.probability = p.host2.probability - 0.05
		p.host3.probability = p.host3.probability + 0.05
	}
	preMemper3 := (p.host3.memload*p.host3.totalMem*1024*1024*1024 + conMemory) / p.host3.totalMem * 1024 * 1024 * 1024
	if preMemper3 > 0.6 {
		if p.host3.probability > 0 {
			p.host3.probability = 0
			p.host1.probability = p.host1.probability + 0.05
		}

	}

}
func bindMachine(w http.ResponseWriter, r *http.Request) {

	m := r.FormValue("method")
	i := r.FormValue("index")
	d := r.FormValue("demand")
	method, _ := strconv.Atoi(m)
	index, _ := strconv.Atoi(i)
	demand, _ := strconv.Atoi(d)
	fmt.Println("index=", index)
	fmt.Println("demand=", demand)
	if demand == 0 {
		fmt.Println("true")
	}
	fmt.Println("method=", method)

	run(method, index, demand)

	//fmt.Println("method=",method);

	//fmt.Println(varpac.Concurrency);
	//sleepTest()
	//onePlusGap(options,defaultHeaders,num)
	//...storeSession(userid,response,hostport) 更新etcd存储
}
func sleepTest() {
	time.Sleep(time.Duration(5 * time.Second))
}
func run(method int, index int, demand int) {
	if method == 1 {
		//timestamp.Timestamp("swarm",i)
		fmt.Println("method=", method)
		swarmSpread(index, demand)
	} else if method == 2 { //论文中的策略
		//timestamp.Timestamp("weight",i)
		weightSchedule(index, demand)
	} else if method == 3 { //我的方法
		//timestamp.Timestamp("combine",i)
		combine(index, demand)
	}
}

//顺序请求

//goroutine模拟并发请求

//顺序加间隔 请求

func combine(i int, demand int) {

	if varpac.Concurrency == 0 {
		pro.updateProb(options2)
		timestamp.Simulog(i)
		fast(i, pro, demand)
	} else if varpac.Concurrency == 1 {
		timestamp.Simulog(i)
		accurate(i, demand)
	} else if varpac.Concurrency == 2 {
		timestamp.Simulog(i)
		typeBsed(i, demand)
	}
}
func fast(i int, proold probability, cpudemand int) {
	//pro = make(probability,1)
	pro := proold

	var section [3]float64
	//分配扇区
	section[0] = pro.host1.probability
	section[1] = section[0] + pro.host2.probability
	section[2] = section[1] + pro.host3.probability
	seed := rand.NewSource(time.Now().Unix() + int64(i))
	newrand := rand.New(seed)
	randnum := newrand.Float64()
	var option types.ContainerCreateConfig

	if cpudemand == 0 {
		option = options1
	} else {
		option = options2
	}
	if randnum < section[0] {
		createContainer("1", option)
	} else if randnum < section[1] {
		createContainer("2", option)
	} else {
		createContainer("3", option)
	}
	varpac.FastVol()
	timestamp.Timestamp("fast", i)
	timestamp.Probalog(section, randnum)

}
func accurate(i int, cpudemand int) {
	var host1, host2, host3 float64

	resp1, err := http.Get("http://11.0.57.1:9095/bindMachine")
	if err != nil {

	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {

	}
	host1, _ = strconv.ParseFloat(string(body), 32)
	resp2, err := http.Get("http://11.0.57.2:9095/bindMachine")
	if err != nil {

	}
	defer resp2.Body.Close()
	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {

	}
	host2, _ = strconv.ParseFloat(string(body), 32)
	resp3, err := http.Get("http://11.0.57.3:9095/bindMachine")
	if err != nil {

	}
	defer resp3.Body.Close()
	body, err = ioutil.ReadAll(resp3.Body)
	if err != nil {

	}
	host3, _ = strconv.ParseFloat(string(body), 32)
	//记录权值
	timestamp.Weightlog(host1, host2, host3)
	var option types.ContainerCreateConfig

	if cpudemand == 0 {
		option = options1
	} else {
		option = options2
	}

	if host1 <= host2 {
		if host1 <= host3 {
			//选择主机1
			createContainer("1", option)
		} else {
			//选择主机3
			createContainer("3", option)
		}
	} else {
		if host2 <= host3 {
			//选择主机2
			createContainer("2", option)
		} else {
			//选择主机3
			createContainer("3", option)
		}
	}
	varpac.AccurateVol()
	timestamp.Timestamp("accurate", i)
}

func typeBsed(i int, cpudemand int) {

	var pace1 float64
	var pace2 float64
	var pace3 float64
	var ip1 = "11.0.57.1"
	var ip2 = "11.0.57.2"
	var ip3 = "11.0.57.3"

	//获取容器内存分配
	pace1 = getpace(ip1, cpudemand)
	pace2 = getpace(ip2, cpudemand)
	pace3 = getpace(ip3, cpudemand)
	var option types.ContainerCreateConfig

	if cpudemand == 0 {
		option = options1
	} else {
		option = options2
	}

	fmt.Println(pace1)
	fmt.Println(pace2)
	fmt.Println(pace3)
	if pace1 > pace2 {
		if pace1 > pace3 {
			createContainer("1", option)
			fmt.Println("type1")
		} else {
			createContainer("3", option)
			fmt.Println("type3")
		}
	} else {
		if pace2 > pace3 {
			createContainer("2", option)
			fmt.Println("type2")
		} else {
			createContainer("3", option)
			fmt.Println("type3")
		}
	}

	varpac.TypeBased()
	timestamp.Timestamp("type", i)

}
func getpace(ip string, cpudemand int) float64 {
	var cpuratio, memratio, diff_before, diff_after, pace float64
	//conmemory := options.HostConfig.Resources.Memory
	var conMemory, conCpu float64
	memratio = getMemload(ip)
	cpuratio = getCpuload(ip)
	if memratio >= 0.8 || cpuratio >= 0.8 {
		return -1.0
	}

	if cpudemand == 0 {
		conMemory = float64(0.0)
		conCpu = 0.0625
	} else {
		conMemory = float64(options2.HostConfig.Resources.Memory)
		conCpu = 0.0
	}
	diff_before = math.Abs(memratio - cpuratio)

	var totalmem float64
	if strings.EqualFold(ip, "11.0.57.1") {
		totalmem = pro.host1.totalMem
	} else if strings.EqualFold(ip, "11.0.57.2") {
		totalmem = pro.host2.totalMem
	} else {
		totalmem = pro.host3.totalMem
	}
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
func getCpuload(ip string) float64 {
	var result float64
	resp, err := http.Get("http://" + ip + ":9095/cpuload")
	if err != nil {
		fmt.Println("Get cpuload failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	result, err = strconv.ParseFloat(string(body), 32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	return result
}
func getMemload(ip string) float64 {
	var result float64
	resp, err := http.Get("http://" + ip + ":9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	result, err = strconv.ParseFloat(string(body), 32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	return result
}

func swarmSpread(i int, cpudemand int) {
	var option types.ContainerCreateConfig

	if cpudemand == 0 {
		option = options1
	} else {
		option = options2
	}
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://11.0.57.2:3375", "v1.23", nil, defaultHeaders)
	if err != nil {
		log.Print(err)
	}

	response, err := cli.ContainerCreate(context.Background(), option.Config, option.HostConfig, option.NetworkingConfig, option.Name)
	if err != nil {
		fmt.Print("fail to create this container")
		fmt.Println(err)
	}
	fmt.Println(response)
	err = cli.ContainerStart(context.Background(), response.ID, types.ContainerStartOptions{})
	timestamp.Timestamp("swarm", i)
	if err != nil {
		fmt.Print("fail to start this container")
		fmt.Println(err)
	}
}

func weightSchedule(i int, cpudemand int) {

	var host1, host2, host3 float64

	resp1, err := http.Get("http://11.0.57.1:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		fmt.Println(err)
	}
	host1, _ = strconv.ParseFloat(string(body), 32)
	resp2, err := http.Get("http://11.0.57.2:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp2.Body.Close()
	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {
		fmt.Println(err)
	}
	host2, _ = strconv.ParseFloat(string(body), 32)
	resp3, err := http.Get("http://11.0.57.3:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp3.Body.Close()
	body, err = ioutil.ReadAll(resp3.Body)
	if err != nil {
		fmt.Println(err)
	}
	host3, _ = strconv.ParseFloat(string(body), 32)
	//记录权值
	timestamp.Weightlog(host1, host2, host3)
	var option types.ContainerCreateConfig

	if cpudemand == 0 {
		option = options1
	} else {
		option = options2
	}
	//if host1 <= host2 {
	//	createContainer("1",option)
	//}else {
	//	createContainer("2",option)
	//}
	if host1 <= host2 {
		if host1 <= host3 {
			//选择主机1
			createContainer("1", option)
		} else {
			//选择主机3
			createContainer("3", option)
		}
	} else {
		if host2 <= host3 {
			//选择主机2
			createContainer("2", option)
		} else {
			//选择主机3
			createContainer("3", option)
		}
	}
	timestamp.Timestamp("weight", i)

}
func createContainer(host string, newoption ...types.ContainerCreateConfig) {
	var option types.ContainerCreateConfig

	if newoption != nil {
		option = newoption[0]
	} else {
		option = options1
	}
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://11.0.57."+host+":2375", "v1.23", nil, defaultHeaders)
	if err != nil {
		log.Print(err)
	}
	response, err := cli.ContainerCreate(context.Background(), option.Config, option.HostConfig, option.NetworkingConfig, option.Name)
	if err != nil {
		fmt.Println("fail to create this container")
		fmt.Println(err)
	}
	fmt.Println(response)
	err = cli.ContainerStart(context.Background(), response.ID, types.ContainerStartOptions{})
	if err != nil {
		fmt.Println("fail to start this container")
		fmt.Println(err)
	}
}
