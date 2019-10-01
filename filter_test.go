package gtmcdc

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_InitInputAndOutput(t *testing.T) {
	defer func() {
		_ = os.Remove("input.txt")
		_ = os.Remove("output.txt")
	}()

	fin, fout := InitInputAndOutput("stdin", "stdout")
	assert.Equal(t, os.Stdin, fin)
	assert.Equal(t, os.Stdout, fout)

	fin, fout = InitInputAndOutput("filter_test.go", "output.txt")
	assert.NotNil(t, fin)
	assert.NotNil(t, fout)
	assert.NotEqual(t, os.Stdin, fin)
	assert.NotEqual(t, os.Stdout, fout)
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

	conf := LoadConfig("./filter.env", false)

	assert.Equal(t, brokers, conf.KafkaBrokerList)        // env overrides config file
	assert.Equal(t, "localhost:10101", conf.PromHttpAddr) // from config file
}

func Test_InitLogging(t *testing.T) {
	tmplog := "temp.log"
	tmpmsg := "test logging"
	defer func() {
		_ = os.Remove(tmplog)
	}()

	InitLogging(tmplog, "debug")
	log.Info(tmpmsg)
	f, err1 := os.Open(tmplog)
	msg, err2 := ioutil.ReadAll(f)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.True(t, strings.Contains(string(msg), tmpmsg))
}

func Test_DoFilter_NoKafka(t *testing.T) {
	defer func() {
		_ = os.Remove("tmp_output.txt")
	}()

	fin, fout := InitInputAndOutput("testdata/test1.txt", "tmp_output.txt")
	DoFilter(fin, fout, "off", "")

	assert.Equal(t, 6.00, GetCounterValue("lines_read_from_input"))
	assert.Equal(t, 1.00, GetCounterValue("lines_parse_error"))
	assert.Equal(t, 5.00, GetCounterValue("lines_output_written"))
}

func Test_DefaultConfigFile(t *testing.T) {
	assert.True(t, strings.Contains(DefaultConfigFile(), "/filter.env"))
}
