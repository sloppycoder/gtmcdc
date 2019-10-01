package gtmcdc

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/Shopify/sarama/mocks"
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

func Test_DoFilter_MockKafka(t *testing.T) {
	defer func() {
		_ = os.Remove("tmp_output.txt")
	}()

	sp := mocks.NewSyncProducer(t, nil)
	defer CleanupProducer()

	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndFail(errors.New("send message failed"))
	SetProducer(sp)

	readCounter := GetCounterValue("lines_read_from_input")
	errorCounter := GetCounterValue("lines_parse_error")
	writeCounter := GetCounterValue("lines_output_written")
	messageCounter := GetCounterValue("lines_parsed_and_published")
	publishFailedCounter := GetCounterValue("lines_parsed_but_not_published")

	fin, fout := InitInputAndOutput("testdata/test1.txt", "tmp_output.txt")
	DoFilter(fin, fout)

	readDelta := GetCounterValue("lines_read_from_input") - readCounter
	errorDelta := GetCounterValue("lines_parse_error") - errorCounter
	writeDelta := GetCounterValue("lines_output_written") - writeCounter
	messageDelta := GetCounterValue("lines_parsed_and_published") - messageCounter
	publishFailedDelta := GetCounterValue("lines_parsed_but_not_published") - publishFailedCounter

	assert.Equal(t, 3.00, readDelta)
	assert.Equal(t, 1.00, errorDelta)
	assert.Equal(t, 2.00, writeDelta)
	assert.Equal(t, 1.00, messageDelta)
	assert.Equal(t, 1.00, publishFailedDelta)
}
