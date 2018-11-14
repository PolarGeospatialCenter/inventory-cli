package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/PolarGeospatialCenter/inventory/pkg/inventory/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdImport)
	rootCmd.AddCommand(cmdExport)
}

type InventoryBackup struct {
	BackupDate time.Time
	Nodes      []*types.Node
	Networks   []*types.Network
	Systems    []*types.System
}

var cmdImport = &cobra.Command{
	Use:        "import filename",
	ArgAliases: []string{"filename"},
	Args:       cobra.MinimumNArgs(1),
	Short:      "import objects from a backup",
	Run: func(cmd *cobra.Command, args []string) {
		api, err := apiConnect()
		if err != nil {
			log.Fatalf("unable to connect to api: %v", err)
		}
		backupData := &InventoryBackup{}

		backupDataBytes, err := ioutil.ReadFile(args[0])
		if err != nil {
			log.Fatalf("unable to open backup file: %v", err)
		}

		err = json.Unmarshal(backupDataBytes, backupData)
		if err != nil {
			log.Fatalf("unable to unmarshal backup data: %v", err)
		}

		for _, node := range backupData.Nodes {
			n, err := api.Node().Get(node.ID())
			if err != nil {
				err = api.Node().Create(node)
				if err != nil {
					log.Fatalf("Unable to create node: %v", err)
				}
			} else if n.Timestamp() < node.Timestamp() {
				err = api.Node().Update(node)
				if err != nil {
					log.Fatalf("Unable to update node: %v", err)
				}
			}
		}

		for _, system := range backupData.Systems {
			n, err := api.System().Get(system.ID())
			if err != nil {
				err = api.System().Create(system)
				if err != nil {
					log.Fatalf("Unable to create system: %v", err)
				}
			} else if n.Timestamp() < system.Timestamp() {
				err = api.System().Update(system)
				if err != nil {
					log.Fatalf("Unable to update system: %v", err)
				}
			}
		}

		for _, network := range backupData.Networks {
			n, err := api.Network().Get(network.ID())
			if err != nil {
				err = api.Network().Create(network)
				if err != nil {
					log.Fatalf("Unable to create network: %v", err)
				}
			} else if n.Timestamp() < network.Timestamp() {
				err = api.Network().Update(network)
				if err != nil {
					log.Fatalf("Unable to update network: %v", err)
				}
			}
		}

	},
}

var cmdExport = &cobra.Command{
	Use:        "export filename",
	ArgAliases: []string{"filename"},
	Args:       cobra.MinimumNArgs(1),
	Short:      "export objects to a backup file",
	Run: func(cmd *cobra.Command, args []string) {
		api, err := apiConnect()
		if err != nil {
			log.Fatalf("unable to connect to api: %v", err)
		}

		backupData := &InventoryBackup{BackupDate: time.Now()}

		nodes, err := api.Node().GetAll()
		if err != nil {
			log.Fatalf("unable to get nodes: %v", err)
		}
		backupData.Nodes = nodes

		systems, err := api.System().GetAll()
		if err != nil {
			log.Fatalf("unable to get systems: %v", err)
		}
		backupData.Systems = systems

		networks, err := api.Network().GetAll()
		if err != nil {
			log.Fatalf("unable to get networks: %v", err)
		}
		backupData.Networks = networks

		backupFile, err := os.OpenFile(args[0], os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			log.Fatalf("unable to open backup file: %v", err)
		}

		jsonData, err := json.Marshal(backupData)
		if err != nil {
			log.Fatalf("unable to marshal backup data: %v", err)
		}

		_, err = backupFile.Write(jsonData)
		if err != nil {
			log.Fatalf("error writing backup to file: %v", err)
		}
	},
}
