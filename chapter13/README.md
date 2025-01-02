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


### Logging with Different Outputs and Encodings

Zap includes useful functions that allow you to **concurrently log to different outputs, using different encodings, at different log levels.**

Checkout the `concurrent.go` file for the example, it creates a logger that writes JSON to a log file and console encoding to standard output. The logger writes only the debug log entries to the console.

### Sampling Log Entries

One of my warnings to you with regard to logging is to consider how it impacts your application from a CPU and I/O perspective. You don’t want logging to become your application’s bottleneck. This normally means taking special care when logging in the busy parts of your application.

One method to mitigate the logging overhead in critical code paths, such as a loop, is to sample log entries. It may not be necessary to log each entry, especially if your logger is outputting many duplicate log entries. Instead, try logging every nth occurrence of a duplicate entry.

Conveniently, Zap has a logger that does just that.

Checkout the `sampling.go` file for the example, creates a logger that will constrain its CPU and I/O overhead by logging a subset of log entries.

### Performing On-Demand Debug Logging

If debug logging introduces an unacceptable burden on your application
under normal operation, or if the sheer amount of debug log data overwhelms your available storage space, you might want the ability to enable debug logging on demand.

One technique is to use a semaphore file to enable debug logging. A semaphore file is an empty file whose existence is meant as a signal to the logger to change its behavior. If the semaphore file is present, the logger outputs debug-level logs. Once you remove the semaphore file, the logger reverts to its previous log level.

Let’s use the `fsnotify` package to allow your application to watch for filesystem notifications. In addition to the standard library, the `fsnotify` package uses the `x/sys` package. Before you start writing code, let’s make sure
our `x/sys` package is current:

```bash
go get -u golang.org/x/sys
```

Not all logging packages provide safe methods to asynchronously modify log levels. Be aware that you may introduce a race condition if you attempt to modify a logger’s level at the same time that the logger is reading the log level. The Zap logger allows you to retrieve a sync/atomic-based leveler to dynamically modify a logger’s level while avoiding race conditions. You’ll pass the atomic leveler to the zapcore.NewCore function in place of a log level, as you’ve previously done.

The `zap.AtomicLevel` struct implements the `http.Handler` interface. You can integrate it into an API and dynamically change the log level over HTTP instead of using a semaphore.

checkout the `dynamic.go` file for the example. dynamic logging using a semaphore file.

### Scaling Up with Wide Event Logging

Wide event logging is a technique that creates `a single, structured log entry per event to summarize the transaction, instead of logging numerous entries as the transaction progresses.`

This technique is most applicable to requestresponse loops, such as API calls, but it can be adapted to other use cases.

When you summarize transactions in a structured log entry, you reduce the logging overhead while conserving the ability to index and search for transaction details.

One approach to wide event logging is to wrap an API handler in middleware. But first, since the `http.ResponseWriter` is a bit stingy with its output,
you need to create your own response writer type to collect and log the response code and length.

checkout the `wide.go` file for the example.

### Log Rotation with Lumberjack

If you elect to output log entries to a file, you could leverage an application like logrotate to keep them from consuming all available hard drive space.
The downside to using a third-party application to manage log files is that the third-party application will need to signal to your application to reopen its log file handle lest your application keep writing to the rotated log file.

A less invasive and more reliable option is to add log file management directly to your logger by using a library like Lumberjack. Lumberjack handles log rotation in a way that is transparent to the logger, because your logger treats Lumberjack as any other io.Writer. Meanwhile, Lumberjack keeps track of the log entry accounting and file rotation for you.

checkout `rotation.go` file for the example.

## Instrumenting Your Code

Instrumenting your code is the process of collecting metrics for the purpose
of making inferences about the current state of your service—such as the duration of each request-response loop, the size of each response, the number of connected clients, the latency between your service and a third-party API, and so on. Whereas logs provide a record of how your service got into a certain state, metrics give you insight into that state itself.

Instrumentation is easy, so much so that I’m going to give you the opposite advice I did for logging: instrument everything (initially). Fine-grained instrumentation involves hardly any overhead, it’s efficient to ship, and it’s inexpensive to store. Plus, instrumentation can solve one of the challenges of logging I mentioned earlier: that you won’t initially know all the questions you’ll want to ask, particularly for complex systems. An insidious problem may be ready to ruin your weekend because you lack critical metrics to give you an early warning that something is wrong.

This section will introduce you to metric types and show you the basics for using those types in your services. You will learn about Go kit’s metrics package, which is an abstraction layer that provides useful interfaces for popular metrics platforms. 

### Counters

Counters are used for tracking values that only increase, such as request
counts, error counts, or completed task counts. You can use a counter to
calculate the rate of increase for a given interval, such as the number of
connections per minute.

### Gauges

Gauges allow you to track values that increase or decrease, such as the current memory usage, in-flight requests, queue sizes, fan speed, or the number of ThinkPads on my desk. Gauges do not support rate calculations, such as the number of connections per minute or megabits transferred per second, while counters do.

### Histograms and Summaries

A histogram places values into predefined buckets. Each bucket is associated with a range of values and named after its maximum one. When a value is observed, the histogram increments the maximum value of the smallest bucket into which the value fits. In this way, a histogram tracks the frequency of observed values for each bucket.

Let’s look at a quick example. Assuming you have three buckets valued at 0.5, 1.0, and 1.5, if a histogram observes the value 0.42, it will increment the counter associated with bucket 0.5, because 0.5 is the smallest bucket that can contain 0.42. It covers the range of 0.5 and smaller values. If the histogram observes the value 1.23, it will increment the counter associated with the bucket 1.5, which covers values in the range of above 1.0 through 1.5. Naturally, the 1.0 bucket covers the range of above 0.5 through 1.0.

You can use a histogram’s distribution of observed values to estimate a percentage or an average of all values. For example, you could use a histogram to calculate the average request sizes or response sizes observed by your service.

A summary is a histogram with a few differences. First, a histogram requires predefined buckets, whereas a summary calculates its own buckets. Second, the metrics server calculates averages or percentages from histograms, whereas your service calculates the averages or percentages from summaries. As a result, you can aggregate histograms across services on the metrics server, but you cannot do the same for summaries.

The general advice is to use summaries when you don’t know the range
of expected values, but I’d advise you to use histograms whenever possible
so that you can aggregate histograms on the metrics server.

### Instrumenting a Basic HTTP Server

The biggest challenges here are determining what you want to instrument, where best to instrument it, and what metric type is most appropriate for each value you want to track.

`server.go` details the initial code needed for an application that comprises an HTTP server to serve the metrics endpoint and another HTTP server to pass all requests to an instrumented handler.

The `promhttp` package includes an `http.Handler` that a Prometheus server can use to scrape metrics from your application. This handler serves not only your metrics but also metrics related to the runtime, such as the Go version, number of cores, and so on. At a minimum, you can use the metrics provided by the Prometheus handler to gain insight into your service’s memory utilization, open file descriptors, heap and stack details, and more.


