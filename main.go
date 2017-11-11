package main

/*server 入口*/
import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"runtime/debug"
	"varpac"
)

var templates = make(map[string]*template.Template)

func init() {
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
		log.Println("Loading template:", templatesPath)
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
		log.Fatal("ListenAndServe:", err.Error())
	}

}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Printf("warn: panic in %v - %v", fn, e)

				log.Println(string(debug.Stack()))
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
func hasImage() {
	db, err := sql.Open("mysql", "root:abcd1234!@tcp(localhost:3306)/cloudlab?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("select * from labimage")
	if err != nil {
		log.Fatalln(err)
	}
	columns, _ := rows.Columns()
	if len(columns) == 0 {
		stmt, err := db.Prepare(`insert labimage (title,imagename) values (?,?)`)
		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(varpac.title, "")
		if err != nil {
			log.Fatal(err)
		}
	}
}
func setImage(w http.ResponseWriter, r *http.Request) {
	imageName := r.FormValue("imagename")
	//写入数据库
	db, err := sql.Open("mysql", "root:abcd1234!@tcp(localhost:3306)/cloudlab?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`update labimage set imagename=? where title= ?`)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := stmt.Exec(imageName, varpac.title)
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatal(imageName)
}
func createContainer(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("userid")
	//创建容器请求，返回容器的url
	resp, err := http.Get(varpac.Master.IP + ":9092/dispatch?userid=" + userid)

	check("failed to dispatch new container ", err)

	defer resp.Body.Close()
	url, err := ioutil.ReadAll(resp.Body)
	check("", err)
	w.Write([]byte("http://" + string(url) + "?password=" + varpac.pasword))
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
		log.Fatal(logstr)
		panic(err)
	}
}

func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{}) {

	err := templates[tmpl].Execute(w, locals)

	check("failed to execute template", err)
}
