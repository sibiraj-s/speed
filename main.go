package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/m-lab/ndt7-client-go"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/spf13/viper"
	"github.com/yarlson/pin"
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

// calculateMbps computes Mbps from total bytes and elapsed microseconds
func calculateMbps(bytes int64, elapsedMicroseconds int64) float64 {
	if elapsedMicroseconds == 0 {
		return 0
	}
	seconds := float64(elapsedMicroseconds) / 1e6
	mbps := (float64(bytes) * 8) / seconds / 1e6
	return mbps
}

// promptUserAgreement prompts the user to agree to the M-Lab privacy policy before running the test
func promptUserAgreement() bool {
	configDir := os.Getenv("HOME") + "/.config"
	configPath := configDir + "/.speed.yml"

	viper.SetConfigName(".speed")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	err := viper.ReadInConfig()
	if err == nil && viper.GetBool("agreed_to_policy") {
		return true
	}

	fmt.Println("Before running the speed test, please review the Measurement Lab privacy policy:")
	fmt.Println("https://www.measurementlab.net/privacy")
	fmt.Println("\nBy continuing, you agree to the data policy, which includes retention and publication of IP addresses.")
	fmt.Print("Type 'yes' to agree and continue: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "yes" {
		fmt.Println("You must agree to the data policy to proceed. Exiting.")
		return false
	}

	viper.Set("agreed_to_policy", true)
	err = viper.WriteConfigAs(configPath)
	if err != nil {
		fmt.Println("Warning: could not save agreement to config file.")
	}

	return true
}

func main() {
	if !promptUserAgreement() {
		return
	}

	fmt.Println("\n✔ Agreed to privacy policy")
	fmt.Println("")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use pin spinner for status
	spinner := pin.New("Retrieving speedtest configuration...",
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorCyan),
	)
	cancelSpinner := spinner.Start(context.Background())
	defer cancelSpinner()

	client := ndt7.NewClient("speed", "1.0.0")
	targets, err := client.Locate.Nearest(ctx, "ndt/ndt7")
	spinner.Stop()
	if err != nil {
		log.Fatalf("Failed to locate nearest server: %v", err)
	}

	if len(targets) == 0 {
		log.Fatal("No servers found")
	}

	target := targets[0]
	fmt.Printf("Server found: %s at %s, %s\n\n", target.Machine, target.Location.City, target.Location.Country)

	// --- Ping Test ---
	fmt.Printf("%s %-18s %8d ms\n", "↔", "Ping (avg)    :", ping(target.Machine))

	// --- Download Test ---
	downloadChan, err := client.StartDownload(ctx)
	if err != nil {
		log.Fatalf("Download error: %v", err)
	}
	var totalDownloadBytes int64
	var totalDownloadElapsed int64
	fmt.Printf("%s %-18s", "↓", "Download speed:")
	for m := range downloadChan {
		if m.AppInfo != nil {
			totalDownloadBytes = m.AppInfo.NumBytes
			totalDownloadElapsed = m.AppInfo.ElapsedTime
			speed := calculateMbps(totalDownloadBytes, totalDownloadElapsed)
			fmt.Printf("%8.2f Mbps\r%s %-18s", speed, "↓", "Download speed:")
		}
	}
	finalDownloadSpeed := calculateMbps(totalDownloadBytes, totalDownloadElapsed)
	fmt.Printf("%8.2f Mbps\n", finalDownloadSpeed)

	// --- Upload Test ---
	uploadChan, err := client.StartUpload(ctx)
	if err != nil {
		log.Fatalf("Upload error: %v", err)
	}
	var totalUploadBytes int64
	var totalUploadElapsed int64
	fmt.Printf("%s %-18s", "↑", "Upload speed  :")
	for m := range uploadChan {
		if m.AppInfo != nil {
			totalUploadBytes = m.AppInfo.NumBytes
			totalUploadElapsed = m.AppInfo.ElapsedTime
			speed := calculateMbps(totalUploadBytes, totalUploadElapsed)
			fmt.Printf("%8.2f Mbps\r%s %-18s", speed, "↑", "Upload speed  :")
		}
	}
	finalUploadSpeed := calculateMbps(totalUploadBytes, totalUploadElapsed)
	fmt.Printf("%8.2f Mbps\n", finalUploadSpeed)

	fmt.Println("\n🚀 Test complete!")
}
