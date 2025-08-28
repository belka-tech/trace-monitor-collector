package counter

import "sync"

type CounterStruct struct {
	value uint64
	mu    sync.Mutex
}

func (c *CounterStruct) Increment() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *CounterStruct) Decrement() {
	c.mu.Lock()
	c.value--
	c.mu.Unlock()
}

func (c *CounterStruct) Count() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *CounterStruct) Reset() {
	c.mu.Lock()
	c.value = 0
	c.mu.Unlock()
}
