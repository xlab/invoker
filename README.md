## invoker

A package providing supercharged primitive for `exec.Cmd` that redirects StdErr and StdOut for asynchronous and thread-safe reading.

### Example

```go
return func(ctx *gin.Context) {
    out := invoker.Run(ctx, "list", "-fmt", "json")
    defer func() {
        go invoker.DrainOut(out)
    }()

    select {
    case <-ctx.Done():
        ctx.AbortWithStatus(504)
        return
    case r := <-out:
        if r.Error != nil {
            ctx.String(500, "%+v", r.Error)
            return
        }

        ctx.Data(200, "application/json", r.StdOut())
    }
}
```

### Redirect to your logging

```go
ctx := context.Background()

logFn := func(data []byte) (stop bool) {
    logger.Info("Output:", string(data))
    return
}

stdErr := invoker.NewWatchedSafeBuffer(ctx, logFn, nil)

out := invoker.RunWithOutputs(ctx, nil, stdErr, "list", "-fmt", "json")

go invoker.DrainOut(out)
```

### License

[MIT](/LICENSE)
