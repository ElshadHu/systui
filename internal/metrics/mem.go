package metrics

import "github.com/shirou/gopsutil/v4/mem"

type MemStats struct {
	Total         uint64
	Used          uint64
	Free          uint64
	Active        uint64
	Buffers       uint64
	Cached        uint64
	UsedPercent   float64
	SwapTotal     uint64
	SwapUsed      uint64
	Compressed    uint64
	HasCompressed bool
}

// ReadMem returns the current memory, swap and compressor usage
func ReadMem() (MemStats, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return MemStats{}, err
	}
	swap, err := mem.SwapMemory()
	if err != nil {
		return MemStats{}, err
	}
	compressed, hasCompressed := readCompressed()
	return MemStats{
		Total:         vm.Total,
		Used:          vm.Used,
		Free:          vm.Free,
		Active:        vm.Active,
		Buffers:       vm.Buffers,
		Cached:        vm.Cached,
		UsedPercent:   vm.UsedPercent,
		SwapTotal:     swap.Total,
		SwapUsed:      swap.Used,
		Compressed:    compressed,
		HasCompressed: hasCompressed,
	}, nil
}
