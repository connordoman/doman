package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
	"github.com/connordoman/doman/internal/web"
	"github.com/spf13/cobra"
)

type IPCommandJSON struct {
	LocalIP  string `json:"local_ip"`
	PublicIP string `json:"public_ip"`
}

var IPCommand = &cobra.Command{
	Use:   "ip",
	Short: "View your IP address",
	Run:   executeIPCommand,
}

func init() {
	IPCommand.Flags().BoolP("json", "j", false, "Output in JSON format")
}

func executeIPCommand(cmd *cobra.Command, args []string) {
	asJSON, _ := cmd.Flags().GetBool("json")

	localIP, err := getLocalIP()
	if err != nil {
		pkg.FailAndExit("Failed to retrieve local IP address: %v", err)
	}

	publicIP, err := getPublicIP()
	if err != nil {
		pkg.FailAndExit("Failed to retrieve public IP address: %v", err)
	}

	if asJSON {
		ipInfo := IPCommandJSON{
			LocalIP:  localIP,
			PublicIP: publicIP,
		}
		jsonOutput, err := json.MarshalIndent(ipInfo, "", "  ")
		if err != nil {
			pkg.FailAndExit("Failed to marshal IP information to JSON: %v", err)
		}
		fmt.Println(string(jsonOutput))
		return
	}

	localIPString := formatIP("\ueb06 local", localIP)
	publicIPString := formatIP("\ueb01 public", publicIP)

	fmt.Printf("%s\n%s\n", localIPString, publicIPString)
}

func getPublicIP() (string, error) {
	body, err := web.Fetch("https://icanhazip.com")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP address found")
}

func formatIP(context, ip string) string {
	return fmt.Sprintf("%s%s", txt.Greyf("%-12s", context), txt.Boldf("%s", ip))
}
