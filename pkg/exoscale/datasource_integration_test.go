//go:build integration

package exoscale_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/exoscale"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
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
		Network: &config.Network{
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
	fmt.Println(res)
}

func (suite *DataSourceTestSuite) TestFindFlavorID() {
	// Act
	res, err := suite.impl.FindFlavorID(context.Background(), host.FlavorName)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
	fmt.Println(res)
}

func (suite *DataSourceTestSuite) TestFindNetworkID() {
	// Act
	res, err := suite.impl.FindNetworkID(context.Background(), cloud.Network.Name)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
	fmt.Println(res)
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
		suite.apiKey,
		suite.apiSecret,
		suite.zone,
	)
}
func TestDataSourceTestSuite(t *testing.T) {
	if err := godotenv.Load(".env.test"); err != nil {
		// Skip test if not defined
		logger.I.Error("Error loading .env.test file", zap.Error(err))
	} else {
		suite.Run(t, &DataSourceTestSuite{
			apiKey:    os.Getenv("EXO_API_KEY"),
			apiSecret: os.Getenv("EXO_API_SECRET"),
			zone:      os.Getenv("EXO_ZONE"),
		})
	}
}
