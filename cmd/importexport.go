package cmd

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/PolarGeospatialCenter/inventory/pkg/inventory/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdExport)
}

type InventoryBackup struct {
	BackupDate time.Time
	Nodes      []*types.Node
	Networks   []*types.Network
	Systems    []*types.System
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
