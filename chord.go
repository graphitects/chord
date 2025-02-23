/*
Package chord implements a flexible concurrent system for handling and managing
"threads" (functions that serve as analogies to handlers) and composite chords
(collections of these thread-handlers). It provides the capability to register
thread-handlers and chords with unique keys, apply middleware wrappers in a
pipeline pattern, and safely access shared resources concurrently using sync.Map.

Note: The term "thread" in this package is used as an analogy to what is typically
known as a handler function. It does not refer to an operating system thread.

The design allows for:
- Dynamic registration and unregistration of thread-handlers.
- Nesting of chords, allowing composite structures.
- Middleware support to wrap thread-handlers in FIFO order, enhancing modularity.
*/

package chord

import (
	"bufio"
	"sync"
)

// Input represents the input to a thread, including a key, arguments, and flags.
type Input struct {
	Key   string            // Identifier for the thread execution context.
	Args  []string          // Arguments to be passed to the thread.
	Flags map[string]string // Optional flags to control thread behavior.
}

// Output represents the output from a thread, using a buffered read-writer.
type Output struct {
	*bufio.ReadWriter // Embedded buffered read-writer for thread output.
}

// Thread is a function type that takes an Input and an Output.
// This defines the basic execution unit in the chord system.
type Thread func(Input, Output)

// Chord holds a collection of threads and composite chords, managed via sync.Map
// for safe concurrent access. It also supports middleware that can be applied
// to threads and chords.
type Chord struct {
	// threads is a sync map that maps keys to threads.
	// Key: string -> thread name
	// Value: Thread -> the thread function
	threads sync.Map
	
	// chords is a sync map that maps keys to composite chords.
	// Key: string     -> chord name
	// Value: *Chord   -> pointer to the chord itself
	chords sync.Map

	// middlewares is a slice of thread wrappers that allow threads/chords to be
	// wrapped in a pipeline pattern. The wrapping is applied in FIFO order,
	// where the first middleware is the outermost wrapper.
	middlewares []ThreadWrapper
}

// FetchThread retrieves a thread from the threads map using its key.
// Returns the thread and true if found, or nil and false otherwise.
func (c *Chord) FetchThread(key string) (Thread, bool) {
	thread, ok := c.threads.Load(key)
	if !ok {
		return nil, false
	}

	return thread.(Thread), true
}

// FetchChord retrieves a chord (composite type) from the chords map using its key.
// Returns the chord pointer and true if found, or nil and false otherwise.
func (c *Chord) FetchChord(key string) (*Chord, bool) {
	chord, ok := c.chords.Load(key)
	if !ok {
		return nil, false
	}

	return chord.(*Chord), true
}

// FetchMiddlewares returns a copy of the current slice of middleware wrappers.
// A copy is provided to avoid external modifications to the original slice.
func (c *Chord) FetchMiddlewares() []ThreadWrapper {
	md := make([]ThreadWrapper, len(c.middlewares))
	copy(md, c.middlewares)
	return md
}

// Register adds a thread to the threads map with the given key.
// Optionally, additional thread wrappers (middleware) can be provided and are
// applied in FIFO order.
func (c *Chord) Register(key string, thread Thread, tw ...ThreadWrapper) {
	thread = WrapThreads(thread, tw...)
	c.threads.Store(key, thread)
}

// Unregister removes a thread from the threads map using its key.
// The provided thread parameter is not used for verification in this implementation.
func (c *Chord) Unregister(key string, thread Thread) {
	c.threads.Delete(key)
}

// Mount adds a composite chord (nested chord) to the chords map with the given key.
func (c *Chord) Mount(key string, chord *Chord) {
	c.chords.Store(key, chord)
}

// Unmount removes a composite chord from the chords map using its key.
func (c *Chord) Unmount(key string) {
	c.chords.Delete(key)
}

// Use registers one or more thread wrappers (middleware) to the chord's middleware chain.
// These wrappers will be applied to threads in the order they were added.
func (c *Chord) Use(tw ...ThreadWrapper) {
	c.middlewares = append(c.middlewares, tw...)
}

// ThreadWrapper is a function type that wraps a Thread.
// It enables modifying or augmenting the behavior of a thread.
type ThreadWrapper func(Thread) Thread

// WrapThreads builds the fully wrapped thread as a pipeline in FIFO order.
// The thread is wrapped by the provided wrappers, with the last wrapper in the slice
// being applied first.
func WrapThreads(thread Thread, tw ...ThreadWrapper) Thread {
	for i := len(tw) - 1; i > 0; i-- {
		thread = tw[i](thread)
	}
	return thread
}

// Match recursively traverses the chord structure to find and wrap the thread
// corresponding to the given path. The path represents the keys to traverse.
// If a valid thread is found, it is wrapped with its associated middleware.
func Match(node *Chord, path []string) (Thread, bool) {
	// Limit case: no keys in path.
	if len(path) == 0 {
		return nil, false
	}
	// Leaf case: single key in path implies direct thread lookup.
	if len(path) == 1 {
		thread, ok := node.FetchThread(path[0])
		if !ok {
			return nil, false
		}
		thread = WrapThreads(thread, node.FetchMiddlewares()...)
		return thread, true
	}

	// Recursive case: traverse to the next chord in the path.
	chord, ok := node.FetchChord(path[0])
	if !ok {
		return nil, false
	}
	// Recursively attempt to match the remaining path.
	thread, ok := Match(chord, path[1:])
	if !ok {
		return nil, false
	}
	// Wrap the matched thread with the middleware from the nested chord.
	thread = WrapThreads(thread, chord.FetchMiddlewares()...)
	return thread, true
}
