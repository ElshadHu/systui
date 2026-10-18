# systui

systui is a terminal program that shows what your computer is doing right
now. It groups processes into apps, so you see one row for Chrome instead
of forty rows of Chrome helpers. It refreshes every second.

It runs on macOS, Linux and Windows. Some numbers are only available on
macOS, and those are marked below.

## Build and run

Requires Go 1.26 or newer.

```
go build -o systui .
./systui
```

| Flag | What it does |
|------|--------------|
| `-theme dark` | Dark colors. This is the default. |
| `-theme light` | Light colors. |
| `-theme ansi` | Uses the 16 colors of your terminal. |

## The screen

Every view has the same frame:

1. Status line: memory pressure, swap, and on macOS compressed memory.
2. Insight line: one or two short sentences about what is wrong, or
   nothing when all is fine.
3. Tab bar with the views.
4. The view.
5. Footer with the keys that work right now.

### Status line

| Field | Meaning |
|-------|---------|
| `Pressure` | GREEN, YELLOW or RED. On macOS this is the kernel's pressure level. On Linux it is read from Pressure Stall Information. Windows uses the amount of free memory. |
| `Swap` | How much swap is in use. |
| `Compressed` | macOS only. How much memory the kernel has compressed. |

## Views

| Key | View | Status |
|-----|------|--------|
| `1` | Apps | Built |
| `2` | Overview | Built |
| `3` | Lifetime | Built |

### Apps view

One row per app, tool or group of processes, sorted by memory or CPU.

How rows are grouped:

- On macOS, every process inside a `.app` bundle belongs to that app.
  Helpers get a short role name, such as `renderer` or `gpu`.
- On Linux and Windows, a process and its children belong to the top-most
  parent below your shell.
- Processes from system folders, and processes owned by root, go into one
  `system` group.
- Everything under 50 MB and 1% CPU goes into one `small processes` row
  at the bottom.

Columns:

| Column | Meaning |
|--------|---------|
| APP | Group name |
| MEM | Memory of the whole group. On macOS this is the same footprint number Activity Monitor shows. Elsewhere it is resident memory. |
| CPU | CPU of the whole group, as a percent of one core |
| PROCS | How many processes are in the group |
| TREND | The last 20 seconds, one bar per second, newest on the right. Shows CPU when sorted by CPU and memory when sorted by memory. A memory trend shows the growth, for example `+140MB`, when the group grew by 100 MB or more in those 20 seconds. |

The PROCS and TREND columns disappear when the terminal is too narrow.

Marks next to a name:

| Mark | Meaning |
|------|---------|
| `◇` | System group |
| `‖` | A process in the group is stopped |
| `†` | A process in the group is a zombie |
| `◌` | The process belongs to another user |

Keys:

| Key | What it does |
|-----|--------------|
| `↑` `↓` | Move the cursor one row |
| `PgUp` `PgDn` or `ctrl+u` `ctrl+d` | Move one page |
| `g` `G` | Jump to the first or last row |
| `→` | Expand the group under the cursor |
| `←` | Collapse it |
| `space` | Expand or collapse |
| `E` | Expand all groups |
| `C` | Collapse all groups |
| `enter` | On a `… N more` row, show the hidden processes |
| `m` | Sort by memory |
| `c` | Sort by CPU |
| `p` | Sort by peak memory |
| `t` | Sort by CPU time |
| `/` | Type a filter. Rows whose app or process name contains the text stay. |
| `esc` | Clear the filter |
| `q` or `ctrl+c` | Quit |

An expanded group shows its five biggest processes and a `… N more` row.

After any key or click, the sort is paused for 2 seconds so rows do not
move under your cursor. The tab bar shows `sort paused` while this is on.

### Lifetime view

The same rows, cursor, filter and sort as the Apps view. The columns show
what each app has used since it started.

| Column | Meaning |
|--------|---------|
| MEM | Memory right now |
| PEAK | Highest memory seen. For an app, since systui started. For a process, since the process started, as reported by the operating system. |
| CPU | CPU right now |
| CPU TIME | CPU seconds since the process started, including children that have exited |
| SINCE | When the app's main process started: a time today, `yest 22:40`, or a date |

The small processes row shows a dash in these columns, because its members
change every second.

When an app is below half of a peak of 1 GB or more, the insight line says
so, for example `Google Chrome peaked at 4.6 GB at 09:31, now 1.6 GB`.

### Overview view

Three panels side by side. They move onto a new row when the terminal is
too narrow.

| Panel | Contents |
|-------|----------|
| % Usage | CPU and memory as bars with a percent |
| CPU | user, sys, idle, nice, iowait, irq, softirq, steal and guest percentages |
| MEM | total, used, free, active, buffers and cached |

Keys here are `1` to `5` to switch view and `q` to quit.

## Colors and symbols

Every color has a symbol next to it, so the screen still makes sense in
black and white.
