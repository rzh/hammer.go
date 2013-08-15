package scenario

import (
	"sync"
)

type Session struct {
	_totalWeight float32
	_calls       []*Call
	_count       int

	State        int
	StepLock     chan int
	InternalLock sync.Mutex
	Storage      map[string]string
}

func (s *Session) UpdateStateAndStorage(st int, storage ...string) {
	// atomic.AddInt64(&s.State, st)
	// State <- st
	s.InternalLock.Lock()
	defer s.InternalLock.Unlock()

	// log.Println(storage[0], storage[1])
	s.State += st
	s.Storage[storage[0]] = storage[1]
	s.StepLock <- s.State
}

// to add a new call to traffic profiles with Random Function
func (s *Session) addCall(weight float32, gp GenCall, cb GenCallBack) {
	s._totalWeight = s._totalWeight + weight
	s._calls[s._count] = new(Call)
	s._calls[s._count].RandomWeight = s._totalWeight
	s._calls[s._count].GenParam = gp
	s._calls[s._count].CallBack = cb
	s._calls[s._count].SePoint = s

	s._calls[s._count].normalize()
	s._count++
}
