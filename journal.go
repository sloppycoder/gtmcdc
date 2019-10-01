package gtmcdc

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var OpCodes = map[string]string{
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

const (
	ErrorNotHorologFormat = "input is not horolog time format"
	ErrorInvalidRecord    = "invalid journal record format"
)

type Header struct {
	timestamp time.Time
	pid       int16
	clientPid int16
}

type Repl struct {
	streamNum  int8
	streamSeq  int
	journalSeq int
}

type Transaction struct {
	token     string
	tokenSeq  int
	num       string
	partners  string
	updateNum int
	tag       string
}

type Expr struct {
	nodeFlags string
	value     string
}

type JournalRecord struct {
	opcode string
	header Header
	repl   Repl
	tran   Transaction
	detail Expr
}

type JournalEvent struct {
	Operand         string `json:"operand,omitempty"`
	TransactionNum  string `json:"transaction_num,omitempty"`
	Token           string `json:"token,omitempty"`
	TokenSeq        int    `json:"token_seq"`
	UpdateNum       int    `json:"update_num"`
	StreamNum       int8   `json:"stream_num"`
	StreamSeq       int    `json:"stream_seq"`
	JournalSeq      int    `json:"journal_seq"`
	Partners        string `json:"partners,omitempty"`
	TransactionTag  string `json:"transaction_tag,omitempty"`
	ProcessId       int16  `json:"pid,omitempty"`
	ClientProcessId int16  `json:"client_pid,omitempty"`
	Node            string `json:"node,omitempty"`
	Value           string `json:"value,omitempty"`
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Warnf("%s is not an integer", s)
		i = 0
	}

	return i
}

//
// parse a GT.M journal extract text string into JournalRecord
//
func Parse(raw string) (*JournalRecord, error) {
	log.Debugf("parsing:%s", raw)

	s := strings.Split(raw, "\\")
	if len(s) < 5 {
		return nil, errors.New(ErrorInvalidRecord)
	}

	ts, err := parseHorologTime(s[1])
	if err != nil {
		return nil, err
	}

	rec := JournalRecord{
		opcode: OpCodes[s[0]],
		header: Header{},
		repl:   Repl{},
		tran:   Transaction{},
		detail: Expr{},
	}

	rec.header.pid = int16(atoi(s[3]))
	rec.header.timestamp = ts
	rec.tran.num = s[2]

	if OpCodes[s[0]] == "PINI" && len(s) >= 8 {
		rec.header.clientPid = int16(atoi(s[7]))
	} else {
		rec.header.clientPid = int16(atoi(s[4]))
	}

	switch rec.opcode {
	case "SET", "KILL", "ZKILL", "ZTRIG":
		rec.tran.tokenSeq, rec.tran.updateNum = atoi(s[5]), atoi(s[8])
		rec.repl.streamNum, rec.repl.streamSeq = int8(atoi(s[6])), atoi(s[7])

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
		rec.repl.streamNum, rec.repl.streamSeq = int8(atoi(s[6])), atoi(s[7])
		if rec.opcode == "TCOM" {
			// must be TCOM
			rec.tran.partners = s[8]
			rec.tran.tag = s[9]
		}

	case "NULL", "EOF", "LGTRIG", "PINI", "PFIN":
		log.Debugf("journal entry ignored: %s", raw)

	default:
		log.Warnf("unknown journal entry: %s", raw)
	}

	return &rec, nil
}

func (rec *JournalRecord) Json() string {
	event := JournalEvent{
		Operand:        rec.opcode,
		TransactionNum: rec.tran.num,
		Token:          rec.tran.token,
		TokenSeq:       rec.tran.tokenSeq,
		UpdateNum:      rec.tran.updateNum,
		StreamNum:      rec.repl.streamNum,
		StreamSeq:      rec.repl.streamSeq,
		JournalSeq:     rec.repl.journalSeq,
		Node:           rec.detail.nodeFlags,
		Value:          rec.detail.value,
	}

	bytes, err := json.Marshal(&event)
	if err != nil {
		return ""
	}

	return string(bytes)
}

// parse a timestamp in GT.M $HOROLOG format, ddddd,sssss format and return a time.Time
//
// ddddd is the number of days after January 1, 1841
// sssss is number of seconds after the midnight of the day
//
func parseHorologTime(horolog string) (time.Time, error) {
	now := time.Now()

	s := strings.Split(horolog, ",")
	if len(s) != 2 {
		return now, errors.New(ErrorNotHorologFormat)
	}

	day, err1 := strconv.Atoi(s[0])
	sec, err2 := strconv.Atoi(s[1])
	if err1 != nil || err2 != nil ||
		day < 0 || day > 2980013 ||
		sec < 0 || sec > 86399 {
		return now, errors.New(ErrorNotHorologFormat)
	}

	horologBaseTime := time.Date(1841, 1, 1, 0, 0, 0, 0, now.Location())
	seconds := day*86400 + sec

	return horologBaseTime.Add(time.Duration(seconds) * time.Second), nil
}
