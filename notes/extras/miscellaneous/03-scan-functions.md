In Go, “Scan functions” usually refers to input-reading functions from the `fmt` package:

```go
fmt.Scan()
fmt.Scanln()
fmt.Scanf()
```

There are also related versions that read from strings or any `io.Reader`:

```go
fmt.Sscan()
fmt.Sscanln()
fmt.Sscanf()

fmt.Fscan()
fmt.Fscanln()
fmt.Fscanf()
```

They all parse text input into variables.

---

# 1. `fmt.Scan`

`fmt.Scan` reads values from **standard input**.

```go
package main

import "fmt"

func main() {
	var name string
	var age int

	fmt.Print("Enter name and age: ")
	fmt.Scan(&name, &age)

	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
}
```

Example input:

```text
Pedro 25
```

Output:

```text
Name: Pedro
Age: 25
```

Important: you must pass **pointers**:

```go
fmt.Scan(&name, &age)
```

Not:

```go
fmt.Scan(name, age) // wrong
```

Because `Scan` needs to modify the variables.

---

## How `Scan` reads input

`fmt.Scan` separates input by **whitespace**.

Whitespace means:

```text
space
tab
newline
```

So these inputs are equivalent:

```text
Pedro 25
```

```text
Pedro
25
```

```text
Pedro        25
```

All work with:

```go
fmt.Scan(&name, &age)
```

---

## Limitation of `Scan`: it cannot easily read full lines with spaces

Example:

```go
var fullName string

fmt.Print("Enter your full name: ")
fmt.Scan(&fullName)

fmt.Println(fullName)
```

Input:

```text
Pedro Lucas
```

Output:

```text
Pedro
```

`Scan` stops reading `fullName` when it reaches the first whitespace.

So `fmt.Scan` is good for simple token-based input, like:

```go
var age int
var salary float64
var active bool
var name string
```

But it is not good for full sentences or names with spaces.

---

# 2. `fmt.Scanln`

`fmt.Scanln` is similar to `Scan`, but it stops scanning at the end of the current line.

```go
package main

import "fmt"

func main() {
	var name string
	var age int

	fmt.Print("Enter name and age: ")
	fmt.Scanln(&name, &age)

	fmt.Println(name, age)
}
```

Input:

```text
Pedro 25
```

Works.

But this input:

```text
Pedro
25
```

does **not** work the same way, because `Scanln` expects the values to be on the same line.

---

## Difference between `Scan` and `Scanln`

### `Scan`

```go
fmt.Scan(&a, &b)
```

Can read across multiple lines:

```text
10
20
```

Works.

### `Scanln`

```go
fmt.Scanln(&a, &b)
```

Usually expects both values before the newline:

```text
10 20
```

Works.

But:

```text
10
20
```

does not satisfy both values in the same scan operation.

---

## Limitation of `Scanln`

Despite the name, `Scanln` does **not** read an entire line as a string.

Example:

```go
var sentence string

fmt.Scanln(&sentence)
```

Input:

```text
hello world
```

Output:

```text
hello
```

It still scans token by token. It just stops at the newline.

So `Scanln` is also not ideal for full-line text input.

---

# 3. `fmt.Scanf`

`fmt.Scanf` reads input according to a **format string**.

It is similar to C’s `scanf`.

```go
package main

import "fmt"

func main() {
	var name string
	var age int

	fmt.Print("Enter data: ")
	fmt.Scanf("%s %d", &name, &age)

	fmt.Println(name, age)
}
```

Input:

```text
Pedro 25
```

Output:

```text
Pedro 25
```

---

## Common `Scanf` verbs

| Verb | Meaning               |
| ---- | --------------------- |
| `%s` | string without spaces |
| `%d` | integer               |
| `%f` | floating-point number |
| `%t` | boolean               |
| `%c` | character/rune        |
| `%v` | default format        |

Example:

```go
var id int
var price float64
var active bool

fmt.Scanf("%d %f %t", &id, &price, &active)
```

Input:

```text
10 49.90 true
```

---

## `Scanf` with literal text

You can force the user input to match a pattern.

```go
var day, month, year int

fmt.Scanf("%d/%d/%d", &day, &month, &year)
```

Input:

```text
27/06/2026
```

Result:

```go
day = 27
month = 6
year = 2026
```

Another example:

```go
var hour, minute int

fmt.Scanf("%d:%d", &hour, &minute)
```

Input:

```text
14:30
```

---

## Limitation of `Scanf`

`Scanf` is stricter than `Scan`.

If the input does not match the expected format, it stops scanning.

Example:

```go
var day, month, year int

n, err := fmt.Scanf("%d/%d/%d", &day, &month, &year)
fmt.Println(n, err)
```

Input:

```text
27-06-2026
```

This does not match `%d/%d/%d`, because the format expects `/`, not `-`.

---

# 4. Return values of Scan functions

All scan functions return:

```go
n, err
```

Where:

```go
n   = number of successfully scanned items
err = error, if something went wrong
```

Example:

```go
var age int

n, err := fmt.Scan(&age)

if err != nil {
	fmt.Println("Error:", err)
	return
}

fmt.Println("Scanned items:", n)
fmt.Println("Age:", age)
```

Input:

```text
abc
```

Output may be similar to:

```text
Error: expected integer
```

You should usually check the error.

---

# 5. `fmt.Sscan`, `fmt.Sscanln`, `fmt.Sscanf`

These read from a **string**, not from the terminal.

## `Sscan`

```go
package main

import "fmt"

func main() {
	input := "Pedro 25"

	var name string
	var age int

	fmt.Sscan(input, &name, &age)

	fmt.Println(name, age)
}
```

Output:

```text
Pedro 25
```

---

## `Sscanf`

Useful when parsing structured strings:

```go
input := "27/06/2026"

var day, month, year int

fmt.Sscanf(input, "%d/%d/%d", &day, &month, &year)

fmt.Println(day, month, year)
```

Output:

```text
27 6 2026
```

---

# 6. `fmt.Fscan`, `fmt.Fscanln`, `fmt.Fscanf`

These read from any `io.Reader`.

That means they can read from:

```go
os.Stdin
files
network connections
strings.Reader
bytes.Buffer
```

Example with `os.Stdin`:

```go
fmt.Fscan(os.Stdin, &name, &age)
```

This is similar to:

```go
fmt.Scan(&name, &age)
```

But more general.

---

## Example with a file

Suppose `data.txt` contains:

```text
Pedro 25
Ana 30
```

Code:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("data.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var name string
	var age int

	for {
		_, err := fmt.Fscan(file, &name, &age)
		if err != nil {
			break
		}

		fmt.Println(name, age)
	}
}
```

Output:

```text
Pedro 25
Ana 30
```

---

# 7. Summary of `fmt` Scan functions

| Function  |     Reads from | Uses format? | Stops at newline? |
| --------- | -------------: | -----------: | ----------------: |
| `Scan`    | standard input |           No |                No |
| `Scanln`  | standard input |           No |               Yes |
| `Scanf`   | standard input |          Yes | Depends on format |
| `Sscan`   |         string |           No |                No |
| `Sscanln` |         string |           No |               Yes |
| `Sscanf`  |         string |          Yes | Depends on format |
| `Fscan`   |    `io.Reader` |           No |                No |
| `Fscanln` |    `io.Reader` |           No |               Yes |
| `Fscanf`  |    `io.Reader` |          Yes | Depends on format |

---

# 8. Main limitations of `fmt.Scan` functions

## 1. They are whitespace-based

This is the biggest limitation.

```go
var name string
fmt.Scan(&name)
```

Input:

```text
Pedro Lucas
```

Only reads:

```text
Pedro
```

Because the space ends the token.

---

## 2. They are inconvenient for full-line input

For full-line input, prefer:

```go
bufio.NewReader(os.Stdin).ReadString('\n')
```

Example:

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your full name: ")
	name, _ := reader.ReadString('\n')

	name = strings.TrimSpace(name)

	fmt.Println("Full name:", name)
}
```

Input:

```text
Pedro Lucas
```

Output:

```text
Full name: Pedro Lucas
```

---

## 3. They can leave unread input behind

Example:

```go
var age int
var name string

fmt.Scan(&age)
fmt.Scanln(&name)
```

Input:

```text
25
Pedro
```

This can behave unexpectedly because the newline after `25` may still affect the next scan.

This is one reason why mixing different input-reading methods can become confusing.

---

## 4. They require correct input types

Example:

```go
var age int
fmt.Scan(&age)
```

Input:

```text
twenty
```

This fails because Go expects an integer.

You should handle errors:

```go
n, err := fmt.Scan(&age)
if err != nil {
	fmt.Println("Invalid input:", err)
	return
}

if n != 1 {
	fmt.Println("Expected one value")
	return
}
```

---

## 5. `Scanf` requires the input to match the format

```go
fmt.Scanf("%d/%d/%d", &day, &month, &year)
```

Expected input:

```text
27/06/2026
```

This does not match:

```text
27-06-2026
```

Because the format expects `/`.

---

## 6. They are not great for robust CLI applications

For serious command-line programs, prefer:

```go
bufio.Reader
bufio.Scanner
flag package
cobra package
```

Use `fmt.Scan` mostly for small examples, exercises, and quick scripts.

---

# 9. `bufio.Scanner`

Apart from `fmt.Scan`, Go also has `bufio.Scanner`.

This is very common for reading input line by line.

```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your full name: ")

	if scanner.Scan() {
		name := scanner.Text()
		fmt.Println("Full name:", name)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}
}
```

Input:

```text
Pedro Lucas
```

Output:

```text
Full name: Pedro Lucas
```

Unlike `fmt.Scan`, this reads the whole line.

---

## Reading multiple lines

```go
scanner := bufio.NewScanner(os.Stdin)

for scanner.Scan() {
	line := scanner.Text()
	fmt.Println("You typed:", line)
}

if err := scanner.Err(); err != nil {
	fmt.Println("Error:", err)
}
```

This keeps reading until EOF.

In the terminal, EOF is usually:

```text
Ctrl+D on Linux/macOS
Ctrl+Z then Enter on Windows
```

---

## Scanner with words instead of lines

By default, `bufio.Scanner` reads line by line.

You can make it read word by word:

```go
scanner := bufio.NewScanner(os.Stdin)
scanner.Split(bufio.ScanWords)

for scanner.Scan() {
	word := scanner.Text()
	fmt.Println(word)
}
```

Input:

```text
hello world from go
```

Output:

```text
hello
world
from
go
```

---

# 10. Limitation of `bufio.Scanner`

`bufio.Scanner` has an important default limitation: the maximum token size is limited.

By default, it can scan tokens up to **64 KiB**.

So this can fail if one line is very long:

```go
scanner := bufio.NewScanner(file)

for scanner.Scan() {
	line := scanner.Text()
	fmt.Println(line)
}

if err := scanner.Err(); err != nil {
	fmt.Println("Error:", err)
}
```

Possible error:

```text
bufio.Scanner: token too long
```

You can increase the buffer:

```go
scanner := bufio.NewScanner(os.Stdin)

buf := make([]byte, 1024)
scanner.Buffer(buf, 1024*1024) // max token size: 1 MB
```

But for very large input, prefer `bufio.Reader`.

---

# 11. When to use each one

## Use `fmt.Scan`

For simple exercises:

```go
var age int
fmt.Scan(&age)
```

Good for:

```text
25
```

---

## Use `fmt.Scanf`

For structured input:

```go
var day, month, year int
fmt.Scanf("%d/%d/%d", &day, &month, &year)
```

Good for:

```text
27/06/2026
```

---

## Use `bufio.Scanner`

For line-by-line input:

```go
scanner := bufio.NewScanner(os.Stdin)
scanner.Scan()
line := scanner.Text()
```

Good for:

```text
Pedro Lucas
```

---

## Use `bufio.Reader`

For full-line input where lines may be long:

```go
reader := bufio.NewReader(os.Stdin)
line, err := reader.ReadString('\n')
```

Good for:

```text
long text with many words
```

---

# 12. Practical rule

For beginner Go programs:

```go
fmt.Scan
```

is fine for simple values.

For real CLI input:

```go
bufio.Scanner
```

is usually better.

For large files or very long lines:

```go
bufio.Reader
```

is safer.

---

# 13. Common beginner mistake

This is wrong:

```go
var age int
fmt.Scan(age)
```

This is correct:

```go
var age int
fmt.Scan(&age)
```

Because `Scan` needs the address of the variable.

---

# 14. Another common issue: reading strings with spaces

This does not read a full name:

```go
var name string
fmt.Scan(&name)
```

Input:

```text
Pedro Lucas
```

Only reads:

```text
Pedro
```

Correct solution:

```go
reader := bufio.NewReader(os.Stdin)

fmt.Print("Enter your name: ")
name, _ := reader.ReadString('\n')
name = strings.TrimSpace(name)
```

---

# 15. Best mental model

Think of `fmt.Scan` like this:

```text
Read the next token.
Convert it to the requested type.
Store it in the variable.
```

Think of `bufio.Scanner` like this:

```text
Read the next line or token.
Give me the raw text.
Then I decide what to do with it.
```

So:

```go
fmt.Scan(&age)
```

means:

```text
Read one value and convert it to int.
```

While:

```go
scanner.Scan()
line := scanner.Text()
```

means:

```text
Read one line as text.
```

Then you can manually parse it with:

```go
strconv.Atoi(line)
strconv.ParseFloat(line, 64)
strings.Fields(line)
fmt.Sscanf(line, ...)
```

For learning Go, the main point is:

```text
fmt.Scan functions are simple but limited.
bufio.Scanner and bufio.Reader are usually better for real input handling.
```
