package functions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"varpac"
)

func GetMemload(ip string) float64 {
	var result float64
	resp, err := http.Get("http://" + ip + ":" + varpac.AgentPort + "/memload")
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

func GetCpuload(ip string) float64 {
	var result float64
	resp, err := http.Get("http://" + ip + ":" + varpac.AgentPort + "/cpuload")
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
