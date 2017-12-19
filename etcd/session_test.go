package etcd

import (
	"testing"
)

func Test_Set(t *testing.T) {
	conid = "000"
	hostip = "11.0.0.172"
	url = "11.0.0.172:9500"
	// set lab session
	labSession := etcd.Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
		Url:    url,
	}
	labSession.Set(userid)
}

func Test_GetPort(t *testing.T) {
	conid = "000"
	hostip = "11.0.0.172"
	url = "11.0.0.172:9500"
	// set lab session
	labSession := etcd.Session{
		IP:     hostip,
		ConID:  conid,
		Status: "started",
		Url:    url,
	}
	labSession.Set(userid)
	port := labSession.getPort()
	t.Log(port)
}
