package gtmcdc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_InitProducer(t *testing.T) {
	// clear side effect of previous test cases
	SetProducer(nil)

	err := InitProducer("off", "blah")
	assert.NotNil(t, err)
	assert.False(t, IsKafkaAvailable())
}

func Test_InitPromHttp(t *testing.T) {
	err := InitPromHttp(":0")
	time.Sleep(100 * time.Millisecond) // slight delay to allow gather test coverage
	assert.Nil(t, err)
}
