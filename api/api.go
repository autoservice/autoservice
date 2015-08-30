package api

import (
	"net/http"
)

func init() {
	server.Post("/register", RegisterService)
	server.Post("/unregister", UnRegisterService)
	server.Get("/query", QueryService)
}

func RegisterService(req *http.Request) {

}

func UnRegisterService(req *http.Request) {

}

func QueryService(req *http.Request) {

}
