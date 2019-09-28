package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gtmcdc"
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

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		gtmcdc.Parse(line)
	}
}
