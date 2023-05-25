//go:build integration

package openstack_test

import (
	"context"
	"os"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
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
	image, err := suite.impl.FindNetworkID(cloud.Network.Name)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(image)
}

func (suite *DataSourceTestSuite) TestFindSubnetIDByNetwork() {
	// Arrange
	networkID, err := suite.impl.FindNetworkID(cloud.Network.Name)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(networkID)

	// Act
	res, err := suite.impl.FindSubnetIDByNetwork(cloud.Network.SubnetCIDR, networkID)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestCreatePort() {
	// Arrange
	networkID, err := suite.impl.FindNetworkID(cloud.Network.Name)
	suite.NoError(err)
	suite.NotEmpty(networkID)
	subnetID, err := suite.impl.FindSubnetIDByNetwork(cloud.Network.SubnetCIDR, networkID)
	suite.NoError(err)
	suite.NotEmpty(subnetID)

	res, err := suite.impl.CreatePort(host.IP, networkID, subnetID)

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
	err := suite.impl.Create(context.Background(), &host, &cloud)

	// Assert
	suite.NoError(err)

	// Cleanup
	err = suite.impl.Delete(context.Background(), host.Name)

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
	if user == "" || password == "" || endpoint == "" || tenantName == "" || tenantID == "" ||
		region == "" ||
		domainID == "" {
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
