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
	// threads is a sync map that maps keys to threads
	// - key: string -> name thread
	// - val: thread -> thread itself
	threads sync.Map
	
	// chords is a sync map that maps keys to composite type
	// - key: string     -> name chord
	// - val: ref chord  -> ref chord itself
	chords sync.Map

	// middlewares is a slice of thread wrappers allowing to get wrapped in a pipeline pattern
	// usable for both the threads and chords
	// - wrapping is in FIFO order (first is the bigger wrapper)
	middlewares []ThreadWrapper
}

// FetchThread retrieves a thread from threads
func (c *Chord) FetchThread(key string) (Thread, bool) {
	thread, ok := c.threads.Load(key)
	if !ok {
		return nil, false
	}

	return thread.(Thread), true
}

// FetchChord retreves a chord from chords
func (c *Chord) FetchChord(key string) (*Chord, bool) {
	chord, ok := c.chords.Load(key)
	if !ok {
		return nil, false
	}

	return chord.(*Chord), true
}

// FetchMiddlewares returns a copy of the slice of middlewares
func (c *Chord) FetchMiddlewares() []ThreadWrapper {
	md := make([]ThreadWrapper, len(c.middlewares))
	copy(md, c.middlewares)
	return md
}

// Register adds a thread to the threads with the given key.
// - it also allows to add thread-wrappers in FIFO order
func (c *Chord) Register(key string, thread Thread, tw ...ThreadWrapper) {
	thread = WrapThreads(thread, tw...)
	c.threads.Store(key, thread)
}

// Unregister removes a thread from the threads with the given key.
func (c *Chord) Unregister(key string, thread Thread) {
	c.threads.Delete(key)
}

// Mount adds a composite chord to the chords with the given key
func (c *Chord) Mount(key string, chord *Chord) {
	c.chords.Store(key, chord)
}

// Unmount removes a composite chord from the chords with the given key
func (c *Chord) Unmount(key string) {
	c.chords.Delete(key)
}

// Use allows to register a thread wrapper in the middlewares
func (c *Chord) Use(tw ...ThreadWrapper) {
	c.middlewares = append(c.middlewares, tw...)
}

// ThreadWrapper is a function type that wraps a Thread.
type ThreadWrapper func(Thread) Thread

// WrapThreads builds the fully wrapped thread as a pipeline in FIFO order
func WrapThreads(thread Thread, tw ...ThreadWrapper) Thread {
	for i := len(tw) - 1; i > 0; i-- {
		thread = tw[i](thread)
	}
	return thread
}