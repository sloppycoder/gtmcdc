package gtmcdc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_InitProducer(t *testing.T) {
	p, err := InitProducer("off", "blah")
	assert.NotNil(t, err)
	assert.False(t, p.IsKafkaAvailable())
}

func Test_InitPromHttp(t *testing.T) {
	err := InitPromHTTP(":0")
	time.Sleep(100 * time.Millisecond) // slight delay to allow gather test coverage
	assert.Nil(t, err)
}
