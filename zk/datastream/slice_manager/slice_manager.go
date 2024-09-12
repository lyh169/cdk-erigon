package slice_manager

import (
	"sync"
)

type SliceManager struct {
	items []interface{}
	iMu   sync.RWMutex
}

func NewSliceManager() *SliceManager {
	return &SliceManager{}
}

func (s *SliceManager) AddItem(item interface{}) {
	s.iMu.Lock()
	defer s.iMu.Unlock()
	s.items = append(s.items, item)
}

func (s *SliceManager) AddItems(items []interface{}) {
	s.iMu.Lock()
	defer s.iMu.Unlock()
	s.items = append(s.items, items...)
}

func (s *SliceManager) ClearItems() {
	s.iMu.Lock()
	defer s.iMu.Unlock()
	s.items = nil
}

func (s *SliceManager) ConsumeCurrentItems() []interface{} {
	s.iMu.Lock()
	defer s.iMu.Unlock()

	// copy the slice
	dCopy := make([]interface{}, len(s.items))
	copy(dCopy, s.items)

	// empty it
	s.items = nil

	return dCopy
}
