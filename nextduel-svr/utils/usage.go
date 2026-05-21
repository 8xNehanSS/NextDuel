package utils

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/cpu"
)

func GetMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mb := float64(m.Alloc) / 1024 / 1024
	return fmt.Sprintf("%.2f MB", mb)
}

func GetCPUUsage() string {
	percent, _ := cpu.Percent(0, false)
	return fmt.Sprintf("%.2f%%", percent[0])
}
