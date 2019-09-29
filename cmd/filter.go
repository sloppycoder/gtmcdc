package main

import (
	"bufio"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	pkg "gtmcdc"
	"os"
	"strings"
)

//
// the main processing loop that reads journal extract and publish messages
//
func doFilter(input, output, brokers, topic string, useKafka bool) {
	var fin, fout *os.File
	var err error

	// the primary reason for this read input from file logic is that
	// in my Goland IDE there's no way to redirect STDIN and STDOUT
	// in Run configuration, so I couldn't debug the code without this.

	if input != "" {
		fin, err = os.OpenFile(input, os.O_RDONLY, 0666)
		if err != nil {
			log.Fatalf("Unable to input file %s", input)
		} else {
			log.Debugf("Input file: %s", input)
		}
		defer fin.Close()
	} else {
		fin = os.Stdin
	}

	if output != "" {
		fout, err = os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to output file %s", output)
		} else {
			log.Debugf("Output file: %s", output)
		}
		defer fout.Close()
	} else {
		fout = os.Stdout
	}

	if useKafka {
		brokerList := strings.Split(brokers, ",")
		pkg.NewCDCProducer(brokerList)
		defer pkg.CleanupProducer()
	}

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := scanner.Text()
		rec, err := pkg.Parse(line)
		if err != nil {
			log.Info("Unable to parse record %s", line)
		} else {
			if useKafka {
				err = pkg.PublishMessage(topic, rec.Json())
				if err != nil {
					log.Warnf("Unable to publish message for journal record %s", line)
				}
			}
			// send to output only after a message is successfully published
			fmt.Fprintln(fout, line)
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
		fmt.Fprintln(os.Stderr, "can't open log file for write.")
		os.Exit(1)
	}
	log.SetOutput(file)
}

func main() {
	var inputFile, outputFile, logFile, brokers, topic string
	var useKafka bool

	flag.StringVar(&inputFile, "i", "", "input file, default to STDIN")
	flag.StringVar(&outputFile, "o", "", "output file, default to STDOUT")
	flag.StringVar(&logFile, "log", "filter.log", "log file, default to filter.log")
	flag.StringVar(&brokers, "brokers", "localhost:9092", "Kafka broker list, default to localhost:9092")
	flag.StringVar(&topic, "topic", "cdc-test", "Kafka topic to publish events to, default to cdc-test")
	flag.BoolVar(&useKafka, "kafka", false, "Enable publishing to Kafa, defaults to false")
	flag.Parse()

	initLogging(logFile)
	log.Infof("filter started with i=%s, o=%s, log=%s, brokers=%s, topic=%s, kafka=%t",
		inputFile, outputFile, logFile, brokers, topic, useKafka)

	doFilter(inputFile, outputFile, brokers, topic, useKafka)
}
