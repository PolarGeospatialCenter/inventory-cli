package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/PolarGeospatialCenter/inventory-cli/pkg/ingestlib"
	"github.com/PolarGeospatialCenter/inventory/pkg/inventory/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var cmdNode = &cobra.Command{
	Use:        "node nodeId...",
	ArgAliases: []string{"nodeId"},
	Args:       cobra.MinimumNArgs(1),
	Short:      "interact with node objects",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var systemName, roleName string

func init() {
	cmdNodeList.Flags().StringVarP(&systemName, "system", "s", "", "list only nodes from system")
	cmdNodeList.Flags().StringVarP(&roleName, "role", "", "", "list only nodes from role")
	cmdNode.AddCommand(cmdNodeList)
	cmdNode.AddCommand(cmdNodeInteractiveCreate)
	cmdNode.AddCommand(cmdNodeInteractiveUpdate)
	cmdNode.AddCommand(cmdNodeResetNetworks)
	cmdNode.AddCommand(cmdNodeDetectNetworks)
	cmdNode.AddCommand(cmdNodeShow)
	cmdNode.AddCommand(cmdSetSerialConsole)
	cmdNode.AddCommand(cmdNodeSetIP)
	rootCmd.AddCommand(cmdNode)
}

var cmdNodeList = &cobra.Command{
	Use:   "list",
	Short: "list all nodes",
	Run:   ListNodes,
}

func ListNodes(cmd *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("Unable to connect to api: %v", err)
	}

	nodes, err := apiClient.Node().GetAll()
	if err != nil {
		log.Fatalf("unable to get nodes: %v", err)
	}

	for _, node := range nodes {
		if systemName != "" && node.System != systemName {
			continue
		}

		if roleName != "" && node.Role != roleName {
			continue
		}
		fmt.Printf("%25s - %s\n", node.ID(), node.Hostname())
	}
}

var cmdNodeInteractiveCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a node interactively",
	Run:   NodeInteractiveCreate,
}

func NodeInteractiveCreate(_ *cobra.Command, _ []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("unable to connect to api: %v", err)
	}

	systems, err := apiClient.System().GetAll()
	if err != nil {
		log.Fatalf("unable to get systems: %v", err)
	}

	networks, err := apiClient.Network().GetAll()
	if err != nil {
		log.Fatalf("unable to get systems: %v", err)
	}

	node := &types.Node{}
	p := &ingestlib.NodePopulator{Node: node, Systems: systems, Networks: networks}
	err = p.PopulateNode()
	if err != nil {
		log.Fatalf("Unable to populate node data: %v", err)
	}

	txt, err := json.MarshalIndent(p.Node, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal node: %v", err)
	}

	fmt.Printf("---------\n")
	fmt.Printf("%s\n", string(txt))
	prompt := promptui.Prompt{Label: "Create this node?", IsConfirm: true}
	_, err = prompt.Run()
	if err != nil {
		log.Fatalf("Exiting without creating node.")
	}

	err = apiClient.Node().Create(p.Node)
	if err != nil {
		log.Fatalf("unable to create node: %v", err)
	}
}

var cmdNodeInteractiveUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update a node interactively",
	Run:   NodeInteractiveUpdate,
}

func NodeInteractiveUpdate(_ *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("unable to connect to api: %v", err)
	}

	systems, err := apiClient.System().GetAll()
	if err != nil {
		log.Fatalf("unable to get systems: %v", err)
	}

	networks, err := apiClient.Network().GetAll()
	if err != nil {
		log.Fatalf("unable to get systems: %v", err)
	}

	for _, nodeId := range args {
		node, err := apiClient.Node().Get(nodeId)
		if err != nil {
			log.Fatalf("Unable to get node %s: %v", nodeId, err)
		}
		p := &ingestlib.NodePopulator{Node: node, Systems: systems, Networks: networks}
		err = p.PopulateNode()
		if err != nil {
			log.Fatalf("Unable to populate node data: %v", err)
		}

		txt, err := json.MarshalIndent(p.Node, "", "  ")
		if err != nil {
			log.Fatalf("Unable to marshal node: %v", err)
		}

		fmt.Printf("---------\n")
		fmt.Printf("%s\n", string(txt))
		prompt := promptui.Prompt{Label: "Update this node?", IsConfirm: true}
		_, err = prompt.Run()
		if err != nil {
			log.Printf("Continuing without updating node.")
			continue
		}

		err = apiClient.Node().Update(p.Node)
		if err != nil {
			log.Fatalf("unable to create node: %v", err)
		}
	}
}

var cmdNodeShow = &cobra.Command{
	Use:   "show",
	Short: "show node",
	Run:   ShowNode,
}

func ShowNode(_ *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("unable to connect to api: %v", err)
	}

	for _, nodeId := range args {
		node, err := apiClient.NodeConfig().Get(nodeId)
		if err != nil {
			log.Fatalf("Unable to get node config for %s: %v", nodeId, err)
		}

		nodeJson, err := json.Marshal(node)
		if err != nil {
			log.Fatalf("Unable to marshal json for node %s: %v", nodeId, err)
		}
		fmt.Print(string(nodeJson))
	}
}

var cmdNodeResetNetworks = &cobra.Command{
	Use:   "reset-networks",
	Short: "Reset networks configured for the node(s)",
	Run:   NodeResetNetworks,
}

func NodeResetNetworks(_ *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("unable to connect to api: %v", err)
	}

	for _, nodeId := range args {
		node, err := apiClient.Node().Get(nodeId)
		if err != nil {
			log.Fatalf("Unable to lookup node '%s': %v", nodeId, err)
		}
		node.Networks = make(types.NICInfoMap, 0)
		err = apiClient.Node().Update(node)
		if err != nil {
			log.Fatalf("Unable to reset networks for node '%s': %v", nodeId, err)
		}
	}
}

var cmdNodeSetIP = &cobra.Command{
	Use:   "set-ip",
	Short: "Attempts to claim the specified IP for the node",
	Run:   NodeSetIP,
}

func NodeSetIP(_ *cobra.Command, args []string) {
	apiClient, err := apiConnect()
	if err != nil {
		log.Fatalf("unable to connect to api: %v", err)
	}

	if len(args) != 2 {
		log.Fatalf("arguments must be <node-id> <ip>")
	}

	ipAddr := net.ParseIP(args[1])
	if ipAddr == nil {
		log.Fatalf("arguments must be <node-id> <ip>")
	}

	existingReservation, err := apiClient.IPAM().GetIPReservation(ipAddr)
	if existingReservation != nil && err == nil {
		log.Fatalf("a reservation for this ip already exists: %v", existingReservation)
	}

	node, err := apiClient.Node().Get(args[0])
	if err != nil {
		log.Fatalf("unable to lookup node '%s': %v", args[0], err)
	}

	networks, err := apiClient.Network().GetAll()
	if err != nil {
		log.Fatalf("unable to lookup networks: %v", err)
	}

	var matchingSubnet *net.IPNet
	var matchingNetwork string

	for _, network := range networks {
		if nicInfo, ok := node.Networks[network.ID()]; !ok || len(nicInfo.NICs) == 0 {
			// no nic on this network, continue
			continue
		}

		for _, subnet := range network.Subnets {
			if subnet.Cidr.Contains(ipAddr) {
				matchingSubnet = subnet.Cidr
				matchingNetwork = network.ID()
				break
			}
		}

		if matchingSubnet != nil {
			break
		}
	}

	if matchingSubnet == nil {
		log.Fatalf("unable to find matching subnet for ip on any network this host is connected to")
	}

	// Found matching subnet with attached nic, now try to claim the IP
	macAddr := node.Networks[matchingNetwork].NICs[0]
	// Delete existing reservations for this mac
	existingReservations, err := apiClient.IPAM().GetIPReservationsByMAC(macAddr)
	if existingReservations != nil {
		for _, reservation := range existingReservations {
			if matchingSubnet.Contains(reservation.IP.IP) {
				err = apiClient.IPAM().DeleteIPReservation(reservation)
				if err != nil {
					log.Fatalf("unable to delete reservation for existing ip (%s): %v", reservation.IP.String(), err)
				}
			}
		}
	}
	_, err = apiClient.IPAM().CreateIPReservation(&types.IpamIpRequest{HwAddress: macAddr.String()}, ipAddr)
	if err != nil {
		log.Fatalf("error reserving requested IP: %v", err)
	}
}

var cmdSetSerialConsole = &cobra.Command{
	Use:   "set-console <serial_console> <node>...",
	Short: "Set serial console metadata for node",
	Run: func(_ *cobra.Command, args []string) {
		apiClient, err := apiConnect()
		if err != nil {
			log.Fatalf("unable to connect to api: %v", err)
		}

		if len(args) < 2 {
			log.Fatalf("Must specify a serial configuration and at least one node")
		}

		serialConsole := args[0]

		for _, nodeId := range args[1:] {
			node, err := apiClient.Node().Get(nodeId)
			if err != nil {
				log.Fatalf("error getting node %s: %v", nodeId, err)
			}
			node.Metadata["serial_console"] = serialConsole
			err = apiClient.Node().Update(node)
			if err != nil {
				log.Fatalf("error updating serial console for node %s: %v", nodeId, err)
			}
		}
	},
}

var cmdNodeDetectNetworks = &cobra.Command{
	Use:   "detect-networks nodeId",
	Short: "Detect networks connected to this node and update it",
	Run: func(cmd *cobra.Command, args []string) {
		var nodeId string
		if nodeIdFile := os.Getenv("NODEID_FILE"); nodeIdFile != "" {
			nodeIdRaw, err := ioutil.ReadFile(nodeIdFile)
			if err != nil {
				log.Fatalf("Unable to read nodeid from NODEID_FILE=%s: %v", nodeIdFile, err)
			}
			nodeId = strings.TrimSpace(string(nodeIdRaw))
		}

		if len(args) == 1 {
			nodeId = args[0]
		}

		if nodeId == "" {
			log.Fatalf("please supply a node id either on the command line or via NODEID_FILE")
		}

		apiClient, err := apiConnect()
		if err != nil {
			log.Fatalf("unable to connect to api server: %v", err)
		}

		networks, err := apiClient.Network().GetAll()
		if err != nil {
			log.Fatalf("unable to get networks: %v", err)
		}

		node, err := apiClient.Node().Get(nodeId)
		if err != nil {
			log.Fatalf("Unable to lookup node '%s': %v", nodeId, err)
		}

		p := &ingestlib.NodePopulator{Node: node, Networks: networks}
		detected, err := p.DetectNetworks()
		if err != nil {
			log.Fatalf("Unable to detect networks: %v", err)
		}

		if detected > 0 {
			log.Printf("Detected %d updated networks, writing updated node:", detected)
			txt, err := json.MarshalIndent(p.Node, "", "  ")
			if err != nil {
				log.Fatalf("Unable to marshal updated node: %v", err)
			}
			log.Printf("%s\n", string(txt))
			p.Node.SetTimestamp(time.Now())
			err = apiClient.Node().Update(node)
			if err != nil {
				log.Fatalf("Unable to update node: %v", err)
			}
		}
	},
}
