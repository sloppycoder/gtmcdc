package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	pkg "gtmcdc"
	"os"
)

func main() {
	// initialize log file
	file, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't open log file for write.")
		os.Exit(1)
	}
	log.SetOutput(file)

	f, _ := os.OpenFile("/home/lee/Projects/git/ydbtests/repl_procedures/B/filter.log", os.O_RDONLY, 0600)
	defer f.Close()

	pkg.NewCDCProducer([]string{"localhost:9092"})
	defer pkg.CleanupProducer()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		rec, err := pkg.Parse(line)
		if err != nil {
			log.Info("Unable to parse record %s", line)
		} else {
			err = pkg.PublishMessage("cdc-test", rec.Json())
			if err != nil {
				log.Warnf("Unable to publish message for journal record %s", line)
			}
		}
	}
}
