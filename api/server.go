package api

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Config struct {
	Addr string `default:"localhost:8000"`
}

var server = martini.Classic()

func init() {
	server.Use(render.Renderer(render.Options{
		Charset: "UTF-8"}))
}

func RunServer(cfg *Config) {
	server.RunOnAddr(cfg.Addr)
}
