package cmd

import (
	"fmt"
	"net/url"

	"github.com/PolarGeospatialCenter/inventory/pkg/api/client"
	"github.com/PolarGeospatialCenter/vaulthelper/pkg/vaulthelper"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

func apiConnect() (*client.InventoryApi, error) {
	baseUrl, err := url.Parse(viper.GetString("baseurl"))
	if err != nil {
		return nil, fmt.Errorf("unable to parse base url: %v", err)
	}

	awsConfig := &aws.Config{}
	awsConfig.WithRegion(viper.GetString("aws.region"))

	vaultClient, err := vaulthelper.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to vault: %v", err)
	}
	credProvider := &vaulthelper.VaultAwsStsCredentials{
		VaultClient: vaultClient,
		VaultRole:   viper.GetString("aws.vault_role"),
	}
	awsConfig.WithCredentials(credentials.NewCredentials(credProvider))

	apiClient := client.NewInventoryApi(baseUrl, awsConfig)
	return apiClient, nil
}
