# Slog

This example shows that go-app also configures slog and it is fully functional.

Run the example with:

``` sh
go run ./examples/slog/ -app-log-human
```

You can change the log level and see the INFO messages are not printed anymore.

``` sh
go run ./examples/slog/ -app-log-human -app-log-level=warn
```
