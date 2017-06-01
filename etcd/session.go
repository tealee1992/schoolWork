package etcd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"varpac"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{varpac.Master.IP + ":2379"}
)

type session struct {
	IP     string
	Port   string
	ConID  string
	Status string
}

func (s session) Set(userid int) {
	
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer etcdcli.Close()
	useridStr := strconv.Itoa(userid)
	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{$p}} -> {{(index $conf 0).HostPort}} {{end}}'
	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}'
	_, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/IP", s.IP)
	if err != nil {
		log.Fatal(err)
		return
	}

	port := s.getPort()

	_, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/Port", port)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/ConID", s.ConID)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/Status", "connected")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (s *session) Get(userid int) {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		DialTimeout: dialTimeout,
		})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer etcdcli.Close()
	useridStr := strconv.Itoa(userid)

	s.IP, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/IP")
	if err != nil {
		log.Fatal(err)
		return
	}

	Port, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/Port")
	if err != nil {
		log.Fatal(err)
		return
	}
	s.ConID, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/ConID")
	if err != nil {
		log.Fatal(err)
		return
	}
	s.Status, err = etcdcli.Put(context.TODO(), "/user/"+useridStr+"/Status")
	if err != nil {
		log.Fatal(err)
		return
	}
}
func (s session) isZero() bool {
	if(s.IP!=""&&s.ConID!=""&&s.Port!=""&&s.Status!=""){
		return false
	}
	return true
}
//get the port one container is listening on
func (s session) getPort() port string{
	postCMD:="docker -H " + varpac.Master.IP + " :3375 " +
		"inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}'"

	out, err := exec.Command("/bin/bash", "-c", postCMD+s.ConID).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	inputstring, err := outReader.ReadString('\n')
	slice := strings.Split(inputstring, " ")
	port = slice[1]
}