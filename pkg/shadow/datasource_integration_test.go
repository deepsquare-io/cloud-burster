package shadow_test

import (
	"context"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/shadow"
	"github.com/stretchr/testify/suite"
)

var (
	host = config.Host{
		Name:       "delete-me-integration-test",
		DiskSize:   64,
		FlavorName: "VM-A4500-7543P-R2",
		ImageName:  "https://sos-ch-dk-2.exo.io/squareos-shadow/",
		IP:         "",
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
	username string
	password string
	zone     string
	sshKey   string
	impl     *shadow.DataSource
}

func (suite *DataSourceTestSuite) TestCreateBlockDevice() {
	// Act
	res, err := suite.impl.CreateBlockDevice(context.Background(), &host)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(res)
}

func (suite *DataSourceTestSuite) TestCreate() {
	// Act
	err := suite.impl.Create(context.Background(), &host, &cloud)

	// Assert
	suite.NoError(err)

}

func (suite *DataSourceTestSuite) TestDelete() {
	err := suite.impl.Delete(context.Background(), "2cca2979-71e3-4749-8c02-73f3038241a0")

	suite.NoError(err)
}

func (suite *DataSourceTestSuite) BeforeTest(suiteName, testName string) {
	suite.impl = shadow.New(
		suite.username,
		suite.password,
		suite.zone,
		suite.sshKey,
	)
}
func TestDataSourceTestSuite(t *testing.T) {
	username := "c_deepsquare_demo01"
	password := "kN1wkkvhofvuXi@czukFtsrGtVU6SwVA"
	zone := "camtl01"
	sshKey := "LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFBQUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUFNd0FBQUF0emMyZ3RaVwpReU5UVXhPUUFBQUNBM3ZIdVdETEpGV2doSWpmYnFtWWJmcFdIV1ZPcE1BUm50VTVyZ09sb3dHUUFBQUpDNVRMcGd1VXk2CllBQUFBQXR6YzJndFpXUXlOVFV4T1FBQUFDQTN2SHVXRExKRldnaElqZmJxbVliZnBXSFdWT3BNQVJudFU1cmdPbG93R1EKQUFBRUIybmFzb25ZSFY0b2wzekFIZVc1UVIyZFpaTDdrSklOTzMzSGlGdWI3UWR6ZThlNVlNc2tWYUNFaU45dXFaaHQrbApZZFpVNmt3QkdlMVRtdUE2V2pBWkFBQUFDWEp2YjNSQVVERTBVd0VDQXdRPQotLS0tLUVORCBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0K"
	// Skip test if not defined
	if username == "" || password == "" || zone == "" || sshKey == "" {
		logger.I.Warn("mandatory variables are not set!")
	} else {
		suite.Run(t, &DataSourceTestSuite{
			username: username,
			password: password,
			zone:     zone,
			sshKey:   sshKey,
		})
	}
}
