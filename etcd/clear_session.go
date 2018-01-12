package etcd

import (
	"fmt"
)

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
}
