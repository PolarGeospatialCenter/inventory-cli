package cmd

import (
	"github.com/PolarGeospatialCenter/inventory-client/pkg/api/client"
)

func apiConnect() (*client.InventoryApi, error) {
	return client.NewInventoryApiDefaultConfig(inventoryCfgProfile)
}
