//go:build integration

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
	}

	cloud = config.Cloud{
		Network: config.Network{
			Name:       "cf-net",
			SubnetCIDR: "172.28.0.0/20",
			DNS:        "1.1.1.1",
			Gateway:    "172.28.0.2",
		},
		PostScripts: config.PostScriptsOpts{
			Git: config.GitOpts{
				Key: `LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFBQUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUFNd0FBQUF0emMyZ3RaVwpReU5UVXhPUUFBQUNEdXo5UmkwRndoUEhncnJDcmJPbkxWYkNoNDhCenhPSFZVSVJvaUhVR3V3d0FBQUpCK0IyRUlmZ2RoCkNBQUFBQXR6YzJndFpXUXlOVFV4T1FBQUFDRHV6OVJpMEZ3aFBIZ3JyQ3JiT25MVmJDaDQ4Qnp4T0hWVUlSb2lIVUd1d3cKQUFBRUFvTFI4b3liMW1mTktuRHZTOUVrMDJsVytldjJPZHlxWEw4aHphcW8xMWNPN1AxR0xRWENFOGVDdXNLdHM2Y3RWcwpLSGp3SFBFNGRWUWhHaUlkUWE3REFBQUFERzFoY21OQWJXRnlZeTF3WXdFPQotLS0tLUVORCBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0K`,
				URL: "git@github.com:SquareFactory/compute-configs.git",
				Ref: "main",
			},
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
	err := suite.impl.Delete(context.Background(), "2404fc69-bde4-4411-ba22-fe2074fbf200")

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
