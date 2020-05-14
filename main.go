package main

import (
	"fmt"
	"k8scheck/httpd"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/miiniper/loges"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("starting   server ...")
	viper.SetConfigName("conf")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		loges.Loges.Info("Config file changed: ", zap.Any("", e.Name))
	})

	httpd.Init()

	service, err := httpd.New(viper.GetString("server.ip"))
	if err != nil {
		panic(err)
	}
	err = service.Start()
	if err != nil {
		panic(err)
	}
	defer service.Close()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	<-terminate

}
