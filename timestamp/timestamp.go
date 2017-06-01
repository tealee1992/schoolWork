package timestamp

import (
	"os"
	"fmt"
	"strconv"
	"time"
	"varpac"
)

//记录每次请求的时间
func Timestamp(method string,i int) {

	outputfile, err := os.OpenFile("./log/"+method+"-time.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err!= nil {
		fmt.Println("an error occurred with file opening or creation")
		return
	}
	defer outputfile.Close()
	//outputWriter := bufio.NewWriter(outputfile)
	outputString := time.Now().Format("2006-01-02 15:04:05")
	concurrency:=strconv.Itoa(varpac.Concurrency)
	fmt.Fprintf(outputfile,strconv.Itoa(i)+"-"+outputString+concurrency+"\n")
	//outputWriter.WriteString(strconv.Itoa(i)+"-"+outputString+"\n")
}

//记录请求中计算出的权值。（分析是否为实时负载）
func Weightlog(weight1 float64,weight2 float64,weight3 float64) {

	outputfile, err := os.OpenFile("./log/log-weight.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err!= nil {
		fmt.Println("an error occurred with file opening or creation")
		return
	}
	defer outputfile.Close()
	//outputWriter := bufio.NewWriter(outputfile)
	outputString := time.Now().Format("2006-01-02 15:04:05")
	w1 := strconv.FormatFloat(weight1,'f',-1,32)
	w2 := strconv.FormatFloat(weight2,'f',-1,32)
	w3 := strconv.FormatFloat(weight3,'f',-1,32)
	fmt.Fprintf(outputfile,outputString+"|"+w1+"|"+w2+"|"+w3+"\n")
	//outputWriter.WriteString(outputString+"|"+w1+"|"+w2+"|"+w3+"\n")
}

func Probalog(section [3]float64,randnum float64) {
	outputfile, err := os.OpenFile("./log/log-proba.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err!= nil {
		fmt.Println("an error occurred with file opening or creation")
		return
	}
	defer outputfile.Close()
	w1 := strconv.FormatFloat(section[0],'f',-1,32)
	w2 := strconv.FormatFloat(section[1],'f',-1,32)
	w3 := strconv.FormatFloat(section[2],'f',-1,32)
	w4 := strconv.FormatFloat(randnum,'f',-1,32)
	fmt.Fprintf(outputfile,w1+"|"+w2+"|"+w3+"|"+w4+"\n")
}

func Simulog(i int){
	outputfile, err := os.OpenFile("./log/"+"simulog-time.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if err!= nil {
		fmt.Println("an error occurred with file opening or creation")
		return
	}
	defer outputfile.Close()
	//outputWriter := bufio.NewWriter(outputfile)
	outputString := time.Now().Format("2006-01-02 15:04:05")
	concurrency:=strconv.FormatBool(varpac.Concurrency)
	fmt.Fprintf(outputfile,strconv.Itoa(i)+"-"+outputString+concurrency+"\n")
	//outputWriter.WriteString(strconv.Itoa(i)+"-"+outputString+"\n")
}