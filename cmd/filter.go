package main

import (
	"bufio"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	pkg "gtmcdc"
	"os"
	"strings"
	"time"
)

//
// the main processing loop that reads journal extract and publish messages
//
func doFilter(fin, fout *os.File, brokers, topic string) {
	useKafka := false

	brokerList := strings.Split(brokers, ",")
	if len(brokerList) >= 1 && brokerList[0] != "off" {
		pkg.NewCDCProducer(brokerList)
		defer pkg.CleanupProducer()
		useKafka = true
	}

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := scanner.Text()
		pkg.IncrCounter("lines_read_from_input")

		rec, err := pkg.Parse(line)
		if err != nil {
			log.Infof("Unable to parse record %s", line)
			pkg.IncrCounter("lines_parse_error")
		} else {
			pkg.IncrCounter("lines_parsed")

			jstr := rec.Json()
			log.Debugf("line parsed to json |%s|=>|%s|", line, jstr)

			if useKafka {
				start := time.Now()

				err = pkg.PublishMessage(topic, jstr)
				if err != nil {
					log.Warnf("Unable to publish message for journal record %s", line)
					pkg.IncrCounter("lines_parsed_but_not_published")
				} else {
					pkg.IncrCounter("lines_parsed_and_published")
					elapsed := time.Since(start)
					pkg.HistoObserve("message_publish_to_kafka", float64(elapsed/time.Microsecond))
				}
			}

			// send to output only after a message is successfully published
			_, err = fmt.Fprintln(fout, line)
			if err != nil {
				pkg.IncrCounter("lines_output_write_error")
				log.Infof("Unable to write to output")
			} else {
				pkg.IncrCounter("lines_output_written")
			}
		}
	}
}

func initLogging(logFile, logLevel string) {
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't open log file for write.")
		os.Exit(1)
	}
	log.SetOutput(file)

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.InfoLevel
		log.Warnf("invalid loglevel %s, defaults to info", logLevel)
	}

	log.SetLevel(level)
}

func main() {
	var inputFile, outputFile, logFile, logLevel, brokers, topic, promHttpAddr string
	var devMode bool

	flag.StringVar(&inputFile, "i", "", "input file, default to STDIN")
	flag.StringVar(&outputFile, "o", "", "output file, default to STDOUT")
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

	if devMode {
		brokers = "off"
		promHttpAddr = "off"
		logLevel = "debug"
	}

	initLogging(logFile, logLevel)
	log.Infof("filter started with i=%s, o=%s, log=%s, loglevel=%s, brokers=%s, topic=%s, prom=%s",
		inputFile, outputFile, logFile, logLevel, brokers, topic, promHttpAddr)

	if promHttpAddr != "off" {
		pkg.InitPromHttp(promHttpAddr)
	}

	// initialize input and output files for this filter
	fin := os.Stdin
	fout := os.Stdout
	var err error

	// the primary reason for this read input from file logic is that
	// in my Goland IDE there's no way to redirect STDIN and STDOUT
	// in Run configuration, so I couldn't debug the code without this.

	if inputFile != "" {
		fin, err = os.OpenFile(inputFile, os.O_RDONLY, 0666)
		if err != nil {
			log.Fatalf("Unable to input file %s", inputFile)
		} else {
			log.Debugf("Input file: %s", inputFile)
		}
		defer fin.Close()
	}

	if outputFile != "" {
		fout, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to output file %s", outputFile)
		} else {
			log.Debugf("Output file: %s", outputFile)
		}
		defer fout.Close()
	}

	doFilter(fin, fout, brokers, topic)

	log.Error("doFilter returned, should have happened")
}
