//go:build integration

package openstack_test

import (
	"os"
	"testing"

	"github.com/gophercloud/gophercloud"
	ostack "github.com/gophercloud/gophercloud/openstack"
	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var (
	image      = "Ubuntu 22.04"
	flavor     = "s1-2"
	network    = "cf-net"
	subnetCIDR = "172.28.0.0/20"
	ip         = "172.28.1.254"

	hostConfig = config.Host{
		Name:       "delete-me-integration-test",
		OS:         "rhel9",
		DiskSize:   5,
		FlavorName: flavor,
		ImageName:  image,
		IP:         "172.28.1.254",
		Netmask:    "20",
		DNS:        "8.8.8.8",
		Search:     "",
	}
	networkConfig = config.Network{
		Name:       network,
		SubnetCIDR: subnetCIDR,
	}
	userData = ""
)

type DataSourceTestSuite struct {
	suite.Suite
	endpoint    string
	user        string
	password    string
	region      string
	projectName string
	projectID   string
	impl        *openstack.DataSource
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
	res, err := suite.impl.Create(
		hostConfig,
		networkConfig,
		[]byte(userData),
	)

	// Assert
	suite.NoError(err)
	suite.NotNil(res)

	if res == nil {
		return
	}

	// Cleanup
	err = suite.impl.Delete(hostConfig.Name)

	// Assert
	suite.NoError(err)
}

func (suite *DataSourceTestSuite) BeforeTest(suiteName, testName string) {
	client, err := ostack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: suite.endpoint,
		Username:         suite.user,
		Password:         suite.password,
		TenantID:         suite.projectID,
		TenantName:       suite.projectName,
		DomainID:         "default",
	})
	if err != nil {
		logger.I.Panic("failed to authenticate client", zap.Error(err))
	}
	suite.impl = openstack.New(client, suite.region)
}
func TestDataSourceTestSuite(t *testing.T) {
	user := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	endpoint := os.Getenv("OS_AUTH_URL")
	region := os.Getenv("OS_REGION_NAME")
	projectName := os.Getenv("OS_PROJECT_NAME")
	projectID := os.Getenv("OS_PROJECT_ID")
	// Skip test if not defined
	if user == "" || password == "" || endpoint == "" || projectName == "" || projectID == "" || region == "" {
		logger.I.Warn("mandatory variables are not set!")
	} else {
		suite.Run(t, &DataSourceTestSuite{
			endpoint: endpoint,
			user:     user,
			password: password,
			region:   region,
		})
	}
}
