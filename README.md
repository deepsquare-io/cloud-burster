# Cloud Burster

A cloud bursting utility tool for [Cluster Factory](https://github.com/SquareFactory/ClusterFactory-CE).

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloud-burster ./cmd
```

## Usage

1. Create a `config.yaml` in the `/etc/cloud-burster` directory, with permission `600`.

```yaml
apiVersion: 'cloud-burster.squarefactory.io/v1alpha1'
clouds:
  - network:
      name: 'name'
      subnetCIDR: '172.28.0.0/20'
    groupsHost:
      - namePattern: cn-s-[1-50].example.com
        ipCIDR: 172.28.0.0/20
        template:
          diskSize: 50
          flavorName: 'd2-2'
          imageName: 'Rocky Linux 9'
    cloudConfig:
      authorizedKeys:
        - 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDUnXMBGq6bV6H+c7P5QjDn1soeB6vkodi6OswcZsMwH nguye@PC-DARKNESS4'
      dns: 1.1.1.1
      search: example.com
      postScripts:
        git:
          key: key
          url: git@github.com:SquareFactory/compute-configs.git
          ref: main
    openstack:
      enabled: true
      identityEndpoint: https://auth.cloud.ovh.net/
      username: user
      password: ''
      region: GRA9
      tenantID: tenantID
      tenantName: 'tenantName'
      domainID: default
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
