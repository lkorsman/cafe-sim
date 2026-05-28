# cafe-sim

A coffee shop event-driven simulation CLI, built to demonstrate the [`github.com/lkorsman/pqueue`](https://pkg.go.dev/github.com/lkorsman/pqueue) generic priority queue library.

Customers arrive at random times throughout a shift. A min-heap priority queue orders every event (arrivals, orders, completions) by time, so the simulation always processes the next thing that happens — exactly how real event-driven systems work.

## Install

```bash
go install github.com/lkorsman/cafe-sim@latest
```

## Usage

```bash
cafe-sim [flags]

Flags:
  --customers  int     number of customers to simulate (default 20)
  --baristas   int     number of baristas working (default 2)
  --duration   float   how long the shop is open in minutes (default 60)
  --seed       int     random seed for reproducibility (default 42)
```

## Example

```bash
cafe-sim --customers 10 --baristas 2 --duration 60 --seed 42
```

```
☕  Cafe Simulation Starting
   10 customers, 2 baristas, 60 minute shift

[08:02] Customer 3 arrives
[08:02] Barista 1 starts making Customer 3's order
[08:05] Customer 7 arrives
[08:05] Barista 2 starts making Customer 7's order
[08:06] Customer 3's order is ready (waited 4.1m)
...

☕  Simulation complete
```

## How it works

Every event is a struct with a timestamp:

```go
type event struct {
    time     float64
    kind     eventType  // arrive, order, finish
    customer string
    arriveAt float64
}
```

Events are pushed into a min-heap priority queue ordered by `time`. The main loop pops the earliest event, handles it, and schedules any follow-up events:

```go
queue := pqueue.New[event](func(a, b event) bool {
    return a.time < b.time
})
```

This is the standard pattern for discrete event simulation — the priority queue is what makes it efficient.

## License

MIT