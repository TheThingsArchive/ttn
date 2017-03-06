// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"os"
	"runtime"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var startTime = time.Now()

// GetSystem gets statistics about the system
func GetSystem() *api.SystemStats {
	status := new(api.SystemStats)
	if load, err := load.Avg(); err == nil {
		status.Load = &api.SystemStats_Loadstats{
			Load1:  float32(load.Load1),
			Load5:  float32(load.Load5),
			Load15: float32(load.Load15),
		}
	}
	if cpu, err := cpu.Times(false); err == nil && len(cpu) == 1 {
		status.Cpu = &api.SystemStats_CPUStats{
			User:   float32(cpu[0].User),
			System: float32(cpu[0].System),
			Idle:   float32(cpu[0].Idle),
		}
	}
	if mem, err := mem.VirtualMemory(); err == nil {
		status.Memory = &api.SystemStats_MemoryStats{
			Total:     mem.Total,
			Available: mem.Available,
			Used:      mem.Used,
		}
	}
	return status
}

// GetComponent gets statistics about this component
func GetComponent() *api.ComponentStats {
	status := new(api.ComponentStats)
	status.Uptime = uint64(time.Now().Sub(startTime).Seconds())
	process, err := process.NewProcess(int32(os.Getpid()))
	if err == nil {
		if memory, err := process.MemoryInfo(); err == nil {
			status.Memory = &api.ComponentStats_MemoryStats{
				Memory: memory.RSS,
				Swap:   memory.Swap,
			}
		}
		if cpu, err := process.Times(); err == nil {
			status.Cpu = &api.ComponentStats_CPUStats{
				User:   float32(cpu.User),
				System: float32(cpu.System),
				Idle:   float32(cpu.Idle),
			}
		}
	}
	status.Goroutines = uint64(runtime.NumGoroutine())
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	status.GcCpuFraction = float32(memstats.GCCPUFraction)
	if status.Memory == nil {
		status.Memory = new(api.ComponentStats_MemoryStats)
	}
	status.Memory.Heap = memstats.HeapInuse
	status.Memory.Stack = memstats.StackInuse
	return status
}
