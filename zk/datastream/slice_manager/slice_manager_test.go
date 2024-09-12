package slice_manager

import (
	"sync"
	"testing"
	"time"
	"sync/atomic"
)

func TestSliceManager_AddItem(t *testing.T) {
	sm := &SliceManager{}

	// Test adding a single item
	sm.AddItem("item1")

	if len(sm.ConsumeCurrentItems()) != 1 {
		t.Errorf("Expected 1 item, got %d", len(sm.ConsumeCurrentItems()))
	}
}

func TestSliceManager_AddItems(t *testing.T) {
	sm := &SliceManager{}

	// multiple items
	items := []interface{}{"item1", "item2", "item3"}
	sm.AddItems(items)

	if len(sm.ConsumeCurrentItems()) != 3 {
		t.Errorf("Expected 3 items, got %d", len(sm.ConsumeCurrentItems()))
	}
}

func TestSliceManager_ClearItems(t *testing.T) {
	sm := &SliceManager{}

	// add/clear
	items := []interface{}{"item1", "item2"}
	sm.AddItems(items)
	sm.ClearItems()

	if len(sm.ConsumeCurrentItems()) != 0 {
		t.Errorf("Expected 0 items after clear, got %d", len(sm.ConsumeCurrentItems()))
	}
}

func TestSliceManager_ConsumeCurrentItems(t *testing.T) {
	sm := &SliceManager{}

	// consume
	items := []interface{}{"item1", "item2"}
	sm.AddItems(items)
	consumed := sm.ConsumeCurrentItems()

	if len(consumed) != 2 {
		t.Errorf("Expected 2 items consumed, got %d", len(consumed))
	}

	if len(sm.ConsumeCurrentItems()) != 0 {
		t.Errorf("Expected slice to be empty after consumption")
	}
}

func TestSliceManager_ConcurrentAccess(t *testing.T) {
	sm := &SliceManager{}
	var wg sync.WaitGroup
	const numAdds = 1000

	var added int32

	// add items (goroutine1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numAdds; i++ {
			sm.AddItem(i)
			atomic.AddInt32(&added, 1)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// consume items until no more (goroutine2)
	wg.Add(1)
	go func() {
		seenItems := false
		defer wg.Done()
		for {
			rI := sm.ConsumeCurrentItems()
			if len(rI) == 0 && seenItems {
				break
			}
			if len(rI) > 0 {
				seenItems = true
			}
			atomic.AddInt32(&added, -int32(len(rI)))
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()

	// ensure all added items are consumed
	if atomic.LoadInt32(&added) != 0 {
		t.Errorf("Expected all items to be consumed, but %d items remain", atomic.LoadInt32(&added))
	}
}

func TestSliceManager_NoDeadlock(t *testing.T) {
	sm := &SliceManager{}
	var wg sync.WaitGroup

	// add items, check for deadlock
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			sm.AddItem(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			sm.ConsumeCurrentItems()
		}
	}()

	wg.Wait()
}
