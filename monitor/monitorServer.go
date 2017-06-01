package monitor

import (
	"log"
	"encoding/json"
	"github.com/robfig/cron"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

type monitor struct {

}

func (m *monitor)monitorLoop(){
	c := cron.New()
	spec := "*/15 * * * * ?"
	c.AddFunc(spec, func() {
		//go func(){
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err := client.NewClient("http://11.0.57.2:3375", "v1.23", nil, defaultHeaders)
		if err != nil {
			panic(err)
		}

		options := types.ContainerListOptions{All: true}
		containers, err := cli.ContainerList(context.Background(), options)
		if err != nil {
			panic(err)
		}
		i := 0;
		for _, c := range containers {
			//fmt.Println(c.ID)
			//var CPUusage types.CPUUsage
			//var CPUstats types.CPUStats
			var Stats  types.Stats
			body,err := cli.ContainerStats(context.Background(),c.ID,false)
			if err!= nil {
				log.Fatal(err)
			}
			dec := json.NewDecoder(body)
			err = dec.Decode(&Stats)
			if err != nil{
				break
			}
			i++
			log.Println(i)
			log.Println(c.Names)
			log.Println(Stats.CPUStats.SystemUsage)

		}

		//}()
	})
	c.Start()
	select{}
}
