package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{"127.0.0.1:2379"}
)

type Session struct {
	IP     string
	Port   string
	ConID  string
	Status string
	Url    string
}

func main() {
	conid := ""
	hostip := ""
	url := ""
	// set lab session
	labSession := Session{
		IP:     hostip,
		ConID:  conid,
		Status: "none",
		Url:    url,
	}
	userid := "111"
	labSession.Set(userid)
	fmt.Println("end of set")
	var s Session
	s.Get(userid)
	fmt.Println(s.Status)
}

func (s Session) Set(userid string) {

	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("chuang jian shibai")
		return
	}
	defer etcdcli.Close()

	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{$p}} -> {{(index $conf 0).HostPort}} {{end}}'
	//docker inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}'
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/IP", s.IP)
	if err != nil {
		fmt.Println(err)
		fmt.Println("IP shibai")
		return
	}

	//s.Port = s.getPort()
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Port", s.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/ConID", s.ConID)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Status", s.Status)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Url", s.Url)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Session) Get(userid string) {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer etcdcli.Close()
	resp, err := etcdcli.Get(context.TODO(), "/user/"+userid+"/IP")
	if err != nil {
		fmt.Println(err)
		return
	}

	s.IP = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/Port")
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Port = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/ConID")
	if err != nil {
		fmt.Println(err)
		return
	}
	s.ConID = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/Status")
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Status = string(resp.Kvs[0].Value)

	resp, err = etcdcli.Get(context.TODO(), "/user/"+userid+"/Url")
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Url = string(resp.Kvs[0].Value)
}
