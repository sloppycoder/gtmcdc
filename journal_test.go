package gtmcdc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Parse_Horolog_Date(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Singapore")
	assert.Nil(t, err)

	tt, err := parseHorologTime("0,0", loc)
	assert.Nil(t, err)
	assert.Equal(t, 1841, tt.Year(), "year incorrect")

	tt, err = parseHorologTime("65282,59700", loc)
	fmt.Println(tt)
	assert.Nil(t, err)
	assert.Equal(t, 2019, tt.Year())
	assert.Equal(t, "September", tt.Month().String())
	assert.Equal(t, 27, tt.Day())
	assert.Equal(t, 17, tt.Hour())
	assert.Equal(t, 39, tt.Minute())
	assert.Equal(t, 0, tt.Nanosecond())

	_, err = parseHorologTime(",", loc)
	assert.NotNil(t, err)

	_, err = parseHorologTime("29800130,1234", loc)
	assert.NotNil(t, err)
}

func Test_Parse_JournalRecord_1(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Singapore")
	assert.Nil(t, err)

	rec, _ := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^acc("00027")="300.00"`, loc)
	// fmt.Println(rec)
	assert.Equal(t, "SET", rec.opcode)
	assert.Equal(t, "300.00", rec.detail.value)

	// record is too short
	_, err = Parse(`05\65282,59700\28`, loc)
	assert.NotNil(t, err)
}

func Test_Parse_JournalRecord_2(t *testing.T) {
	loc := time.Now().Location()
	rec, _ := Parse(`08\65287,62154\3\0\0\3\0\0`, loc)
	assert.Equal(t, "TSTART", rec.opcode)

	rec, _ = Parse(`09\65287,58606\8\0\0\8\0\0\1\`, loc)
	assert.Equal(t, "TCOM", rec.opcode)
	assert.Equal(t, 8, rec.tran.tokenSeq)
	assert.Equal(t, "", rec.tran.tag)
	assert.Equal(t, "1", rec.tran.partners)
}

func Test_JournalRecord_Json(t *testing.T) {
	expected := `{"operand":"SET","transaction_num":"28",` +
		`"token_seq":28,"update_num":0,"stream_num":0,"stream_seq":0,` +
		`"journal_seq":0,"global":"ACN","key":"1234","subscripts":["51"],` +
		`"node_values":["300.00","61212","1","","","",""],` +
		`"time_stamp":1569577175}`

	loc, err := time.LoadLocation("Asia/Singapore")
	assert.Nil(t, err)

	rec, err := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^ACN(1234,51)="300.00|61212|1||||"`, loc)
	assert.Nil(t, err)

	jstr, err := rec.JSON(loc)

	assert.Nil(t, err)
	assert.Equal(t, expected, jstr)
}

func Test_atli(t *testing.T) {
	i := atoi("100")
	assert.Equal(t, 100, i)

	i = atoi("xx")
	assert.Equal(t, 0, i)
}

func Test_parseNodeFlags(t *testing.T) {
	r, err := parseNodeFlags("^ACN(5877000047,51)")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(r))
	assert.Equal(t, "ACN", r[0])
	assert.Equal(t, "5877000047", r[1])
	assert.Equal(t, "51", r[2])

	r, err = parseNodeFlags("^ACN(5877000047)")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, "ACN", r[0])
	assert.Equal(t, "5877000047", r[1])

	r, err = parseNodeFlags("^acn(5877000047,51,1245)")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(r))
	assert.Equal(t, "ACN", r[0])
	assert.Equal(t, "5877000047", r[1])
	assert.Equal(t, "51", r[2])
	assert.Equal(t, "1245", r[3])

	_, err = parseNodeFlags("^ACN()")
	assert.NotNil(t, err)

	_, err = parseNodeFlags("garbage")
	assert.NotNil(t, err)

	_, err = parseNodeFlags("")
	assert.NotNil(t, err)

}
