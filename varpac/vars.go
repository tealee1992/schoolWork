package varpac
var Concurrency=0
var sem1 = make(chan struct{},1)
var sem2 = make(chan struct{},1)
var Master =host {
	IP:"11.0.57.2",
}
var Cluster []host
type host struct{
	IP string
	totalMem float64
	memload float64
	probability float64
}
func FastVol()  {
	select {
	case sem1<- struct {}{}:
		default:
			Concurrency=(Concurrency+1)%3
	}
}

func AccurateVol()  {
	select {
	case sem2<- struct {}{}:
	default:
		Concurrency=(Concurrency+1)%3
	}
}

func TypeBased(){
	select {
	case <-sem1:
		<-sem2
	default:
		Concurrency=(Concurrency+1)%3
	}

}