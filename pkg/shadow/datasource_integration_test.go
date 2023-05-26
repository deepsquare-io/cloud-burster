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

func (suite *DataSourceTestSuite) TestCreateVM() {

	//VolumeUUID, err := suite.impl.CreateBlockDevice(context.Background(), &host)
	//suite.NoError(err)
	//suite.NotEmpty(VolumeUUID)

	// Act
	res, err := suite.impl.CreateVM(context.Background(), &host, &cloud, "dc541777-ca8b-4e5f-93b0-12558c2af647")

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
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZWQyNTUxOQAAACA3vHuWDLJFWghIjfbqmYbfpWHWVOpMARntU5rgOlowGQAAAJC5TLpguUy6YAAAAAtzc2gtZWQyNTUxOQAAACA3vHuWDLJFWghIjfbqmYbfpWHWVOpMARntU5rgOlowGQAAAEB2nasonYHV4ol3zAHeW5QR2dZZL7kJINO33HiFub7Qdze8e5YMskVaCEiN9uqZht+lYdZU6kwBGe1TmuA6WjAZAAAACXJvb3RAUDE0UwECAwQ=-----END OPENSSH PRIVATE KEY-----"
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
