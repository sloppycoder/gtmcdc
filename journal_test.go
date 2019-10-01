package gtmcdc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Parse_Horolog_DAte(t *testing.T) {
	tt, err := parseHorologTime("0,0")
	assert.Nil(t, err)
	assert.Equal(t, 1841, tt.Year(), "year incorrect")

	tt, err = parseHorologTime("65282,59700")
	fmt.Println(tt)
	assert.Nil(t, err)
	assert.Equal(t, 2019, tt.Year())
	assert.Equal(t, "September", tt.Month().String())
	assert.Equal(t, 27, tt.Day())
	assert.Equal(t, 17, tt.Hour())
	assert.Equal(t, 39, tt.Minute())
	assert.Equal(t, 0, tt.Nanosecond())

	_, err = parseHorologTime(",")
	assert.NotNil(t, err)
}

func Test_Parse_JournalRecord_1(t *testing.T) {
	rec, _ := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^acc("00027")="300.00"`)
	// fmt.Println(rec)
	assert.Equal(t, "SET", rec.opcode)
	assert.Equal(t, "300.00", rec.detail.value)

	// record is too short
	_, err := Parse(`05\65282,59700\28`)
	assert.NotNil(t, err)

}

func Test__JournalRecord_Json(t *testing.T) {
	rec, err := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^acc("00027")="300.00"`)
	assert.Nil(t, err)

	jstr := rec.Json()
	expected := `{"operand":"SET","node":"^acc(\"00027\")","value":"300.00","token_seq":28,"update_num":0,"stream_num":0,"stream_seq":0,"journal_seq":0,"partners":""}`
	assert.Equal(t, expected, jstr)
}
