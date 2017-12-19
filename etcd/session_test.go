package etcd

import (
	"fmt"
	"testing"
)

func Test_Set(t *testing.T) {
	conid := "0945a9b40e4e"
	hostip := "11.0.0.172"
	url := "11.0.0.172:9500"
	// set lab session
	labSession := Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
		Url:    url,
	}
	userid := "000"
	labSession.Set(userid)
	fmt.Println("end of set")
}

func Test_GetPort(t *testing.T) {
	conid := "0945a9b40e4e"
	hostip := "11.0.0.172"
	url := "11.0.0.172:9500"
	// set lab session
	labSession := Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
		Url:    url,
	}
	userid := "000"
	labSession.Set(userid)
	port := labSession.getPort()
	t.Log(port)
	fmt.Println("port:" + port)
	fmt.Println("end of get")
}
