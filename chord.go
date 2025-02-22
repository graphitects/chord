// Package chord provides a framework for managing and executing threads
// with associated input and output. It allows for the registration,
// unregistration, and wrapping of threads using a sync.Map for concurrency
// safety.
//
// The main components of the package are:
// - Input: Represents the input to a thread, including a key, arguments, and flags.
// - Output: Represents the output from a thread, using a buffered read-writer.
// - Thread: A function type that takes an Input and an Output.
// - Chord: Holds a collection of threads, managed by a sync.Map.
// - ThreadWrapper: A function type that wraps a Thread.
//
// The Chord struct provides methods to register, unregister, and register
// wrapped threads, allowing for flexible thread management and execution.
package chord

import (
	"bufio"
	"sync"
)

// Input represents the input to a thread, including a key, arguments, and flags.
type Input struct {
	Key   string
	Args  []string
	Flags map[string]string
}

// Output represents the output from a thread, using a buffered read-writer.
type Output struct {
	*bufio.ReadWriter
}

// Thread is a function type that takes an Input and an Output.
type Thread func(Input, Output)

// Chord holds a collection of threads, managed by a sync.Map.
type Chord struct {
	// multiplexer is a sync map that maps keys to threads
	// - key: string -> name thread
	// - val: thread -> thread itself
	multiplexer sync.Map
}

// Register adds a thread to the multiplexer with the given key.
func (c *Chord) Register(key string, thread Thread) {
	c.multiplexer.Store(key, thread)
}

// Unregister removes a thread from the multiplexer with the given key.
func (c *Chord) Unregister(key string, thread Thread) {
	c.multiplexer.Delete(key)
}

// ThreadWrapper is a function type that wraps a Thread.
type ThreadWrapper func(Thread) Thread

// RegisterWrapped adds a wrapped thread to the multiplexer with the given key.
func (c *Chord) RegisterWrapped(key string, tw ...ThreadWrapper) {
	var thread Thread
	for i := range tw {
		thread = tw[i](thread)
	}
	
	c.multiplexer.Store(key, thread)
}