package cmd

import (
	"log"
	"net"

	"github.com/spf13/cobra"
)

var cmdIp = &cobra.Command{
	Use:        "ip ip...",
	ArgAliases: []string{"IP"},
	Args:       cobra.MinimumNArgs(1),
	Short:      "interact with ip reservations",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	cmdIp.AddCommand(cmdIpShow)
	rootCmd.AddCommand(cmdIp)
}

var cmdIpShow = &cobra.Command{
	Use:   "show",
	Short: "show ip reservation",
	Run:   ShowIP,
}

func ShowIP(_ *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("Unable to connect to api: %v", err)
	}

	for _, ipString := range args {
		ip := net.ParseIP(ipString)
		if ip == nil {
			log.Fatalf("invalid IP: %s", ipString)
		}
		reservation, err := apiClient.IPAM().GetIPReservation(ip)
		if err != nil {
			log.Fatalf("Unable to get reservation for %s: %v", ip, err)
		}
		log.Print(reservation)
	}
}
