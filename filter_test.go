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
