package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	probing "github.com/prometheus-community/pro-bing"
)

func ping(fqdn string) int64 {
	// Remove any URL scheme if present (shouldn't be, but just in case)
	// Only use the hostname for ICMP ping
	pinger, err := probing.NewPinger(fqdn)
	if err != nil {
		return 0
	}
	pinger.Count = 3
	pinger.Timeout = 3 * time.Second
	pinger.SetPrivileged(false) // Use unprivileged mode for macOS compatibility
	err = pinger.Run()
	if err != nil {
		return 0
	}
	stats := pinger.Statistics()
	return int64(stats.AvgRtt / time.Millisecond)
}

func calculateSpeed(i *spec.AppInfo) float64 {
	if i == nil {
		return 0
	}
	elapsedSeconds := float64(i.ElapsedTime) / 1e6
	if elapsedSeconds > 0 {
		return float64(i.NumBytes*8) / elapsedSeconds / 1e6
	}

	return elapsedSeconds
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("\nRetrieving speedtest.net configuration...")

	// create a ndt7 client
	client := ndt7.NewClient("speed", "1.0.0")

	// fetch nearest server details
	targets, err := client.Locate.Nearest(ctx, "ndt/ndt7")
	if err != nil {
		log.Fatalf("Failed to locate nearest server: %v", err)
	}

	if len(targets) == 0 {
		log.Fatal("No servers found")
	}

	target := targets[0]
	fmt.Printf("\nServer found: %s at %s, %s\n\n", target.Machine, target.Location.City, target.Location.Country)

	// --- Ping Test ---
	fmt.Printf("%s %-18s %8d ms\n", "↔", "Ping (avg)    :", ping(target.Machine))

	// --- Download Test ---
	downloadChan, err := client.StartDownload(ctx)
	if err != nil {
		log.Fatalf("Download error: %v", err)
	}
	var downloadSpeed float64
	fmt.Printf("%s %-18s", "↓", "Download speed:")
	for m := range downloadChan {
		speed := calculateSpeed(m.AppInfo)
		if speed > downloadSpeed {
			downloadSpeed = speed
		}
		fmt.Printf("%8.2f Mbps\r%s %-18s", speed, "↓", "Download speed:")
	}
	fmt.Printf("%8.2f Mbps\n", downloadSpeed)

	// --- Upload Test ---
	uploadChan, err := client.StartUpload(ctx)
	if err != nil {
		log.Fatalf("Upload error: %v", err)
	}
	var uploadSpeed float64
	fmt.Printf("%s %-18s", "↑", "Upload speed  :")
	for m := range uploadChan {
		speed := calculateSpeed(m.AppInfo)
		if speed > uploadSpeed {
			uploadSpeed = speed
		}
		fmt.Printf("%8.2f Mbps\r%s %-18s", speed, "↑", "Upload speed  :")
	}
	fmt.Printf("%8.2f Mbps\n", uploadSpeed)

	fmt.Println("\n🚀 Test complete!")
}
