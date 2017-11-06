package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"monitorloop"
)

type memStatus struct {
	totalMem float32
	availMem float32
}

var ioratio float32

/*type logistic struct {
	changed bool
	correct float32
}

var logistic logistic
*/
var Hostname string
var Startdata string //agent启动时的网络流量数据
func main() {
	Startdata = time.Now().Format("2006-01-02 15:04:05") + ";" + getNetbytes()
	go loopfunc()
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	Hostname = host
	//初始化参数correct的修正值为false
	//logistic.changed = false
	http.HandleFunc("/bindMachine", bindMachine)
	http.HandleFunc("/weightedSchedule", weightedSchedule)
	http.HandleFunc("/load", aveload)
	http.HandleFunc("/memload", memload)
	http.HandleFunc("/cpuload", cpuload)
	err = http.ListenAndServe(":9902", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

/********循环每15秒更新网络包数据********/
func loopfunc() {
	spec := "*/15* * * * ?"
	var f = func() {
		Startdata = time.Now().Format("2006-01-02 15:04:05") + ";" + getNetbytes()
		ioratio = ioRatio()
	}
	monitorloop.Execute(spec, f)
}
func bindMachine(w http.ResponseWriter, r *http.Request) {
	var workload float32 = 0
	var cpuratio, mempreratio, netratio, ioratio float32
	//var linknum int
	var alpha, belta, gama, namda float32
	var k1, k2, k3, k4 float32
	alpha = 1
	belta = 1
	gama = 1
	namda = 1
	k1 = 0.4
	k2 = 0.4
	k3 = 0.1
	k4 = 0.1
	cpuratio = cpuRatio()
	mempreratio = mempreRatio()
	//linknum = Getlink()
	netratio = networkLoad()
	//ioratio = ioRatio()
	//containerperc = ContainerUseMemPerc()
	fmt.Println("cpu+", cpuratio)
	fmt.Println("mempre+", mempreratio)
	fmt.Println("netratio+", netratio)
	fmt.Println("ioratio+", ioratio)
	//fmt.Println(linknum)
	//如果内存使用率（包括预测）高于0.5，对correct进行修正
	//if mempreratio >= 0.5 && logistic.changed == false {
	//corrChange(mempreratio, linknum)
	//}
	//netload = logisticTran(linknum)

	if cpuratio > 0.8 {
		alpha = 1.5
		if cpuratio > 0.9 {
			alpha = 1024
		}
	}
	if mempreratio > 0.5 {
		belta = 10
		if mempreratio > 0.8 {
			belta = 1024
		}
	}
	if netratio > 0.8 {
		gama = 1.5
		if netratio > 0.9 {
			gama = 1024
		}
	}
	if ioratio > 0.8 {
		namda = 1.5
		if ioratio > 0.9 {
			namda = 1024
		}
	}
	workload = workload + alpha*k1*cpuratio
	workload = workload + belta*k2*mempreratio
	workload = workload + gama*k3*netratio
	workload = workload + namda*k4*ioratio
	//workload = workload + k4*containerperc
	workload64 := float64(workload)
	result := strconv.FormatFloat(workload64, 'f', -1, 32)
	fmt.Fprint(w, result)
}
func weightedSchedule(w http.ResponseWriter, r *http.Request) {
	var workload float32 = 0.0
	var cpupercent, mempercent, netpercent, containerperc float32
	var alpha, belta, gama float32
	var k1, k2, k3, k4 float32
	alpha = 1
	belta = 1
	gama = 1
	k1 = 0.35
	k2 = 0.35
	k3 = 0.2
	k4 = 0.1
	cpupercent = cpuRatio()
	mempercent = memRatio()
	netpercent = networkLoad()
	containerperc = ConMemPerc()
	if cpupercent > 0.8 {
		alpha = 1.5
		if cpupercent > 0.9 {
			alpha = 1024
		}
	}
	if mempercent > 0.8 {
		belta = 1.5
		if mempercent > 0.9 {
			belta = 1024
		}
	}
	if netpercent > 0.8 {
		gama = 1.5
		if netpercent > 0.9 {
			gama = 1024
		}
	}

	fmt.Println(cpupercent)
	fmt.Println(mempercent)
	fmt.Println(netpercent)
	fmt.Println(containerperc)
	workload = workload + alpha*k1*cpupercent
	workload = workload + belta*k2*mempercent
	workload = workload + gama*k3*netpercent
	workload = workload + k4*containerperc
	workload64 := float64(workload)
	result := strconv.FormatFloat(workload64, 'f', -1, 32)
	fmt.Fprint(w, result)
}
func aveload(w http.ResponseWriter, r *http.Request) {
	var workload1, workload2 float32
	var cpuratio, mempreratio, netratio float32
	workload1 = 0
	workload2 = 0

	var k1, k2, k3, k4 float32
	cpuratio = 0.0
	for i := 0; i < 10; i++ {
		cpuratio = cpuratio + cpuRatio()
	}
	cpuratio = cpuratio / 10.0
	mempreratio = mempreRatio()
	netratio = networkLoad()
	ioratio = ioRatio()

	k1 = 0.2
	k2 = 0.6
	k3 = 0.1
	k4 = 0.1
	//我的评价指标（负载）
	workload1 = workload1 + k1*cpuratio
	workload1 = workload1 + k2*mempreratio
	workload1 = workload1 + k3*netratio
	workload1 = workload1 + k4*ioratio
	//论文的评价指标（负载）
	workload2 = workload2 + 0.4*cpuratio
	workload2 = workload2 + 0.4*mempreratio
	workload2 = workload2 + 0.1*netratio
	workload2 = workload2 + 0.1*ioratio
	workload := strconv.FormatFloat(float64(workload1), 'f', -1, 32)
	workload = workload + ";"
	workload = workload + strconv.FormatFloat(float64(workload2), 'f', -1, 32)
	fmt.Println("cpu=", cpuratio)
	fmt.Println("mem=", mempreratio)
	fmt.Println("aveload=", workload)
	fmt.Fprint(w, workload)
}
func memload(w http.ResponseWriter, r *http.Request) {
	mempercent := memRatio()
	mempercent64 := float64(mempercent)
	result := strconv.FormatFloat(mempercent64, 'f', -1, 32)
	fmt.Fprint(w, result)
}
func cpuload(w http.ResponseWriter, r *http.Request) {
	cpuratio := cpuRatio()
	cpuratio64 := float64(cpuratio)
	result := strconv.FormatFloat(cpuratio64, 'f', -1, 32)
	fmt.Fprint(w, result)
}

/*****对tcp连接数进行logistics转换*****
func logisticTran(numX int) float32 {
	var logRes float32

	memStat := memStat()
	memtotal := memStat.totalMem
	logistic.correct = memtotal / 1024
	pow := -(numX/logistic.correct - 6)
	exp := math.Exp(pow)
	logRes = 1 / (1 + exp)
	return logRes
}
*/
/*****对correct进行修正*****
func corrChange(memratio float32, numX int) {
	if logistic.changed == false {
		lognum := 1/memratio - 1
		ln := math.Log(lognum)
		mom := 6 - ln
		logistic.correct = numX / mom
		logistic.changed = true
	}

}
*/
/*****获取iostat数据%util,io百分比******/
func getIO() string {
	cmd := exec.Command("/bin/bash", "./getIO.sh")
	outbytes, err := cmd.Output()
	if err != nil {
		fmt.Println("iostat", err)
		return ""
	}
	outString := bytestoString(outbytes)
	outString = strings.TrimSpace(outString) //去除空格，换行符
	return outString
}

/*****执行脚本，获取vmstat数据******/
/*
#!/bin/bash
us=`vmstat | awk '{ for (i=1; i<=NF; i++) if ($i=="us") { getline; print $i }}'`
sy=`vmstat | awk '{ for (i=1; i<=NF; i++) if ($i=="sy") { getline; print $i }}'`
id=`vmstat | awk '{ for (i=1; i<=NF; i++) if ($i=="id") { getline; print $i }}'`
echo "$us;$sy;$id"
*/
/*****执行脚本，获取iostat cpu空闲时间数据******/
/*使用vmstat获取的数据不准确，改用iostat -c输出的系统空闲时间%idle
#!/bin/bash
idle=`iostat -c | sed -n '4p'|awk '{  i=NF; print $i }'`
*/
func getCpustat() string {
	cmd := exec.Command("/bin/bash", "./getCpustat.sh")
	outbytes, err := cmd.Output()
	if err != nil {
		fmt.Println("getCpustat", err)
		return ""
	}
	outString := bytestoString(outbytes)
	outString = strings.TrimSpace(outString) //去除空格，换行符
	return outString
}

/*****IO状态******/
func ioRatio() float32 {
	var ratio float64
	outString := getIO()
	outString = strings.Replace(outString, "\n", "", -1) //去除换行符
	ratio, _ = strconv.ParseFloat(outString, 32)
	return float32(ratio / 100.0)

}

/*****CPU使用率vmstat******/
func cpuRatiovm() float32 {
	var us int64 //用户态
	var sy int64 //系统态
	var id int64 //空闲态
	us = 0
	sy = 0
	id = 0

	outString := getCpustat()
	slice := strings.Split(outString, ";")
	slice[2] = strings.Replace(slice[2], "\n", "", -1) //去除换行符
	us, _ = strconv.ParseInt(slice[0], 10, 0)
	sy, _ = strconv.ParseInt(slice[1], 10, 0)
	id, _ = strconv.ParseInt(slice[2], 10, 0)

	if (us + sy + id) == 0 {
		return -1
	} else {
		return float32(us+sy) / float32(us+sy+id)
	}
}

/*****CPU使用率iostat******/
func cpuRatio() float32 {

	//%空闲率0-100
	var idle float64
	outString := getCpustat()

	outString = strings.Replace(outString, "\n", "", -1) //去除换行符

	idle, _ = strconv.ParseFloat(outString, 32)
	//fmt.Println("idel=",idle)
	return float32(1.0 - idle/100.0)
}

/*****计算我的算法第二项******/
func mempreRatio() float32 {

	memStat := memStat()
	memtotal := memStat.totalMem
	memavailable := memStat.availMem
	ConUnuse := ContainerUnuse()
	//容器已申请，但未使用的内存，在未来有一定的可能会被使用，设可能概率为20%
	ConUnuse = ConUnuse * 0.2 / (1024.0 * 1024.0 * 1024.0)
	percent := (float32(memtotal) - float32(memavailable) + ConUnuse) / float32(memtotal)
	return percent
}

/*****主机内存使用率******/
func memRatio() float32 {

	inputFile, inputError := os.Open("/proc/meminfo")
	if inputError != nil {
		fmt.Printf("An error occurred on opening the inputFlie\n" +
			"Have you got access to it?\n")
		return 0
	}
	defer inputFile.Close()

	var colNumber []string
	//读取文件中前两行的第二列内容，分别为MemTotal和MemAvailable的数值
	var i int
	for i = 0; i < 4; i++ {
		var v1, v2, v3 string
		_, err := fmt.Fscanln(inputFile, &v1, &v2, &v3)
		if err != nil {
			break
		}
		colNumber = append(colNumber, v2)
		if strings.EqualFold(v1, "MemAvailable:") {
			break
		}
	}
	memtotal, _ := strconv.Atoi(colNumber[0])
	//memfree,_ := strconv.Atoi(colNumber[1])
	memavailable, _ := strconv.Atoi(colNumber[2])
	if i == 4 {
		cmd := exec.Command("/bin/bash", "./countMem.sh")
		outbytes, err := cmd.Output()
		if err != nil {
			fmt.Println("countMem", err)
		}
		outString := bytestoString(outbytes)
		outString = strings.TrimSpace(outString) //去除空格，换行符
		count, _ := strconv.ParseInt(outString, 10, 64)
		percent1 := float32((int(count)*2*1024*1024))/float32(memtotal) + 0.08

		return percent1
	}

	//var memtotal  int
	//inputFile, inputError := os.Open("/proc/meminfo")
	//if inputError != nil {
	//	fmt.Printf("An error occurred on opening the inputFlie\n" +
	//		"Have you got access to it?\n")
	//	return 0
	//}
	//defer inputFile.Close()
	////读取文件中第一行第二列内容，MemTotal的数值
	//
	//var v1, v2, v3 string
	//_, err := fmt.Fscanln(inputFile, &v1, &v2, &v3)
	//if err != nil {
	//
	//}
	//memtotal, _ = strconv.Atoi(v2)
	//cmd := exec.Command("/bin/bash", "./getAvailable.sh")
	//outbytes, err := cmd.Output()
	//if err != nil {
	//	fmt.Println("getAvailable", err)
	//}
	//outString := bytestoString(outbytes)
	//outString = strings.TrimSpace(outString) //去除空格，换行符
	//memavailable ,_:= strconv.ParseInt(outString,10,64)
	percent := (float32(memtotal) - float32(memavailable)) / float32(memtotal)
	//fmt.Println(percent)
	return percent
}

/*****执行脚本，获取网络流量数据******/
/*
#!/bin/bash
DATE=`date --utc`
RX=`cat /proc/net/dev | grep em1 |awk '{print $2}'`
TX=`cat /proc/net/dev | grep em1 |awk '{print $10}'`
echo "$RX;$TX"
*/
func getNetbytes() string {
	cmd := exec.Command("/bin/bash", "./getNetbytes.sh")
	outbytes, err := cmd.Output()
	if err != nil {
		fmt.Println("getNetbytes", err)
		return ""
	}
	outString := bytestoString(outbytes)
	outString = strings.TrimSpace(outString) //去除空格，换行符
	return outString
}
func bytestoString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

/*****网络平均负载******/
func networkLoad() float32 {
	var loadPercent float32
	var averageLoad float32
	var receiveBytes float32
	var transmitBytes float32

	outString := getNetbytes()

	slice := strings.Split(outString, ";")
	date_now := time.Now()
	recBytes_now, _ := strconv.ParseInt(slice[0], 10, 64)
	transBytes_now, _ := strconv.ParseInt(slice[1], 10, 64)

	//主机网络传输字节的初始值
	slice_start := strings.Split(Startdata, ";")
	date_start := slice_start[0]
	recBytes_start, _ := strconv.ParseInt(slice_start[1], 10, 64)
	transBytes_start, _ := strconv.ParseInt(slice_start[2], 10, 64)
	t2, _ := time.Parse("2006-01-02 15:04:05", date_start)
	t1, _ := time.Parse("2006-01-02 15:04:05", date_now.Format("2006-01-02 15:04:05"))
	diff := t1.Sub(t2)

	duration := float32(diff.Seconds())
	receiveBytes = float32(recBytes_now - recBytes_start)
	transmitBytes = float32(transBytes_now - transBytes_start)

	if duration != 0 {
		averageLoad = (receiveBytes + transmitBytes) * 8 / duration / 1024 / 1024
	} else {
		averageLoad = 0
	}
	//1000Mbps为系统总吞吐量
	loadPercent = averageLoad / 1000.0
	return loadPercent
}
func Getlink() int {
	var linkNum int
	cmd := exec.Command("/bin/bash", "netstat -na|grep ESTABLISHED|wc -l")
	outbytes, err := cmd.Output()
	if err != nil {
		fmt.Println("netstat", err)
		return -1
	}
	outString := bytestoString(outbytes)
	outString = strings.TrimSpace(outString) //去除空格，换行符
	linkNum, err = strconv.Atoi(outString)
	if err != nil {
		fmt.Println("Atoi error", err)
		return -1
	}
	return linkNum
}

//计算主机内存总值和可用值
func memStat() memStatus {
	memStat := new(memStatus)
	inputFile, inputError := os.Open("/proc/meminfo")
	if inputError != nil {
		fmt.Printf("An error occurred on opening the inputFlie\n" +
			"Have you got access to it?\n")

	}
	defer inputFile.Close()

	var colNumber []string
	//读取文件中前两行的第二列内容，分别为MemTotal和MemAvailable的数值
	var i int
	for i = 0; i < 4; i++ {
		var v1, v2, v3 string
		_, err := fmt.Fscanln(inputFile, &v1, &v2, &v3)
		if err != nil {
			break
		}
		colNumber = append(colNumber, v2)
		if strings.EqualFold(v1, "MemAvailable:") {
			break
		}
	}
	memtotal, _ := strconv.Atoi(colNumber[0])
	//memfree,_ := strconv.Atoi(colNumber[1])
	memavailable, _ := strconv.Atoi(colNumber[2])
	if i == 4 {
		cmd := exec.Command("/bin/bash", "./countMem.sh")
		outbytes, err := cmd.Output()
		if err != nil {
			fmt.Println("countMem", err)
		}
		outString := bytestoString(outbytes)
		outString = strings.TrimSpace(outString) //去除空格，换行符
		count, _ := strconv.ParseInt(outString, 10, 64)
		memavailable = memtotal - int(count)*2*1024*1024 - 2633728 //2572占内存8%

	}
	memStat.totalMem = float32(memtotal)
	memStat.availMem = float32(memavailable)
	return *memStat
}

/*****为Docker分配出去但未使用的内存********/
func ContainerUnuse() float32 {
	//先获取容器使用的内存数额
	var total_rss float64
	absDir := "/sys/fs/cgroup/memory/docker/"
	list, err := ioutil.ReadDir(absDir)
	if err != nil {
		fmt.Println("read dir error")
		return 0
	}
	for _, info := range list {
		//判断是否为目录，目录名是container的id，里面为相应container的信息
		if info.IsDir() == true {
			realpath := filepath.Join(absDir, info.Name(), "memory.stat")
			inputFile, inputError := os.Open(realpath)
			if inputError != nil {
				fmt.Printf("An error occurred on opening the inputFlie\n" +
					"Have you got access to it?\n")
				return 0
			}
			defer inputFile.Close()

			for {
				var v1 string
				var v2 float64
				_, err := fmt.Fscanln(inputFile, &v1, &v2)
				if err != nil {
					break
				}
				if strings.EqualFold(v1, "total_rss") {
					total_rss = total_rss + v2
					break
				}
			}

		}
	}
	//fmt.Println(total_rss)
	//获取分配给所有容器的内存
	var reservedMem float64
	out, err := exec.Command("/bin/bash", "-c", "docker -H 11.0.57.2:3375 info").Output()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("cmd output: %s",out)
	//out的type为[]byte，需要转换为io.Reader才能使用Fscanln
	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	flagnode := false

	for {
		inputString, readerError := outReader.ReadString('\n')
		if readerError == io.EOF {
			break
		}
		slice := strings.Split(inputString, " ")
		//for i:=0;i<len(slice);i++ {
		//fmt.Println(slice[i])
		//}
		if flagnode != true {
			//slice[0]是换行符，madan
			if strings.EqualFold(slice[1], string(Hostname+":")) {
				flagnode = true
				//fmt.Println("i am in the 417-2 loop")
			}
		} else {
			if strings.EqualFold(slice[4], "Memory:") {
				reservedMem, _ = strconv.ParseFloat(slice[5], 32)
				//fmt.Println("i am in the Memory: loop")
				if strings.EqualFold(slice[6], "GiB") {
					reservedMem = reservedMem * float64(1024*1024*1024)
				}
				if strings.EqualFold(slice[6], "MiB") {
					reservedMem = reservedMem * float64(1024*1024)
				}
				if strings.EqualFold(slice[6], "KiB") {
					reservedMem = reservedMem * float64(1024)
				}
				break
			}
		}

	}
	//处理字符串rawData，提取出主机为容器分配的内存数值
	fmt.Println("reser", reservedMem)
	fmt.Println("total", total_rss)
	if reservedMem != 0 {
		return float32(reservedMem - total_rss)
	} else {
		return 0
	}
}

/*****为Docker分配出去但未使用的内存占全部分配的比例******/
func ConMemPerc() float32 {

	//先获取容器使用的内存数额

	memStat := memStat()
	ConUnuse := ContainerUnuse()
	ConUnuse = ConUnuse / (1024 * 1024 * 1024)
	if memStat.totalMem != 0 {
		return float32(ConUnuse / memStat.totalMem)
	} else {
		return 0
	}

}
