package id

import (
	"sync"
	"time"
)

type Snowflake struct {
	mutex  sync.Mutex
	stamp  int64
	worker int64
	seq    int64
}

func NewSnowflake(workerId int64) *Snowflake {
	return &Snowflake{
		worker: workerId,
	}
}

func (s *Snowflake) NextId() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now().UnixNano() / 1000000

	if s.stamp == now {
		s.seq = (s.seq + 1) & 4095
		if s.seq == 0 {
			for now <= s.stamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		s.seq = 0
	}

	s.stamp = now
	return (now << 22) | (s.worker << 12) | s.seq
}
