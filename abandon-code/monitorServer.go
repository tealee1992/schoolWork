
package monitor

import (
	"fmt"
	"io/ioutil"
	"net/http" //http server ,http client
	//log "github.com/Sirupsen/logrus"
	//"github.com/codegangsta/negroni"
	//"github.com/gorilla/context"
	"github.com/gorilla/mux"
	/*Package gorilla/mux implements a request router
	and dispatcher for matching incoming requests to
	their respective handler.
	like http.ServeMux*/
	//"github.com/mailgun/oxy/forward"
	"github.com/vulcand/oxy/forward"
	/*
	new addr "vulcand/oxy/forward" Oxy is a Go library with HTTP handlers
	forwards requests to remote location and rewrites headers
	*/
	"context"
)
const (
	leaderServer = "11.0.57.2"
)
type (
	Api struct{
		listenAddr  string
		authwhitelistCIDRs []string
		enableCors	bool
		serverVersion	string
		allowInsecure	bool
		tlsCACertPath	string
		tlsCertPath	string
		tlsKeyPath 	string
		dUrl	string
		fwd	*forward.Forwarder
	}
	ApiConfig struct{
		ListenAddr	string
		AuthWhiteListCIDRs []string
		EnableCORS         bool
		AllowInsecure      bool
		TLSCACertPath      string
		TLSCertPath        string
		TLSKeyPath         string
	}
	
	Credentials struct{
		Username string `json:"username,omitempty"`
		Passward string `json:"passward,omitempty"`
	}
)

func NewApi(config ApiConfig) (*Api, error){
	return &Api{
		listenAddr:	config.ListenAddr,
		authwhitelistCIDRs:	config.AuthWhiteListCIDRs,
		enableCors:	config.EnableCORS,
		allowInsecure:	config.AllowInsecure,
		tlsCertPath:	config.TLSCertPath,
		tlsKeyPath:	config.TLSKeyPath,
		tlsCACertPath:	config.TLSCACertPath,
	},nil
}

func (a *Api) Run() error {

	globalMux := http.NewServeMux()

	//forward for swarm
	var err error
	a.fwd, err = forward.New()
	if err != nil{
		return err
	}

	swarmRedirect := http.HandlerFunc(a.swarmRedirect)

	apiRouter := mux.NewRouter()

	//global handler
	globalMux.Handle("/",http.FileServer(http.Dir("static")))

	//swarm
	swarmRouter := mux.NewRouter()
	// these are pulled from the swarm api code to proxy and allow
	// usage with the standard Docker cli
	m := map[string]map[string]http.HandleFunc{
		"GET":{
			"/_ping":	swarmRedirect,
			"/events":	swarmRedirect,
			"/info":	swarmRedirect,
			"/info":	swarmRedirect,
			"/version":	swarmRedirect,
			"/images/json":	swarmRedirect,
			"/images/viz":	swarmRedirect,
			"/images/search":	swarmRedirect,
			"/images/get":                     swarmRedirect,
			"/images/{name:.*}/get":           swarmRedirect,
			"/images/{name:.*}/history":       swarmRedirect,
			"/images/{name:.*}/json":          swarmRedirect,
			"/containers/ps":                  swarmRedirect,
			"/containers/json":                swarmRedirect,
			"/containers/{name:.*}/export":    swarmRedirect,
			"/containers/{name:.*}/changes":   swarmRedirect,
			"/containers/{name:.*}/json":      swarmRedirect,
			"/containers/{name:.*}/top":       swarmRedirect,
			"/containers/{name:.*}/logs":      swarmRedirect,
			"/containers/{name:.*}/stats":     swarmRedirect,
			"/containers/{name:.*}/attach/ws": swarmHijack,
			"/exec/{execid:.*}/json":          swarmRedirect,
		},
		"POST": {
			"/auth":                         swarmRedirect,
			"/commit":                       swarmRedirect,
			"/build":                        swarmRedirect,
			"/images/create":                swarmRedirect,
			"/images/load":                  swarmRedirect,
			"/images/{name:.*}/push":        swarmRedirect,
			"/images/{name:.*}/tag":         swarmRedirect,
			"/containers/create":            swarmRedirect,
			"/containers/{name:.*}/kill":    swarmRedirect,
			"/containers/{name:.*}/pause":   swarmRedirect,
			"/containers/{name:.*}/unpause": swarmRedirect,
			"/containers/{name:.*}/rename":  swarmRedirect,
			"/containers/{name:.*}/restart": swarmRedirect,
			"/containers/{name:.*}/start":   swarmRedirect,
			"/containers/{name:.*}/stop":    swarmRedirect,
			"/containers/{name:.*}/wait":    swarmRedirect,
			"/containers/{name:.*}/resize":  swarmRedirect,
			"/containers/{name:.*}/attach":  swarmHijack,
			"/containers/{name:.*}/copy":    swarmRedirect,
			"/containers/{name:.*}/exec":    swarmRedirect,
			"/exec/{execid:.*}/start":       swarmHijack,
			"/exec/{execid:.*}/resize":      swarmRedirect,
		},
		"DELETE": {
			"/containers/{name:.*}": swarmRedirect,
			"/images/{name:.*}":     swarmRedirect,
		},
		"OPTIONS": {
			"": swarmRedirect,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				if a.enableCors {
					writeCorsHeaders(w, r)
				}
				localFct(w, r)
			}
			localMethod :=method
			//add the new route
			swarmRouter.Path("/v{version:[0-9.]+}"+localRoute).Methods()(localMethod).HandlerFunc(wrap)
			swarmRouter.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}
	s := &http.Server{
		Addr: a.listenAddr,
		Handler:context.ClearHandler(globalMux),
	}
	var runErr error
	runErr = s.ListenAndServe()
	return runErr
}