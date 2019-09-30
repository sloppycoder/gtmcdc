package gtmcdc

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_InitInputAndOutput(t *testing.T) {
	fin, fout := InitInputAndOutput("stdin", "stdout")
	assert.Equal(t, os.Stdin, fin)
	assert.Equal(t, os.Stdout, fout)
}

func Test_LoadConfig_DevMode(t *testing.T) {
	conf := LoadConfig("_not_exist", true)
	assert.Equal(t, "off", conf.KafkaBrokerList)
	assert.Equal(t, "off", conf.PromHttpAddr)
	assert.Equal(t, "debug", conf.LogLevel)
}

func Test_LoadConfig_Default(t *testing.T) {
	brokers := "anyhost:1000"
	_ = os.Setenv("GTMCDC_KAFKA_BROKERS", brokers)

	conf := LoadConfig(DefaultConfigFile, false)

	assert.Equal(t, brokers, conf.KafkaBrokerList)        // env overrides config file
	assert.Equal(t, "localhost:10101", conf.PromHttpAddr) // from config file
	assert.Equal(t, "info", conf.LogLevel)                // default not in config file
}
