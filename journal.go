package gtmcdc

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Error Messages
const (
	ErrorNotHorologFormat = "input is not horolog time format"
	ErrorInvalidRecord    = "invalid journal record format"
	ErrorDatePriorTo1971  = "date is prior to 1971/1/1"
)

type header struct {
	timestamp int64
	pid       int16
	clientPid int16
}

type repl struct {
	streamNum  int
	streamSeq  int
	journalSeq int
}

type transaction struct {
	token     string
	tokenSeq  int
	num       string
	partners  string
	updateNum int
	tag       string
}

type expr struct {
	nodeFlags string
	value     string
}

// JournalRecord represent content of a GT.M journal log entry
type JournalRecord struct {
	opcode string
	header header
	repl   repl
	tran   transaction
	detail expr
}

// JournalEvent is an event published to Kafka that
// is associated to a JournalRecord
type JournalEvent struct {
	Operand         string   `json:"operand,omitempty"`
	TransactionNum  string   `json:"transaction_num,omitempty"`
	Token           string   `json:"token,omitempty"`
	TokenSeq        int      `json:"token_seq"`
	UpdateNum       int      `json:"update_num"`
	StreamNum       int      `json:"stream_num"`
	StreamSeq       int      `json:"stream_seq"`
	JournalSeq      int      `json:"journal_seq"`
	Partners        string   `json:"partners,omitempty"`
	TransactionTag  string   `json:"transaction_tag,omitempty"`
	ProcessID       int16    `json:"pid,omitempty"`
	ClientProcessID int16    `json:"client_pid,omitempty"`
	Global          string   `json:"global,omitempty"`
	Key             string   `json:"key,omitempty"`
	Subscripts      []string `json:"subscripts,omitempty"`
	NodeValues      []string `json:"node_values,omitempty"`
	TimeStamp       int64    `json:"time_stamp,omitempty"`
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Warnf("%s is not an integer", s)
		i = 0
	}

	return i
}

// given 2 digits numeric and return the operand name
func OpCode(numeric string) string {
	var codes map[string]string
	func() {
		if codes == nil {
			codes = map[string]string{
				"00": "NULL",
				"01": "PINI",
				"02": "PFIN",
				"03": "EOF",
				"04": "KILL",
				"05": "SET",
				"06": "ZTSTART",
				"07": "ZTCOM",
				"08": "TSTART",
				"09": "TCOM",
				"10": "ZKILL",
				"11": "ZTWORM",
				"12": "ZTRIG",
				"13": "LGTRIG",
			}
		}
	}()

	if operand, ok := codes[numeric]; ok {
		return operand
	}

	return ""
}

// Parse a GT.M journal extract text string into JournalRecord
// an journal extract entry is
// NULL    = "00"\time\tnum\pid\clntpid\jsnum\strm_num\strm_seq\salvaged
// PINI    = "01"\time\tnum\pid\nnam\unam\term\clntpid\clntnnam\clntunam\clntterm
// PFIN    = "02"\time\tnum\pid\clntpid
// EOF     = "03"\time\tnum\pid\clntpid\jsnum
// KILL    = "04"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\nodeflags\node
// SET     = "05"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\nodeflags\node=sarg
// ZTSTART = "06"\time\tnum\pid\clntpid\token
// ZTCOM   = "07"\time\tnum\pid\clntpid\token\partners
// TSTART  = "08"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq
// TCOM    = "09"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\partners\tid
// ZKILL   = "10"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\nodeflags\node
// ZTWORM  = "11"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\ztwormhole
// ZTRIG   = "12"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\nodeflags\node
// LGTRIG  = "13"\time\tnum\pid\clntpid\token_seq\strm_num\strm_seq\updnum\trigdefinition
func Parse(raw string) (*JournalRecord, error) {
	// log with fields
	logf := log.WithFields(log.Fields{"journal": raw})

	s := strings.Split(raw, "\\")
	if len(s) < 5 {
		return nil, errors.New(ErrorInvalidRecord)
	}

	ts, err := Horolog2Timestamp(s[1])
	if err != nil {
		return nil, err
	}

	rec := JournalRecord{
		opcode: OpCode(s[0]),
		header: header{},
		repl:   repl{},
		tran:   transaction{},
		detail: expr{},
	}

	rec.header.pid = int16(atoi(s[3]))
	rec.header.timestamp = ts
	rec.tran.num = s[2]

	if OpCode(s[0]) == "PINI" && len(s) >= 8 {
		rec.header.clientPid = int16(atoi(s[7]))
	} else {
		rec.header.clientPid = int16(atoi(s[4]))
	}

	switch rec.opcode {
	case "":
		// ignore an empty line

	case "SET", "KILL", "ZKILL", "ZTRIG":
		rec.tran.tokenSeq, rec.tran.updateNum = atoi(s[5]), atoi(s[8])
		rec.repl.streamNum, rec.repl.streamSeq = atoi(s[6]), atoi(s[7])

		s2 := strings.Split(s[len(s)-1], "=")
		rec.detail.nodeFlags = s2[0]
		if len(s2) > 1 {
			val := s2[1]
			// remove leading and end double quote characters
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			rec.detail.value = val
		}

	case "TSTART", "TCOM":
		rec.tran.tokenSeq = atoi(s[5])
		rec.repl.streamNum, rec.repl.streamSeq = atoi(s[6]), atoi(s[7])
		if rec.opcode == "TCOM" {
			// must be TCOM
			rec.tran.partners = s[8]
			rec.tran.tag = s[9]
		}

	case "NULL", "EOF", "LGTRIG", "PINI", "PFIN":
		logf.Debugf("journal entry ignored. %s", rec.opcode)

	default:
		logf.Info("unknown journal entry")
	}

	return &rec, nil
}

// JSON representation of a journal log entry
func (rec *JournalRecord) JSON() (string, error) {
	var r []string
	var err error
	switch rec.opcode {
	case "SET", "KILL", "ZKILL", "ZTRIG":
		r, err = parseNodeFlags(rec.detail.nodeFlags)
		if err != nil {
			return "", errors.New("unable to parse")
		}
	default:
		// for other type of operands the node flags are all empty
		r = []string{"", "", ""}
	}

	event := JournalEvent{
		Operand:        rec.opcode,
		TransactionNum: rec.tran.num,
		Token:          rec.tran.token,
		TokenSeq:       rec.tran.tokenSeq,
		UpdateNum:      rec.tran.updateNum,
		StreamNum:      rec.repl.streamNum,
		StreamSeq:      rec.repl.streamSeq,
		JournalSeq:     rec.repl.journalSeq,
		Global:         r[0],
		Key:            r[1],
		Subscripts:     r[2:],
		NodeValues:     strings.Split(rec.detail.value, "|"),
		TimeStamp:      rec.header.timestamp,
	}

	bytes, err := json.Marshal(&event)
	if err != nil {
		return "", errors.New("unable to parse")
	}

	return string(bytes), nil
}

// parse a timestamp in GT.M $HOROLOG format, ddddd,sssss format
// returns a timestamp of int64 that is number of seconds since
// 1971/1/1.
//
// ddddd is the number of days after January 1, 1841
// sssss is number of seconds after the midnight of the day
//
// no timezone is used here
// date prior to 1971/1/1 will return error
func Horolog2Timestamp(horolog string) (int64, error) {
	// GTM 6.3 will send 0,0 during replication
	// for some reason YottaDB doesn't
	// just return 0 in this case
	if horolog == "" || horolog == "," || horolog == "0" || horolog == "0,0" {
		return int64(0), nil
	}

	s := strings.Split(horolog, ",")
	if s == nil {
		return -1, errors.New(ErrorNotHorologFormat)
	}

	day, err := strconv.Atoi(s[0])
	if err != nil {
		return -1, errors.New(ErrorNotHorologFormat)
	}

	var sec int
	if len(s) > 1 {
		sec, err = strconv.Atoi(s[1])
	} else {
		sec = 0
	}

	if err != nil ||
		day < 0 || day > 2980013 ||
		sec < 0 || sec > 86399 {
		return -1, errors.New(ErrorNotHorologFormat)
	}

	seconds := (day-47117)*86400 + sec
	if seconds < 0 {
		return -1, errors.New(ErrorDatePriorTo1971)
	}

	return int64(seconds), nil
}

func parseNodeFlags(node string) ([]string, error) {
	var nodeRegex *regexp.Regexp
	func() {
		if nodeRegex == nil {
			nodeRegex = regexp.MustCompile(`\^(?P<global>.*?)\((?P<index>.+)\)`)
		}
	}()

	m := nodeRegex.FindStringSubmatch(node)
	if m == nil || len(m) < 2 {
		return nil, errors.New("invalid node")
	}

	subscripts := strings.Split(m[2], ",")
	if subscripts == nil || len(m) < 1 {
		return nil, errors.New("invalid node")
	}

	ret := []string{strings.ToUpper(m[1]), subscripts[0]}
	ret = append(ret, subscripts[1:]...)

	return ret, nil
}
