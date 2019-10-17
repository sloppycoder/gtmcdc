package main

import (
	"flag"
	pkg "gtmcdc"
	"os"

	log "github.com/sirupsen/logrus"
)

func closeFile(f *os.File) {
	if f != nil {
		_ = f.Close()
	}
}

func main() {
	var inputFile, outputFile, envFile string
	flag.StringVar(&inputFile, "i", "stdin", "input file")
	flag.StringVar(&outputFile, "o", "stdout", "output file")
	flag.StringVar(&envFile, "env", "", "config env file")
	flag.Parse()

	conf := pkg.LoadConfig(envFile)

	pkg.InitLogging(conf.LogFile, conf.LogLevel)
	log.Infof("Starting cdcfilter with conf=%s, i=%s, o=%s, %+v",
		envFile, inputFile, outputFile, conf)

	producer, err := pkg.InitProducer(conf.KafkaBrokerList, conf.KafkaTopic)
	if err != nil {
		log.Infof("Kafka producer not available. %v", err)
	}
	defer producer.CleanupProducer()

	if conf.PromHTTPAddr != "off" {
		err = pkg.InitPromHTTP(conf.PromHTTPAddr)
		if err != nil {
			log.Warn(err)
		}
	}

	fin, fout := pkg.InitInputAndOutput(inputFile, outputFile)
	defer closeFile(fin)
	defer closeFile(fout)

	metrics := pkg.InitMetrics()
	pkg.DoFilter(fin, fout, producer, metrics)

	log.Info("done")
}
