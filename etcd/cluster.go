package etcd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"log"
	"os/exec"
	"time"
	"varpac"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{varpac.Master.IP + ":2379"}
)

//通过docker client 启动swarm 容器执行swarm list，containerstart没有-a选项，不能使用
func UpdateClusterClient() {
	var option = types.ContainerCreateConfig{
		Config: &container.Config{
			//AttachStdin: true,
			//Tty:         true,
			Image: "swarm",
			// Volumes: map[string]struct{}{
			// 	"/tempfiles": {},
			// },
			Cmd: []string{"list","etcd://"+varpac.Master.IP+"2379/swarm"}
		},
		HostConfig: &container.HostConfig{
			//AutoRemove : true,
			//因为docker create 没有提供--rm flag 所以，先正常创建，启动，在删除
		},
	}
	defaultHeaders := map[string]string{"User-Agent":"engine-api-cli-1.0"}
	cli, err := client.NewClient("http://"+varpac.Master.IP+"3375", "v1.23", nil,defaultHeaders)
	if err != nil {
		log.Fatal(err)
	}

	response, err := cli.ContainerCreate(context.Background(),option.Config, option.HostConfig, option.NetworkingConfig, option.Name)
	if err != nil {
		log.Println("fail to create this container")
		log.Fatal(err)
	}
	//处理response
	err = cli.ContainerStart(context.Background(), response.ID, types.ContainerStartOptions{})
	if err != nil {
		fmt.Println("fail to start this container")
		fmt.Println(err)
	}

}

//直接执行cmd命令 swarm list 获取cluster nodes
func UpdateCluster() {
	CMD := "docker run --rm swarm list " + varpac.Master.IP + ":2379/swarm"
	out, err := exec.Command("/bin/bash", "-c", CMD).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)
	inputstring, err := outReader.ReadString('\n')
	//处理
}
func UpdtClusbyEtcd() {
	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer etcdcli.Close()

	resp, err := etcdcli.Get(context.TODO(), "/swarm")
	if err != nil {
		log.Fatal(err)
		return
	} else {
		fmt.Println("resp: ", resp)
	}

	outBuffer := bytes.NewBuffer(out)
	outReader := bufio.NewReader(outBuffer)

}
