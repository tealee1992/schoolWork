// test project main.go
package main

import (
	"time"
)
import (
	"net/http"

	"strconv"
)

func main() {
	//test(60, 0, 1)
	//test(30, 30, 1)
	//test(0, 60, 1)
	//test(60, 0, 2)
	//test(30, 30, 2)
	//test(0, 60, 2)
	//test(20, 0, 3)
	//test(30, 30, 3)
	///test(0, 60, 3)
	//test(180, 0, 1)
	//test(120, 0, 1)
	//test(120, 0, 2)
	//test(120, 0, 3)
	//test(0, 120, 1)
	//test(0, 120, 2)
	//test(0, 120, 3)
	//test(0, 60, 2)
	//test(0, 60, 3)
	typecreate(40, 3)
	//typecreate(40, 2)
	//typecreate(40, 1)
}
func test(num1 int, num2 int, method int) {
	if num1 > 0 {
		onebyonecreate(num1, method)
	}
	if num2 > 0 {
		goroutinecreate(num2, method)
	}
}

//类型请求
func typecreate(num int, method int) {
	var demand string
	for i := 0; i < num; i++ {
		//swarm的Spread策略
		if i%2 == 0 {
			demand = "0"
		} else {
			demand = "1"
		}
		index := strconv.Itoa(i)
		if method == 1 {
			http.Get("http://11.0.57.2:9090/containers/create?method=1&index=" + index + "&demand=" + demand)
			//http.Get("http://localhost:9090/containers/create?method=1&index=" + index)

		} else if method == 2 { //论文中的策略
			http.Get("http://11.0.57.2:9090/containers/create?method=2&index=" + index + "&demand=" + demand)

		} else if method == 3 { //我的方法
			http.Get("http://11.0.57.2:9090/containers/create?method=3&index=" + index + "&demand=" + demand)
		}
		//sleepTest()
	}
}

//顺序请求
func onebyonecreate(num int, method int) {
	for i := 0; i < num; i++ {
		//swarm的Spread策略

		index := strconv.Itoa(i)
		if method == 1 {
			http.Get("http://11.0.57.2:9090/containers/create?method=1&index=" + index)
			//http.Get("http://localhost:9090/containers/create?method=1&index=" + index)

		} else if method == 2 { //论文中的策略
			http.Get("http://11.0.57.2:9090/containers/create?method=2&index=" + index)

		} else if method == 3 { //我的方法
			http.Get("http://11.0.57.2:9090/containers/create?method=3&index=" + index)
		}
		//sleepTest()
	}
}

//goroutine模拟并发请求
func goroutinecreate(num int, method int) {
	//chs := make([]chan int,num)
	chs := make(chan int)
	for i := 0; i < num; i++ {
		//swarm的Spread策略
		index := strconv.Itoa(i)
		if method == 1 {
			//http.Get("http://localhost:9090/containers/create?method=1&index=" + index)
			go http.Get("http://11.0.57.2:9090/containers/create?method=1&index=" + index)
			//swarmSpread(i)
		} else if method == 2 { //论文中的策略
			go http.Get("http://11.0.57.2:9090/containers/create?method=2&index=" + index)
			//weightSchedule(i)
		} else if method == 3 { //我的方法
			go http.Get("http://11.0.57.2:9090/containers/create?method=3&index=" + index)
			//combine(i)
		}
	}
	time.Sleep(time.Duration(2) * time.Minute)
	chs <- 1
	<-chs
}
func sleepTest() {
	time.Sleep(time.Duration(5 * time.Second))
}
