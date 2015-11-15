# terraform-provider-lxc

LXC provider plugin for Terraform.

## Installation

### Requirements

1. [Terraform](http://terraform.io). Make sure you have it installed and it's accessible from your `$PATH`.
2. LXC

### From Source (only method right now)

* Install the `lxc-dev` package appropriate for your distribution.
* Install Go and [configure](https://golang.org/doc/code.html) your workspace.
* Install `godep`:

```shell
$ go get github.com/tools/godep
```

* Download this repo:

```shell
$ go get github.com/jtopjian/terraform-provider-lxc
```

* Install the dependencies:

```shell
$ cd $GOPATH/src/github.com/jtopjian/terraform-provider-lxc
$ godep restore
```

* Compile it:

```shell
$ go build -o terraform-provider-lxc
```

* Copy it to a directory:

```shell
$ sudo cp terraform-provider-lxc ~/lxc-demo
```

## Usage

Here's a simple Terraform file to get you started:

```ruby
provider "lxc" {}

resource "lxc_container" "ubuntu" {
  name = "ubuntu"
}

resource "lxc_clone" "ubuntu_clone" {
  name   = "ubuntu_clone"
  source = "${lxc_container.ubuntu.name}"
}
```

Here's a more complete example that does the following:

* Creates a new bridge called `my_bridge`.
* Creates an Ubuntu container with two interfaces: one on the default `lxcbr0` and one on `my_bridge`.
* Creates an Ubuntu container with one interface on the `my_bridge` bridge.

```ruby
provider "lxc" {}

resource "lxc_bridge" "my_bridge" {
  name = "my_bridge"
}

resource "lxc_container" "ubuntu" {
  name = "ubuntu"
  template_name = "ubuntu"
  template_release = "trusty"
  template_arch = "amd64"
  template_extra_args = ["--auth-key", "/root/.ssh/id_rsa.pub"]
  network_interface {
    type = "veth"
    options {
      link = "lxcbr0"
      flags = "up"
      hwaddr = "00:16:3e:xx:xx:xx"
    }
  }
  network_interface {
    type = "veth"
    options {
      link = "${lxc_bridge.my_bridge.name}"
      flags = "up"
      hwaddr = "00:16:3e:xx:xx:xx"
      veth.pair = "foobar"
      ipv4 = "192.168.255.1/24"
    }
  }
}

resource "lxc_container" "ubuntu2" {
  name = "ubuntu2"
  template_name = "ubuntu"
  template_release = "trusty"
  template_arch = "amd64"
  template_extra_args = ["--auth-key", "/root/.ssh/id_rsa.pub"]
  network_interface {
    type = "veth"
    options {
      link = "${lxc_bridge.my_bridge.name}"
      flags = "up"
      hwaddr = "00:16:3e:xx:xx:xx"
      veth.pair = "barfoo"
      ipv4 = "192.168.255.2/24"
    }
  }
}
```

For either example, save it to a `.tf` file and run:

```shell
$ terraform plan
$ terraform apply
$ terraform show
```

## Reference

### Provider

#### Example

```ruby
provider "lxc" {
  lxc_path = "/var/lib/lxc"
}
```

#### Parameters

* `lxc_path`: Optional. Explicitly set the path to where containers will be built.

### lxc_bridge

#### Example

```ruby
resource "lxc_bridge" "my_bridge" {
  name = "my_bridge"
}
```

#### Parameters

* `name`: Required. The name of the bridge.

#### Exported Parameters

* `mac`: The MAC address of the new bridge.

### lxc_container

#### Example

```ruby
resource "lxc_container" "my_container" {
  name    = "my_container"
  backend = "zfs"
  network_interface {
    type = "veth"
    options {
      link = "lxcbr0"
      flags = "up"
      hwaddr = "00:16:3e:xx:xx:xx"
    }
  }
}
```

#### Parameters

* `name`: Required. The name of the container.
* `backend`: Optional. The storage backend to use. Valid options are: btrfs, directory, lvm, zfs, aufs, overlayfs, loopback, or best. Defaults to `directory`.
* `exec`: Optional. Commands to run after container creation. This won't be interpreted by a shell so use `bash -c "{shellcode}"` if you want a shell.
* `template_name`: Optional. Defaults to `download`. See `/usr/share/lxc/templates` for more template options.
* `template_distro`: Optional. Defaults to `ubuntu`.
* `template_release`: Optional. Defaults to `trusty`.
* `template_arch`: Optional. Defaults to `amd64`.
* `template_variant`: Optional. Defaults to `default`.
* `template_server`: Optional. Defaults to `images.linuxcontainers.org`.
* `template_key_id`: Optional.
* `template_key_server`: Optional.
* `template_flush_cache`: Optional. Defaults to `false`.
* `template_force_cache`: Optional. Defaults to `false`.
* `template_disable_gpg_validation`: Optional. defaults to `false`.
* `template_extra_args`: Optional. A list of extra parameters to pass to the template.
* `options`: Optional. A set of key/value pairs of extra LXC options. See `lxc.container.conf(5)`.
* `network_interface`: Optional. Defines a NIC.
  * `type`: Optional. The type of NIC. Defaults to `veth`.
  * `management`: Optional. Make this NIC the management / accessible NIC.
  * `options`: Optional. A set of key/value `lxc.network.*` pairs for the NIC.

#### Notes

Because `lxc.network.type` _must_ be the first line that denotes a new NIC, a separate `network_interface` parameter is used rather than bundling it all into `options`

#### Exported Parameters

* `address_v4`: The first discovered IPv4 address of the container.
* `address_v6`: The first discovered IPv6 address of the container.

### lxc_clone

#### Example

```ruby
resource "lxc_clone" "my_clone" {
  name    = "my_clone"
  source  = "my_container"
  backend = "zfs"
  network_interface {
    type = "veth"
    options {
      link = "lxcbr0"
      flags = "up"
      hwaddr = "00:16:3e:xx:xx:xx"
    }
  }
}
```

#### Parameters

* `name`: Required. The name of the container.
* `source`: Required. The source of this clone.
* `backend`: Optional. The storage backend to use. Valid options are: btrfs, directory, lvm, zfs, aufs, overlayfs, loopback, or best. Defaults to `directory`.
* `keep_mac`: Optional. Keep the MAC address(es) of the source. Defaults to `false`.
* `snapshot`: Optional. Whether to clone as a snapshot instead of copy. Defaults to `false`.
* `options`: Optional. A set of key/value pairs of extra LXC options. See `lxc.container.conf(5)`.
* `network_interface`: Optional. Defines a NIC.
  * `type`: Optional. The type of NIC. Defaults to `veth`.
  * `management`: Optional. Make this NIC the management / accessible NIC.
  * `options`: Optional. A set of key/value `lxc.network.*` pairs for the NIC.

#### Exported Parameters

* `address_v4`: The first discovered IPv4 address of the container.
* `address_v6`: The first discovered IPv6 address of the container.
