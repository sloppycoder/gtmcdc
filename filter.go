package gtmcdc

import (
	"bufio"
	"fmt"
	"github.com/JeremyLoy/config"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokerList string `config:"GTMCDC_KAFKA_BROKERS"`
	KafkaTopic      string `config:"GTMCDC_KAFKA_TOPIC"`
	PromHttpAddr    string `config:"GTMCDC_PROM_HTTP_ADDR"`
	LogFile         string `config:"GTMCDC_LOG"`
	LogLevel        string `config:"GTMCDC_LOG_LEVEL"`
	InputFile       string `config:"GTMCDC_INPUT"`
	OutputFile      string `config:"GTMCDC_OUTPUT"`
	DevMode         bool   `config:"GTMCDC_DEVMODE"`
}

//
// the main processing loop that reads journal extract and publish messages
//
func DoFilter(fin, fout *os.File, brokers, topic string) {
	useKafka := false

	brokerList := strings.Split(brokers, ",")
	if len(brokerList) >= 1 && brokerList[0] != "off" {
		NewCDCProducer(brokerList)
		defer CleanupProducer()
		useKafka = true
	}

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := scanner.Text()
		IncrCounter("lines_read_from_input")

		rec, err := Parse(line)
		if err != nil {
			log.Info("Unable to parse record")
			IncrCounter("lines_parse_error")
		} else {
			IncrCounter("lines_parsed")

			jsonstr := rec.Json()
			log.Debugf("line parsed to json %s", jsonstr)

			if useKafka {
				start := time.Now()

				err = PublishMessage(topic, jsonstr)
				if err != nil {
					log.Warn("Unable to publish message for journal record")
					IncrCounter("lines_parsed_but_not_published")
				} else {
					IncrCounter("lines_parsed_and_published")
					elapsed := time.Since(start)
					HistoObserve("message_publish_to_kafka", float64(elapsed/time.Microsecond))
				}
			}

			// send to output only after a message is successfully published
			_, err = fmt.Fprintln(fout, line)
			if err != nil {
				IncrCounter("lines_output_write_error")
				log.Infof("Unable to write to output")
			} else {
				IncrCounter("lines_output_written")
			}
		}
	}
}

func InitLogging(logFile, logLevel string) {
	var file *os.File
	var err error

	if strings.EqualFold(logFile, "stderr") {
		file = os.Stderr
	} else {
		file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("can't open log file for write.")
		}
	}

	log.SetOutput(file)

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.InfoLevel
		log.Warnf("invalid loglevel %s, defaults to info", logLevel)
	}

	log.SetLevel(level)
}

func InitInputAndOutput(inputFile, outputFile string) (*os.File, *os.File) {
	// initialize input and output files for this filter
	fin := os.Stdin
	fout := os.Stdout
	var err error

	// the primary reason for this read input from file logic is that
	// in my Goland IDE there's no way to redirect STDIN and STDOUT
	// in Run configuration, so I couldn't debug the code without this.

	if inputFile != "" && !strings.EqualFold(inputFile, "stdin") {
		fin, err = os.OpenFile(inputFile, os.O_RDONLY, 0666)
		if err != nil {
			log.Fatalf("Unable to input file %s", inputFile)
		} else {
			log.Debugf("Input file: %s", inputFile)
		}
	}

	if outputFile != "" && !strings.EqualFold(outputFile, "stdout") {
		fout, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Unable to output file %s", outputFile)
		} else {
			log.Debugf("Output file: %s", outputFile)
		}
	}

	return fin, fout
}

func LoadConfig(configFile string, devMode bool) *Config {
	// defaults
	conf := Config{
		KafkaBrokerList: "off",
		PromHttpAddr:    "off",
		LogFile:         "filter.log",
		LogLevel:        "debug",
	}

	if devMode {
		return &conf
	}

	config.From(configFile).FromEnv().To(&conf)

	if conf.InputFile == "" {
		conf.InputFile = "stdin"
	}

	if conf.OutputFile == "" {
		conf.OutputFile = "stdout"
	}

	if conf.LogFile == "" {
		conf.LogFile = "filter.log"
	}

	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}

	return &conf
}

func DefaultConfigFile() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir + "/filter.env"
}
