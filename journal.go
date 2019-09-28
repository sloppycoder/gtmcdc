package gtmcdc

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	NULL    = "00"
	PINI    = "01"
	PFIN    = "02"
	EOF     = "03"
	KILL    = "04"
	SET     = "05"
	ZTSTART = "06"
	ZTCOM   = "07"
	TSTART  = "08"
	TCOM    = "09"
	ZKILL   = "10"
	ZTWORM  = "11"
	ZTRIG   = "12"
	LGTRIG  = "13"
)

const (
	ErrorNotHorologFormat = "input is not horolog time format"
	ErrorInvalidRecord    = "invalid journal record format"
)

type Header struct {
	timestamp time.Time
	pid       string
	clientPid string
}

type Repl struct {
	streamNum  string
	streamSeq  string
	journalSeq string
}

type Transaction struct {
	token     string
	tokenSeq  string
	partners  string
	updateNum string
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

//
// parse a GT.M journal extract text string into JournalRecord
//
func Parse(raw string) (*JournalRecord, error) {
	s := strings.Split(raw, "\\")
	if len(s) < 5 {
		return nil, errors.New(ErrorInvalidRecord)
	}

	ts, err := parseHorologTime(s[1])
	if err != nil {
		return nil, err
	}

	rec := JournalRecord{
		opcode: s[0],
		header: Header{},
		repl:   Repl{},
		tran:   Transaction{},
		detail: Expr{},
	}

	rec.header.pid = s[2]
	rec.header.timestamp = ts
	if s[0] == PINI && len(s) >= 8 {
		rec.header.clientPid = s[7]
	} else {
		rec.header.clientPid = s[4]
	}

	switch rec.opcode {
	case SET, KILL, TCOM, ZTRIG:
		rec.tran.tokenSeq, rec.tran.updateNum = s[5], s[8]
		rec.repl.streamNum, rec.repl.streamSeq = s[6], s[7]

		s2 := strings.Split(s[len(s)-1], "=")
		rec.detail.nodeFlags = s2[0]
		if len(s2) > 1 {
			rec.detail.value = s2[1]
		}
		break

	default:
		log.Warnf("journal entry ignored: %s", raw)
	}

	return &rec, nil
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
