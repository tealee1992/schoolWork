package main

/*server 入口*/
import (
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

	http.HandleFunc("/containers/set", safeHandler(setImage))
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
func setImage(w http.ResponseWriter, r *http.Request) {
	imageName := r.FormValue("imagename")
	//写入数据库
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

	http.Redirect(w, r, "http://"+string(url)+"", http.StatusFound)
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