package main

import (
	"testing"
)

func TestCollectSystemMetrics(t *testing.T) {
	// Call the function
	collectSystemMetrics()

	// Validate metrics
	if cpuUsage == nil {
		t.Error("cpuUsage metric is not initialized")
	}
	if memoryUsage == nil {
		t.Error("memoryUsage metric is not initialized")
	}
	if diskUsage == nil {
		t.Error("diskUsage metric is not initialized")
	}
}
