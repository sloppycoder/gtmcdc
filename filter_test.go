package gtmcdc

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

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
	assert.Equal(t, "off", conf.PromHTTPAddr)
	assert.Equal(t, "debug", conf.LogLevel)
}

func Test_LoadConfig_Default(t *testing.T) {
	brokers := "anyhost:1000"
	_ = os.Setenv("GTMCDC_KAFKA_BROKERS", brokers)

	conf := LoadConfig("./filter.env", false)

	assert.Equal(t, brokers, conf.KafkaBrokerList)        // env overrides config file
	assert.Equal(t, "localhost:10101", conf.PromHTTPAddr) // from config file
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
	sp := mocks.NewSyncProducer(t, nil)
	defer CleanupProducer()

	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndFail(errors.New("send message failed"))
	SetProducer(sp)

	counters := []string{
		"lines_read_from_input",
		"lines_parse_error",
		"lines_output_written",
		"lines_parsed_and_published",
		"lines_parsed_but_not_published",
	}

	prevValues := getCounters(counters)

	fin, fout := InitInputAndOutput("testdata/test1.txt", nullFile())
	DoFilter(fin, fout)

	currentValues := getCounters(counters)
	deltas, err := deltaCounters(prevValues, currentValues)

	assert.Nil(t, err)
	expected := []float64{3.0, 1.0, 2.0, 1.0, 1.0}
	assert.ElementsMatch(t, expected, deltas)
}

func getCounters(counterNames []string) []float64 {
	values := make([]float64, len(counterNames))
	for i, name := range counterNames {
		values[i] = GetCounterValue(name)
	}

	return values
}

func deltaCounters(prev, current []float64) ([]float64, error) {
	if prev == nil || current == nil || len(prev) != len(current) {
		return nil, errors.New("invalid input")
	}

	deltas := make([]float64, len(prev))
	for i := 0; i < len(prev); i++ {
		deltas[i] = current[i] - prev[i]
	}

	return deltas, nil
}

func nullFile() string {
	if runtime.GOOS == "widnows" {
		return "NUL"
	}
	return "/dev/null"
}
