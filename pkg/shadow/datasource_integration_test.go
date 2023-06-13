//go:build integration

package shadow_test

import (
	"context"
	"os"
	"testing"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/shadow"
	"github.com/stretchr/testify/suite"
)

var (
	host = config.Host{
		Name:       "cn1.ca-bhs-5.deepsquare.run",
		DiskSize:   64,
		RAM:        112,
		GPU:        1,
		FlavorName: "VM-A4500-7543P-R2",
		ImageName:  "https://sos-ch-dk-2.exo.io/squareos-shadow/",
	}

	cloud = config.Cloud{
		PostScripts: config.PostScriptsOpts{
			Git: config.GitOpts{
				Key: os.Getenv("POSTCRIPTS_KEY"),
				URL: "git@github.com:SquareFactory/compute-configs.git",
				Ref: "shadow",
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
	err := suite.impl.Delete(context.Background(), "de505a71-8a6c-43d9-b153-321931c345ff")

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
	username := os.Getenv("BHS_USER")
	password := os.Getenv("BHS_PASSWORD")
	zone := os.Getenv("BHS_ZONE")
	sshKey := os.Getenv("BHS_KEY")
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
