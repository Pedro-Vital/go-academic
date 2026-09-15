### 1. Added

```go
time.Sleep(3 * time.Second)
```

to the WriteResult() method of FileManager.

### 2. Modified `main()`

* Turning the Process() call into a goroutine that receives a channel (one for each job).

* Receiving the value through the channel

* Setting up an error channel.

### 3. Modified `Process()`

Bear in mind: Just because the Process method is started as goroutine doesn't mean that all functions inside Process are started as goroutines as well.

## Error Handling

We can't simply

```go
for _,errorChan := range errorChans {
    <-errorChan
}
```

The program would be expecting any value from error handling. If we have no error it crashes. It doesn't make any sense.

If we have multiple channels, like we have with errorChan and doneChan, where only one of the channels will emit a value for a given goroutine, we can use the **select** control structure:

```go
for index := range taxRates {
    select {
    case err := <-errorChans[index]:
        if err != nil {
            fmt.Println(err)
        }
    case <-doneChans[index]:
        fmt.Println("Done!")
    }
}
```

### **See the important detail in the `SECOND-LOOP-BLOCK.md` file in this folder.**