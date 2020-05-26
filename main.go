package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"

	"awesomeProject/beacon/mqtt-benchmark-sn/common"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	log.SetOutput(colorable.NewColorableStdout())
}

func processConfiguration(cont *cli.Context) (*common.Config, error) {
	path := cont.String("config")
	cfg, err := common.LoadConf(path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return cfg, nil
}

func showConfiguration(cont *cli.Context) {
	_, _ = processConfiguration(cont)
}

func runBenchmark(cont *cli.Context) {
	cfg, _ := processConfiguration(cont)

	ctd := cfg.CalculateData()
	wm := common.NewWorkerMetrics()
	pts := common.NewPublishRunTimeStats()
	sts := common.NewSubscriberThroughputStats()

	if cfg.General.Debug {
		go wm.DebugInfo()
		time.Sleep(1 * time.Second)
	}

	wg := sync.WaitGroup{}
	stopChan := make(chan bool, 1)

	common.StartSubscriberWorker(sts, wm, &wg, &cfg.MqttSubscriber, &cfg.MqttTopic, &stopChan, ctd)
	time.Sleep(2 * time.Second)
	common.StartPublisherWorker(pts, wm, &wg, &cfg.MqttPublisher, &cfg.MqttTopic, &stopChan, ctd)
	wg.Wait()

	pts.LogInfoForSummary()
	sts.LogInfoForSummary(wm)

	mode := ""
	if cfg.MqttPublisher.Enablestaticmessage {
		mode = " ****** Enable Static Content Messages Mode ****** \n"
	}
	line := "**********************************************************"
	broker := fmt.Sprintf("****** Benchmark broker : '%s' ****** \n%s%s", cfg.General.Brokername, mode, line)
	log.Info(wm.GetSummaryInfo(ctd) + broker)
}

func main() {
	app := cli.NewApp()
	app.Name = "MQTT Benchmark"
	app.Version = "7.1.20"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name: "Abe.Chua",
		},
	}
	app.Usage = "A command-line benchmark tool for the MQTT broker."

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Run the benchmark jobs.",
			Action: runBenchmark,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Configuration file path",
					Value: "./conf/mbt.ini",
				},
			},
		},
		{
			Name:   "show",
			Usage:  "Print the configuration information.",
			Action: showConfiguration,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Configuration file path",
					Value: "./conf/mbt.ini",
				},
			},
		},
	}

	_ = app.Run(os.Args)
}
