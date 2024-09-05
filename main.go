package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var IPsPerSubnet = 2 // Number of proxies to select from each subnet if IPsTotal is 0
var IPsTotal = 100   // Total number of proxies to select across all subnets, set to 0 to use IPsPerSubnet

// getSubnet extracts the /24 subnet from an IP address
func getSubnet(ip net.IP) string {
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.0/24", ip[0], ip[1], ip[2])
}

// readProxies reads proxies from a file and categorizes them by /24 subnet
func readProxies(filename string) (map[string][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	proxies := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		if proxy == "" {
			continue
		}

		ip := net.ParseIP(proxy)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", proxy)
		}

		subnet := getSubnet(ip)
		proxies[subnet] = append(proxies[subnet], proxy)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}

// selectRandomProxies selects proxies based on IPsPerSubnet or IPsTotal
func selectRandomProxies(subnetProxies map[string][]string, ipsPerSubnet int, ipsTotal int) map[string][]string {
	rand.Seed(time.Now().UnixNano())
	selected := make(map[string][]string)

	if ipsTotal > 0 {
		// Flatten the list of all proxies
		var allProxies []string
		for _, proxies := range subnetProxies {
			allProxies = append(allProxies, proxies...)
		}

		// Shuffle and select the total number of proxies
		if len(allProxies) > ipsTotal {
			rand.Shuffle(len(allProxies), func(i, j int) { allProxies[i], allProxies[j] = allProxies[j], allProxies[i] })
			selectedProxies := allProxies[:ipsTotal]

			// Group selected proxies by subnet
			subnetMap := make(map[string][]string)
			for _, proxy := range selectedProxies {
				ip := net.ParseIP(proxy)
				if ip != nil {
					subnet := getSubnet(ip)
					subnetMap[subnet] = append(subnetMap[subnet], proxy)
				}
			}

			return subnetMap
		}
	} else {
		// Use IPsPerSubnet if IPsTotal is 0
		for subnet, proxies := range subnetProxies {
			if len(proxies) > ipsPerSubnet {
				rand.Shuffle(len(proxies), func(i, j int) { proxies[i], proxies[j] = proxies[j], proxies[i] })
				selected[subnet] = proxies[:ipsPerSubnet]
			} else {
				selected[subnet] = proxies
			}
		}
	}

	return selected
}

// writeResults writes the selected proxies to a file in the specified format
func writeResults(filename string, selectedProxies map[string][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for subnet, proxies := range selectedProxies {
		subnetPrefix := subnet[:strings.LastIndex(subnet, ".")+1]
		fmt.Fprintf(writer, "%s subnet, %d IPs:\n", subnetPrefix, len(proxies))
		for _, proxy := range proxies {
			fmt.Fprintln(writer, proxy)
		}
		fmt.Fprintln(writer)
	}

	return writer.Flush()
}

// clearResultsFile clears the results file at the start
func clearResultsFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	file.Close() // Closing the file right after creating it clears the content
	return nil
}

// main function
func main() {
	// Define file names and number of proxies per subnet and total
	inputFile := "proxies.txt"
	outputFile := "results.txt"

	// Clear the results file at startup
	if err := clearResultsFile(outputFile); err != nil {
		fmt.Printf("Error clearing results file: %v\n", err)
		return
	}

	proxies, err := readProxies(inputFile)
	if err != nil {
		fmt.Printf("Error reading proxies: %v\n", err)
		return
	}

	selectedProxies := selectRandomProxies(proxies, IPsPerSubnet, IPsTotal)

	if err := writeResults(outputFile, selectedProxies); err != nil {
		fmt.Printf("Error writing results: %v\n", err)
		return
	}

	fmt.Printf("Results written to %s\n", outputFile)
}
