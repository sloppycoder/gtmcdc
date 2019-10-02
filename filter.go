package gtmcdc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JeremyLoy/config"
	log "github.com/sirupsen/logrus"
)

// DefaultConfigFile is the default configuration file name
const DefaultConfigFile = "filter.env"

// Config stores the filter configurations
type Config struct {
	KafkaBrokerList string `config:"GTMCDC_KAFKA_BROKERS"`
	KafkaTopic      string `config:"GTMCDC_KAFKA_TOPIC"`
	PromHTTPAddr    string `config:"GTMCDC_PROM_HTTP_ADDR"`
	LogFile         string `config:"GTMCDC_LOG"`
	LogLevel        string `config:"GTMCDC_LOG_LEVEL"`
	InputFile       string `config:"GTMCDC_INPUT"`
	OutputFile      string `config:"GTMCDC_OUTPUT"`
	DevMode         bool   `config:"GTMCDC_DEVMODE"`
}

// DoFilter is the main processing loop that
//reads journal extract and publish messages
func DoFilter(fin, fout *os.File) {
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

			jsonstr := rec.JSON()
			log.Debugf("line parsed to json %s", jsonstr)

			if IsKafkaAvailable() {
				start := time.Now()

				err = PublishMessage(jsonstr)
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

// InitLogging initialize log output based on configuration
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

// InitInputAndOutput initialize input and output files for the filter
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

// LoadConfig loads the filter configurations from file
// devMode flags will override configurations file
func LoadConfig(configFile string, devMode bool) *Config {
	conf := Config{
		KafkaBrokerList: "off",
		PromHTTPAddr:    "off",
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
