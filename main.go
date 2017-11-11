package main

/*server 入口*/
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"varpac"
)

var templates = make(map[string]*template.Template)
var logFile *os.File
var loger *log.Logger

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
	http.HandleFunc("/containers/create", safeHandler(createContainer))
	http.HandleFunc("/containers/list", safeHandler(listContainer))
	http.HandleFunc("/containers/checkpoint", safeHandler(checkpoint))
	http.HandleFunc("/containers/restore", safeHandler(restore))
	http.HandleFunc("/", safeHandler(homePage))
	err := http.ListenAndServe(":9090", nil)
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
	columns, _ := rows.Columns()
	if len(columns) == 0 {
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
	imageName := r.FormValue("imagename")
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
	_, err = stmt.Exec(imageName, varpac.Title)
	if err != nil {
		loger.Fatalln(err)
	}
	loger.Fatal(imageName)
}
func createContainer(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("userid")
	//创建容器请求，返回容器的url
	resp, err := http.Get(varpac.Master.IP + ":9092/dispatch?userid=" + userid)

	check("failed to dispatch new container ", err)

	defer resp.Body.Close()
	url, err := ioutil.ReadAll(resp.Body)
	check("", err)
	w.Write([]byte("http://" + string(url) + "?password=" + varpac.Password))
	// http.Redirect(w, r, "http://"+string(url)+"", http.StatusFound)
}

func listContainer(w http.ResponseWriter, r *http.Request) {

}

func checkpoint(w http.ResponseWriter, r *http.Request) {

}

func restore(w http.ResponseWriter, r *http.Request) {

}
func check(logstr string, err error) {
	if err != nil {
		loger.Fatal(logstr)
		panic(err)
	}
}

func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{}) {

	err := templates[tmpl].Execute(w, locals)

	check("failed to execute template", err)
}
