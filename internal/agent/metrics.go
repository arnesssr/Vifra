package agent

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// SystemMetrics represents system metrics collected from the OS
type SystemMetrics struct {
	CPUUsage    float64
	MemoryUsed  uint64
	MemoryTotal uint64
	DiskUsed    uint64
	DiskTotal   uint64
	LoadAvg     float64
}

// CollectSystemMetrics collects system metrics from the OS
func CollectSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// Collect CPU usage
	cpuUsage, err := collectCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU usage: %v", err)
	}
	metrics.CPUUsage = cpuUsage

	// Collect memory usage
	memUsed, memTotal, err := collectMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to collect memory usage: %v", err)
	}
	metrics.MemoryUsed = memUsed
	metrics.MemoryTotal = memTotal

	// Collect disk usage
	diskUsed, diskTotal, err := collectDiskUsage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to collect disk usage: %v", err)
	}
	metrics.DiskUsed = diskUsed
	metrics.DiskTotal = diskTotal

	// Collect load average
	loadAvg, err := collectLoadAvg()
	if err != nil {
		return nil, fmt.Errorf("failed to collect load average: %v", err)
	}
	metrics.LoadAvg = loadAvg

	return metrics, nil
}

// collectCPUUsage collects CPU usage percentage
func collectCPUUsage() (float64, error) {
	// This is a simplified implementation
	// A real implementation would read /proc/stat and calculate usage over time
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[0] == "cpu" {
			// Parse CPU stats (user, nice, system, idle, iowait, ...)
			// This is a simplified calculation
			return 0.0, nil // Placeholder
		}
	}

	return 0, fmt.Errorf("failed to parse CPU stats")
}

// collectMemoryUsage collects memory usage
func collectMemoryUsage() (uint64, uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = file.Close() }()

	var memTotal, memFree, memAvailable, memBuffers, memCached uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			switch fields[0] {
			case "MemTotal:":
				_, _ = fmt.Sscanf(fields[1], "%d", &memTotal)
			case "MemFree:":
				_, _ = fmt.Sscanf(fields[1], "%d", &memFree)
			case "MemAvailable:":
				_, _ = fmt.Sscanf(fields[1], "%d", &memAvailable)
			case "Buffers:":
				_, _ = fmt.Sscanf(fields[1], "%d", &memBuffers)
			case "Cached:":
				if fields[1] != "0" { // Skip "Cached:" line that appears twice
					_, _ = fmt.Sscanf(fields[1], "%d", &memCached)
				}
			}
		}
	}

	// Convert from KB to bytes
	memTotal *= 1024
	memFree *= 1024
	memAvailable *= 1024
	memBuffers *= 1024
	memCached *= 1024

	// Calculate used memory (total - free - buffers - cached)
	memUsed := memTotal - memFree - memBuffers - memCached

	return memUsed, memTotal, nil
}

// collectDiskUsage collects disk usage for a given path
func collectDiskUsage(path string) (uint64, uint64, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// Calculate disk usage in bytes
	blockSize := uint64(stat.Bsize)
	total := blockSize * stat.Blocks
	free := blockSize * stat.Bfree
	used := total - free

	return used, total, nil
}

// collectLoadAvg collects system load average
func collectLoadAvg() (float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			load, err := strconv.ParseFloat(fields[0], 64)
			if err != nil {
				return 0, err
			}
			return load, nil
		}
	}

	return 0, fmt.Errorf("failed to parse load average")
}