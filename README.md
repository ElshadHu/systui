# systui

I am currently working on systui, a terminal program that shows useful
metrics of the system in a user-friendly manner.

Right now it shows, refreshing every second:

- a top bar with memory pressure, swap, compressed memory on macOS
  and a one line insight that says what is wrong, or `All normal`
- CPU and memory usage as bars
- a CPU breakdown with user, sys, idle, nice, iowait, irq, softirq, steal
  and guest percentages
- memory stats with total, used, free, active, buffers and cached

The panels sit side by side and move onto a new row when the terminal is
too narrow for all of them.

I have implemented CPU and memory so far. I will keep making it evolve. For now, let's see how it goes.


## Build and run

Requires Go 1.26 or newer.

```
go build -o systui .
./systui
```

Press `q` or `ctrl+c` to quit.
