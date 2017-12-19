package etcd

import (
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
	t.Log(port)
}
