# learning-golang

Go (or Golang) is an open-source programming language created at Google, designed to be fast, simple, and highly concurrent. It has a clean syntax, built-in support for parallelism via goroutines and channels, a powerful standard library, and produces single static binaries. It's commonly used for cloud services, distributed systems, DevOps tooling, and high-performance backends.

## Most Used Go Standard Library Packages

### 📦 Core I/O, OS, and Utilities

- **`fmt`** – Text formatting and printing.
- **`os`** – File system access, environment variables, processes.
- **`io`** – Core I/O primitives (`Reader`, `Writer`, copying streams).
- **`bufio`** – Buffered I/O for improved performance.
- **`path/filepath`** – Cross-platform file path operations.
- **`flag`** – Command-line argument parsing.
- **`log`** – Basic logging utilities.

### 🔤 Strings, Numbers, and Data Handling

- **`strings`** – String manipulation helpers.
- **`strconv`** – Conversions between strings and numeric/bool types.
- **`bytes`** – Efficient operations on byte slices.
- **`encoding/json`** – JSON serialization/deserialization.
- **`encoding/base64`** – Base64 encode/decode operations.
- **`sort`** – Sorting primitives and custom collections.
- **`math`** / **`math/rand`** – Math utilities and randomness.

### ⏱️ Time, Concurrency, and Context

- **`time`** – Time handling, durations, timers, and parsing.
- **`sync`** – Concurrency primitives (mutex, once, wait groups).
- **`context`** – Deadlines, cancellation, and request-scoped data.

### 🌐 Networking and Web

- **`net/http`** – HTTP server and client; widely used in web services.
- **`net`** – TCP/UDP networking primitives.
- **`crypto/*`** – Hashing, encryption, TLS, and other crypto tools.

## Pointers Receivers vs Value Receivers

- Has mutex / shared state / needs mutation / is large? → use \*T.

- Small, immutable-ish value type (like math types)? → use T.

If you’re ever in doubt and the type is used as a service / store / client, default to pointer receivers.

All arguments passed to methods are copied, so using a pointer receiver means only the pointer is copied, not the entire value. This is more efficient for large structs.
