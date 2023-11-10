# Cloud Burster

A cloud bursting utility tool for [Cluster Factory](https://github.com/SquareFactory/ClusterFactory-CE).

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloud-burster ./cmd
```

## Usage

1. Create a `config.yaml` in the `/etc/cloud-burster` directory, with permission `600`.

Example:

```yaml
apiVersion: 'cloud-burster.squarefactory.io/v1alpha1'

## Use suffixSearch to append a suffix when receiving an input.
## For example, if you use the name "cn-s-1" and the suffixSearch "example.com",
## the cloud-burster will search "cn-s-1.example.com", then "cn-s-1".
suffixSearch:
  - '.example.com'
clouds:
  - type: openstack
    network:
      name: 'net'
      subnetCIDR: '172.28.0.0/20'
      dns: 1.1.1.1
      search: example.com
      gateway: 172.28.0.2
    hosts:
      - name: 'host'
        diskSize: 50
        flavorName: 'd2-2'
        imageName: 'Rocky Linux 9'
        ip: 172.28.16.254
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        ipOffset: 256
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    authorizedKeys:
      - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
    postScripts:
      git:
        key: key
        url: git@github.com:SquareFactory/compute-configs.git
        ref: main
    openstack:
      # If you have the openrc.sh file, this corresponds to:
      # identityEndpoint: OS_AUTH_URL
      # username: OS_USERNAME
      # password: OS_PASSWORD
      # region: OS_REGION_NAME
      # tenantID: OS_PROJECT
      # tenantName: OS_PROJECT_NAME
      # domainID: OS_PROJECT_DOMAIN_ID
      identityEndpoint: https://auth.cloud.ovh.net/
      username: user
      password: password
      region: GRA9
      tenantID: tenantID
      tenantName: 'tenantName'
      domainID: default
  - type: exoscale
    network:
      name: 'net'
      subnetCIDR: '172.28.0.0/20'
      dns: 1.1.1.1
      search: example.com
      gateway: 172.28.0.2
    hosts:
      - name: 'host'
        diskSize: 50
        flavorName: 'd2-2'
        imageName: 'Rocky Linux 9'
        ip: 172.28.16.254
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        ipOffset: 256
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    authorizedKeys:
      - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
    postScripts:
      git:
        key: key
        url: git@github.com:SquareFactory/compute-configs.git
        ref: main
    exoscale:
      apiKey: key
      apiSecret: secret
      zone: zone
```

Then, execute the `create` or `delete` command:

```shell
./cloud-burster create cn-s-1.example.com
```

```shell
./cloud-burster create cn-s-1.example.com,cn-s-2.example.com
```

```shell
./cloud-burster create cn-s-[1-2].example.com
```

```shell
./cloud-burster create cn-s-[1-2,5].example.com
```
