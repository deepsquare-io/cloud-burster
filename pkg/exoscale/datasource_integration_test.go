//go:build integration

package exoscale_test

import (
	"context"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/exoscale"
	"github.com/stretchr/testify/suite"
)

var (
	host = config.Host{
		Name:       "delete-me-integration-test",
		DiskSize:   10,
		FlavorName: "Tiny",
		ImageName:  "Rocky Linux 9 (Blue Onyx) 64-bit",
		IP:         "172.24.1.254",
	}

	cloud = config.Cloud{
		Network: config.Network{
			Name:       "exo-connected-gcp",
			SubnetCIDR: "172.24.0.0/20",
			DNS:        "1.1.1.1",
			Gateway:    "172.24.0.2",
		},
		AuthorizedKeys: []string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
		},
		Hosts: []config.Host{host},
	}
)

type DataSourceTestSuite struct {
	suite.Suite
	url       string
	apiKey    string
	apiSecret string
	zone      string
	impl      *exoscale.DataSource
}

func (suite *DataSourceTestSuite) TestFindImageID() {
	// Act
	res, err := suite.impl.FindImageID(context.Background(), host.ImageName)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindFlavorID() {
	// Act
	res, err := suite.impl.FindFlavorID(context.Background(), host.FlavorName)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindNetworkID() {
	// Act
	networkID, err := suite.impl.FindNetworkID(context.Background(), cloud.Network.Name)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(networkID)
}

func (suite *DataSourceTestSuite) TestCreate() {
	// Act
	ctx := context.Background()
	err := suite.impl.Create(ctx, &host, &cloud)

	// Assert
	suite.NoError(err)

	// Cleanup
	err = suite.impl.Delete(ctx, host.Name)

	// Assert
	suite.NoError(err)
}

func (suite *DataSourceTestSuite) BeforeTest(suiteName, testName string) {
	suite.impl = exoscale.New(
		suite.url,
		suite.apiKey,
		suite.apiSecret,
		suite.zone,
	)
}
func TestDataSourceTestSuite(t *testing.T) {
	// url := os.Getenv("EXO_URL")
	// key := os.Getenv("EXO_API_KEY")
	// secret := os.Getenv("EXO_API_SECRET")
	// zone := os.Getenv("EXO_ZONE")
	url := "https://api.exoscale.com/compute/"
	key := "EXO8fe7a4b7eb3d1cce648dc300"
	secret := "v_hGzbWzIq7oFQZMpQvR_DAcL6J9jxJAsAkgGCIJV2M"
	zone := "at-vie-1"
	// Skip test if not defined
	if url == "" || key == "" || secret == "" {
		logger.I.Warn("mandatory variables are not set!")
	} else {
		suite.Run(t, &DataSourceTestSuite{
			url:       url,
			apiKey:    key,
			apiSecret: secret,
			zone:      zone,
		})
	}
}
