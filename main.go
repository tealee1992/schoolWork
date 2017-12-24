package main

/*server 入口*/
import (
	"database/sql"
	"encoding/json"
	"etcd"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"varpac"
)

var templates = make(map[string]*template.Template)
var logFile *os.File
var loger *log.Logger

type Entry struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func init() {
	//设置日志
	setLog()
	//判断是否有labimage
	hasImage()
	//页面模板
	TEMPLATE_DIR := varpac.TEMPLATE_DIR
	fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
	check("failed to read template", err)

	var templateName, templatesPath string
	for _, fileInfo := range fileInfoArr {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}
		templatesPath = TEMPLATE_DIR + "/" + templateName
		loger.Println("Loading template:", templatesPath)
		t := template.Must(template.ParseFiles(templatesPath))
		templates[templateName] = t
	}

}

func main() {

	http.HandleFunc("/containers/setimage", safeHandler(setImage))
	http.HandleFunc("/containers/getlabimage", safeHandler(getLabImage))
	http.HandleFunc("/containers/init", safeHandler(init_student))
	http.HandleFunc("/containers/create", safeHandler(createContainer))
	http.HandleFunc("/containers/list", safeHandler(listContainer))
	http.HandleFunc("/containers/checkpoint", safeHandler(checkpoint))
	http.HandleFunc("/containers/restore", safeHandler(restore))
	http.HandleFunc("/containers/destroy", safeHandler(destroy))
	http.HandleFunc("/", safeHandler(homePage))
	err := http.ListenAndServe(":9901", nil)
	if err != nil {
		loger.Fatal("ListenAndServe:", err.Error())
	}

}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				loger.Printf("warn: panic in %v - %v", fn, e)

				loger.Println(string(debug.Stack()))
			}
		}()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json")
		fn(w, r)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	locals := make(map[string]interface{})
	locals["UserId"] = "111111"
	renderHtml(w, "index.html", locals)
}

func setLog() {
	logFile, err := os.Create("./logs.txt")
	if err != nil {
		fmt.Println(err)
	}
	loger = log.New(logFile, "cloudlab_go_server_", log.Ldate|log.Ltime|log.Lshortfile)
}

func hasImage() {
	db, err := sql.Open("mysql", "root:abcd1234!@tcp(localhost:3306)/cloudlab?parseTime=true")
	if err != nil {
		loger.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("select * from labimage")
	if err != nil {
		loger.Fatalln(err)
	}
	tag := false
	for rows.Next() {
		tag = true
	}
	if !tag {
		stmt, err := db.Prepare(`insert labimage (title,imagename) values (?,?)`)
		if err != nil {
			loger.Fatal(err)
		}
		_, err = stmt.Exec(varpac.Title, "")
		if err != nil {
			loger.Fatal(err)
		}
	}
}
func setImage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	var labimage map[string]interface{}
	json.Unmarshal(data, &labimage)
	imageName := labimage["image"]
	loger.Println(imageName)
	//写入数据库
	db, err := sql.Open("mysql", "root:abcd1234!@tcp(localhost:3306)/cloudlab?parseTime=true")
	if err != nil {
		loger.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`update labimage set imagename=? where title= ?`)
	if err != nil {
		loger.Fatalln(err)
	}
	res, err := stmt.Exec(imageName, varpac.Title)
	if err != nil {
		loger.Fatalln(err)
	}
	num, err := res.RowsAffected()
	check("update image", err)
	entry := Entry{}
	if num == 1 {
		entry.Code = "success"
	} else {
		entry.Code = "failed"
	}
	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}
	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}
func getLabImage(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:abcd1234!@tcp(localhost:3306)/cloudlab?parseTime=true")
	if err != nil {
		loger.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`select imagename from labimage where title= ?`)
	if err != nil {
		loger.Fatalln(err)
	}
	rows, err := stmt.Query(varpac.Title)
	if err != nil {
		loger.Fatalln(err)
	}
	var imagename string
	for rows.Next() {
		rows.Scan(&imagename)
	}
	entry := Entry{}
	if imagename != "" {
		entry.Code = "success"
		entry.Data = imagename
	} else {
		entry.Code = "failed"
	}
	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}
	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}
func init_student(w http.ResponseWriter, r *http.Request) {
	var labSession etcd.Session

	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	loger.Println(data)
	var user map[string]interface{}
	json.Unmarshal(data, &user)
	loger.Println(user)
	userid, ok := user["userid"].(string)
	if !ok {
		loger.Println("type assertion err")
	}
	loger.Println(userid)
	if userid == "" {
		return
	}
	entry := Entry{}
	if labSession.IsExist(userid) {
		labSession.Get(userid)
	} else {
		labSession.Status = "none"
		labSession.Url = "#"
	}
	entry.Code = "success"
	entry.Data = map[string]string{
		"status": labSession.Status,
		"url":    labSession.Url,
	}
	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}

	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}
func createContainer(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	loger.Println(data)
	var user map[string]interface{}
	json.Unmarshal(data, &user)
	loger.Println(user)
	userid, ok := user["userid"].(string)
	if !ok {
		loger.Println("type assertion err")
	}
	loger.Println(userid)
	if userid == "" {
		return
	}
	entry := Entry{}
	//创建容器请求，返回容器的url
	Resp, err := http.Get("http://" + varpac.Master.IP + ":9903/dispatch?userid=" + userid)

	loger.Println(err)
	if err != nil {
		entry.Code = "fail"
	} else {
		entry.Code = "success"
	}
	defer Resp.Body.Close()
	url, err := ioutil.ReadAll(Resp.Body)
	check("", err)
	// w.Write([]byte("http://" + string(url) + "?password=" + varpac.Password))
	// http.Redirect(w, r, "http://"+string(url)+"", http.StatusFound)
	loger.Println(url)

	entry.Data = "http://" + string(url) + "?password=" + varpac.Password
	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}

	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}

func listContainer(w http.ResponseWriter, r *http.Request) {

}

func checkpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	loger.Println(data)
	var user map[string]interface{}
	json.Unmarshal(data, &user)
	loger.Println(user)
	userid, ok := user["userid"].(string)
	if !ok {
		loger.Println("type assertion err")
	}
	loger.Println(userid)
	if userid == "" {
		return
	}
	entry := Entry{}
	var labSession etcd.Session
	labSession.Get(userid)
	//通过docker exec 执行保存eclipse的脚本
	saveCMD := "docker -H " + varpac.Master.IP + ":3375 " +
		"exec -i " + labSession.ConID + " bash -c \"export DISPLAY=:1 && bash /tempfiles/eclipseReload.sh\""

	loger.Println(saveCMD)
	_, err := exec.Command("/bin/bash", "-c", saveCMD).Output()

	if err != nil {
		loger.Println(err)
		entry.Code = "fail"
	} else {
		//暂停容器
		stopCMD := "docker -H " + varpac.Master.IP + ":3375 " +
			"stop " + labSession.ConID
		_, err = exec.Command("/bin/bash", "-c", stopCMD).Output()
		if err != nil {
			loger.Println(err)
			entry.Code = "fail"
		} else {
			entry.Code = "success"
			setStatus(userid, "saved")
		}
	}
	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}

	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}

func restore(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	loger.Println(data)
	var user map[string]interface{}
	json.Unmarshal(data, &user)
	loger.Println(user)
	userid, ok := user["userid"].(string)
	if !ok {
		loger.Println("type assertion err")
	}
	loger.Println(userid)
	if userid == "" {
		return
	}
	entry := Entry{}
	var labSession etcd.Session
	labSession.Get(userid)
	//启动容器
	startCMD := "docker -H " + varpac.Master.IP + ":3375 " +
		"start " + labSession.ConID
	_, err := exec.Command("/bin/bash", "-c", startCMD).Output()
	if err != nil {
		loger.Println(err)
		entry.Code = "fail"
	} else {
		entry.Code = "success"
		setStatus(userid, "created")
	}

	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}

	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}

//设置容器状态
func setStatus(userid string, status string) {
	etcd.SetStatus(userid, status)
}
func destroy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	loger.Println(data)
	var user map[string]interface{}
	json.Unmarshal(data, &user)
	loger.Println(user)
	userid, ok := user["userid"].(string)
	if !ok {
		loger.Println("type assertion err")
	}
	loger.Println(userid)
	if userid == "" {
		return
	}
	entry := Entry{}
	var labSession etcd.Session
	labSession.Get(userid)
	//暂停容器
	stopCMD := "docker -H " + varpac.Master.IP + ":3375 " +
		"stop " + labSession.ConID
	loger.Println(stopCMD)
	_, err := exec.Command("/bin/bash", "-c", stopCMD).Output()
	if err != nil {
		loger.Println(err)
		entry.Code = "fail"
	} else {
		//移除容器
		rmCMD := "docker -H " + varpac.Master.IP + ":3375 " +
			"rm " + labSession.ConID
		loger.Println(rmCMD)
		_, err = exec.Command("/bin/bash", "-c", rmCMD).Output()
		if err != nil {
			loger.Println(err)
		} else {
			entry.Code = "success"
			setStatus(userid, "none")
		}
	}

	resp, err := json.Marshal(entry)
	if err != nil {
		loger.Println(err)
	}

	loger.Println(string(resp))
	fmt.Fprint(w, string(resp))
}
func check(logstr string, err error) {
	if err != nil {
		loger.Fatal(logstr)
		loger.Println(err)
	}
}

func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{}) {

	err := templates[tmpl].Execute(w, locals)

	check("failed to execute template", err)
}
