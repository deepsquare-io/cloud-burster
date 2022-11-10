//go:build integration

package openstack_test

import (
	"os"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
	"github.com/stretchr/testify/suite"
)

var (
	image      = "Ubuntu 22.04"
	flavor     = "s1-2"
	network    = "cf-net"
	subnetCIDR = "172.28.0.0/20"
	ip         = "172.28.1.254"

	hostConfig = config.Host{
		Name:       "delete-me-integration-test",
		DiskSize:   5,
		FlavorName: flavor,
		ImageName:  image,
		IP:         "172.28.1.254",
	}
	networkConfig = config.Network{
		Name:       network,
		SubnetCIDR: subnetCIDR,
	}
	cloudConfig = config.CloudConfigTemplateOpts{
		AuthorizedKeys: []string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
		},
		DNS: "1.1.1.1",
	}
)

type DataSourceTestSuite struct {
	suite.Suite
	endpoint   string
	user       string
	password   string
	region     string
	tenantName string
	tenantID   string
	domainID   string
	impl       *openstack.DataSource
}

func (suite *DataSourceTestSuite) TestFindImageID() {
	// Act
	res, err := suite.impl.FindImageID(image)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindFlavorID() {
	// Act
	res, err := suite.impl.FindFlavorID(flavor)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestFindNetworkID() {
	// Act
	image, err := suite.impl.FindNetworkID(network)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(image)
}

func (suite *DataSourceTestSuite) TestFindSubnetIDByNetwork() {
	// Arrange
	networkID, err := suite.impl.FindNetworkID(network)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(networkID)

	// Act
	res, err := suite.impl.FindSubnetIDByNetwork(subnetCIDR, networkID)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestCreatePort() {
	// Arrange
	networkID, err := suite.impl.FindNetworkID(network)
	suite.NoError(err)
	suite.NotEmpty(networkID)
	subnetID, err := suite.impl.FindSubnetIDByNetwork(subnetCIDR, networkID)
	suite.NoError(err)
	suite.NotEmpty(subnetID)

	res, err := suite.impl.CreatePort(ip, networkID, subnetID)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)

	if res == "" {
		return
	}

	// Cleanup
	err = suite.impl.DeletePort(res)

	// Assert
	suite.NoError(err)
}

func (suite *DataSourceTestSuite) TestCreate() {
	// Act
	err := suite.impl.Create(
		hostConfig,
		networkConfig,
		cloudConfig,
	)

	// Assert
	suite.NoError(err)

	// Cleanup
	err = suite.impl.Delete(hostConfig.Name)

	// Assert
	suite.NoError(err)
}

func (suite *DataSourceTestSuite) BeforeTest(suiteName, testName string) {
	suite.impl = openstack.New(
		suite.endpoint,
		suite.user,
		suite.password,
		suite.tenantID,
		suite.tenantName,
		suite.region,
		suite.domainID,
	)
}
func TestDataSourceTestSuite(t *testing.T) {
	user := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	endpoint := os.Getenv("OS_AUTH_URL")
	region := os.Getenv("OS_REGION_NAME")
	tenantName := os.Getenv("OS_PROJECT_NAME")
	tenantID := os.Getenv("OS_PROJECT_ID")
	domainID := os.Getenv("OS_PROJECT_DOMAIN_ID")
	// Skip test if not defined
	if user == "" || password == "" || endpoint == "" || tenantName == "" || tenantID == "" || region == "" || domainID == "" {
		logger.I.Warn("mandatory variables are not set!")
	} else {
		suite.Run(t, &DataSourceTestSuite{
			endpoint:   endpoint,
			user:       user,
			password:   password,
			region:     region,
			tenantName: tenantName,
			tenantID:   tenantID,
			domainID:   domainID,
		})
	}
}
