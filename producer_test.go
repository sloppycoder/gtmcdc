package gtmcdc

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	assert.Nil(t, err)
}
