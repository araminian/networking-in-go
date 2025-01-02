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

### Leveled Log Entries

I wrote earlier in the chapter that verbose logging may be inefficient in production and can overwhelm you with the sheer number of log entries as your service scales up. One way to avoid this is by instituting logging levels, which assign a priority to each kind of event, enabling you to always log high-priority errors but conditionally log low-priority entries more suited for debugging and development purposes.

For example, you’d always want to know if your service is unable to connect to its database, but you may care to log only details about individual connections while in development or when diagnosing a failure.

I recommend you create just a few log levels to begin with. In my experience, you can address most use cases with just an `error` level and a `debug` level, maybe even an `info` level on occasion.

`Error` log entries should accompany some sort of alert, since these entries indicate a condition that needs your attention.

`Info` log entries typically log non-error information. For example, it may be appropriate for your use case to log a successful database connection or to add a log entry when a listener is ready for incoming connections on a network socket.

`Debug` log entries should be verbose and serve to help you diagnose failures, as well as aid development by helping you reason about the workflow.

Although Go’s log package does not have inherent support for leveled log entries, you can add similar functionality by creating separate loggers for each log level you need.

## Structured Logging

The log entries made by the code you’ve written so far are meant for human consumption. They are easy for you to read, since each log entry is little more than a message. This means that finding log lines relevant to an issue involves liberal use of the grep command or, at worst, manually skimming log entries.

But this could become more challenging if the number of log entries increases. You may find yourself looking for a needle in a haystack.
Remember, logging is useful only if you can quickly find the information you need.

A common approach to solving this problem is to add metadata to your log entries and then parse the metadata with software to help you organize them. This type of logging is called structured logging.

Creating structured log entries involves adding `key-value pairs` to each log entry. In these, you may include the time at which you logged the entry, the part of your application that made the log entry, the log level, the hostname or IP address of the node that created the log entry, and other bits of metadata that you can use for indexing and filtering.

Most structured loggers encode log entries as JSON before writing them to log files or shipping them to centralized logging servers. Structured logging makes the whole process of collecting logs in a centralized server easy, since the additional metadata associated with each log entry allows the server to organize and collate log entries across services.

### Using the Zap Logger

Zap allows you to integrate log file rotation.

`Log file rotation` is the process of closing the current log file, renaming it, and then opening a new log file after the current log file reaches a specific age or size threshold. Rotating log files is a good practice to prevent them from filling up your available hard drive space. Plus, searching through smaller, date-delimited log files is more efficient than searching through a single, monolithic log file.

For example, you may want to rotate your log files every week and keep only eight weeks’ worth of rotated log files. 

I’ve used other structured loggers on large projects, and in my experience, Zap causes the least overhead; I can use it in busy bits of code without a noticeable performance hit, unlike other heavyweight structured loggers.

The Zap logger includes zap.Core and its options. The zap.Core has three components: a `log-level threshold`, an `output`, and an `encoder`. 

The `log-level threshold` sets the minimum log level that Zap will log; Zap will simply ignore any log entry below that level, allowing you to leave debug logging in your code and configure Zap to conditionally ignore it.

`Zap’s output` is a `zapcore.WriteSyncer`, which is an `io.Writer` with an additional `Sync` method. Zap can write log entries to any object that implements this interface.

`Zap’s encoder` can encode the log entry before writing it to the output.

### Writing the Encoder

Although Zap provides a few helper functions, such as `zap.NewProduction` or `zap.NewDevelopment`, to quickly create production and development loggers, you’ll create one from scratch, starting with the encoder configuration.

Checkout the `encoder.go` file for the example.

The `encoder configuration` is independent of the `encoder` itself in that you can use the same encoder configuration no matter whether you’re passing it to a JSON encoder or a console encoder.

The `encoder` will use your configuration to dictate its output format.


### Using the Console Encoder

Let’s instead assume you want to log something a bit more humanreadable, yet that has structure. Zap includes a console encoder that’s essentially a drop-in replacement for its JSON encoder. 

Checkout the `console.go` file for the example.