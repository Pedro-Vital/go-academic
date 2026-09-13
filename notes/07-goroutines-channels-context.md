# Guide 7: Goroutines, Channels & the `context` Package

> Concurrency is the feature that most separates Go's design philosophy from Python's. This guide builds the mental model from scratch: what a goroutine actually is, how channels let goroutines talk to each other safely, when you still need `sync.Mutex`/`sync.WaitGroup`, and how `context` gives you a uniform way to cancel and time out work. By the end, the goal is that `ctx context.Context` as the first argument of a function stops looking like ceremony and starts looking obvious — because you'll be passing it through every handler, every database call, and every outbound HTTP request in Phase E and the capstone.

## Table of Contents

1. [Two Different Concurrency Philosophies](#1-two-different-concurrency-philosophies)
2. [Goroutines: The Basic Unit of Concurrency](#2-goroutines-the-basic-unit-of-concurrency)
3. [The Go Scheduler: M:N Scheduling](#3-the-go-scheduler-mn-scheduling)
4. [Why You Almost Always Need to Wait: `sync.WaitGroup`](#4-why-you-almost-always-need-to-wait-syncwaitgroup)
5. [Race Conditions and the Race Detector](#5-race-conditions-and-the-race-detector)
6. [`sync.Mutex` and `sync.RWMutex`](#6-syncmutex-and-syncrwmutex)
7. [Channels: Goroutines Talking to Each Other](#7-channels-goroutines-talking-to-each-other)
8. [Buffered vs Unbuffered Channels](#8-buffered-vs-unbuffered-channels)
9. [Channel Directions and API Design](#9-channel-directions-and-api-design)
10. [`select`: Waiting on Multiple Channels](#10-select-waiting-on-multiple-channels)
11. [Closing Channels and the Comma-Ok Idiom](#11-closing-channels-and-the-comma-ok-idiom)
12. [Common Concurrency Patterns](#12-common-concurrency-patterns)
13. [The `context` Package](#13-the-context-package)
14. [`context` in Web Handlers: A Preview of Phase E](#14-context-in-web-handlers-a-preview-of-phase-e)
15. [Goroutine Leaks and Other Pitfalls](#15-goroutine-leaks-and-other-pitfalls)
16. [Exercises](#16-exercises)
17. [Key Takeaways](#17-key-takeaways)

---

## 1. Two Different Concurrency Philosophies

Before touching syntax, it's worth being explicit about *why* Go's concurrency model feels so different coming from Python, because the difference isn't cosmetic — it's architectural.

### Python's story: one thread at a time, cooperatively or not

Python gives you three real options, and each has a sharp edge:

- **`threading`** — real OS threads, but the Global Interpreter Lock (GIL) means only one thread executes Python bytecode at any instant. Threads help when you're waiting on I/O (the GIL is released during blocking calls), but they don't give you parallel CPU-bound execution.
- **`multiprocessing`** — sidesteps the GIL by using separate processes, each with its own interpreter and memory space. You get real parallelism, but communication between processes means serialization (pickling) and higher overhead.
- **`asyncio`** — a single-threaded event loop with cooperative multitasking. You get high concurrency for I/O-bound work, but it requires an entirely separate ecosystem (`async def`, `await`, `asyncio.gather`, async-flavored libraries) and one blocking call anywhere in the call graph stalls the whole event loop.

The practical consequence: in Python, you *choose your concurrency model up front* and it infects your whole codebase. A library written for `asyncio` doesn't compose with one written for `threading` without adapter code. You've lived this — FastAPI route handlers are `async def`, and calling a synchronous, blocking database driver inside one silently blocks the entire event loop unless you're careful to run it in a thread pool executor.

### Go's story: one model, and it scales down to simple and up to complex

Go has exactly one concurrency primitive that matters day-to-day: the **goroutine**. There's no separate "async" flavor of the language, no colored functions (a "colored function" is the pattern where an `async def` function can only be called from other `async def` functions — Go simply doesn't have this problem). Any function can be launched as a goroutine with the `go` keyword, full stop.

Underneath, Go's runtime multiplexes potentially hundreds of thousands of goroutines onto a small number of OS threads, and the scheduler itself knows how to move a goroutine off a thread when it blocks on I/O — so you get the throughput characteristics of `asyncio` (cheap, plentiful concurrent units) with the programming model of plain, ordinary, synchronous-looking code. You never write `await`. You just write a function, and if you want it to run concurrently, you put `go` in front of the call.

The philosophy Go encourages for coordinating those goroutines comes from Tony Hoare's Communicating Sequential Processes (CSP), and it's usually summarized as:

> **"Do not communicate by sharing memory; instead, share memory by communicating."**

In practice this means: rather than having multiple goroutines all reach into the same shared variable (which is what `threading.Lock` is protecting against in Python), you have goroutines send values to each other over **channels** — typed conduits that are safe for concurrent use by construction. You still have shared-memory tools (`sync.Mutex`, `sync.WaitGroup`) for the cases where they're the right fit, and you'll use them constantly — but they're the fallback, not the default mental model.

| Concept | Python | Go |
|---|---|---|
| Concurrent unit | OS thread (`threading`) / coroutine (`asyncio`) / process (`multiprocessing`) | Goroutine (one primitive) |
| Real parallelism? | Only via `multiprocessing` (GIL blocks `threading`) | Yes — goroutines run on real OS threads across CPU cores |
| Cost to start one | Threads: ~MB-scale stack; processes: heavier still | Goroutines: ~2KB starting stack, grows as needed |
| Syntax marker | `async def` / `await`, or `Thread(target=...)` | `go functionCall()` |
| Function coloring | Yes — async and sync code don't mix freely | No — any function can be `go`-launched |
| Default coordination style | Shared memory + locks, or `asyncio` primitives | Channels (CSP) first, shared memory + `sync` as needed |
| Cancellation | `asyncio.CancelledError`, manual flags | `context.Context` propagated explicitly through call chains |

Keep this table in your head as you read the rest of the guide — nearly every section maps onto one of these rows.

---

## 2. Goroutines: The Basic Unit of Concurrency

Launching a goroutine is almost aggressively simple:

```go
package main

import (
	"fmt"
	"time"
)

func sayHello(name string) {
	fmt.Println("Hello,", name)
}

func main() {
	go sayHello("Pedro") // runs concurrently, doesn't block main

	// Without this, main() might exit before the goroutine runs at all.
	time.Sleep(100 * time.Millisecond)
	fmt.Println("main finished")
}
```

Three things to internalize immediately:

1. **`go` is not a function call, it's a statement modifier.** You write `go f(args)` and Go schedules `f(args)` to run concurrently; the calling goroutine does not wait for it, and does not receive its return value. If `f` returns something, that return value is simply discarded — you cannot write `result := go f()`. This is a common early mistake for people used to Python's `future = executor.submit(f)`, where you get a handle back immediately.

2. **`main()` is itself running on a goroutine** — usually called the *main goroutine*. When `main()` returns, the entire program exits, **even if other goroutines are still running**. This is the single most common beginner bug in Go concurrency, and it's why the example above needs `time.Sleep` (a crude fix we'll replace with `sync.WaitGroup` in the next section). Python has no direct equivalent to this danger: if you `await` a coroutine, you're implicitly blocked until it's done; if you fire-and-forget with `asyncio.create_task`, the event loop itself doesn't exit under you mid-script the same way.

3. **Goroutines are cheap, not free.** A goroutine starts with roughly a 2KB stack (versus megabytes for an OS thread, and non-trivial memory for a Python `threading.Thread`), and that stack grows and shrinks as needed. This is why idiomatic Go code launches goroutines liberally — per incoming request, per unit of work — in a way that would be reckless with OS threads in other languages. It's the same intuition as spawning many `asyncio` tasks, except goroutines can also achieve genuine CPU parallelism across cores, which `asyncio` tasks on a single event loop never can.

### A closure gotcha you will hit at least once

```go
func main() {
	words := []string{"alpha", "beta", "gamma"}

	var wg sync.WaitGroup
	for _, w := range words {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(w) // Go 1.22+: safe, w is per-iteration
		}()
	}
	wg.Wait()
}
```

Prior to **Go 1.22**, loop variables were reused across iterations, so every goroutine closure captured the *same* variable, and by the time the goroutines actually ran, the loop might have already finished — printing the last value three times, or some nondeterministic mix, instead of `alpha`, `beta`, `gamma`. Go 1.22 changed the language so that `for` loop variables are scoped per-iteration by default, which fixes this class of bug at the language level. Since your toolchain (per Guide 1) is on a modern Go version, you get this for free — but you will see the old pattern `w := w` inside loop bodies in any codebase or blog post written before 2024, and now you know exactly what problem it was solving.

Python sidesteps this differently: a `for` loop variable in a list comprehension or generator has its own scoping rules, and `asyncio.gather(*(coro(w) for w in words))` captures `w` by value at each call, so this particular footgun doesn't have a direct Python analogue — but the *general* lesson (closures capture variables, not values, unless you're careful) is one you already know from JavaScript-style closures if you've touched `useEffect` dependency arrays in your React work.

---

## 3. The Go Scheduler: M:N Scheduling

You don't have to manage this directly, but understanding it demystifies *why* Go's concurrency feels the way it does.

Go uses what's called **M:N scheduling**: M goroutines are multiplexed onto N OS threads, where N is typically `GOMAXPROCS` (defaulting to the number of CPU cores). The runtime scheduler itself decides which goroutine runs on which OS thread at any given moment, and — critically — it knows how to **park** a goroutine that's blocked on I/O (a network call, a channel receive, a mutex) and run a *different* goroutine on that freed-up thread instead.

This gives you a combination that neither Python option gives you alone:

- Like `asyncio`, you get very cheap, very numerous concurrent units, because the scheduler — not the OS — is managing them.
- Unlike `asyncio`, CPU-bound goroutines *can* run truly in parallel across multiple cores, because there's no GIL-equivalent lock serializing execution of Go code. `asyncio` concurrency is entirely about I/O overlap on one core; Go concurrency is about I/O overlap *and* genuine multi-core parallelism, unified in the same programming model.
- Unlike `threading`, you're not paying OS-thread costs (context-switch overhead, megabyte-scale stacks) per unit of concurrency.

The mental shorthand: **an `asyncio` event loop is a single-threaded scheduler you interact with explicitly (`await`); the Go scheduler is a multi-threaded scheduler you interact with implicitly (blocking calls just work).** You never write "yield control back to the scheduler" in Go — a channel receive, a mutex lock, a network read, or a `time.Sleep` all naturally yield the underlying thread to other goroutines without you doing anything special.

You can inspect and tune this if you ever need to:

```go
import "runtime"

fmt.Println(runtime.NumCPU())       // physical/logical CPUs available
fmt.Println(runtime.GOMAXPROCS(0))  // current GOMAXPROCS, 0 = just read it
```

In practice you will almost never touch `GOMAXPROCS` — the default (number of CPUs) is correct for the overwhelming majority of programs, including web servers.

---

## 4. Why You Almost Always Need to Wait: `sync.WaitGroup`

The `time.Sleep` hack from Section 2 is not a real solution — it's guessing how long concurrent work takes. `sync.WaitGroup` is the idiomatic way to block until a known number of goroutines have finished.

```go
package main

import (
	"fmt"
	"sync"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done() // decrement the counter when this goroutine returns
	fmt.Printf("worker %d starting\n", id)
	// ... do work ...
	fmt.Printf("worker %d done\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)        // increment the counter before launching
		go worker(i, &wg) // pass by pointer — see note below
	}

	wg.Wait() // blocks until the counter returns to zero
	fmt.Println("all workers done")
}
```

The API is three methods, and the mental model is a simple counter:

- **`wg.Add(n)`** — increments the internal counter by `n`. Call this *before* launching the goroutine, not inside it (otherwise `Wait()` might race ahead and return before the `Add` call happens).
- **`wg.Done()`** — decrements the counter by 1. Conventionally called via `defer` at the top of the goroutine's function, so it fires even if the function panics or has multiple return paths.
- **`wg.Wait()`** — blocks the calling goroutine until the counter hits zero.

**Note on the pointer:** `sync.WaitGroup` (like `sync.Mutex`, which you'll see next) must never be copied after first use — you always pass it around by pointer (`*sync.WaitGroup`), never by value. If you're storing one on a struct, this is one of the few cases in idiomatic Go where holding it as a value field, then always referencing it via a pointer receiver method, is the standard advice from Guide 3.

### Python side-by-side

The direct analogue in `threading` is joining a list of threads:

```python
import threading

def worker(worker_id):
    print(f"worker {worker_id} starting")
    # ... do work ...
    print(f"worker {worker_id} done")

threads = []
for i in range(1, 6):
    t = threading.Thread(target=worker, args=(i,))
    threads.append(t)
    t.start()

for t in threads:
    t.join()  # block until this thread finishes

print("all workers done")
```

In `asyncio`, the equivalent is `asyncio.gather`:

```python
import asyncio

async def worker(worker_id):
    print(f"worker {worker_id} starting")
    await asyncio.sleep(0)  # yield point
    print(f"worker {worker_id} done")

async def main():
    await asyncio.gather(*(worker(i) for i in range(1, 6)))
    print("all workers done")

asyncio.run(main())
```

`asyncio.gather` is actually the closer conceptual match — it's declarative ("wait for all of these to finish") rather than imperative ("loop over threads calling `.join()`"), which is the spirit `sync.WaitGroup` is going for, just expressed with an explicit counter instead of a collected list of awaitables.

---

## 5. Race Conditions and the Race Detector

Once two or more goroutines can touch the same memory, you can have a **data race**: unsynchronized concurrent access to the same variable where at least one access is a write. This is the exact problem CSP-style channels are designed to make rare, and that `sync.Mutex` exists to solve when you do need shared memory.

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // DATA RACE: read-modify-write, not atomic
		}()
	}

	wg.Wait()
	fmt.Println("counter:", counter) // almost never prints 1000
}
```

`counter++` looks like one operation but is really three: read the value, add one, write it back. When many goroutines interleave those three steps on the same variable, increments get lost. This isn't a Go-specific danger — the identical bug exists in Python `threading` code touching a shared `int` without a lock — but Go gives you a first-class tool to *detect* it that Python doesn't have built into the standard toolchain:

```bash
go run -race main.go
go test -race ./...
```

The `-race` flag builds an instrumented binary that tracks memory accesses across goroutines at runtime and reports the exact file and line of any race it observes, including a stack trace for each of the conflicting accesses. It has real runtime and memory overhead, so it's a development/CI tool, not something you ship to production — but running your test suite with `-race` regularly (or in CI) is considered standard practice in professional Go codebases, precisely because data races are undefined behavior: your program might work fine for months and then corrupt data unpredictably under load. Treat `-race` the way you'd treat `mypy --strict` or a linter gate: cheap to run, and it catches an entire class of bug before it reaches production.

---

## 6. `sync.Mutex` and `sync.RWMutex`

When goroutines genuinely need to share mutable state (rather than communicate via channels), `sync.Mutex` is the lock that makes access to that state safe.

```go
package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func main() {
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()
	fmt.Println("counter:", counter.Value()) // always 1000
}
```

Patterns worth calling out:

- **`defer mu.Unlock()` immediately after `mu.Lock()`** is close to a universal idiom. It guarantees the unlock happens on every exit path, including panics, mirroring how you'd use `defer conn.Close()` from Guide 3/4 patterns.
- **The mutex is embedded as a field, not held externally.** Bundling the lock with the data it protects (here, `mu` sits right next to `count` inside `SafeCounter`) is the idiomatic Go convention — it makes it visually obvious, right at the struct definition, which lock protects which data. Compare this to Python, where a `threading.Lock()` is often a separate object you have to remember to associate with the right shared state by convention alone; Go's struct embedding enforces the association structurally.
- **`sync.RWMutex`** is a variant with two lock modes: `Lock()`/`Unlock()` for exclusive write access, and `RLock()`/`RUnlock()` for shared read access — any number of readers can hold an `RLock` simultaneously, but a `Lock()` call blocks until all readers release. Use this when reads vastly outnumber writes (e.g., an in-memory cache read on every request, written to rarely):

```go
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
```

### Python comparison

`sync.Mutex` maps directly onto `threading.Lock`:

```python
import threading

class SafeCounter:
    def __init__(self):
        self._lock = threading.Lock()
        self.count = 0

    def increment(self):
        with self._lock:   # context manager ~ defer mu.Unlock()
            self.count += 1
```

Python's `with lock:` and Go's `defer mu.Unlock()` are solving the identical problem (guarantee release on every exit path) with different language mechanisms — a context manager versus a deferred call. The `RWMutex`/reader-writer distinction doesn't have a standard-library equivalent in Python threading at all (you'd typically reach for a third-party library or roll your own), which is one case where Go's standard library is more batteries-included for concurrent code than Python's.

Important: **none of this applies to `asyncio`.** Because `asyncio` is single-threaded cooperative multitasking, a plain `counter += 1` inside an `async def` is *not* subject to the same interleaving danger unless you `await` in the middle of the read-modify-write — the GIL plus the lack of true parallelism means asyncio code doesn't need locks nearly as often. This is a real source of confusion moving between the two ecosystems: Go's true parallelism means you need locks in more places than idiomatic `asyncio` code does, but fewer places than idiomatic `threading` code, because Go pushes you toward channels first.

---

## 7. Channels: Goroutines Talking to Each Other

A channel is a typed conduit for passing values between goroutines. You declare a channel's type the same way you'd declare a slice's element type:

```go
ch := make(chan int)      // unbuffered channel of ints
ch := make(chan string)   // unbuffered channel of strings
ch := make(chan Order)    // unbuffered channel of your own struct type
```

Sending and receiving both use the `<-` operator, with the arrow pointing in the direction data flows:

```go
ch <- 42       // send 42 into the channel
value := <-ch  // receive a value from the channel, assign to value
```

Here's the smallest complete example — one goroutine computes a result, sends it back over a channel, and `main` waits to receive it:

```go
package main

import "fmt"

func square(n int, result chan<- int) {
	result <- n * n
}

func main() {
	result := make(chan int)
	go square(7, result)

	value := <-result // blocks until square() sends
	fmt.Println(value) // 49
}
```

The critical property to internalize: **an unbuffered channel send blocks until another goroutine is ready to receive, and a receive blocks until another goroutine sends.** This is a *rendezvous* — both sides must be present at the same moment for the operation to complete. That blocking behavior is itself a synchronization mechanism: in the example above, you don't need a `sync.WaitGroup` at all, because `<-result` in `main` naturally blocks until `square`'s goroutine has produced its value.

### Why this matters more than it looks

This is the CSP idea made concrete. Instead of `square` writing into some shared variable that `main` reads (which would need a mutex to be safe), `square` *sends* its result directly to whoever wants it. There's no shared memory to protect — the channel itself is safe for concurrent use, guaranteed by the language runtime, the same way Python's `queue.Queue` is thread-safe by design.

Which is, in fact, the closest Python analogue:

```python
import threading
import queue

def square(n, result_queue):
    result_queue.put(n * n)

result_queue = queue.Queue()
t = threading.Thread(target=square, args=(7, result_queue))
t.start()

value = result_queue.get()  # blocks until square() puts a value
print(value)  # 49
```

`queue.Queue` is Python's closest built-in match to a Go channel — both are safe for concurrent producers/consumers without you adding your own lock. The difference is that channels are a **first-class language construct** in Go (with dedicated syntax, compiler-checked types, and direct integration with `select`, covered next), whereas `queue.Queue` is a library class you opt into. Go nudges you toward this pattern by making it the path of least resistance; Python leaves it as one option among several.

---

## 8. Buffered vs Unbuffered Channels

`make(chan int)` creates an **unbuffered** channel — capacity zero, rendezvous semantics as just described. `make(chan int, n)` creates a **buffered** channel with room for `n` values in flight before a send blocks:

```go
ch := make(chan int, 3) // buffer capacity 3

ch <- 1 // succeeds immediately, buffer now holds [1]
ch <- 2 // succeeds immediately, buffer now holds [1, 2]
ch <- 3 // succeeds immediately, buffer now holds [1, 2, 3]
ch <- 4 // BLOCKS — buffer is full, waits for a receive to free a slot

fmt.Println(len(ch), cap(ch)) // 3 3 — current count, total capacity
```

A buffered channel decouples the sender and receiver in time: the sender can get ahead of the receiver by up to the buffer size before it has to wait. This is useful, but it's easy to reach for it as a quick fix for a design problem rather than understanding what you're actually choosing:

- **Unbuffered channels give you a synchronization guarantee**: "the receiver has definitely gotten this value" the instant the send returns. This is often exactly the guarantee you want — e.g., "don't proceed until this worker has picked up the task."
- **Buffered channels give you a decoupling/throughput tool**, useful when you want a bounded queue between a fast producer and a slower consumer, but they weaken that guarantee — a send succeeding only means "the value is in the buffer," not "someone has processed it."

A common, low-risk starting point: **default to unbuffered channels** unless you have a specific, articulable reason for a buffer size (e.g., "I know exactly 5 workers will be reading from this, so a buffer of 5 avoids unnecessary blocking on bursts"). An arbitrary buffer size chosen to "make a deadlock go away" is usually masking a design bug rather than fixing one.

There isn't a clean Python equivalent to *unbuffered* channels — `queue.Queue(maxsize=0)` in Python means *unlimited* buffering, the opposite of Go's zero-capacity meaning. If you want a bounded, buffered queue in Python, `queue.Queue(maxsize=n)` is the analogue of `make(chan T, n)`; if you want a rendezvous with no buffering at all, Python doesn't give you that directly — you'd typically use `queue.Queue(maxsize=1)` plus a bit of extra coordination, or reach for `asyncio`'s zero-capacity primitives, as an approximation.

---

## 9. Channel Directions and API Design

You saw this above without a callout: `result chan<- int` in the `square` function signature. Channel types can be restricted to send-only or receive-only in a function signature, and the compiler enforces the restriction:

```go
func producer(out chan<- int) { // out: can only send
	for i := 0; i < 5; i++ {
		out <- i
	}
	close(out)
}

func consumer(in <-chan int) { // in: can only receive
	for v := range in {
		fmt.Println(v)
	}
}

func main() {
	ch := make(chan int) // bidirectional at creation
	go producer(ch)
	consumer(ch) // ch implicitly converts to <-chan int here
}
```

Reading the syntax: `chan<- int` — the arrow points *into* `chan`, so this is a channel you send into. `<-chan int` — the arrow points *out of* `chan`, so this is a channel you receive from. A plain `chan int` is bidirectional and can be passed to either kind of parameter; Go converts it implicitly (the reverse — passing a directional channel where a bidirectional one is expected — is a compile error).

This is a documentation-and-safety mechanism with no real analogue in Python's type system: it tells the reader of a function signature exactly how that channel will be used, and the compiler rejects code that tries to send on something declared receive-only or vice versa. There's nothing forcing you to use directional types (a plain `chan int` parameter would work too), but doing so is considered good practice for any function whose whole job is to be one side of a producer/consumer relationship — it's the concurrency equivalent of accepting an `io.Reader` instead of a concrete file type (Guide 4): you're expressing the minimum capability your function actually needs.

---

## 10. `select`: Waiting on Multiple Channels

`select` lets a goroutine wait on several channel operations at once and proceeds with whichever one is ready first — it's Go's answer to "wait for the first of several things to happen," a shape of problem Python solves with `asyncio.wait(..., return_when=FIRST_COMPLETED)` or `concurrent.futures.wait`.

```go
func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "from ch1"
	}()
	go func() {
		time.Sleep(2 * time.Second)
		ch2 <- "from ch2"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Println("received:", msg1)
		case msg2 := <-ch2:
			fmt.Println("received:", msg2)
		}
	}
}
```

`select` blocks until *one* of its `case`s can proceed; if multiple are ready simultaneously, it picks one at random (deliberately, to prevent programs from accidentally depending on ordering). Two extra forms you'll use constantly:

**A `default` case makes the whole `select` non-blocking** — if no channel is ready right away, `default` fires instead of blocking:

```go
select {
case msg := <-ch:
	fmt.Println("got:", msg)
default:
	fmt.Println("no message ready, moving on")
}
```

**A `time.After` case gives you a timeout on a channel wait**, which is the pattern you'll use before you reach for `context` (and which `context.WithTimeout`, covered in Section 13, generalizes and makes cancelable/propagatable):

```go
select {
case msg := <-ch:
	fmt.Println("got:", msg)
case <-time.After(3 * time.Second):
	fmt.Println("timed out waiting for message")
}
```

`time.After(d)` returns a channel that receives a single value after duration `d` elapses — so racing it against your real channel inside `select` is an idiomatic, if slightly manual, timeout mechanism. You'll see in Section 13 why `context.WithTimeout` is usually the better choice once cancellation needs to propagate through several function calls, but understanding this raw `select` + `time.After` pattern first makes it obvious what `context` is actually doing under the hood.

---

## 11. Closing Channels and the Comma-Ok Idiom

`close(ch)` signals that no more values will ever be sent on a channel. Receivers can detect this:

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch)

v, ok := <-ch
fmt.Println(v, ok) // 1 true  — value received normally
v, ok = <-ch
fmt.Println(v, ok) // 2 true  — value received normally
v, ok = <-ch
fmt.Println(v, ok) // 0 false — channel is closed and drained; v is the zero value
```

The two-value receive form (`v, ok := <-ch`) is the channel equivalent of the comma-ok idiom you already know from map lookups (Guide 2) and type assertions (Guide 4) — `ok` tells you whether you got a "real" value or the zero-value-because-closed case. Note that receiving from an already-closed, already-drained channel never blocks and never panics — it just keeps returning `(zeroValue, false)` forever. This is deliberate: it makes closed channels usable as broadcast "done" signals (see the pipeline pattern below).

The far more common way to consume a channel until it's closed is `for range`, which stops automatically:

```go
for v := range ch {
	fmt.Println(v) // loop exits automatically once ch is closed and drained
}
```

**Rules that will save you real debugging time:**

- **Only the sender should close a channel, never the receiver.** The receiver generally doesn't know whether more sends are coming.
- **Sending on a closed channel panics.** This is not recoverable gracefully in the way a Python exception is — it's a program-crashing panic, so closing discipline matters.
- **Closing an already-closed channel also panics.**
- **You do not have to close every channel.** Closing is a signal ("I'm done sending"), not cleanup like `file.Close()` — if nothing depends on detecting the end of the stream (e.g., you know exactly how many values to expect and read exactly that many), you can leave a channel unclosed and let the garbage collector reclaim it once nothing references it.

Python has no direct equivalent to "closing" a `queue.Queue` — the conventional pattern is a sentinel value (commonly `None`) pushed onto the queue to signal "no more items," which consumers check for explicitly:

```python
SENTINEL = None

def producer(q):
    for i in range(5):
        q.put(i)
    q.put(SENTINEL)  # manual "I'm done" signal

def consumer(q):
    while True:
        item = q.get()
        if item is SENTINEL:
            break
        print(item)
```

Go's `close()` + `for range` replaces this hand-rolled sentinel convention with a language-level mechanism — one less thing every producer/consumer pair has to independently agree on.

---

## 12. Common Concurrency Patterns

These three patterns cover the overwhelming majority of real-world goroutine/channel usage, and you'll draw on all three during the Phase F capstone (a REST API handling concurrent requests, each potentially doing concurrent downstream work).

### Worker pool

A fixed number of goroutines pull work items off a shared channel, so you get bounded concurrency instead of spawning one goroutine per task (dangerous if the number of tasks is large or attacker-controlled):

```go
func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		results <- j * j // pretend this is expensive work
	}
}

func main() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	for w := 1; w <= 3; w++ { // 3 workers, however many jobs
		go worker(w, jobs, results)
	}

	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // workers' range loops exit once jobs is drained

	for i := 0; i < 9; i++ {
		fmt.Println(<-results)
	}
}
```

This is directly analogous to Python's `concurrent.futures.ThreadPoolExecutor` or `multiprocessing.Pool` — a fixed pool size, work distributed across it — except here you're wiring the pool together yourself out of channels and goroutines rather than calling into a library class. That extra explicitness is a running theme in Go: the standard library gives you primitives (goroutines, channels) rather than a high-level `Pool` abstraction, and you compose the abstraction yourself when you need it.

### Fan-out, fan-in

Multiple goroutines ("fan-out") read from the same input channel to parallelize work, then a single goroutine ("fan-in") merges their output channels into one:

```go
func fanIn(channels ...<-chan int) <-chan int {
	merged := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				merged <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(merged) // close only after all sources are drained
	}()

	return merged
}
```

Note the `wg.Wait()` running inside its own goroutine before `close(merged)` — this is the standard idiom for "close a channel once N producers are all finished," since you can't call `close` until you're certain nothing else will send.

### Pipeline

Chaining stages together, each consuming from one channel and producing to the next — conceptually close to Unix pipes, or to composing generator functions in Python:

```go
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	for v := range square(generate(1, 2, 3, 4)) {
		fmt.Println(v) // 1 4 9 16
	}
}
```

Each stage is a function that takes a receive-only channel and returns a receive-only channel, closing its own output when its input is drained. This composability — chaining `square(generate(...))` — is structurally similar to chaining Python generators (`(n*n for n in gen())`), except each Go stage is running concurrently in its own goroutine rather than being lazily pulled one item at a time by a single-threaded caller.

---

## 13. The `context` Package

Every pattern so far handles *coordination* (waiting for goroutines, protecting shared state, passing values). None of them handle **cancellation** — telling a goroutine (or a whole tree of goroutines it spawned) "stop, we don't need this anymore," possibly because a client disconnected, a timeout elapsed, or a sibling operation already failed. That's what `context.Context` is for, and it's the single package you'll import more than almost any other once you reach Phase E, because it's the standard way HTTP request lifecycles get propagated through your handler, your database calls, and any outbound requests you make.

### The interface itself

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}
```

You will almost never implement this interface yourself — you'll construct contexts using the constructor functions below, and consume them via `Done()` and `Err()`.

### Creating contexts

```go
ctx := context.Background() // root context — start of a call chain, e.g. in main()
ctx := context.TODO()        // placeholder when you haven't decided yet / are migrating old code
```

`context.Background()` is the context you use when there's no parent — the top of a request's lifecycle. `context.TODO()` signals "this should have a real context eventually" and is mostly seen mid-refactor; you'll rarely write it deliberately in new code.

Everything else **derives** a new context from a parent, forming a tree:

```go
ctx, cancel := context.WithCancel(parent)
defer cancel() // ALWAYS call the cancel function, even if you don't cancel manually — see below

ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

ctx, cancel := context.WithDeadline(parent, someSpecificTime)
defer cancel()

ctx = context.WithValue(parent, someKey, someValue)
```

### `WithCancel`: manual cancellation

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx)

	time.Sleep(2 * time.Second)
	cancel() // signal the worker to stop
	time.Sleep(1 * time.Second)
}

func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker: received cancellation signal:", ctx.Err())
			return
		default:
			fmt.Println("worker: working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
```

`ctx.Done()` returns a channel that's closed exactly once, at the moment the context is canceled (whether by explicit `cancel()`, a timeout firing, or a deadline passing) — which means every goroutine holding that context can `select` on `<-ctx.Done()` and react immediately, and (per Section 11) a closed channel can be received from by an unlimited number of goroutines without blocking, which is exactly the "broadcast to everyone watching" behavior cancellation needs. `ctx.Err()` tells you *why* it was canceled: `context.Canceled` for explicit cancellation, `context.DeadlineExceeded` for a timeout/deadline.

**Why you must always call `cancel()`, even on the "success" path:** every `WithCancel`/`WithTimeout`/`WithDeadline` call allocates resources associated with tracking that context (internally, a goroutine or timer watching for the deadline). If you never call `cancel()`, those resources leak until the parent context itself is canceled or the deadline passes on its own — for a `WithTimeout` context this eventually self-cleans when the timer fires, but for `WithCancel` it never will unless you call it. `defer cancel()` immediately after creation is the idiom precisely because it guarantees cleanup on every exit path, success or failure, exactly like `defer mu.Unlock()` and `defer file.Close()` before it.

### `WithTimeout` and `WithDeadline`: the ones you'll use constantly

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

select {
case result := <-doSomethingSlow(ctx):
	fmt.Println("got result:", result)
case <-ctx.Done():
	fmt.Println("gave up:", ctx.Err()) // prints "context deadline exceeded"
}
```

`WithTimeout(parent, d)` is really just `WithDeadline(parent, time.Now().Add(d))` — a convenience wrapper. Both automatically cancel the context (closing `Done()`, setting `Err()` to `context.DeadlineExceeded`) once the clock runs out, with no manual `time.After` racing required — this is the generalization of the raw `select`/`time.After` pattern from Section 10, but with one crucial upgrade: **the timeout propagates automatically to every function you pass the context into**, rather than being local to one `select` statement. If your handler calls a database function, and that function calls an external API, and you passed the same `ctx` all the way down, a single timeout set at the top cancels the *entire chain* the moment it fires — every layer just needs to respect `ctx.Done()` (or, more commonly, pass `ctx` into a context-aware standard-library or driver function that already respects it for you, like `http.NewRequestWithContext` or most database drivers' `QueryContext`).

### `WithValue`: request-scoped data (use sparingly)

```go
type contextKey string

const userIDKey contextKey = "userID"

ctx := context.WithValue(context.Background(), userIDKey, "user-123")

// deeper in the call chain:
userID, ok := ctx.Value(userIDKey).(string)
```

`WithValue` lets you attach request-scoped data (a request ID, an authenticated user ID, a trace ID) that flows down through function calls without adding an explicit parameter everywhere. The Go community's near-universal guidance is to use this **sparingly** — it's not a general-purpose way to avoid passing parameters, only for genuinely cross-cutting, request-scoped metadata like tracing/auth identity, because `ctx.Value` lookups are stringly/interface-typed (as you see above, retrieved via a type assertion) and bypass the compiler's usual type-checking that makes Go valuable in the first place. If a value is meaningfully part of a function's contract, it should be an explicit parameter, not smuggled through the context.

### Python comparison

Nothing in Python's standard library unifies "cancellation," "timeout," and "request-scoped values" the way `context.Context` does. The closest pieces are scattered:

- `asyncio.wait_for(coro, timeout=5)` gives you a timeout on a single awaitable, raising `asyncio.TimeoutError` — similar in spirit to `WithTimeout`, but it doesn't automatically propagate down an arbitrary call chain the way a passed-around `ctx` parameter does; you'd need to thread a cancellation signal through manually, or rely on the fact that canceling an `asyncio.Task` propagates `CancelledError` into whatever it's currently `await`-ing.
- `threading.Event` is the closest analogue to `ctx.Done()` as a broadcast-cancellation signal — a flag that multiple threads can check via `.is_set()` or block on via `.wait()`.
- Request-scoped values (like FastAPI's request-scoped dependencies, or `contextvars.ContextVar`) cover roughly the same ground as `WithValue`, and `contextvars` in particular is actually a close structural match — it's part of *why* `asyncio` code can have "current request" state without explicit parameter threading.

The unification is the point: in Go, one type (`context.Context`) and one idiom (pass it as the first parameter, named `ctx`) covers cancellation, timeouts, deadlines, and scoped values everywhere in the standard library and virtually every third-party library, which is why you'll see `ctx context.Context` as the first parameter of practically every function that does I/O, from here through the rest of the roadmap.

---

## 14. `context` in Web Handlers: A Preview of Phase E

You're not building the web layer yet (that's Phase E), but since you'll use `context` "constantly in web handlers," it's worth previewing *why*, concretely, so Section 13 doesn't stay abstract.

`net/http`'s `*http.Request` carries a context automatically — accessible via `r.Context()` — and that context is **canceled automatically when the client disconnects** before the handler finishes. A shape you'll write often:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // the request's own context — canceled if the client disconnects

	// Derive a bounded timeout for a downstream call:
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := fetchFromDatabase(ctx, someQuery)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "upstream timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func fetchFromDatabase(ctx context.Context, query string) (Result, error) {
	// A context-aware driver call — e.g. db.QueryContext(ctx, query) —
	// returns early with ctx.Err() if ctx is canceled or times out
	// mid-query, instead of blocking until the query naturally finishes.
	// ...
}
```

Two properties fall out of this for free, and both map onto patterns you already care about from your FastAPI background:

1. **Automatic client-disconnect cancellation.** If a client closes the connection mid-request, `r.Context()` is canceled, and any downstream call that respects the context (database query, outbound HTTP call) can stop early instead of finishing work nobody will ever read — directly analogous to FastAPI's `Request.is_disconnected()` check, except Go's version is threaded automatically through everything downstream that accepts a `ctx`, rather than something you have to poll for yourself.
2. **Composable timeouts.** Deriving `context.WithTimeout(ctx, 2*time.Second)` from the request's own context means "give this downstream call 2 seconds, *or* stop early if the whole request context is already canceled for some other reason (client disconnect, an outer deadline) — whichever comes first." This is the same idea as setting a `timeout=` on an `httpx.AsyncClient` call in FastAPI, generalized so it composes across every layer of your call stack instead of being configured per-library.

This is also exactly why `ctx context.Context` shows up as the first parameter, by convention, of virtually every function signature that talks to a database, calls another service, or does any blocking I/O in idiomatic Go web code — it's the mechanism by which "the client is gone, stop working" and "this has taken too long, stop working" reach every layer of your application uniformly, without each layer inventing its own cancellation flag.

---

## 15. Goroutine Leaks and Other Pitfalls

A goroutine that blocks forever — on a channel nobody will ever send to or receive from, or on a `ctx.Done()` that will never fire — never gets garbage collected, because Go can't know it will never be used again. This is a **goroutine leak**, and unlike a leaked Python object (which the garbage collector reclaims once nothing references it, `asyncio` task or not), a leaked goroutine consumes stack memory and scheduler bookkeeping for the lifetime of the process, silently, with no exception or traceback pointing at the cause.

Common causes and their fixes:

- **Sending on a channel nobody will ever read from again.** If a consumer goroutine exits early (say, due to an error) while a producer goroutine is still trying to send, the producer blocks forever. Fix: give the producer a way to notice the consumer is gone — typically by selecting on `ctx.Done()` alongside the channel send.

  ```go
  select {
  case out <- value:
  	// sent successfully
  case <-ctx.Done():
  	return ctx.Err() // don't block forever if nobody's listening anymore
  }
  ```

- **Forgetting to call `cancel()`** from `WithCancel`/`WithTimeout`/`WithDeadline` — covered in Section 13, but worth repeating here as a leak, specifically: the internal timer/goroutine tracking that context outlives the function that created it if you skip `defer cancel()`.

- **A worker pool reading from a channel that's never closed.** If `close(jobs)` never happens, every worker goroutine's `for range jobs` loop blocks forever waiting for either a new job or the close signal, even after the program logically has no more work.

- **Unbuffered channel sends with no matching receiver**, e.g., in error-handling paths where you send an error onto a results channel but the receiving goroutine already returned after seeing an earlier error.

Python's rough equivalent is a "hung" `threading.Thread` that never returns (which likewise prevents clean process shutdown) or an `asyncio.Task` that nobody ever `await`s or cancels — the Python runtime warns about the latter ("Task was destroyed but it is pending"), which is more visible than Go's silence on this exact failure mode. The defensive habit worth building now, before Phase E and the capstone: **any goroutine that can block indefinitely on a channel operation should also be selecting on a context's `Done()` channel**, so there's always an escape hatch.

---

## 16. Exercises

Work through these in order — each builds on the primitive introduced just before it.

1. **Goroutine + WaitGroup basics.** Write a program that launches 10 goroutines, each of which sleeps for a random duration between 100–500ms and then prints its own ID. Use `sync.WaitGroup` (not `time.Sleep` in `main`) to wait for all 10 to finish before printing `"all done"`.

2. **Fix the race.** Take the unsynchronized `counter++` example from Section 5, run it with `go run -race`, confirm it reports a race, then fix it two different ways: once using `sync.Mutex`, and once by having each goroutine send its increment over a channel to a single goroutine that owns the counter (no mutex at all). Compare the two solutions — which one feels more "Go-idiomatic" per the CSP philosophy, and why?

3. **Unbuffered vs. buffered.** Write a producer goroutine that sends 5 integers into a channel and a consumer that receives and prints them. First run it with an unbuffered channel; then change only the channel creation line to give it a buffer of 5. Explain, in your own words, what specifically changes about the *timing* of when the producer's sends return — not just whether the program still works.

4. **Worker pool.** Build a worker pool (Section 12) with 4 workers that each "process" jobs by computing whether a given integer is prime, reading job integers from a `jobs` channel and writing `(int, bool)` results to a `results` channel. Feed it 50 jobs and print all results once collected.

5. **Timeout with `select` + `time.After`.** Write a function that simulates a slow operation (sleeps for a random 0–2 second duration, then sends a result on a channel). Call it from `main` using `select` racing against `time.After(1 * time.Second)`, printing either the result or a timeout message.

6. **The same thing, with `context`.** Rewrite exercise 5 using `context.WithTimeout` instead of raw `time.After`, threading `ctx` into the slow function so it can check `ctx.Done()` itself rather than `main` racing a bare timer. What's different about how the "give up" decision is made compared to exercise 5?

7. **Cancellation propagation.** Write three functions, `a(ctx)`, `b(ctx)`, and `c(ctx)`, where `a` calls `b`, `b` calls `c`, and `c` is a loop that checks `ctx.Done()` on every iteration and returns `ctx.Err()` when canceled. In `main`, create a `context.WithTimeout` of 2 seconds, call `a(ctx)`, and confirm that canceling the context at the top causes `c`, three calls deep, to stop — without `a` or `b` needing any timeout logic of their own.

8. **Goroutine leak, on purpose.** Deliberately write a goroutine that leaks (a send on an unbuffered channel with no receiver, similar to Section 15). Use `runtime.NumGoroutine()` before and after to confirm the leaked goroutine is still alive and never cleaned up, even after the "main" work is done. Then fix it using a `select` with `ctx.Done()` as an escape hatch, and confirm the goroutine count returns to baseline.

---

## 17. Key Takeaways

- Go has **one** concurrency primitive — the goroutine — launched with `go f(...)`, unlike Python's split between `threading` (real threads, GIL-limited), `multiprocessing` (real parallelism, heavier processes), and `asyncio` (single-threaded cooperative multitasking with its own function coloring). There's no `async`/`await` and no "colored functions" in Go.
- The **Go scheduler** does M:N multiplexing of goroutines onto OS threads and automatically parks goroutines that block on I/O, giving you `asyncio`-like cheap concurrency *and* real multi-core parallelism in one model — you get this for free without writing explicit yield points.
- **`main()` exiting kills all other goroutines immediately**, regardless of whether they've finished — always synchronize with `sync.WaitGroup` (or a channel) rather than guessing with `time.Sleep`.
- **`sync.WaitGroup`** (`Add`/`Done`/`Wait`) is how you block until a known number of goroutines finish; **always pass it by pointer**, never copy it after first use.
- **Data races** — unsynchronized concurrent access with at least one writer — are undefined behavior. Catch them early and often with `go run -race` / `go test -race`.
- **`sync.Mutex`** (and `sync.RWMutex` for read-heavy workloads) protects genuinely shared mutable state; the idiom is to embed the lock as a struct field right next to the data it guards, and `defer mu.Unlock()` immediately after `mu.Lock()`.
- **Channels** are CSP's core idea made concrete: goroutines communicate by sending typed values rather than sharing memory directly. Unbuffered channels are a rendezvous (send blocks until received); buffered channels decouple sender and receiver up to a fixed capacity — default to unbuffered unless you have a specific reason for a buffer.
- **Channel direction types** (`chan<-`, `<-chan`) in function signatures document and compiler-enforce whether a function only sends or only receives — express the minimum capability a function actually needs.
- **`select`** waits on multiple channel operations at once, optionally with a non-blocking `default` or a `time.After`-based timeout — this is the manual version of what `context.WithTimeout` generalizes and automates.
- **`close(ch)`** signals "no more sends are coming"; only the sender should close, sending on a closed channel panics, and `for range ch` is the idiomatic way to consume until closure. Not every channel needs to be closed.
- **Worker pools, fan-out/fan-in, and pipelines** are the three patterns that cover most real concurrent Go code — all built by composing goroutines and channels, rather than reaching for a high-level `Pool` class the way you would in Python.
- **`context.Context`** unifies cancellation, timeouts, deadlines, and (sparingly) request-scoped values into one type, passed as the conventional first parameter (`ctx context.Context`) through every function that does I/O. `ctx.Done()` returns a channel closed on cancellation; `ctx.Err()` tells you why.
- **Always `defer cancel()`** immediately after any `WithCancel`/`WithTimeout`/`WithDeadline` call, on every code path, or you leak the resources tracking that context.
- In web handlers (previewing Phase E), `r.Context()` is automatically canceled on client disconnect, and deriving `context.WithTimeout` from it lets a single cancellation signal propagate through your entire downstream call chain — database queries, outbound API calls — with no per-layer polling required.
- A goroutine blocked forever on a channel or an unfired `ctx.Done()` is a **goroutine leak** — silent, with no traceback, unlike Python's more visible "pending task" warnings. Any goroutine that can block indefinitely should also select on `ctx.Done()` as an escape hatch.

---

**Next up: Guide 8**, continuing Phase C, will build on this foundation with more advanced concurrency patterns and the tools for reasoning about them precisely — or, depending on how you've scoped the rest of the roadmap, moving into Phase D (testing and data). Let me know which is next and I'll pick up from here in the same format.
