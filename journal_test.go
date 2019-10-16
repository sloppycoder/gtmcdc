package gtmcdc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Horolog2UnixTime(t *testing.T) {
	_, tmp := time.Now().Zone()
	offset := int64(tmp)
	loc, _ := time.LoadLocation("Asia/Singapore")

	// run the following command in GT.M to check day values
	// write $$CDS^%H(day)

	// 1971/1/1 is 47117 days after 1840/12/31
	// any day that results in a date prior to 1971/1/1 in the
	// local timezone will return an error because
	// the timestamp value will be negative
	// Unix system cannot handle this sort of thing
	// 47118 should get a positive timestamp value in any timezone
	ts, err := Horolog2UnixTime("47118")
	assert.Nil(t, err)
	assert.Equal(t, 24*3600-offset, ts)

	expected, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-09-26 16:35:00", loc)
	ts, err = Horolog2UnixTime("65282,59700")
	assert.Nil(t, err)
	assert.Equal(t, expected.Unix(), ts)

	expected, _ = time.ParseInLocation("2006-01-02 15:04:05", "2019-10-04 00:24:45", loc)
	ts, err = Horolog2UnixTime("65290,1485")
	assert.Nil(t, err)
	assert.Equal(t, expected.Unix(), ts)

	_, err = Horolog2UnixTime(",")
	assert.NotNil(t, err)

	_, err = Horolog2UnixTime("29800130,1234")
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

func Test_Parse_JournalRecord_2(t *testing.T) {
	rec, _ := Parse(`08\65287,62154\3\0\0\3\0\0`)
	assert.Equal(t, "TSTART", rec.opcode)

	rec, _ = Parse(`09\65287,58606\8\0\0\8\0\0\1\`)
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
		`"time_stamp":1569486900}`

	rec, err := Parse(`05\65282,59700\28\0\0\28\0\0\0\0\^ACN(1234,51)="300.00|61212|1||||"`)
	assert.Nil(t, err)

	jayson, err := rec.JSON()

	assert.Nil(t, err)
	assert.Equal(t, expected, jayson)
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
