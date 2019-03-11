// Start an Apigw.
package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/apigw/gateway"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Version as set during build.
	Version string

	// cli flags
	configPath = flag.String("config", "config.yaml",
		`config file path.`)
)

func main() {
	defer glog.Flush()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.Parse() // glog needs flag otherwise it will prefix 'ERROR: logging before flag.Parse:' to each message.

	s := fmt.Sprintf("Start gateway %s ", Version)
	pflag.VisitAll(func(flag *pflag.Flag) {
		s = fmt.Sprintf("%s %s=%q", s, flag.Name, flag.Value)
	})
	glog.Info(s)

	// Read yaml
	configYaml, err := ioutil.ReadFile(*configPath)
	if err != nil {
		glog.Exit(err)
	}

	// Read config file.
	cfg := &gateway.Config{}
	err = yaml.Unmarshal(configYaml, cfg)
	if err != nil {
		glog.Exit("parsing config", configPath, err)
	}

	//TODO Use contour workgroup or https://github.com/oklog/run to manage go routines

	// Create gateway.
	gw := gateway.NewWithConfig(cfg)
	defer gw.ShutdownWithTimeout(10 * time.Second)
	go func() {
		err := gw.Run()
		if err != nil {
			glog.Fatal("running gateway", err)
		}
	}()

	// Create Prometheus endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		glog.Fatal(http.ListenAndServe(cfg.Management.Bind, nil))
	}()

	// Wait for SIGINT
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT)
	<-c
}
