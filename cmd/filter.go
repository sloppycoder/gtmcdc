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
func doFilter(fin, fout *os.File, brokers, topic string, useKafka bool) {

	if useKafka {
		brokerList := strings.Split(brokers, ",")
		pkg.NewCDCProducer(brokerList)
		defer pkg.CleanupProducer()
	}

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := scanner.Text()
		pkg.IncrCounter("lines_read_from_input")

		rec, err := pkg.Parse(line)
		if err != nil {
			log.Info("Unable to parse record %s", line)
			pkg.IncrCounter("lines_parse_error")
		} else {
			pkg.IncrCounter("lines_parsed")

			if useKafka {
				start := time.Now()
				err = pkg.PublishMessage(topic, rec.Json())

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
			} else {
				pkg.IncrCounter("lines_output_written")
			}
		}
	}
}

//  DO NOT DELETE, will be restored later when configuration becomes more complex
//
// 	"github.com/olebedev/config"
// func readConfig(configFile string) *config.Config {
//	f, err := ioutil.ReadFile(configFile)
//	if err != nil {
//		log.Fatalf("Unable to read config file %s", configFile)
//	}
//	yamlString := string(f)
//
//	conf, err := config.ParseYaml(yamlString)
//	return conf
//}

func initLogging(logFile string) {
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can't open log file for write.")
		os.Exit(1)
	}
	log.SetOutput(file)
	log.SetLevel(log.DebugLevel)
}

func main() {
	var inputFile, outputFile, logFile, brokers, topic string
	var noKafka bool
	var promHttpAddr string

	flag.StringVar(&inputFile, "i", "", "input file, default to STDIN")
	flag.StringVar(&outputFile, "o", "", "output file, default to STDOUT")
	flag.StringVar(&logFile, "log", "filter.log", "log file, default to filter.log")
	flag.StringVar(&brokers, "brokers", "localhost:9092", "Kafka broker list, default to localhost:9092")
	flag.StringVar(&topic, "topic", "cdc-test", "Kafka topic to publish events to, default to cdc-test")
	flag.BoolVar(&noKafka, "nokafka", false, "Enable publishing to Kafka, defaults to false")
	flag.StringVar(&promHttpAddr, "prom", "127.0.0.1:10101",
		`expose metrics on this address to be scraped by Prometheus, 
defaults to 127.0.0.1:101010. specify off to disable the HTTP listener`)

	flag.Parse()

	initLogging(logFile)
	log.Infof("filter started with i=%s, o=%s, log=%s, brokers=%s, topic=%s, nokafka=%t, prom=%s",
		inputFile, outputFile, logFile, brokers, topic, noKafka, promHttpAddr)

	if promHttpAddr != "off" {
		pkg.InitPromHttp(promHttpAddr)
	}

	// initialize input and output files for this filter

	var fin, fout *os.File
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
	} else {
		fin = os.Stdin
	}

	if outputFile != "" {
		fout, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to output file %s", outputFile)
		} else {
			log.Debugf("Output file: %s", outputFile)
		}
		defer fout.Close()
	} else {
		fout = os.Stdout
	}

	doFilter(fin, fout, brokers, topic, !noKafka)
}
