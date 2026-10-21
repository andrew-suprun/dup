package main

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	// Get virtual memory statistics
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Failed to get memory stats: %v", err)
	}

	// Convert bytes to megabytes (MB)
	const mb = 1024 * 1024

	fmt.Printf("Total Memory:     %d MB\n", v.Total/mb)
	fmt.Printf("Available Memory: %d MB\n", v.Available/mb)
	fmt.Printf("Used Memory:      %d MB\n", v.Used/mb)
	fmt.Printf("Memory Usage:     %.2f%%\n", v.UsedPercent)
}
