In prices.go, I'm not letting the process method know about failures of LoadData() and WriteResult(). It should not be the case, because if something goes wrong we want the Process() to stop. 

Bear in mind that the "02-advanced-version" approach calls `return` within a function other than `main()`, so program execution is not interrupted.

LoadData() should return error.
Process() should return error too, to whichever place called it.
`main()` call Process().

---
This relates to a question I asked the AI:

> # PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP

```
func (job *TaxIncludedPricesJob) LoadData() {
	file, err := os.Open("prices.txt")
	
	if err != nil {
		fmt.Println("Could not open the file!")
		fmt.Println(err)
		return
	}

	scanner := bufio.NewScanner(file)
	
	var lines []string
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()

	if err != nil {
		fmt.Println("Reading the content file failed!")
		fmt.Println(err)
		file.Close()
		return
	}

	prices := make([]float64, len(lines))

	for lineIndex, line := range lines {
		floatPrice, err := strconv.ParseFloat(line, 64)

		if err != nil {
			fmt.Println("Converting price to float failed!")
			fmt.Println(err)
			file.Close()
			return
		}

		prices[lineIndex] = floatPrice

	}

	job.InputPrices = prices

}
```

Shouldn't all this error handling be delegated to the main function by returning error? How can I tell whether an error should be returned or handled immediately?

> # PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP

Yes. **In this particular example, returning the errors is generally the better Go design.**

The key principle is:

> **A function should usually return an error when its caller can reasonably decide what to do about that error.**

Your `LoadData()` method currently decides to print errors and silently return. That makes the function harder to reuse and harder for `main` (or another caller) to control.

### 1. Your current approach

You have:

```go
func (job *TaxIncludedPricesJob) LoadData() {
    file, err := os.Open("prices.txt")

    if err != nil {
        fmt.Println("Could not open the file!")
        fmt.Println(err)
        return
    }

    // ...
}
```

The problem is that `LoadData()` is doing **two jobs**:

1. Performing the operation.
2. Deciding how the application should respond to failure.

For example, perhaps `main()` wants to:

* print the error,
* retry,
* log it,
* terminate the application,
* return an HTTP 500,
* report it to monitoring,
* or ignore it.

`LoadData()` shouldn't necessarily make that decision.

---

## 2. Better design

Return the error:

```go
func (job *TaxIncludedPricesJob) LoadData() error {
    file, err := os.Open("prices.txt")

    if err != nil {
        return err
    }

    scanner := bufio.NewScanner(file)

    var lines []string

    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        file.Close()
        return err
    }

    prices := make([]float64, len(lines))

    for lineIndex, line := range lines {
        floatPrice, err := strconv.ParseFloat(line, 64)

        if err != nil {
            file.Close()
            return err
        }

        prices[lineIndex] = floatPrice
    }

    job.InputPrices = prices

    file.Close()

    return nil
}
```

Then:

```go
func main() {
    job := TaxIncludedPricesJob{}

    if err := job.LoadData(); err != nil {
        fmt.Println("Loading prices failed:")
        fmt.Println(err)
        return
    }

    // Continue...
}
```

Now the responsibilities are clearer:

```text
LoadData()
   │
   ├── tries to load data
   │
   └── returns error
          │
          ▼
        main()
          │
          ├── print it
          ├── log it
          ├── retry
          └── terminate
```

---

# 3. But should `main` handle every error?

Not exactly.

This is an important distinction:

> **Return an error when the caller can make a meaningful decision about it. Handle an error locally when the function itself is the appropriate place to recover from it.**

Consider two different situations.

### Situation A — return it

```go
func LoadConfig() error {
    file, err := os.Open("config.json")
    if err != nil {
        return err
    }

    // ...
}
```

The function can't really know what the application should do if the configuration file doesn't exist.

So:

```go
func main() {
    if err := LoadConfig(); err != nil {
        log.Fatal(err)
    }
}
```

Makes sense.

---

### Situation B — handle it locally

Suppose you're processing a list of optional files:

```go
for _, filename := range filenames {
    file, err := os.Open(filename)

    if err != nil {
        fmt.Println("Skipping:", filename)
        continue
    }

    // process file
}
```

Here, **skipping the file is part of the function's intended behavior**.

There's no reason to bubble every error up to `main`.

---

# 4. A useful question to ask yourself

When you encounter an error, ask:

> **"Does this function know what the correct response to this error is?"**

If **no** → return the error.

If **yes** → handle it.

For example:

```go
func ReadPrices() ([]float64, error)
```

If the file doesn't exist, does `ReadPrices()` know whether the application should terminate?

Probably not.

So:

```go
return nil, err
```

is appropriate.

But imagine:

```go
func TryParsePrice(s string) float64
```

where the application explicitly defines invalid prices as `0`:

```go
func TryParsePrice(s string) float64 {
    price, err := strconv.ParseFloat(s, 64)

    if err != nil {
        return 0
    }

    return price
}
```

Here, the function **knows the desired behavior**: invalid input becomes zero.

That's local error handling.

---

# 5. There's another important issue in your code

Your function should probably use `defer` for `file.Close()`.

Currently you have:

```go
file, err := os.Open("prices.txt")

if err != nil {
    return
}
```

Once the file is successfully opened, you can immediately do:

```go
defer file.Close()
```

Then you don't need:

```go
file.Close()
```

at every error path.

So:

```go
func (job *TaxIncludedPricesJob) LoadData() error {
    file, err := os.Open("prices.txt")
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    var lines []string

    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return err
    }

    prices := make([]float64, len(lines))

    for lineIndex, line := range lines {
        floatPrice, err := strconv.ParseFloat(line, 64)
        if err != nil {
            return err
        }

        prices[lineIndex] = floatPrice
    }

    job.InputPrices = prices

    return nil
}
```

This is much more idiomatic Go.

---

# 6. You can also add context to errors

Rather than simply:

```go
return err
```

you'll often see:

```go
return fmt.Errorf("failed to open prices file: %w", err)
```

and:

```go
return fmt.Errorf("failed to parse price %q: %w", line, err)
```

Then `main` gets useful context:

```go
if err := job.LoadData(); err != nil {
    log.Fatal(err)
}
```

Potential output:

```text
failed to parse price "abc": strconv.ParseFloat: parsing "abc": invalid syntax
```

The `%w` is important because it **wraps** the original error while preserving it for further inspection with `errors.Is` / `errors.As`.

---

## 7. One subtle point: don't blindly push everything to `main`

It's tempting to think:

> "All errors should eventually reach `main`."

That's not the rule.

A better mental model is:

```text
                 ┌─────────────────────┐
                 │   Can I recover     │
                 │   meaningfully here?│
                 └──────────┬──────────┘
                            │
                    ┌───────┴───────┐
                   YES              NO
                    │                │
                    ▼                ▼
             Handle locally      return error
                                     │
                                     ▼
                              caller decides
```

And the error can travel through several layers:

```text
ParsePrice()
     │
     │ error
     ▼
LoadData()
     │
     │ wraps/returns
     ▼
RunJob()
     │
     │ wraps/returns
     ▼
main()
     │
     │ decides application behavior
     ▼
log.Fatal(...)
```

This is one of the central patterns in Go:

**lower-level functions provide information about failure; higher-level functions decide what that failure means for the application.**

For your `LoadData()` specifically, I would **definitely return the errors rather than print them**, because it is a data-loading operation and doesn't have enough context to decide how the application should react.
