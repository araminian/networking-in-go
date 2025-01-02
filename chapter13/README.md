# LOGGING AND METRICS

## Event Logging

You need to strike a balance in order to log the right information to answer those questions without overwhelming yourself with irrelevant log lines.

Overzealous logging may suit you fine in development, where you control the scale of testing and overall entropy of your service, but it will quickly degrade your ability to find the needle in the haystack when you need to diagnose production failures.

In addition to figuring out what to log, you need to consider that logging isn’t free. It consumes CPU and I/O time your application could otherwise use. A log entry added to a busy for loop while in development may help you understand what your service is doing. But it may become a bottleneck in production, insidiously adding latency to your service.

### The log Package

These require us to go beyond the package-level logger and
instantiate our own *log.Logger instance. You can do this using the log.New
function:

```go
// func New(out io.Writer, prefix string, flag int) *Logger
logger := log.New(os.Stdout, "myapp: ", log.LstdFlags)
```

Accepting an `io.Writer` means the logger can write to anything that satisfies that interface, including an in-memory buffer or a network socket.

The default logger writes its output to `os.Stderr, standard error.`

```go
func Example_log() {
l := log.New(os.Stdout, "example: ", log.Lshortfile)
l.Print("logging to standard output")
// Output:
// example: log_test.go:12: logging to standard output
}
```

The flags of the default logger are `log.Ldate` and `log.Ltime`, collectively `log.LstdFlags`, which print the timestamp of each log entry. Since you want to simplify the output for testing purposes when you run the example on the command line, you omit the timestamp and configure the logger to write the source code filename and line of each log entry 

`log.Lshortfile` is a flag that tells the logger to write the source code filename and line of each log entry.

Recognizing that the logger accepts an `io.Writer`, you may realize this allows you to use multiple writers, such as a log file and standard output or an in-memory ring buffer and a centralized logging server over a network. Unfortunately, the `io.MultiWriter` is not ideal for use in logging. An `io.MultiWriter` writes to each of its writers in sequence, aborting if it receives an error from any `Write` call. This means that if you configure your `io.MultiWriter` to write to a log file and standard output in that order, standard output will never receive the log entry if an error occurred when writing to the log file.

Fear not. It’s an easy problem to solve. Let’s create our own `io.MultiWriter` implementation, which sustains writes across its writers and accumulates any errors it encounters.

