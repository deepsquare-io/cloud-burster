package shadow

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/squarefactory/cloud-burster/pkg/config"
)

type CloudConfigOpts struct {
	PostScripts config.PostScriptsOpts
	Hostname    string
}

const cloudConfigTemplate = `#!/bin/bash
set -ex

# Inject hostname
hostnamectl set-hostname {{ .Hostname }}

cat << 'EOF' >> /key
{{ .PostScripts.Git.Key | b64dec }}
EOF
chmod 600 /key

# Cloning git repo containing postscripts.
mkdir -p /configs
GIT_SSH_COMMAND='ssh -i /key -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null' git clone -b {{ .PostScripts.Git.Ref }} {{ .PostScripts.Git.URL }} /configs
if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then
	cd /configs || exit 1
	./post.sh "$1"
fi
rm -f /key

# Security
chmod -R g-rwx,o-rwx .
`

func GenerateCloudConfig(options *CloudConfigOpts) ([]byte, error) {
	t, err := template.New("cloud-config").Funcs(sprig.TxtFuncMap()).Parse(cloudConfigTemplate)
	if err != nil {
		return []byte{}, err
	}

	var out bytes.Buffer

	if err := t.Execute(&out, options); err != nil {
		return []byte{}, err
	}

	return out.Bytes(), nil
}
