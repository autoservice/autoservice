package main

import (
	"flag"
	"os"

	"github.com/codingxyz/autoservice/api"
	"github.com/codingxyz/autoservice/db"
	"github.com/codingxyz/autoservice/utils"
	"github.com/golang/glog"
)

type Config struct {
	API *api.Config
	DB  *db.Config
}

var config = flag.String("config", "config.yaml", "config file path")

func main() {
	flag.Parse()

	var cfg Config
	if err := utils.LoadConfigFromFile(&cfg, *config); err != nil {
		glog.Errorf("load config file fail: %v", err)
		os.Exit(-1)
	}
	if err := db.InitDB(cfg.DB); err != nil {
		glog.Errorf("init db fail: %v", err)
		os.Exit(-1)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			glog.Errorf("close db fail: %v", err)
		}
	}()
	api.RunServer(cfg.API)
}
