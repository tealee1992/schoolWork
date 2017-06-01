package monitorloop

import (

	"github.com/robfig/cron"
)


func Execute(spec string,f func()){
	c := cron.New()
	c.AddFunc(spec,f);
	c.Start()
	select{}
}