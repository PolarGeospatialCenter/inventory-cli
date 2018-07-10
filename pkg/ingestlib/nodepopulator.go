package ingestlib

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	inventorytypes "github.com/PolarGeospatialCenter/inventory/pkg/inventory/types"

	"github.com/manifoldco/promptui"
)

func regexValidator(regex, value string) error {
	match, err := regexp.MatchString(regex, value)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("Provided value must match the regex %s", regex)
	}
	return nil
}

func nonEmpty(value string) error {
	if value != "" {
		return nil
	}
	return fmt.Errorf("Value cannot be empty")
}

func validInventoryID(value string) error {
	return regexValidator(`[a-zA-Z]+-\d{4}`, value)
}

func validRack(value string) error {
	return regexValidator(`[a-zA-Z]{2}\d{2}`, value)
}

func validInt(value string) error {
	_, err := strconv.Atoi(value)
	return err
}

func validRackSpace(value string) error {
	number, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if number < 0 || number > 42 {
		return fmt.Errorf("invalid rack space, must be in [0,42]")
	}
	return nil
}

type NodePopulator struct {
	Node     *inventorytypes.Node
	Systems  []*inventorytypes.System
	Networks []*inventorytypes.Network
}

func SelectLoop(sel promptui.Select) (int, string) {
	err := fmt.Errorf("Not run")
	var idx int
	var value string
	for err != nil {
		idx, value, err = sel.Run()
		if err != nil && err.Error() == "^C" {
			log.Fatalf("Got interrupt, exiting...")
		}
	}
	return idx, value
}

func ReadString(prompt promptui.Prompt) string {
	err := fmt.Errorf("Not run")
	var result string
	for err != nil {
		result, err = prompt.Run()
		if err != nil && err.Error() == "^C" {
			log.Fatalf("Got interrupt, exiting...")
		}
	}
	return result
}

func ReadInt(prompt promptui.Prompt) int {
	if prompt.Validate == nil {
		prompt.Validate = validInt
	}

	value := ReadString(prompt)
	result, _ := strconv.Atoi(value)
	return result
}

func (p *NodePopulator) ReadInventoryID() string {
	return ReadString(promptui.Prompt{Label: "Inventory ID", Default: p.Node.InventoryID, Validate: validInventoryID})
}

func (p *NodePopulator) ReadChassisLocation() *inventorytypes.ChassisLocation {
	location := &inventorytypes.ChassisLocation{}
	if p.Node.ChassisLocation == nil {
		p.Node.ChassisLocation = &inventorytypes.ChassisLocation{Building: "wbob", Room: "30"}
	}

	location.Building = ReadString(promptui.Prompt{Label: "Building", Default: p.Node.ChassisLocation.Building, Validate: nonEmpty})
	location.Room = ReadString(promptui.Prompt{Label: "Room", Default: p.Node.ChassisLocation.Room, Validate: nonEmpty})
	location.Rack = ReadString(promptui.Prompt{Label: "Rack", Default: p.Node.ChassisLocation.Rack, Validate: validRack})
	location.BottomU = uint(ReadInt(promptui.Prompt{Label: "Bottom Rack Space", Default: fmt.Sprintf("%d", p.Node.ChassisLocation.BottomU), Validate: validRackSpace}))

	return location
}

func (p *NodePopulator) ReadChassisSubIndex() string {
	return ReadString(promptui.Prompt{Label: "Chassis Sub-index", Default: p.Node.ChassisSubIndex})
}

func (p *NodePopulator) ReadSystem(systems []*inventorytypes.System) *inventorytypes.System {
	systemNames := []string{}
	for _, system := range systems {
		systemNames = append(systemNames, system.Name)
	}
	prompt := promptui.Select{
		Label: "Please select a system",
		Items: systemNames,
	}
	systemIdx, _ := SelectLoop(prompt)
	return systems[systemIdx]
}

func (p *NodePopulator) ReadRole(system *inventorytypes.System) string {
	prompt := promptui.Select{
		Label: "Please select a role",
		Items: system.Roles,
	}
	_, role := SelectLoop(prompt)
	return role
}

func (p *NodePopulator) ReadEnvironment(system *inventorytypes.System) string {
	environments := make([]string, 0, len(system.Environments))
	for env, _ := range system.Environments {
		environments = append(environments, env)
	}
	prompt := promptui.Select{
		Label: "Please select an environment",
		Items: environments,
	}
	_, environment := SelectLoop(prompt)
	return environment
}

func (p *NodePopulator) PopulateNode() error {
	p.Node.InventoryID = p.ReadInventoryID()
	p.Node.ChassisLocation = p.ReadChassisLocation()
	p.Node.ChassisSubIndex = p.ReadChassisSubIndex()
	system := p.ReadSystem(p.Systems)
	p.Node.System = system.ID()
	p.Node.Role = p.ReadRole(system)
	p.Node.Environment = p.ReadEnvironment(system)
	p.Node.SetTimestamp(time.Now())
	return nil
}

func (p *NodePopulator) DetectNetworks() (int, error) {
	detected, err := DetectNetworks(p.Networks)
	if err != nil {
		return 0, err
	}
	if p.Node.Networks == nil {
		p.Node.Networks = map[string]*inventorytypes.NICInfo{}
	}
	var updates int
	for networkId, detected := range detected {
		_, ok := p.Node.Networks[networkId]
		if !ok {
			p.Node.Networks[networkId] = &inventorytypes.NICInfo{}
		}

		if p.Node.Networks[networkId].MAC.String() != detected.MAC.String() {
			p.Node.Networks[networkId].MAC = detected.MAC
			updates += 1
		}
	}
	return updates, nil
}

func LookupNetworkByIp(networks []*inventorytypes.Network, ip net.IP) *inventorytypes.Network {
	for _, network := range networks {
		for _, subnet := range network.Subnets {
			if subnet.Cidr.Contains(ip) {
				return network
			}
		}
	}
	return nil
}

func DetectNetworks(networks []*inventorytypes.Network) (map[string]*inventorytypes.NICInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Unable to list interfaces: %v", err)
	}
	detected := make(map[string]*inventorytypes.NICInfo, len(ifaces))
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Unable to get addresses for interface: %v", err)
		}

		for _, addr := range addrs {
			ipAddr, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				// unparsable ip, continue
				continue
			}

			network := LookupNetworkByIp(networks, ipAddr)
			if network != nil {
				log.Printf("Found network %s at %s", network.ID(), iface.Name)
				detected[network.ID()] = &inventorytypes.NICInfo{MAC: iface.HardwareAddr}
				break
			}
		}
	}
	return detected, nil
}
