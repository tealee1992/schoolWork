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
	"log"
	"net/http"
	"io/ioutil"
	"strconv"
	"fmt"
	"github.com/docker/engine-api/types/container"
	//"github.com/docker/go-connections/nat"
	"varpac"
	"time"
	"math/rand"
	"net"
	"netlimit"
	"timestamp"
	"monitorloop"

)
//container config
var options = types.ContainerCreateConfig{
	Config : &container.Config{
		AttachStdin:true,
		Tty	:true,
		Image:"stress-bigger",
		Volumes:map[string]struct{}{
			"/tempfiles":{},
		},
	},
	HostConfig:&container.HostConfig{
		Binds:[]string{"/home/docker/GoWorkspace/tempfiles/:/tempfiles"},
		Resources:container.Resources{
			//CPUShares:1,
			Memory:314572800,//Memory:314572800,//300M内存
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
	totalMem float64
	memload float64
	probability float64
}
type probability struct {
	host1 host
	host2 host
	host3 host
}

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
	l,err := net.Listen("tcp","11.0.57.2:9090")
	//l,err := net.Listen("tcp","127.0.0.1:9090")
	if err != nil {
		log.Fatal("Listen: %v",err)
	}
	defer l.Close()

	criticalPoint := 10
	l = netlimit.LimitListener(l,criticalPoint)


	handler := http.HandlerFunc(bindMachine)
	http.Handle("/containers/create",handler)
	http.Serve(l,handler)

	/*http.HandleFunc("/containers/create",bindMachine)
	err1 := http.ListenAndServe(":9090",nil)
	if err1 != nil{
		log.Fatal("ListenAndServe:",err)
	}*/


}
func loopfunc(){
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
		i := 0;
		for _, c := range containers {
			//fmt.Println(c.ID)
			//var CPUusage types.CPUUsage
			//var CPUstats types.CPUStats
			var Stats  types.Stats
			body,err := cli.ContainerStats(context.Background(),c.ID,false)
			if err!= nil {
				log.Fatal(err)
			}
			dec := json.NewDecoder(body)
			err = dec.Decode(&Stats)
			if err != nil{
				break
			}
			i++
			log.Println(i)
			log.Println(c.Names)
			log.Println(Stats.CPUStats.SystemUsage)
		}
	}
	monitorloop.Execute(spec,f)
}
//更新host的probability
func (p *probability)updateProb(options types.ContainerCreateConfig){
	var mempercent1 ,mempercent2 ,mempercent3 float64
	//var preMemper1,preMemper2,preMemper3 float64
	//获取容器内存分配
	conmemory := options.HostConfig.Resources.Memory
	conMemory := float64(conmemory)
	//获取主机内存使用率
	resp,err :=http.Get("http://11.0.57.1:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent1,err = strconv.ParseFloat(string(body),32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	resp,err =http.Get("http://11.0.57.2:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent2,err = strconv.ParseFloat(string(body),32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	resp,err =http.Get("http://11.0.57.3:9095/memload")
	if err != nil {
		fmt.Println("Get memload failed")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil read failed")
	}
	mempercent3,err = strconv.ParseFloat(string(body),32)
	if err != nil {
		fmt.Println("ParseFloat failed")
	}
	p.host1.memload=mempercent1
	p.host2.memload=mempercent2
	p.host3.memload=mempercent3
	//获取主机总内存
	p.host1.totalMem=32
	p.host2.totalMem=32
	p.host3.totalMem=4
	totalMem := p.host1.totalMem+p.host2.totalMem+p.host3.totalMem
	//计算各主机分配概率
	//
	p.host1.probability = p.host1.totalMem/totalMem
	p.host2.probability = p.host2.totalMem/totalMem
	p.host3.probability = p.host3.totalMem/totalMem
	//
	preMemper1 := (p.host1.memload*p.host1.totalMem*1024*1024*1024+60*p.host1.probability*conMemory)/p.host1.totalMem*1024*1024*1024
	if preMemper1 > 0.8 {
		p.host1.probability=p.host1.probability-0.05
		p.host2.probability=p.host2.probability+0.05
	}
	preMemper2 := (p.host2.memload*p.host2.totalMem*1024*1024*1024+60*p.host2.probability*conMemory)/p.host2.totalMem*1024*1024*1024
	if preMemper2 > 0.8 {
		p.host2.probability=p.host2.probability-0.05
		p.host3.probability=p.host3.probability+0.05
	}
	preMemper3 := (p.host3.memload*p.host3.totalMem*1024*1024*1024+60*p.host3.probability*conMemory)/p.host3.totalMem*1024*1024*1024
	if preMemper3 > 0.8 {
		p.host3.probability=p.host3.probability-0.05
		p.host1.probability=p.host1.probability+0.05
	}

}
func bindMachine(w http.ResponseWriter, r *http.Request){

	m :=r.FormValue("method")
	i :=r.FormValue("index")
	method,_ := strconv.Atoi(m)
	index,_ := strconv.Atoi(i)
	fmt.Println("index=",index);
	run(method,index)

	//fmt.Println("method=",method);

	//fmt.Println(varpac.Concurrency);
	//sleepTest()
	//onePlusGap(options,defaultHeaders,num)
	//...storeSession(userid,response,hostport) 更新etcd存储
}
func  sleepTest(flag bool)  {
	if(flag){
		time.Sleep(time.Duration(1*time.Second))
	}else{
		time.Sleep(time.Duration(5*time.Second))
	}

}
func run(method int,index int){
	if method == 1 {
		//timestamp.Timestamp("swarm",i)
		fmt.Println("method=",method);
		swarmSpread(index)
	}else if method == 2 {   //论文中的策略
		//timestamp.Timestamp("weight",i)
		weightSchedule(index)
	}else if method == 3 {    //我的方法
		//timestamp.Timestamp("combine",i)
		combine(index)
	}
}

//顺序请求
func onebyonecreate(num int,method int) {
	for i:=0 ;i<num;i++ {
		//swarm的Spread策略
		if method == 1 {
			//timestamp.Timestamp("swarm",i)
			swarmSpread(i)
		}else if method == 2 {   //论文中的策略
			//timestamp.Timestamp("weight",i)
			weightSchedule(i)
		}else if method == 3 {    //我的方法
			//timestamp.Timestamp("combine",i)
			combine(i)
		}
	}
}
//goroutine模拟并发请求
func goroutinecreate(num int,method int) {
	//chs := make([]chan int,num)
	var pro probability
	pro.updateProb(options)
	chs := make(chan int)
	for i:=0 ;i<num;i++ {
		//swarm的Spread策略
		if method == 1 {
			//timestamp.Timestamp("swarm",i)
			go swarmSpread(i)
		}else if method == 2 {   //论文中的策略
			//timestamp.Timestamp("weight",i)
			go weightSchedule(i)
		}else if method == 3 {    //我的方法
			//timestamp.Timestamp("combine",i)
			go combine(i)
		}
	}
	time.Sleep(time.Duration(5)*time.Minute)
	chs<-1
	<-chs

}
//顺序加间隔 请求
func onePlusGap(num int,method int){
	for i:=0 ;i<num;i++ {
		//swarm的Spread策略
		if method == 1 {
			timestamp.Timestamp("swarm",i)
			swarmSpread(i)
		}
		//论文中的策略
		if method == 2 {
			timestamp.Timestamp("weight",i)
			weightSchedule(i)
		}
		//我的方法
		if method == 3 {
			timestamp.Timestamp("combine",i)
			combine(i)
		}
		time.Sleep(time.Duration(10)*time.Second)//各请求之间相差10秒
	}
}
func combine(i int){


	if varpac.Concurrency {
		var pro probability
		pro.updateProb(options)
		timestamp.Simulog(i);
		sleepTest(true)
		timestamp.Timestamp("fast",i)
		//fast(i,pro)
	}else {
		timestamp.Simulog(i);
		sleepTest(false)
		timestamp.Timestamp("combine",i)
		//accurate(i)
	}
}
func fast(i int,proold probability){
	//pro = make(probability,1)
	pro := proold

	var section [3]float64
	//分配扇区
	section[0]=pro.host1.probability
	section[1]=section[0]+pro.host2.probability
	section[2]=section[1]+pro.host3.probability
	seed := rand.NewSource(time.Now().Unix()+int64(i))
	newrand := rand.New(seed)
	randnum := newrand.Float64()
	if randnum<section[0] {
		createContainer("1")
	}else if randnum<section[1] {
		createContainer("2")
	}else {
		createContainer("3")
	}
	timestamp.Timestamp("fast",i)
	timestamp.Probalog(section,randnum)
}
func accurate(i int) {
	var host1,host2,host3 float64

	resp1,err :=http.Get("http://11.0.57.1:9095/bindMachine")
	if err != nil {

	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {

	}
	host1,_ = strconv.ParseFloat(string(body),32)
	resp2,err :=http.Get("http://11.0.57.2:9095/bindMachine")
	if err != nil {

	}
	defer resp2.Body.Close()
	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {

	}
	host2,_ = strconv.ParseFloat(string(body),32)
	resp3,err :=http.Get("http://11.0.57.3:9095/bindMachine")
	if err != nil {

	}
	defer resp3.Body.Close()
	body, err = ioutil.ReadAll(resp3.Body)
	if err != nil {

	}
	host3,_ = strconv.ParseFloat(string(body),32)
	//记录权值
	timestamp.Weightlog(host1,host2,host3)

	if host1 <= host2 {
		if host1 <= host3 {
			//选择主机1
			createContainer("1")
		}else {
			//选择主机3
			createContainer("3")
		}
	}else {
		if host2 <= host3 {
			//选择主机2
			createContainer("2")
		}else {
			//选择主机3
			createContainer("3")
		}
	}
	timestamp.Timestamp("combine",i)
}

func swarmSpread(i int) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://11.0.57.2:3375", "v1.23", nil, defaultHeaders)
	if err != nil {
		log.Print(err)
	}

	response, err :=cli.ContainerCreate(context.Background(),options.Config,options.HostConfig,options.NetworkingConfig,options.Name)
	if err != nil {
		fmt.Print("fail to create this container")
		fmt.Println(err)
	}
	fmt.Println(response)
	err = cli.ContainerStart(context.Background(),response.ID,types.ContainerStartOptions{})
	timestamp.Timestamp("swarm",i)
	if err != nil {
		fmt.Print("fail to start this container")
		fmt.Println(err)
	}
}

func weightSchedule(i int) {


	var host1,host2,host3 float64

	resp1,err :=http.Get("http://11.0.57.1:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		fmt.Println(err)
	}
	host1,_ = strconv.ParseFloat(string(body),32)
	resp2,err :=http.Get("http://11.0.57.2:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp2.Body.Close()
	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {
		fmt.Println(err)
	}
	host2,_ = strconv.ParseFloat(string(body),32)
	resp3,err :=http.Get("http://11.0.57.3:9095/weightedSchedule")
	if err != nil {
		fmt.Println(err)
	}
	defer resp3.Body.Close()
	body, err = ioutil.ReadAll(resp3.Body)
	if err != nil {
		fmt.Println(err)
	}
	host3,_ = strconv.ParseFloat(string(body),32)
	//记录权值
	timestamp.Weightlog(host1,host2,host3)

	if host1 <= host2 {
		if host1 <= host3 {
			//选择主机1
			createContainer("1")
		}else {
			//选择主机3
			createContainer("3")
		}
	}else {
		if host2 <= host3 {
			//选择主机2
			createContainer("2")
		}else {
			//选择主机3
			createContainer("3")
		}
	}
	timestamp.Timestamp("weight",i)

}
func createContainer(host string) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://11.0.57."+host+":2375", "v1.23", nil, defaultHeaders)
	if err != nil {
		log.Print(err)
	}
	response, err :=cli.ContainerCreate(context.Background(),options.Config,options.HostConfig,options.NetworkingConfig,options.Name)
	if err != nil {
		fmt.Print("fail to create this container")
		fmt.Println(err)
	}
	fmt.Println(response)
	err = cli.ContainerStart(context.Background(),response.ID,types.ContainerStartOptions{})
	if err != nil {
		fmt.Print("fail to start this container")
		fmt.Println(err)
	}
}
