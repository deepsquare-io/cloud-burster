//go:build unit

package openstack_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/squarefactory/cloud-burster/pkg/openstack"
	"github.com/stretchr/testify/suite"
)

type CloudConfigTestSuite struct {
	suite.Suite
}

func (suite *CloudConfigTestSuite) TestGenerateCloudConfig() {
	// Arrange
	opts := openstack.CloudConfigOpts{
		AuthorizedKeys: []string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4",
		},
		DNS:    "1.1.1.1",
		Search: "example.com",
		PostScripts: config.PostScriptsOpts{
			Git: config.GitOpts{
				Key: "key",
				URL: "url",
				Ref: "ref",
			},
		},

		CustomCloudConfig: `users:
  - name: user
    passwd: $6$rounds=4096$im4bWTNrEwWBTJy/$4xuVSLiNd56v9Pxk7tHehxgFDLgmqxod78qV0484ys.Duu1mXZ9dq4w1vIjrNOWh25ewWQ6N8E6MLxdvXxv3x1
`,
	}
	expected := `#cloud-config
disable_root: false

ssh_authorized_keys:
  - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4

write_files:
  - path: /etc/systemd/resolved.conf
    content: |
      [Resolve]
      DNS=1.1.1.1
      DNSStubListener=no

  - path: /etc/NetworkManager/NetworkManager.conf
    content: |
      [main]
      plugins = ifcfg-rh
      dns = none

      [logging]

  - path: /etc/resolv.conf
    content: |
      nameserver 1.1.1.1
      search example.com

  - path: /key
    content: key
    encoding: b64
    permissions: '0600'

runcmd:
  - [ systemctl, restart, NetworkManager ]
  - [ systemctl, stop, firewalld ]
  - [ systemctl, disable, firewalld ]
  - [ sed, "-i", "-e", 's/SELINUX=enforcing/SELINUX=disabled/g', /etc/selinux/config]
  - [ setenforce, "0" ]

  - [mkfs.xfs, "/dev/sdb"]
  - [mkdir, -p, "/mnt/storage"]
  - [mount, "/dev/sdb", "/mnt/storage"]

  - mkdir -p /configs && GIT_SSH_COMMAND='ssh -i /key -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o IdentitiesOnly=yes' git clone -b ref url /configs
  - if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then cd /configs && ./post.sh compute; fi
  - [ rm, -f, /key ]
  - [ chmod, -R, "g-rwx,o-rwx", /configs ]

  - [ touch, /etc/cloud/cloud-init.disabled ]

users:
  - name: user
    passwd: $6$rounds=4096$im4bWTNrEwWBTJy/$4xuVSLiNd56v9Pxk7tHehxgFDLgmqxod78qV0484ys.Duu1mXZ9dq4w1vIjrNOWh25ewWQ6N8E6MLxdvXxv3x1

`

	res, err := openstack.GenerateCloudConfig(&opts)
	suite.NoError(err)
	suite.Equal(expected, string(res))
}

func TestCloudConfigTestSuite(t *testing.T) {
	suite.Run(t, &CloudConfigTestSuite{})
}
