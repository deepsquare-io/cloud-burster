//go:build integration

package exoscale_test

import (
	"os"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/exoscale"
	"github.com/stretchr/testify/suite"
)

var (
	host = config.Host{
		Name:       "delete-me-integration-test",
		DiskSize:   5,
		FlavorName: "s1-2",
		ImageName:  "Rocky Linux 9",
		IP:         "172.28.1.254",
	}

	cloud = config.Cloud{
		Network: config.Network{
			Name:       "cf-net",
			SubnetCIDR: "172.28.0.0/20",
			DNS:        "1.1.1.1",
			Gateway:    "172.28.0.2",
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
	res, err := suite.impl.FindImageID(host.ImageName)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindFlavorID() {
	// Act
	res, err := suite.impl.FindFlavorID(host.FlavorName)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindNetworkID() {
	// Act
	networkID, err := suite.impl.FindNetworkID(cloud.Network.Name)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(networkID)
}

func (suite *DataSourceTestSuite) TestCreate() {
	// Act
	err := suite.impl.Create(&host, &cloud)

	// Assert
	suite.NoError(err)

	// Cleanup
	err = suite.impl.Delete(host.Name)

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
	url := os.Getenv("EXO_URL")
	key := os.Getenv("EXO_API_KEY")
	secret := os.Getenv("EXO_API_SECRET")
	zone := os.Getenv("EXO_ZONE")
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
