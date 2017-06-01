package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"strings"
	"time"
	"varpac"

)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{varpac.Master.IP + ":4001"}
)

type session struct {
	IP     string
	Port   string
	ConID  string
	Status string
}

func main() {
	//var port string
	//ip:= "11.0.57.2"
	conid:="5c4625c062f6"
	getport := "docker -H " + varpac.Master.IP + ":3375" +
		" inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}' "+conid

	fmt.Println(getport)
	out, err := exec.Command("/bin/bash", "-c", getport).Output()
	if err != nil {
		log.Fatal(err)
		fmt.Println("err from exec command")
		return
	}

	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	inputstring, err := outReader.ReadString('\n')
	slice := strings.Split(inputstring, " ")
	fmt.Println(slice[1])

}
