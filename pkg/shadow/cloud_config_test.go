package shadow_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/shadow"
	"github.com/stretchr/testify/suite"
)

type CloudConfigTestSuite struct {
	suite.Suite
}

func (suite *CloudConfigTestSuite) TestGenerateCloudConfig() {
	// Arrange
	opts := shadow.CloudConfigOpts{
		Hostname: "test",
		PostScripts: config.PostScriptsOpts{
			Git: config.GitOpts{
				Key: `LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFB
QUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUFNd0FBQUF0emMyZ3RaVwpReU5UVXhP
UUFBQUNCWnVxejEzQS91MWtIcW1adEI4TjUzbmN5d0JqMC9kY2FXTWpabVFUcWVaZ0FBQUpCR1hX
dEdSbDFyClJnQUFBQXR6YzJndFpXUXlOVFV4T1FBQUFDQlp1cXoxM0EvdTFrSHFtWnRCOE41M25j
eXdCajAvZGNhV01qWm1RVHFlWmcKQUFBRURJL1RnVkp3M0FvUjA5bG52WDFZZVhPQWxlS1A2TGdh
Vi9zRmhiaXRNLzBsbTZyUFhjRCs3V1FlcVptMEh3M25lZAp6TEFHUFQ5MXhwWXlObVpCT3A1bUFB
QUFEVzFoY21OQWRIVnVaM04wWlc0PQotLS0tLUVORCBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0K`,
				URL: "url",
				Ref: "ref",
			},
		},
	}

	expected := `#!/bin/bash
set -ex

# Inject hostname
hostnamectl set-hostname test

cat << 'EOF' >> /key
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBZuqz13A/u1kHqmZtB8N53ncywBj0/dcaWMjZmQTqeZgAAAJBGXWtGRl1r
RgAAAAtzc2gtZWQyNTUxOQAAACBZuqz13A/u1kHqmZtB8N53ncywBj0/dcaWMjZmQTqeZg
AAAEDI/TgVJw3AoR09lnvX1YeXOAleKP6LgaV/sFhbitM/0lm6rPXcD+7WQeqZm0Hw3ned
zLAGPT91xpYyNmZBOp5mAAAADW1hcmNAdHVuZ3N0ZW4=
-----END OPENSSH PRIVATE KEY-----

EOF
chmod 600 /key

# Cloning git repo containing postscripts.
mkdir -p /configs
GIT_SSH_COMMAND='ssh -i /key -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null' git clone -b ref url /configs
if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then
	cd /configs || exit 1
	./post.sh "$1"
fi
rm -f /key

# Security
chmod -R g-rwx,o-rwx .
`

	res, err := shadow.GenerateCloudConfig(&opts)
	suite.NoError(err)
	suite.Equal(expected, string(res))
}

func TestCloudConfigTestSuite(t *testing.T) {
	suite.Run(t, &CloudConfigTestSuite{})
}
