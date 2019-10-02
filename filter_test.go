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

func Test_LoadConfig_Default(t *testing.T) {
	conf := LoadConfig("")
	assert.Equal(t, "off", conf.KafkaBrokerList)
	assert.Equal(t, "off", conf.PromHTTPAddr)
	assert.Equal(t, "debug", conf.LogLevel)
}

func Test_LoadConfig_Override(t *testing.T) {
	// write a test env file
	tmpFile, err := testTempFileWithContent([]byte("GTMCDC_KAFKA_BROKERS=myhost:10000"))
	assert.Nil(t, err)
	defer os.Remove(tmpFile)

	os.Setenv("GTMCDC_ENV", tmpFile)
	conf := LoadConfig("")

	assert.Equal(t, "myhost:10000", conf.KafkaBrokerList) // value from tmp file
	assert.Equal(t, "off", conf.PromHTTPAddr)             // default value
}

func Test_InitLogging(t *testing.T) {
	tmplog, err := testTempFileWithContent([]byte("GTMCDC_KAFKA_BROKERS=myhost:10000"))
	assert.Nil(t, err)
	defer os.Remove(tmplog)

	InitLogging(tmplog, "debug")
	log.Info("test logging")

	f, err1 := os.Open(tmplog)
	msg, err2 := ioutil.ReadAll(f)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.True(t, strings.Contains(string(msg), "test logging"))
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

	// the file contains 3 records
	// #1 is good
	// #2 cannot be parsed
	// #3 is a TCOM, the mock producer will fail when this
	//    message is published
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

func testTempFileWithContent(content []byte) (string, error) {
	tmpfile, err := ioutil.TempFile("", "tmp_test")
	if err != nil {
		return "", err
	}

	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	return tmpfile.Name(), nil
}
