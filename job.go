package worker

import (
	"bytes"
	"net/url"
	"strconv"
	"time"
)

const DefaultMaxRetryCount = 5

type JobStatus int

const (
	SDoing JobStatus = iota
	SSuccess
	SRetrying //
	SFailed   //
	SFinish   //
)

type JobFlag int

const (
	FRetryWait JobFlag = iota // wait 2+3*second and retry
	FRetryNow                 // retry now
	FSuccess
	FFailed
	FDelete // delete job, not do never
)

type Job struct {
	topic   []byte // topic name
	channel []byte // channel name, on listener and counter use it
	keys    [][]byte
	values  [][]byte

	MaxRetry int

	count    int // retry count
	Status   JobStatus
	interval time.Duration
}

var maxRetryKey = []byte("__MaxRetry")

func NewJob(topic string) *Job {
	return &Job{
		topic:    s2B(topic),
		MaxRetry: DefaultMaxRetryCount,
		interval: time.Minute,
	}
}

// encode to []byte like "key=value&name=bysir"
func (p *Job) encode() []byte {
	var queryBuf bytes.Buffer

	for i, k := range p.keys {
		if queryBuf.Len() != 0 {
			queryBuf.WriteByte('&')
		}
		queryBuf.Write(k)
		queryBuf.WriteByte('=')
		queryBuf.Write(p.values[i])
	}

	// add self params
	if queryBuf.Len() != 0 {
		queryBuf.WriteByte('&')
	}
	queryBuf.Write(maxRetryKey)
	queryBuf.WriteByte('=')
	queryBuf.WriteString(strconv.Itoa(p.MaxRetry))

	return queryBuf.Bytes()
}

func (p *Job) decode(data []byte) bool {
	// now, count is not save to queue
	//vAndC := bytes.Split(data, []byte{'#'})
	//values := vAndC[0]
	//if len(vAndC) == 2 && len(vAndC[1]) != 0 {
	//	p.count = vAndC[1][0]
	//}
	values := data
	if len(values) != 0 {
		vs := bytes.Split(values, []byte{'&'})
		vsLen := len(vs)
		p.values = make([][]byte, vsLen)
		p.keys = make([][]byte, vsLen)
		for i, v := range vs {
			kv := bytes.Split(v, []byte{'='})
			if len(kv) != 2 {
				return false
			}

			if bytes.Equal(kv[0], maxRetryKey) {
				p.MaxRetry, _ = strconv.Atoi(b2S(kv[1]))
			} else {
				p.keys[i] = kv[0]
				p.values[i] = kv[1]
			}
		}
	}

	return true
}

func (p *Job) String() string {
	var buf bytes.Buffer
	switch p.Status {
	case SFailed:
		buf.WriteString("FAILED ")
	case SSuccess:
		buf.WriteString("SUCCESS ")
	case SDoing:
		buf.WriteString("DOING ")
	case SRetrying:
		buf.WriteString("RETRYING ")
	case SFinish:
		buf.WriteString("FINISHED ")
	}
	buf.Write(p.topic)
	buf.WriteByte(',')
	buf.Write(p.channel)
	buf.WriteByte(':')
	buf.Write(p.encode())
	buf.WriteByte('#')
	buf.WriteString(strconv.Itoa(p.count))
	return buf.String()
}

func (p *Job) Channel() string {
	return b2S(p.channel)
}

func (p *Job) Topic() string {
	return b2S(p.topic)
}
func (p *Job) Count() int {
	return p.count
}

func (p *Job) setCount(count int) {
	p.count = count
}

func (p *Job) addCount() {
	p.count++
}

func (p *Job) Param(key string) (value string, ok bool) {
	if p.keys == nil {
		return
	}
	kb := s2B(key)
	key, _ = url.QueryUnescape(key)
	for i, k := range p.keys {
		if bytes.Equal(k, kb) {
			value = b2S(p.values[i])
			value, _ = url.QueryUnescape(value)
			ok = true
			return
		}
	}
	return
}

func (p *Job) SetParam(key string, value string) {
	key = url.QueryEscape(key)
	value = url.QueryEscape(value)
	kb := s2B(key)
	vb := s2B(value)
	p.SetParamByte(kb, vb)
	return
}

func (p *Job) SetParamByte(kb, vb []byte) {
	if p.keys == nil {
		p.values = [][]byte{}
		p.keys = [][]byte{}
	}

	for i, k := range p.keys {
		if bytes.Equal(k, kb) {
			p.values[i] = vb
			return
		}
	}

	p.keys = append(p.keys, kb)
	p.values = append(p.values, vb)

	return
}

func (p *Job) SetInterval(duration time.Duration) {
	p.interval = duration
}
