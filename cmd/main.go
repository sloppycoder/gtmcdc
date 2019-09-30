package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	pkg "gtmcdc"
	"os"
)

func closeFile(f *os.File) {
	if f != nil {
		_ = f.Close()
	}
}

func main() {
	var inputFile, configFile string
	var devMode bool
	flag.StringVar(&inputFile, "i", "", "input file")
	flag.StringVar(&configFile, "conf", pkg.DefaultConfigFile, "filter config file")
	flag.BoolVar(&devMode, "dev", false, "Developer mode, internal use only.")
	flag.Parse()

	if os.Getenv("GTMCDC_DEVMODE") != "" {
		devMode = true
	}

	conf := pkg.LoadConfig(configFile, devMode)
	// input file from command line overrides config file
	if inputFile == "" {
		conf.InputFile = inputFile
	}

	pkg.InitLogging(conf.LogFile, conf.LogLevel)
	log.Infof("Starting filter with dev=%t, conf=%s, %v", devMode, configFile, conf)

	if conf.PromHttpAddr != "off" {
		pkg.InitPromHttp(conf.PromHttpAddr)
	}

	fin, fout := pkg.InitInputAndOutput(conf.InputFile, conf.OutputFile)
	defer closeFile(fin)
	defer closeFile(fout)

	pkg.DoFilter(fin, fout, conf.KafkaBrokerList, conf.KafkaTopic)

	log.Error("doFilter returned, should not have happened")
}
