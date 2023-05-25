package shadow

import (
	"bytes"
	"text/template"

	"github.com/squarefactory/cloud-burster/pkg/config"
)

type CloudConfigOpts struct {
	PostScripts config.PostScriptsOpts
}

const cloudConfigTemplate = `#!/bin/bash
set -ex

# Fetch encrypted deploy key
curl --retry 5 -fsSL { .PostScripts.Git.Key }} -o /key.enc
chmod 600 /key.enc

# Decrypt deploy key
echo "my_decrypt_password_is_long" | openssl aes-256-cbc -d -a -pbkdf2 -in /key.enc -out /key -pass stdin
chmod 600 /key

# Cloning git repo containing postscripts.
mkdir -p /configs
GIT_SSH_COMMAND='ssh -i /key -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null' git clone -b {{ .PostScripts.Git.Ref }} {{ .PostScripts.Git.URL }} /configs
if [ -f /configs/post.sh ] && [ -x /configs/post.sh ]; then
	cd /configs || exit 1
	./post.sh "$1"
fi
rm -f /key /key.env

# Security
chmod -R g-rwx,o-rwx .
`

func GenerateCloudConfig(options *CloudConfigOpts) ([]byte, error) {
	t, err := template.New("cloud-config").Parse(cloudConfigTemplate)
	if err != nil {
		return []byte{}, err
	}

	var out bytes.Buffer

	if err := t.Execute(&out, options); err != nil {
		return []byte{}, err
	}

	return out.Bytes(), nil
}
