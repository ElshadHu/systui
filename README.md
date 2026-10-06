# systui

I am currently working on systui, a terminal program that shows useful
metrics of the system in a user-friendly manner.

Right now it shows CPU usage as a bar that refreshes every second:

```
systui - press q to quit

CPU [|||||||                                 ]  17.3%
```

I have implemented CPU usage so far. I will keep making it evolve, and I
plan to add memory and disk usage next. For now, let's see how it goes.


## Build and run

Requires Go 1.26 or newer.

```
go build -o systui .
./systui
```

Press `q` or `ctrl+c` to quit.
