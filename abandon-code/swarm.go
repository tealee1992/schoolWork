package monitor

import (
	"bytes"
	"net/http"
	"net/url"
)

func (a *Api)  swarmRedirect(w http.ResponseWriter, req *http.Request){
	var err error
	req.URL,err=url.ParseRequestURI(a.dUrl)
	if err != nil{
		http.Error(w, err.Error(),http.StatusInternalServerError)
		return
	}
	a.fwd.ServeHTTP(w,req)
}

type  proxyWrite struct{
	Body	*bytes.Buffer
	Headers	*map[string][]string
	StatusCode	*int
}

func (p proxyWrite) Header() http.Header  {
	return *p.Headers
}
func (p proxyWrite) Write(data []byte) (int, error){
	return p.Body.Write(data)
}
func (p proxyWrite) WriteHeader(code int)  {
	*p.StatusCode = code
}