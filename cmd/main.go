package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	pkg "gtmcdc"
	"os"
)

func main() {

	var inputFile, outputFile, logFile, logLevel, brokers, topic, promHttpAddr string
	var devMode bool

	flag.StringVar(&inputFile, "i", "stdin", "input file")
	flag.StringVar(&outputFile, "o", "stdout", "output file")
	flag.StringVar(&logFile, "log", "filter.log", "log file")
	flag.StringVar(&logLevel, "loglevel", "debug", "log level")
	flag.StringVar(&brokers, "brokers", "localhost:9092", "Kafka broker list")
	flag.StringVar(&topic, "topic", "cdc-test", "Kafka topic to publish events to")
	flag.StringVar(&promHttpAddr, "prom", "127.0.0.1:10101",
		`expose metrics on this address to be scraped by Prometheus, 
specify "off" to disable the HTTP listener
`)
	// this is for developer use only
	flag.BoolVar(&devMode, "dev", false, "Developer mode, internal use only.")

	flag.Parse()

	if promHttpAddr != "off" {
		pkg.InitPromHttp(promHttpAddr)
	}

	fin, fout := pkg.InitInputAndOutput(inputFile, outputFile)
	defer func(fin, fout *os.File) {
		if fin != nil {
			_ = fin.Close()
		}

		if fout != nil {
			_ = fout.Close()
		}
	}(fin, fout)

	pkg.DoFilter(fin, fout, brokers, topic)

	log.Error("doFilter returned, should have happened")
}
