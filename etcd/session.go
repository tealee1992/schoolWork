package etcd

/*管理端口会话*/
import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os/exec"
	"strings"
	"time"
	"varpac"
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

//保存端口会话信息
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

//获取端口信息
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

//判断是否存在该user的session
//这个函数理想的做法是 利用etcd的api询问是否有该userid的记录
//但是etcdcli没有这样的操作，只能以报错来作为判断依据了，没时间了。。。
func (s Session) IsExist(userid string) bool {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer etcdcli.Close()
	resp, err := etcdcli.Get(context.TODO(), "/user/"+userid+"/IP")
	if err != nil {
		fmt.Println(err)
		fmt.Println("etcd Get err")
		return false
	}
	if resp.Kvs == nil || len(resp.Kvs) == 0 {
		fmt.Println("there is no log of this user")
		return false
	}
	return true
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
	postCMD := "docker -H " + varpac.Master.IP + ":3375 " +
		"inspect --format='{{range $p, $conf := .NetworkSettings.Ports}} {{(index $conf 0).HostPort}} {{end}}' "

	out, err := exec.Command("/bin/bash", "-c", postCMD+" "+s.ConID).Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	inputstring, err := outReader.ReadString('\n')
	slice := strings.Split(inputstring, " ")
	port := slice[1]
	return port
}
func (s Session) getUrl() {

}

//设置容器状态
func SetStatus(userid string, status string) {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer etcdcli.Close()
	_, err = etcdcli.Put(context.TODO(), "/user/"+userid+"/Status", status)
	if err != nil {
		fmt.Println(err)
		return
	}
}
