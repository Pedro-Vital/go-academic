**Source: https://youtu.be/N_WEcmNk15w?si=HVxm2PPUwTZMyT9b**

---

This is a good example of combining **goroutines + channels + buffered channels + `WaitGroup` + channel closing**.

The key idea is:

> Two goroutines produce messages (`pinger` and `ponger`), one goroutine consumes them (`printer`), and `main` coordinates when everything is finished.

---

## 1. The channel

```go
canal := make(chan string, 40)
```

This creates a **buffered channel** with capacity 40.

```text
canal
┌──────────────────────────────────────────────┐
│  [ ][ ][ ][ ][ ] ... [ ][ ][ ][ ]           │
│              capacity = 40                   │
└──────────────────────────────────────────────┘
```

The channel can temporarily hold up to **40 strings** without requiring a receiver to be ready.

There will be:

* 20 `"ping!"`
* 20 `"pong!"`

So exactly **40 messages** will be sent.

---

# 2. `pinger`

```go
func pinger(c chan string, wg *sync.WaitGroup) {
    defer wg.Done()

    for i := 1; i <= 20; i++ {
        c <- "ping!"
        log.Println("Mensagem enviada! (ping)")
    }

    fmt.Println("Envio de Pinger Finalizado!!")
}
```

This goroutine sends 20 messages:

```go
c <- "ping!"
```

Each iteration sends one message into the channel.

Because the channel has capacity 40, the sender can initially put messages into the buffer without waiting for `printer`.

At the end:

```go
defer wg.Done()
```

signals:

> "The pinger goroutine has finished."

---

# 3. `ponger`

It works exactly the same way:

```go
func ponger(c chan string, wg *sync.WaitGroup) {
    defer wg.Done()

    for i := 1; i <= 20; i++ {
        c <- "pong!"
        log.Println("Mensagem enviada! (pong)")
    }

    fmt.Println("Envio de Ponger Finalizado!!")
}
```

It produces another 20 messages.

Therefore:

```text
pinger → 20 × "ping!"
ponger → 20 × "pong!"

Total → 40 messages
```

---

# 4. `printer`

This is the consumer:

```go
func printer(c chan string, wg *sync.WaitGroup) {
    defer wg.Done()

    for msg := range c {
        log.Println("Mensagem recebida!", msg)
        time.Sleep(time.Millisecond * 250)
    }

    log.Println("Todas mensagens retiradas do canal!")
}
```

The interesting part is:

```go
for msg := range c
```

This means:

> Keep receiving messages from `c` until `c` is closed **and empty**.

It is essentially equivalent to repeatedly doing:

```go
msg, ok := <-c

if !ok {
    break
}
```

The `range` version is much cleaner.

---

# 5. Why does `printer` sleep?

```go
time.Sleep(time.Millisecond * 250)
```

This simulates a slow consumer.

Every message takes approximately 250 ms to process.

There are 40 messages, so processing them takes roughly:

```text
40 × 250 ms = 10 seconds
```

But importantly, the **senders don't necessarily take 10 seconds**.

That's where the buffered channel becomes important.

---

# 6. The `WaitGroup`s

There are actually **two different WaitGroups**:

```go
wgSenders := sync.WaitGroup{}
wgPrinter := sync.WaitGroup{}
```

This is intentional.

### Sender WaitGroup

```go
wgSenders.Add(1)
go pinger(canal, &wgSenders)

wgSenders.Add(1)
go ponger(canal, &wgSenders)
```

So:

```text
wgSenders
counter = 2
```

Eventually:

```go
pinger → wg.Done() → counter 1
ponger → wg.Done() → counter 0
```

Then:

```go
wgSenders.Wait()
```

unblocks.

---

### Printer WaitGroup

```go
wgPrinter.Add(1)
go printer(canal, &wgPrinter)
```

So:

```text
wgPrinter
counter = 1
```

The printer calls:

```go
defer wg.Done()
```

only when its entire function finishes.

---

# 7. The most important part: why `close(canal)` happens here

Look at `main`:

```go
wgSenders.Wait()
close(canal)

wgPrinter.Wait()
```

This ordering is **very important**.

First:

```go
wgSenders.Wait()
```

means:

> Wait until both `pinger` and `ponger` have completely finished sending.

Only then:

```go
close(canal)
```

This tells the receiver:

> No more messages will ever be sent to this channel.

Then `printer` eventually finishes:

```go
for msg := range c {
    ...
}
```

because the channel is closed **and all buffered messages have been consumed**.

Finally:

```go
wgPrinter.Wait()
```

waits for the printer to finish.

---

# 8. Why can't we close the channel immediately?

Imagine doing:

```go
go pinger(canal, &wgSenders)
go ponger(canal, &wgSenders)

close(canal)
```

That's dangerous.

The goroutines might still be executing:

```go
c <- "ping!"
```

after the channel has been closed.

Sending to a closed channel causes a panic:

```text
panic: send on closed channel
```

Therefore, the producer/consumer pattern is:

```text
             PRODUCERS
          ┌──────────────┐
          │              │
      pinger          ponger
          │              │
          └──────┬───────┘
                 │
                 ▼
          ┌──────────────┐
          │   channel    │
          │   buffer 40  │
          └──────┬───────┘
                 │
                 ▼
              printer
```

And the rule is:

> **The sender side is responsible for closing the channel, after all senders are finished.**

Here `main` performs that responsibility.

---

# 9. What happens chronologically?

A simplified execution looks like this:

### Step 1

`main` starts `printer`:

```go
go printer(canal, &wgPrinter)
```

Now printer is waiting:

```go
for msg := range c
```

There aren't necessarily messages yet.

---

### Step 2

`main` starts `pinger` and `ponger`:

```go
go pinger(...)
go ponger(...)
```

Now we have three goroutines:

```text
main
 │
 ├── printer
 ├── pinger
 └── ponger
```

---

### Step 3

The producers start filling the channel:

```text
pinger → "ping!"
ponger → "pong!"
pinger → "ping!"
ponger → "pong!"
...
```

The actual order is **not guaranteed**.

You might see:

```text
ping
ping
pong
ping
pong
pong
...
```

or:

```text
pong
ping
pong
pong
ping
...
```

The Go scheduler determines when each goroutine runs.

---

### Step 4

`printer` removes messages:

```go
msg := <-c
```

and processes them:

```go
time.Sleep(250 * time.Millisecond)
```

So while `printer` is processing one message, the producers can continue putting messages into the buffer.

---

# 10. Why does the buffer matter?

Suppose `printer` is currently processing:

```text
"ping!"
```

It sleeps for 250 ms.

Without a buffer:

```go
make(chan string)
```

the next:

```go
c <- "pong!"
```

would block until `printer` receives it.

With:

```go
make(chan string, 40)
```

the producer can put the message into the buffer and continue.

Conceptually:

```text
Producer                    Consumer

pinger ──┐
          ├──> [ channel buffer ] ──> printer
ponger ──┘
```

This decouples the producer speed from the consumer speed, **up to the buffer capacity**.

---

# 11. What happens when the buffer becomes full?

This is an important point.

The channel has capacity 40:

```go
make(chan string, 40)
```

There are exactly 40 messages total.

If `printer` were not consuming anything, the producers could collectively fill the channel:

```text
[ping][pong][ping][pong]...[ping][pong]
                  40
```

Then another:

```go
c <- "..."
```

would **block**.

That's the fundamental behavior of a buffered channel:

```text
buffer has space
      ↓
send succeeds immediately

buffer is full
      ↓
send blocks until space becomes available
```

In this program, `printer` is continuously consuming, so the senders can usually proceed.

---

# 12. Why two WaitGroups instead of one?

You could technically design this differently, but having two makes the synchronization responsibilities very clear.

The program needs to know two separate things:

### Question 1

> Have all producers finished?

```go
wgSenders.Wait()
```

Once yes:

```go
close(canal)
```

### Question 2

> Has the consumer finished processing everything?

```go
wgPrinter.Wait()
```

So the sequence is:

```text
pinger ────────┐
               │
ponger ────────┤
               ▼
         wgSenders.Wait()
               │
               ▼
          close(canal)
               │
               ▼
         printer drains
         remaining messages
               │
               ▼
         wgPrinter.Wait()
               │
               ▼
             END
```

That's the core architecture.

---

# 13. One subtle point: closing doesn't delete messages

This is extremely important when learning channels.

Suppose:

```text
channel buffer:

[ping][pong][ping][pong]
```

and then:

```go
close(canal)
```

The buffered messages **remain**.

`printer` can still receive them.

So:

```go
close(canal)
```

does **not** mean:

> "Stop receiving."

It means:

> "No more values can be sent."

The receiver can continue receiving already-buffered values.

That's precisely why this works:

```go
for msg := range c {
    ...
}
```

The loop ends only after:

1. the channel is closed, **and**
2. all remaining messages have been received.

---

# 14. The entire program in one mental model

Think of it as a factory:

```text
       PRODUCERS
       
   ┌─────────────┐
   │   pinger    │──┐
   │ 20 messages │  │
   └─────────────┘  │
                    │
                    ▼
              ┌───────────┐
              │  CHANNEL  │
              │ buffer 40 │
              └─────┬─────┘
                    │
                    ▼
              ┌───────────┐
              │  printer  │
              │  consumer │
              └───────────┘
                    │
                    ▼
                  DONE

   ┌─────────────┐
   │   ponger    │──┘
   │ 20 messages │
   └─────────────┘
```

`WaitGroup` controls the lifecycle:

```text
                    wgSenders
                       │
              "Are both producers done?"
                       │
                       ▼
                    YES
                       │
                       ▼
                 close(channel)
                       │
                       ▼
              "No more messages."
                       │
                       ▼
                  printer drains
                       │
                       ▼
                   wgPrinter
                       │
                       ▼
                    main ends
```

### The three concepts to remember

**`WaitGroup` answers:**

> "Has this goroutine/task finished?"

**Channel answers:**

> "How do goroutines exchange data?"

**`close(channel)` answers:**

> "Will there be any more values sent?"

And in this particular pattern, the crucial rule is:

> **Wait for all producers → close the channel → let the consumer range over it until empty → wait for the consumer.**
