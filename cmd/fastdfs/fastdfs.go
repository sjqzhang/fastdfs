package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/luoyunpeng/go-fastdfs/internal/config"
	"github.com/luoyunpeng/go-fastdfs/internal/model"
	"github.com/luoyunpeng/go-fastdfs/internal/server"
	"github.com/luoyunpeng/go-fastdfs/pkg"
)

var (
	Version     string
	BuildTime   string
	GoVersion   string
	GitVersion  string
	versionInfo = flag.Bool("v", false, "display version")
)

func init() {
	flag.Parse()
	if *versionInfo {
		fmt.Printf("%s\n%s\n%s\n%s\n", Version, BuildTime, GoVersion, GitVersion)
		os.Exit(0)
	}

}

func main() {
	conf := config.NewConfig()
	model.Svr = model.NewServer(conf)
	model.Svr.InitComponent(false, conf)
	svr := model.Svr
	go func() {
		for {
			svr.CheckFileAndSendToPeer(pkg.GetToDay(), conf.Md5ErrorFile(), false, conf)
			//fmt.Println("CheckFileAndSendToPeer")
			time.Sleep(time.Second * time.Duration(conf.RefreshInterval()))
			//svr.pkg.RemoveEmptyDir(STORE_DIR)
		}
	}()
	go svr.CleanAndBackUp(conf)
	go model.CheckClusterStatus(conf)
	go svr.LoadQueueSendToPeer(conf)
	go svr.ConsumerPostToPeer(conf)
	go svr.ConsumerLog(conf)
	go svr.ConsumerDownLoad(conf)
	go svr.ConsumerUpload(conf)
	go svr.RemoveDownloading(conf)
	if conf.EnableFsNotify() {
		go svr.WatchFilesChange(conf)
	}
	//go svr.LoadSearchDict()
	if conf.EnableMigrate() {
		go svr.RepairFileInfoFromFile(conf)
	}
	if conf.AutoRepair() {
		go func() {
			for {
				time.Sleep(time.Minute * 3)
				svr.AutoRepair(false, conf)
				time.Sleep(time.Minute * 60)
			}
		}()
	}
	go func() { // force free memory
		for {
			time.Sleep(time.Minute * 1)
			debug.FreeOSMemory()
		}
	}()

	server.Start(conf)
}