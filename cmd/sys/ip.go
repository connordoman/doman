package sys

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

type IPCommandJSON struct {
	LocalIP  string `json:"local_ip"`
	PublicIP string `json:"public_ip"`
}

var IPCommand = &cobra.Command{
	Use:   "sys.ip",
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

	localIPString := formatIP("local", localIP)
	publicIPString := formatIP("public", publicIP)

	fmt.Printf("%s\n%s\n", localIPString, publicIPString)
}

func getPublicIP() (string, error) {
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
	return fmt.Sprintf("%s\t%s", txt.Greyf(context), txt.Boldf(ip))
}
