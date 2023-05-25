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
				Key: "key",
				URL: "url",
				Ref: "ref",
			},
		},
	}

	expected := `#!/bin/bash
set -ex

# Inject hostname
hostnamectl set-hostname test

# Fetch encrypted deploy key
curl --retry 5 -fsSL key -o /key.enc
chmod 600 /key.enc

# Decrypt deploy key
echo "my_decrypt_password_is_long" | openssl aes-256-cbc -d -a -pbkdf2 -in /key.enc -out /key -pass stdin
chmod 600 /key

# Cloning git repo containing postscripts.
mkdir -p /configs
GIT_SSH_COMMAND='ssh -i /key -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null' git clone -b ref url /configs
if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then
	cd /configs || exit 1
	./post.sh "$1"
fi
rm -f /key /key.env

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
