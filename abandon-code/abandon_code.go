package monitor

import (
	"encoding/json"
	"strings"
	"net/http"
	"github.com/robfig/cron"
)


//打印 docker events
body, err := cli.Events(context.Background(), types.EventsOptions{})
if err != nil {
log.Fatal(err)
}

dec := json.NewDecoder(body)
for {
var event events.Message
err := dec.Decode(&event)
if err != nil && err == io.EOF {
break
}

log.Println(event)
}
//解析容器json数据的类型
type Port struct{
	PrivatePort	int
	PublicPort	int
}
type Container struct {
	ID 	string
	Names   []string
	Images	string
	Status	string
	Ports	[]Port
}

type Containers struct {
	Containers []Container
}
type Stats struct{
	read 		string
	network 	Network
	memory_stats	Memory
	//blkio_stats	IO
	cpu_stats	CPU
}
type Network struct {
	rx_bytes	int
	tx_bytes	int
}
type Memory struct {
	max_usage	uint64
	usage		uint64
	limit		uint64
}
type IO struct {

}
type CPU struct {
	cpu_usage	stats_cpu
	system_cpu_usage	float32
}
type stats_cpu struct {
	total_usage	float32
}
func main() {

	c := cron.New()
	spec := "* */15 * * * ?"
	c.AddFunc(spec, func() {
		go func(){
			var containers Containers
			cons, err := http.Get("/containers/json")
			if err != nil{
				//
			}
			//解析获取的容器json数据
			json.Unmarshal([]byte(cons), &containers)
			//依此判断容器的运行状态
			for i:=0;i<len(containers.Containers) ; i++ {
				var con_stats Stats
				ID := containers.Containers[i].ID
				url := strings.Join([]string{"/containers/", "/stats"}, string(ID))
				stats, err := http.Get(url)
				if err != nil{
					//
				}
				json.Unmarshal([]byte(stats),&con_stats)

			}
			//i++
			//log.Println("running:",i)

		}()
	})
	c.Start()
	select{}

	/*直接使用select阻塞，加上time，也可以实现每隔一段时间执行一次检测
	for {
		select {
		case <-time.After(2 * time.Second):
			i++
			log.Println("running:",i)
		}
	}*/
}