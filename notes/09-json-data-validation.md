# Guide 9: JSON & Data Validation

> Pydantic does two jobs at once so smoothly that it's easy to forget they're separate: it **parses** JSON into typed Python objects, and it **validates** that the data actually satisfies your constraints, raising a single, richly-detailed error if not. Go's standard library only does the first job — `encoding/json` marshals and unmarshals, full stop, with no opinion about whether an email field actually looks like an email. Validation is a separate, deliberate step, usually via a dedicated library. This guide covers both halves in the depth you'll need for the Phase F capstone, where every request body your REST API accepts will go through exactly this pipeline: unmarshal, then validate.

## Table of Contents

1. [Two Philosophies: Implicit Validation vs. Explicit Everything](#1-two-philosophies-implicit-validation-vs-explicit-everything)
2. [`encoding/json` Basics: `Marshal` and `Unmarshal`](#2-encodingjson-basics-marshal-and-unmarshal)
3. [Struct Tags: Controlling JSON Field Names](#3-struct-tags-controlling-json-field-names)
4. [`omitempty` and Optional Fields](#4-omitempty-and-optional-fields)
5. [Unexported Fields Are Invisible to JSON](#5-unexported-fields-are-invisible-to-json)
6. [The Nil vs. Empty Slice/Map Quirk](#6-the-nil-vs-empty-slicemap-quirk)
7. [Embedded Structs and Field Promotion in JSON](#7-embedded-structs-and-field-promotion-in-json)
8. [Decoding into `any`: The `float64` Trap](#8-decoding-into-any-the-float64-trap)
9. [Strict Decoding: Rejecting Unknown Fields](#9-strict-decoding-rejecting-unknown-fields)
10. [Streaming JSON: `Encoder` and `Decoder`](#10-streaming-json-encoder-and-decoder)
11. [Custom Marshaling: `MarshalJSON` and `UnmarshalJSON`](#11-custom-marshaling-marshaljson-and-unmarshaljson)
12. [`time.Time` and Custom Date Formats](#12-timetime-and-custom-date-formats)
13. [Validation with `go-playground/validator`](#13-validation-with-go-playgroundvalidator)
14. [Combining Unmarshal + Validate: Your Pydantic Equivalent](#14-combining-unmarshal--validate-your-pydantic-equivalent)
15. [Exercises](#15-exercises)
16. [Key Takeaways](#16-key-takeaways)

---

## 1. Two Philosophies: Implicit Validation vs. Explicit Everything

A Pydantic model does an enormous amount of work from a small amount of declaration:

```python
from pydantic import BaseModel, EmailStr, Field

class CreateUserRequest(BaseModel):
    name: str = Field(min_length=1)
    email: EmailStr
    age: int = Field(ge=0, le=150)

user = CreateUserRequest.model_validate_json(raw_json)
# Parsing AND validation happen in this one call.
# Malformed JSON, a missing field, an invalid email, or age=-5
# all raise the same kind of error: pydantic.ValidationError,
# with a structured list of every problem found.
```

One class definition gets you: JSON parsing, type coercion, field-level constraints, and a unified error type describing every violation at once. This is genuinely one of Pydantic's best qualities, and FastAPI leans on it hard — it's likely a big part of why request validation in your FastAPI projects has felt almost invisible.

Go splits this into two unrelated concerns, each solved by a different tool:

```go
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

var req CreateUserRequest
if err := json.Unmarshal(rawJSON, &req); err != nil {
	// This ONLY catches malformed JSON or type mismatches
	// (e.g., "age": "not a number") — never business-rule violations.
}

validate := validator.New()
if err := validate.Struct(req); err != nil {
	// THIS catches required/email/range violations —
	// a completely separate call, completely separate error type.
}
```

Two struct tags, two library calls, two different error types to handle. This isn't Go being needlessly verbose for its own sake — it's the same "explicit over implicit" philosophy you've seen in every guide so far (explicit error returns instead of exceptions in Guide 5, explicit `context` propagation instead of implicit cancellation in Guide 7): **parsing untrusted bytes into a typed struct** and **checking that struct against business rules** are treated as two separate, independently-testable responsibilities, not one magic call. Once you've internalized that split, the rest of this guide is just filling in the mechanics of each half.

---

## 2. `encoding/json` Basics: `Marshal` and `Unmarshal`

Two functions cover the majority of your usage, and the naming is worth memorizing precisely because Go's convention is the reverse of what "serialize/deserialize" terminology might suggest at first glance:

- **`json.Marshal`** — Go struct **→** JSON bytes (serialization).
- **`json.Unmarshal`** — JSON bytes **→** Go struct (deserialization/parsing).

```go
package main

import (
	"encoding/json"
	"fmt"
)

type Product struct {
	Name  string
	Price float64
	InStock bool
}

func main() {
	p := Product{Name: "Widget", Price: 9.99, InStock: true}

	// Marshal: struct -> JSON bytes
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// {"Name":"Widget","Price":9.99,"InStock":true}

	// Unmarshal: JSON bytes -> struct
	var p2 Product
	jsonInput := []byte(`{"Name":"Gadget","Price":19.99,"InStock":false}`)
	if err := json.Unmarshal(jsonInput, &p2); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", p2)
	// {Name:Gadget Price:19.99 InStock:false}
}
```

Notice immediately: **`Unmarshal` takes a pointer** (`&p2`), because it needs to mutate the struct you pass it — a direct parallel to the pointer-receiver discussion from Guide 6. Passing a non-pointer value here compiles (the parameter type is `any`/`interface{}`) but fails at runtime with `json: Unmarshal(non-pointer main.Product)`, which is a mistake you will make at least once and immediately recognize once you've seen the error text.

Also notice the default field names in the marshaled output: `Name`, `Price`, `InStock` — Go capitalizes them because it just uses your Go field names verbatim, and your fields have to be exported (capitalized, per Guide 3) to be visible to `encoding/json` at all (more on this in Section 5). Since idiomatic JSON convention is `camelCase` or `snake_case`, not `PascalCase`, you'll want struct tags to fix this — which is exactly Section 3.

---

## 3. Struct Tags: Controlling JSON Field Names

Guide 3 introduced struct tags briefly as string metadata attached to fields; here's their most common real use:

```go
type Product struct {
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	InStock bool    `json:"in_stock"`
}
```

```go
p := Product{Name: "Widget", Price: 9.99, InStock: true}
data, _ := json.Marshal(p)
fmt.Println(string(data))
// {"name":"Widget","price":9.99,"in_stock":true}
```

The tag `json:"name"` tells `encoding/json` "use `name` as the JSON key for this field, both when marshaling and unmarshaling" — the same job a Pydantic `Field(alias="name")` or a `model_config = ConfigDict(alias_generator=to_camel)` does, except applied per-field rather than via a global config option (Go doesn't have a project-wide "convert all field names to snake_case" switch — you tag each field explicitly, or write a small code-generation step if you truly want to automate it across a large struct set).

A tag of `json:"-"` excludes a field from JSON entirely, in both directions — the equivalent of a Pydantic `Field(exclude=True)` or marking a field with a leading underscore convention:

```go
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // never marshaled out, never expected in input
}
```

This is the idiomatic way to keep a sensitive field (a password hash, an internal-only ID) present on your Go struct for internal use but categorically absent from any JSON your API produces or accepts — no risk of it leaking into a response body by accident, since it's not a runtime check, it's baked into how the (de)serializer itself behaves.

---

## 4. `omitempty` and Optional Fields

Add `,omitempty` after the field name in the tag to **skip that field in the marshaled output if it holds its zero value**:

```go
type Product struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
}

p := Product{Name: "Widget", Price: 9.99} // Description left as zero value ""
data, _ := json.Marshal(p)
fmt.Println(string(data))
// {"name":"Widget","price":9.99}   <-- "description" key is entirely absent
```

This maps onto Pydantic's `Optional[str] = None` combined with `model_dump(exclude_none=True)` — "don't include this field in the output if it wasn't meaningfully set." But there's a sharp edge worth calling out explicitly: **`omitempty` triggers on the type's zero value, not on some separate "was this explicitly set" flag.** For a `string`, that's `""`; for an `int`, that's `0`; for a `bool`, that's `false`; for a slice/map/pointer, that's `nil`. This means **a legitimately-meaningful zero — a price of `0.0`, a quantity of `0`, a boolean explicitly set to `false`** — is indistinguishable from "not set" and gets omitted too:

```go
type Product struct {
	Name     string `json:"name"`
	OnSale   bool   `json:"on_sale,omitempty"` // explicitly false is indistinguishable from "unset"
}

p := Product{Name: "Widget", OnSale: false} // deliberately false, not "unset"
data, _ := json.Marshal(p)
fmt.Println(string(data))
// {"name":"Widget"}   <-- on_sale silently vanished, even though it was meaningfully false
```

This is a real, common source of bugs, and the idiomatic fix is the same one you'd reach for in Pydantic to distinguish "explicitly null" from "not provided" — **use a pointer** for fields where the zero value is ambiguous with "not set":

```go
type Product struct {
	Name   string `json:"name"`
	OnSale *bool  `json:"on_sale,omitempty"` // nil = "not provided"; &false = "explicitly false"
}
```

```go
onSale := false
p := Product{Name: "Widget", OnSale: &onSale}
data, _ := json.Marshal(p)
fmt.Println(string(data))
// {"name":"Widget","on_sale":false}   <-- now correctly present
```

With `*bool`, the only value that triggers `omitempty` is `nil` (genuinely "not set") — an explicit `false`, wrapped in a pointer, is a distinct, present value. This mirrors exactly the reasoning behind Pydantic's `Optional[bool] = None` vs. a `bool` field with a default — the Go version just makes the "is this pointer nil" check syntactically visible everywhere the field is used, rather than hidden inside Pydantic's model machinery. The tradeoff, predictably: you now have to dereference (`*p.OnSale`) and nil-check everywhere you *use* the field, which is real, ongoing ergonomic cost for the precision you're buying.

---

## 5. Unexported Fields Are Invisible to JSON

This follows directly from Guide 3's encapsulation rules, but it's worth stating as its own rule because it silently breaks JSON handling in a way that produces no error at all:

```go
type Order struct {
	ID       string  // exported — visible to JSON
	total    float64 // unexported (lowercase) — INVISIBLE to JSON, always
}

o := Order{ID: "abc123", total: 99.99}
data, _ := json.Marshal(o)
fmt.Println(string(data))
// {"ID":"abc123"}   <-- "total" is silently gone, no error, no warning
```

`encoding/json` uses Go's reflection package to inspect struct fields at runtime, and reflection can only see **exported** fields (Guide 3's capitalization rule) from outside the field's own package — there is no tag, no configuration flag, no workaround that makes an unexported field visible to `Marshal`/`Unmarshal`. If you need a field present in JSON, it must be exported (capitalized), full stop. This has no direct Pydantic parallel, since Python doesn't have a language-level exported/unexported distinction the way Go's capitalization convention creates one — the nearest mental model is a Python attribute prefixed with a single underscore that you've also excluded via `Field(exclude=True)`, except in Go this exclusion is automatic and unconditional rather than something you opt into per-field.

The practical consequence for API design: any field you want serialized in a request or response body needs to be an exported struct field, which is one more reason Go structs meant to cross a JSON boundary tend to be flat, plainly-named, and entirely public — private computation or derived state lives on *different* internal types, not hidden fields on your wire-format struct.

---

## 6. The Nil vs. Empty Slice/Map Quirk

This one trips up almost everyone coming from Python, because Python doesn't distinguish "no list" from "empty list" at the type level the way Go distinguishes a `nil` slice from an empty-but-non-nil one:

```go
type Response struct {
	Items []string `json:"items"`
}

var r1 Response                          // Items is nil (zero value for a slice)
data1, _ := json.Marshal(r1)
fmt.Println(string(data1))
// {"items":null}

r2 := Response{Items: []string{}}        // Items is an empty, non-nil slice
data2, _ := json.Marshal(r2)
fmt.Println(string(data2))
// {"items":[]}
```

**A `nil` slice marshals to JSON `null`; a non-nil, empty slice marshals to JSON `[]`.** These look nearly identical in Go source (`var r1 Response` vs. explicitly initializing `Items: []string{}`), but they produce meaningfully different JSON — and for an API consumer, `"items": null` and `"items": []` often mean genuinely different things ("no data" vs. "confirmed empty result"), the same distinction Pydantic would let you express via `Optional[list[str]] = None` vs. a plain `list[str] = []` default.

The practical rule worth internalizing: **if an API response should always return a JSON array (even an empty one) rather than sometimes returning `null`, explicitly initialize the slice** — e.g., `items := []string{}` rather than `var items []string` — before marshaling it. This is an easy detail to get wrong the first several times, and it's exactly the kind of thing that looks fine in Go source, passes `go vet`, passes compilation, and only shows up as a bug when a frontend or another service chokes on an unexpected `null` where it expected `[]`.

The same nil/non-nil distinction applies to maps (`map[string]int` zero value is `nil`, marshaling to `null`; `map[string]int{}` marshals to `{}`), for exactly the same underlying reason — both are reference types whose zero value is genuinely "no underlying data structure at all," not "an empty one" (a detail Guide 2 touched on for maps and slices generally, now showing up concretely at a JSON boundary).

---

## 7. Embedded Structs and Field Promotion in JSON

Guide 3 covered struct embedding for composition. `encoding/json` respects the same field-promotion rules by default — an embedded struct's fields are flattened into the parent's JSON representation, not nested under a key matching the embedded type's name:

```go
type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type Customer struct {
	Name string `json:"name"`
	Address      // embedded, no field name — fields promoted
}

c := Customer{
	Name: "Pedro",
	Address: Address{Street: "123 Main St", City: "Lisbon"},
}
data, _ := json.Marshal(c)
fmt.Println(string(data))
// {"name":"Pedro","street":"123 Main St","city":"Lisbon"}
//                  ^^^^^^ flattened, NOT nested under an "address" key
```

If you *want* nesting (a JSON key that mirrors the Go-level structural grouping), use a **named** field instead of an embedded one — the same choice Guide 3 framed as "composition (named field) vs. embedding (promoted fields)," now with a directly visible JSON consequence:

```go
type Customer struct {
	Name    string  `json:"name"`
	Address Address `json:"address"` // named field, not embedded — nests under "address"
}
```

```go
// {"name":"Pedro","address":{"street":"123 Main St","city":"Lisbon"}}
```

This is a genuinely useful lever once you know it's there: whether a related group of fields should flatten into the parent JSON object or nest under their own key is decided purely by whether you embed the struct or declare it as a named field — the same Go-level choice, two different wire formats, no extra annotation needed. Pydantic's closest equivalent would be manually flattening fields with something like a custom `model_dump` override or restructuring your model hierarchy — Go gets the flattening behavior for free from the same embedding mechanic you already learned for method promotion.

---

## 8. Decoding into `any`: The `float64` Trap

Sometimes you genuinely don't know the shape of incoming JSON ahead of time — you're decoding into `map[string]any` (Guide 4's `any`) rather than a concrete struct:

```go
var data map[string]any
jsonInput := []byte(`{"name": "Widget", "quantity": 5, "price": 9.99, "active": true, "tags": ["a", "b"]}`)
json.Unmarshal(jsonInput, &data)

for key, value := range data {
	fmt.Printf("%s: %v (%T)\n", key, value, value)
}
// name: Widget (string)
// quantity: 5 (float64)      <-- NOT int!
// price: 9.99 (float64)
// active: true (bool)
// tags: [a b] ([]interface {})
```

**Every JSON number decoded into `any` becomes a Go `float64`, never an `int`** — even `5`, which looks like an obvious integer in the source JSON. This is because JSON itself has only one numeric type (it doesn't distinguish "integer" from "float" the way Go's type system does), and `encoding/json`'s default behavior when it doesn't know your intended Go type is to pick the most general numeric type that can hold any JSON number. This means code like the following is a subtle, easy-to-write bug:

```go
quantity := data["quantity"].(int) // PANICS: interface conversion, data["quantity"] is float64, not int
```

```go
quantity := data["quantity"].(float64) // correct — then convert if you need an int
quantityInt := int(quantity)
```

Python's `json.loads` avoids this entirely because Python's own numeric tower already distinguishes `int` from `float` at parse time and preserves the distinction (`json.loads('{"x": 5}')['x']` really is a Python `int`) — this is one case where Go's stricter, single-numeric-JSON-type reality surfaces a wrinkle Python's dynamic typing simply doesn't have to confront. The general lesson, consistent with Guide 4's nil-interface-gotcha and type-assertion coverage: **decoding into `any` should be treated as an escape hatch for genuinely unknown shapes, not a convenience**, and the moment you know the expected shape, define a concrete struct and unmarshal into that instead — you get real `int`/`float64`/`string` typing back, and the compiler (not a runtime panic) tells you if you use a field incorrectly.

---

## 9. Strict Decoding: Rejecting Unknown Fields

By default, `json.Unmarshal` silently ignores any JSON key that doesn't match a field on your target struct — permissive by default, the opposite of Pydantic's default (which, depending on version and config, often rejects or warns on unrecognized fields, and definitely lets you opt into `model_config = ConfigDict(extra="forbid")`):

```go
type CreateUserRequest struct {
	Name string `json:"name"`
}

var req CreateUserRequest
jsonInput := []byte(`{"name": "Pedro", "isAdmn": true}`) // typo'd field, silently ignored
json.Unmarshal(jsonInput, &req)
fmt.Printf("%+v\n", req)
// {Name:Pedro}   <-- no error, no warning that "isAdmn" was never used
```

This silence is a real risk in an API context — a client with a typo'd or subtly wrong field name gets no feedback that their request didn't do what they thought. To opt into strict, Pydantic-`extra="forbid"`-style behavior, you need `json.NewDecoder` (Section 10) plus `DisallowUnknownFields`:

```go
decoder := json.NewDecoder(bytes.NewReader(jsonInput))
decoder.DisallowUnknownFields()

var req CreateUserRequest
if err := decoder.Decode(&req); err != nil {
	fmt.Println(err)
	// json: unknown field "isAdmn"
}
```

Note this is only available via the `Decoder` API, not plain `json.Unmarshal` — one more reason the streaming decoder (next section) tends to be the version you actually reach for in HTTP handlers, where you're decoding a single request body exactly once anyway.

---

## 10. Streaming JSON: `Encoder` and `Decoder`

`json.Marshal`/`json.Unmarshal` operate on an entire `[]byte` buffer already fully in memory. `json.NewEncoder`/`json.NewDecoder` operate directly on any `io.Writer`/`io.Reader` (Guide 4's standard-library interfaces) — which matters for HTTP handlers, where your request body arrives as a stream (`r.Body`, an `io.ReadCloser`) and you'd otherwise have to read it all into a `[]byte` yourself first with `io.ReadAll` before calling `Unmarshal`:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	decoder := json.NewDecoder(r.Body) // reads directly from the request stream
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CreateUserResponse{ID: "generated-id", Name: req.Name}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w) // writes directly to the response stream
	if err := encoder.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
```

This is the shape you'll write constantly once you reach Phase E — `json.NewDecoder(r.Body).Decode(&req)` to parse an incoming request, `json.NewEncoder(w).Encode(resp)` to write a response — and it's directly analogous to how FastAPI/Starlette read the request body as a stream under the hood, except here it's something you wire up explicitly rather than something a framework does invisibly via Pydantic model parameters.

One practical note: `Decoder.Decode` reads exactly one JSON value from the stream and stops — useful for a stream containing multiple, back-to-back JSON values (e.g., newline-delimited JSON logs), where you'd call `Decode` repeatedly in a loop until it returns `io.EOF`, a pattern with no single-call `json.loads` equivalent in Python (you'd typically split on newlines yourself, or use a library, for JSONL).

---

## 11. Custom Marshaling: `MarshalJSON` and `UnmarshalJSON`

Sometimes a type's default JSON representation isn't the one you want — a custom enum-like type that should serialize as a string rather than its underlying int, a value that needs bespoke formatting. `encoding/json` checks whether your type implements one of two interfaces (Guide 4's implicit interface satisfaction, now doing real work) and, if so, defers to your implementation entirely:

```go
type json.Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type json.Unmarshaler interface {
	UnmarshalJSON([]byte) error
}
```

A worked example — a `Status` type backed by an `int` that should read/write as a human-readable string in JSON:

```go
type Status int

const (
	StatusPending Status = iota
	StatusActive
	StatusClosed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusActive:
		return "active"
	case StatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String()) // delegate to json.Marshal for a plain string
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "pending":
		*s = StatusPending
	case "active":
		*s = StatusActive
	case "closed":
		*s = StatusClosed
	default:
		return fmt.Errorf("unknown status: %q", str)
	}
	return nil
}
```

```go
type Order struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

o := Order{ID: "abc123", Status: StatusActive}
data, _ := json.Marshal(o)
fmt.Println(string(data))
// {"id":"abc123","status":"active"}   <-- not "1", thanks to MarshalJSON
```

Note `UnmarshalJSON` is defined on `*Status` (a pointer receiver — required, since it must mutate the value, exactly per Guide 3/6's pointer-receiver rules) while `MarshalJSON` is defined on `Status` (a value receiver is fine, since it only reads). This pairing is the direct structural equivalent of a Pydantic custom validator/serializer pair (`@field_validator` for parsing, a custom `model_serializer`/`field_serializer` for output) — Go just expresses it as two interface methods rather than decorator-annotated methods on a `BaseModel` subclass, consistent with the theme running through this whole guide: the same job, done via explicit interface implementation instead of framework-level annotation magic.

---

## 12. `time.Time` and Custom Date Formats

`time.Time` (standard library) already implements `MarshalJSON`/`UnmarshalJSON`, defaulting to **RFC 3339** format (`"2026-09-12T14:30:00Z"`) — which is also the format FastAPI/Pydantic defaults to for `datetime` fields, so in the common case, nothing special is needed at all:

```go
type Event struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
}

e := Event{Name: "Launch", StartTime: time.Date(2026, 9, 12, 14, 30, 0, 0, time.UTC)}
data, _ := json.Marshal(e)
fmt.Println(string(data))
// {"name":"Launch","start_time":"2026-09-12T14:30:00Z"}
```

If an upstream API gives you a non-RFC3339 date format (a bare date like `"2026-09-12"`, or a Unix timestamp), you need a custom type wrapping `time.Time` with its own `UnmarshalJSON`, since you can't attach a second `UnmarshalJSON` implementation directly to the standard library's `time.Time`:

```go
type DateOnly struct {
	time.Time
}

const dateLayout = "2006-01-02" // Go's reference-date layout string (Guide 2 territory: no strftime-style %Y-%m-%d)

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	t, err := time.Parse(dateLayout, str)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}
```

This embeds `time.Time` (field promotion from Section 7 — `DateOnly` gets all of `time.Time`'s methods like `.Year()`, `.Before()`, for free) while overriding just the JSON-parsing behavior with a narrower format. The `"2006-01-02"` layout string is Go's famously unusual date-formatting convention — rather than `strftime`-style `%Y-%m-%d` placeholder codes, Go uses a specific reference date (`Mon Jan 2 15:04:05 MST 2006`, easy to remember as `01/02 03:04:05PM '06 -0700`) whose components you rearrange to describe your desired layout. It has no Python parallel worth drawing (Python's `strftime` codes are close to universal across languages) — just budget a few minutes the first time you write a custom layout string to look up or recall the reference date, since guessing it from first principles isn't really the intended workflow.

---

## 13. Validation with `go-playground/validator`

Back to Section 1's split: parsing is done, now for the "does this data actually satisfy business rules" half — the [`go-playground/validator`](https://github.com/go-playground/validator) package, by a wide margin the most widely used validation library in the Go ecosystem, and structurally the closest thing to Pydantic's `Field(...)` constraints:

```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

func main() {
	validate := validator.New()

	req := CreateUserRequest{Name: "", Email: "not-an-email", Age: 200}

	if err := validate.Struct(req); err != nil {
		for _, fieldErr := range err.(validator.ValidationErrors) {
			fmt.Printf("field %s failed on %q\n", fieldErr.Field(), fieldErr.Tag())
		}
		// field Name failed on "required"
		// field Email failed on "email"
		// field Age failed on "lte"
	}
}
```

Notice the `validate:"..."` tag lives right alongside the `json:"..."` tag on the same field — the same struct carries both the wire-format mapping and the business-rule constraints, so there's still just one type to look at, even though (per Section 1) two entirely separate function calls process it. Common validation tags, mapped against their nearest Pydantic equivalent:

| `validator` tag | Meaning | Nearest Pydantic equivalent |
|---|---|---|
| `required` | Must not be the zero value | `Field(...)` with no default (required field) |
| `email` | Must be a syntactically valid email | `EmailStr` |
| `min=N` / `max=N` | String length / slice length / numeric bounds | `Field(min_length=N)` / `Field(max_length=N)` |
| `gte=N` / `lte=N` | Numeric ≥ / ≤ | `Field(ge=N)` / `Field(le=N)` |
| `oneof=a b c` | Value must be one of a fixed set | `Literal["a", "b", "c"]` |
| `url` | Must be a syntactically valid URL | Custom validator or `AnyUrl` |
| `dive` | Apply validation to each element of a slice/map | Pydantic validates list element types automatically |

`err.(validator.ValidationErrors)` is a type assertion (Guide 4) unpacking into a slice of individual field errors — each with `.Field()` (which field failed), `.Tag()` (which rule failed), and a few other accessors — giving you the same "here's a structured list of every problem, not just the first one" experience Pydantic's `ValidationError.errors()` provides, just via a type-asserted slice rather than a built-in exception's method. As with `testify` in Guide 8, this is a widely-used, well-regarded, production-grade library — reaching for it is the idiomatic default for anything beyond the most trivial validation, not a shortcut you should feel you're "cheating" by using.

---

## 14. Combining Unmarshal + Validate: Your Pydantic Equivalent

Putting Sections 10, 11, and 13 together into the shape you'll actually write in every Phase F handler — this is the closest Go gets, end to end, to a single Pydantic `model_validate_json` call:

```go
func decodeAndValidate(r *http.Request, dst any, validate *validator.Validate) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err) // %w wraps the error — Guide 5 territory
	}

	if err := validate.Struct(dst); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeAndValidate(r, &req, sharedValidator); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// req is now guaranteed well-formed JSON AND business-rule-valid —
	// exactly the guarantee a Pydantic model gives you at the top of a FastAPI route.
	// ... create the user ...
}
```

Two errors can occur — malformed JSON (Section 9/10's `Decoder`, wrapped with `%w`) or a failed business rule (Section 13's `validator`) — and both are surfaced through the same explicit, two-step, two-`if`-statement shape this whole guide has been building toward. It's more lines than `CreateUserRequest.model_validate_json(raw_json)`, but every step is visible, individually testable (you could unit test the decode step and the validate step in complete isolation, using Guide 8's table-driven tests), and there's no framework-level magic translating a raised exception into an HTTP 422 behind your back — in Phase E, you'll be the one deciding exactly which HTTP status code each of these two failure modes maps to.

---

## 15. Exercises

1. **Marshal round-trip with tags.** Define a `Book` struct with `Title`, `Author`, `PublishedYear`, and `ISBN` fields, tagged with `snake_case` JSON names. Marshal an instance, print the JSON, then unmarshal that JSON back into a fresh `Book` and confirm the two structs are equal.

2. **The `omitempty` zero-value trap.** Define a struct with an `int` field called `Discount` tagged `json:"discount,omitempty"`. Create one instance with `Discount: 0` (genuinely no discount) and marshal it; confirm the key vanishes. Now change the field to `*int`, set it to a pointer to `0` (an explicit, meaningful zero discount), and confirm the key now appears as `"discount":0`. Write a short comment explaining when each version is the correct modeling choice.

3. **Unexported field silently dropped.** Define a struct with one exported and one unexported field, marshal an instance, and confirm — by printing the JSON string — that the unexported field is simply absent, with no error raised anywhere.

4. **Nil vs. empty slice.** Write a function `SearchResults(query string) []string` that returns `nil` if `query == ""` and `[]string{}` otherwise (regardless of whether any results were "found" in this exercise). Marshal both return values from two separate calls and print both JSON strings side by side, confirming one is `null` and the other is `[]`.

5. **Embedding vs. named field.** Take the `Customer`/`Address` example from Section 7 and write it both ways (embedded and as a named field). Marshal both and confirm one flattens `street`/`city` into the top level and the other nests them under `"address"`.

6. **The `float64` trap, on purpose.** Unmarshal a JSON object containing an integer field into a `map[string]any`. Attempt a type assertion to `int` and confirm it panics; catch the panic with `recover` (or just observe it once, then fix it) and correct the code to assert to `float64` instead, converting to `int` only afterward.

7. **Strict decoding.** Define a small struct and a JSON payload with one extra, unexpected field. First decode with plain `json.Unmarshal` and confirm no error occurs. Then decode again using `json.NewDecoder` with `DisallowUnknownFields()` and confirm you get an explicit error naming the offending field.

8. **Custom `MarshalJSON`/`UnmarshalJSON`.** Define your own enum-like `Priority` type (e.g., `Low`, `Medium`, `High`, backed by an `int`) with `MarshalJSON`/`UnmarshalJSON` methods so it serializes as a lowercase string. Write a table-driven test (Guide 8) covering both directions: marshaling each `Priority` value to its expected string, and unmarshaling each valid string back to the correct value, plus one case asserting an unmarshal error for an invalid string like `"urgent"`.

9. **Full pipeline.** Define a `CreateOrderRequest` struct with at least three fields and appropriate `validate` tags (mixing `required`, a numeric bound, and one `oneof`). Write a `decodeAndValidate`-style function (Section 14) and three test cases: valid input (no error), malformed JSON (decode error), and well-formed-but-invalid JSON (validation error) — confirming each produces the right *kind* of failure, not just "some error occurred."

---

## 16. Key Takeaways

- Go splits what Pydantic does in one step into two explicit, independent steps: **`encoding/json`** parses bytes into a typed struct (or vice versa); a separate library, typically **`go-playground/validator`**, checks that struct against business rules. No single call does both.
- **`json.Marshal`** (struct → JSON bytes) and **`json.Unmarshal`** (JSON bytes → struct, always via a pointer argument) are the two core functions; struct tags like `` `json:"name"` `` control the JSON key used for each field, and `` `json:"-"` `` excludes a field entirely.
- **`,omitempty`** omits a field from marshaled output only when it holds its *zero value* — which makes a meaningful `0`, `false`, or `""` indistinguishable from "not set." Use a pointer (`*bool`, `*int`) when that distinction matters, exactly as you'd use `Optional[T] = None` in Pydantic to separate "unset" from "the zero value."
- **Unexported (lowercase) struct fields are completely invisible to `encoding/json`** — no tag or configuration can change this; any field you want serialized must be exported.
- **A `nil` slice/map marshals to JSON `null`; a non-nil, empty slice/map marshals to `[]`/`{}`.** These look nearly identical in Go source but produce different JSON — initialize explicitly (`[]string{}`, not `var x []string`) when your API should always return an array.
- **Embedding a struct promotes and flattens its fields into the parent's JSON**; using a named field instead nests them under their own key — the same embedding-vs-composition choice from Guide 3, now with a direct, visible JSON consequence.
- **Every JSON number decoded into `any`/`map[string]any` becomes a Go `float64`, never an `int`**, because JSON has only one numeric type — asserting straight to `int` on such a value panics. Treat decoding into `any` as an escape hatch, not a default; prefer concrete structs whenever the shape is known.
- **`json.Unmarshal` silently ignores unrecognized JSON fields by default.** Opt into strict, Pydantic-`extra="forbid"`-style rejection with `json.NewDecoder(...).DisallowUnknownFields()`.
- **`json.NewDecoder`/`json.NewEncoder`** operate directly on `io.Reader`/`io.Writer` streams (like an HTTP request body or response writer) rather than requiring a fully-buffered `[]byte`, and are the pair you'll use in essentially every Phase E handler.
- Implementing the **`MarshalJSON`/`UnmarshalJSON`** interface methods (the latter on a pointer receiver, since it mutates) lets a type control its own JSON representation entirely — the Go equivalent of a Pydantic custom field validator/serializer pair, done via explicit interface implementation rather than decorators.
- **`time.Time`** already implements both methods, defaulting to RFC 3339 — the same default Pydantic uses for `datetime` — so most date/time fields need no extra code; a custom wrapper type is only needed for non-standard incoming formats.
- **`go-playground/validator`**'s `validate:"..."` struct tags (`required`, `email`, `gte=`/`lte=`, `oneof=`, and others) are the closest Go equivalent to Pydantic's `Field(...)` constraints, living right alongside the `json:"..."` tag on the same struct, and returning a `validator.ValidationErrors` slice describing every violation at once, not just the first.
- The full request-handling idiom you'll use throughout Phase F is: **decode with a strict `Decoder`, then validate with `validator.Struct`**, treating "malformed JSON" and "well-formed but invalid data" as two distinct, separately-handled error cases — a deliberate, visible version of what a single Pydantic `model_validate_json` call does invisibly underneath.

---

**Next up: Guide 10** — let me know what's next in your sequence and I'll pick up in the same format.
