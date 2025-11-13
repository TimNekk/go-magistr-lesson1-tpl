package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/levigross/grequests/v2"
)

var ErrFetchFailed = errors.New("fetch failed")

type Stats struct {
	LoadAverage                 int64
	MemoryTotalBytes            int64
	MemoryUsedBytes             int64
	DiskTotalBytes              int64
	DiskUsedBytes               int64
	NetworkBandwidthBytesPerSec int64
	NetworkUsageBytesPerSec     int64
}

func main() {
	ctx := context.Background()

	monitor(ctx)
}

func monitor(ctx context.Context) {
	for {
		var (
			stats *Stats
			err   error
		)

		for range 3 {
			stats, err = getStats(ctx)
			if err != nil {
				continue
			}
			break
		}

		if stats == nil {
			fmt.Fprintln(os.Stdout, "Unable to fetch server statistic")
			continue
		}

		if stats.LoadAverage > 30 {
			fmt.Fprintf(os.Stdout, "Load Average is too high: %d\n", stats.LoadAverage)
		}

		if stats.MemoryTotalBytes > 0 {
			memoryUsagePercent := (stats.MemoryUsedBytes * 100) / stats.MemoryTotalBytes
			if memoryUsagePercent > 80 {
				fmt.Fprintf(os.Stdout, "Memory usage too high: %d%%\n", memoryUsagePercent)
			}
		}

		if stats.DiskTotalBytes > 0 {
			diskUsagePercent := (stats.DiskUsedBytes * 100) / stats.DiskTotalBytes
			diskLeftMegaBytes := (stats.DiskTotalBytes - stats.DiskUsedBytes) / 1024 / 1024
			if diskUsagePercent > 90 {
				fmt.Fprintf(os.Stdout, "Free disk space is too low: %d Mb left\n", diskLeftMegaBytes)
			}
		}

		if stats.NetworkBandwidthBytesPerSec > 0 {
			NetworkUsagePercent := (stats.NetworkUsageBytesPerSec * 100) / stats.NetworkBandwidthBytesPerSec
			NetworkAvailableMegaBitsPerSec := (stats.NetworkBandwidthBytesPerSec - stats.NetworkUsageBytesPerSec) / 1000000
			if NetworkUsagePercent > 90 {
				fmt.Fprintf(os.Stdout, "Network bandwidth usage high: %d Mbit/s available\n", NetworkAvailableMegaBitsPerSec)
			}
		}
	}
}

func getStats(ctx context.Context) (*Stats, error) {
	resp, err := grequests.Get(ctx, "http://srv.msk01.gigacorp.local/_stats")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, ErrFetchFailed
	}

	metrics := strings.Split(resp.String(), ",")

	if len(metrics) != 7 {
		return nil, ErrFetchFailed
	}

	metricsInt := make([]int64, len(metrics))

	for i, metric := range metrics {
		metricsInt[i], err = strconv.ParseInt(metric, 10, 64)
		if err != nil {
			return nil, ErrFetchFailed
		}
	}

	return &Stats{
		LoadAverage:                 metricsInt[0],
		MemoryTotalBytes:            metricsInt[1],
		MemoryUsedBytes:             metricsInt[2],
		DiskTotalBytes:              metricsInt[3],
		DiskUsedBytes:               metricsInt[4],
		NetworkBandwidthBytesPerSec: metricsInt[5],
		NetworkUsageBytesPerSec:     metricsInt[6],
	}, nil
}
