package etcd

/*管理端口会话*/
import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"os/exec"
	"strings"
	"time"
	"varpac"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{varpac.Master.IP + ":2379"}
)

type Session struct {
	IP     string
	Port   string
	ConID  string
	Status string
}

//保存端口会话信息
func (s Session) Set(userid string) {

	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	check("", err)
	defer etcdcli.Close()

	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{$p}} -> {{(index $conf 0).HostPort}} {{end}}'
	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}'
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/IP", s.IP)
	check("", err)

	s.Port = s.getPort()

	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Port", s.Port)
	check("", err)
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/ConID", s.ConID)
	check("", err)
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Status", "connected")
	check("", err)
}

//获取端口信息
func (s *Session) Get(userid string) {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer etcdcli.Close()

	resp, err := etcdcli.Get(context.TODO(), "/user/"+userid+"/IP")
	if err != nil {
		log.Fatal(err)
		return
	}

	s.IP = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/Port")
	if err != nil {
		log.Fatal(err)
		return
	}
	s.Port = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/ConID")
	if err != nil {
		log.Fatal(err)
		return
	}
	s.ConID = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/Status")
	if err != nil {
		log.Fatal(err)
		return
	}

}

//判断是否不全
func (s Session) isZero() bool {
	if s.IP != "" && s.ConID != "" && s.Port != "" && s.Status != "" {
		return false
	}
	return true
}

//get the port one container is listening on
func (s Session) getPort() string {
	postCMD := "docker -H " + varpac.Master.IP + " :3375 " +
		"inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}'"

	out, err := exec.Command("/bin/bash", "-c", postCMD+s.ConID).Output()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	inputstring, err := outReader.ReadString('\n')
	slice := strings.Split(inputstring, " ")
	port := slice[1]
	return port
}
