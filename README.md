# Chord

Chord is a flexible concurrent system implemented in Go for managing "threads" (functions 
serving as analogies to handlers) and composite chords (collections of these thread-handlers). 
It allows dynamic registration and unregistration of thread-handlers, supports nesting chords 
for complex composite structures, and applies middleware wrappers in a FIFO pipeline pattern. 
It leverages Go's `sync.Map` for safe concurrent access.

**Note:** The term "thread" in this package is used as an analogy to handler functions commonly 
found in web frameworks or similar architectures, and does not refer to operating system threads.

## Features

- **Dynamic Thread-Handler Registration:** Easily register and unregister handler functions with 
  unique keys.
- **Composite Chord Structure:** Organize thread-handlers into nested chords for hierarchical management.
- **Middleware Support:** Wrap thread-handlers with middleware in FIFO order to enhance or modify behavior.
- **Concurrent Safety:** Uses Go’s `sync.Map` for thread-safe resource management.

## Installation

Install using `go get`:

```bash
go get github.com/graphitects/chord
```

## Usage

Below is an example demonstrating how to use Chord in your Go application.

### Example

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/graphitects/chord"
)

func main() {
	// Create a new chord instance.
	rootChord := &chord.Chord{}

	// Define a simple thread-handler that writes a greeting.
	helloHandler := func(input chord.Input, output chord.Output) {
		greeting := "Hello, " + input.Args[0] + "!"
		output.WriteString(greeting)
		output.Flush()
	}

	// Register the thread-handler with the key "hello".
	rootChord.Register("hello", helloHandler)

	// Define a middleware to log thread-handler execution.
	logMiddleware := func(next chord.Thread) chord.Thread {
		return func(input chord.Input, output chord.Output) {
			fmt.Println("Executing handler:", input.Key)
			next(input, output)
			fmt.Println("Finished handler:", input.Key)
		}
	}

	// Apply the middleware to the root chord.
	rootChord.Use(logMiddleware)

	// Setup a buffered read-writer for handler output.
	rw := bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
	output := chord.Output{ReadWriter: rw}

	// Use the Match function to retrieve the handler by its path.
	if handler, ok := chord.Match(rootChord, []string{"hello"}); ok {
		input := chord.Input{
			Key:  "hello",
			Args: []string{"World"},
		}
		handler(input, output)
	} else {
		fmt.Println("Handler not found!")
	}
}
```

### API Overview

- **Chord**
  - `Register(key string, thread Thread, tw ...ThreadWrapper)`: Registers a thread-handler with a given key and applies any provided middleware wrappers.
  - `Unregister(key string, thread Thread)`: Removes a thread-handler using its key.
  - `Mount(key string, chord *Chord)`: Adds a composite chord (nested chord) under the specified key.
  - `Unmount(key string)`: Removes a composite chord.
  - `Use(tw ...ThreadWrapper)`: Adds middleware to the chord.
  - `FetchThread(key string) (Thread, bool)`: Retrieves a thread-handler by its key.
  - `FetchChord(key string) (*Chord, bool)`: Retrieves a nested chord by its key.
  - `FetchMiddlewares() []ThreadWrapper`: Returns a copy of the currently registered middleware.
- **ThreadWrapper**: A function type for wrapping a thread-handler, allowing modification or augmentation of its behavior.
- **WrapThreads(thread Thread, tw ...ThreadWrapper) Thread**: Wraps a thread-handler with the provided middleware wrappers.
- **Match(node *Chord, path []string) (Thread, bool)**: Recursively searches for a thread-handler in a chord structure based on a path of keys, wrapping it with any associated middleware along the way.

## Contributing

Contributions are welcome! To contribute:
1. Fork the repository.
2. Create a feature branch for your changes.
3. Open a pull request with a detailed explanation of your modifications.
4. For significant changes, please open an issue to discuss your ideas first.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.