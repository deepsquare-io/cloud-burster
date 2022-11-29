package exoscale

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/pkg/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type CloudConfigOpts struct {
	AuthorizedKeys []string
	PostScripts    config.PostScriptsOpts
	// AddressCIDR follows the format <ip>/<mask>
	AddressCIDR       string
	Gateway           string
	DNS               string
	Search            string
	CustomCloudConfig string
}

const cloudConfigTemplate = `#cloud-config
disable_root: false

ssh_authorized_keys:
{{- range .AuthorizedKeys }}
  - {{ . }}
{{- end }}

write_files:
  - path: /etc/systemd/resolved.conf
    content: |
      [Resolve]
      DNS={{ .DNS }}
      DNSStubListener=no

  - path: /etc/NetworkManager/NetworkManager.conf
    content: |
      [main]
      plugins = ifcfg-rh
      dns = none

      [logging]

  - path: /etc/resolv.conf
    content: |
      nameserver {{ .DNS }}
{{- if .Search }}
      search {{ .Search }}
{{ end }}

{{- if .PostScripts.Git.Key }}
  - path: /key
    content: {{ .PostScripts.Git.Key }}
    encoding: b64
    permissions: '0600'
{{- end }}

runcmd:
  - [ systemctl, restart, NetworkManager ]
  - [ systemctl, stop, firewalld ]
  - [ systemctl, disable, firewalld ]
  - [ growpart, "/dev/vda", "2" ]
  - [ xfs_growfs, "/" ]
  - [ resize2fs, "/dev/vda2" ]
  - [ nmcli, connection, modify, "Wired connection 1", connection.autoconnect, "yes" ]
  - [ nmcli, connection, modify, "Wired connection 1", ipv4.addresses, "{{ .AddressCIDR }}" ]
  - [ nmcli, connection, modify, "Wired connection 1", ipv4.gateway, "{{ .Gateway }}" ]
  - [ nmcli, connection, modify, "Wired connection 1", ipv4.route-metric, "1" ]
  - [ nmcli, connection, modify, "Wired connection 1", ipv4.never-default, "no" ]
  - [ nmcli, connection, modify, "Wired connection 1", ipv4.method, manual ]
  - [ nmcli, connection, up, "Wired connection 1" ]
  - [ nmcli, connection, down, "System ens3" ]
  - [ nmcli, connection, modify, "System ens3", connection.autoconnect, "no" ]
  - [ sed, "-i", "-e", 's/SELINUX=enforcing/SELINUX=disabled/g', /etc/selinux/config]
  - [ setenforce, "0" ]
{{- if and .PostScripts.Git.URL .PostScripts.Git.Ref }}

  - mkdir -p /configs && GIT_SSH_COMMAND='ssh -i /key -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o IdentitiesOnly=yes' git clone -b {{ .PostScripts.Git.Ref }} {{ .PostScripts.Git.URL }} /configs
  - if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then cd /configs && ./post.sh compute; fi
  - [ rm, -f, /key ]
  - [ chmod, -R, "g-rwx,o-rwx", /configs ]
{{- end }}

  - [ touch, /etc/cloud/cloud-init.disabled ]

{{ .CustomCloudConfig }}
`

func validate(cloudConfig []byte) error {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(cloudConfig, &m)
	if err != nil {
		logger.I.Error("cloud config validation failed", zap.Error(err), zap.String("cloud-config", string(cloudConfig)))
		return fmt.Errorf("cloud config validation failed: %s", err.Error())
	}
	return nil
}

func GenerateCloudConfig(options *CloudConfigOpts) ([]byte, error) {
	t, err := template.New("cloud-config").Parse(cloudConfigTemplate)
	if err != nil {
		return []byte{}, err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, options); err != nil {
		return []byte{}, err
	}

	outb := out.Bytes()

	if err := validate(outb); err != nil {
		return []byte{}, err
	}

	return outb, nil
}
