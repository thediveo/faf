# `faf` Fire and Forget (Crash Later)

![Coverage](https://img.shields.io/badge/Coverage-96.6%25-brightgreen)

**faf** is a Go module **for Linux** that springs from working on the system
discovery service [lxkns](https://github.com/thediveo/lxkns). Some of the hot
paths in this and other services happen to comb the [`procfs` process
filesystem](https://man7.org/linux/man-pages/man5/procfs.5.html) extensively,
such as to gather process and task details, up to CPU affinity lists. As many
elements in `procfs` tend to be of comparably small sizes, constantly allocating
new buffer memory to read and process them, and then throwing away the buffers
without any need or reusing the buffers puts unnecessary pressure on the GC and
also results in avoidable CPU load.

> [!CAUTION]
> 
> If you stumbled upon **faf** and just want to use it because you
> "want to make my programs faster", stay away. As with any optimizations as the
> ones provided by **faf** you need to have hot paths that really benefit from
> optimizing them under clearly understood constraints.

Go's standard library has to cover all walks of life in order to provide a high
quality of (Go) living. Yet hot paths of very specific use cases benefit from
running with open scissors. This is where **faf** comes in: reduce dynamic
allocations considerably where they are basically just trashing the GC by
trading in general comfort by deferring heap escapes to user code and only if
the user code really needs them. **faf** optimizes for no error reporting,
because in some use cases the hot paths never reported them, but just skipped.
That's where **faf** draws its name "Fire and Forget (Crash Later)" from.

Interestingly, now that Go 1.23 introduced the iterator pattern, we can
elegantly use optimizations such as when reading directories. Before, even
skipping unnecessary sorting, the canonical form was along these lines:

```go
// before with lots of smaller allocations;
// error handling was used to just bail out, but never for reporting.
dir, _ := os.Open("/my/directory")
defer dir.Close()
entries, _ := dir.ReadDir(-1)
for entry := range entries {
    _ = entry
}
```

We now turn this into a loop over a dedicated iterator that avoids having to
read all entries first, thus avoiding creating heap trash:

```go
// faf with Go 1.23 or later
for entry := range faf.ReadDir("/my/directory") {
    _ = entry
}
```

If the directory cannot be read, then the loop simply doesn't "loop" (that's
what I think loops do for a living).

And to provide some figures, while benchmarks show a 25% increase in execution
speed on reading directories (which you can easily eat up in your loop body),
the real benefit is a constant single allocation for a full directory read. And
while `os.File.ReadDir` runs with O(n) heap allocation, where n is the number of
directory entries read, `faf.ReadDir` just needs a single small heap allocation.
This is perfect for these use cases where you just process the directory entries
and then forget them.
