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

Call this file `demo.tf` or anything you want and place it in the same directory as `terraform-provider-lxc`. Then run:

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

### lxc_container

#### Example

```ruby
resource "lxc_container" "my_container" {
  name    = "my_container"
  backend = "zfs"
}
```

#### Parameters

* `name`: Required. The name of the container.
* `backend`: Optional. The storage backend to use. Valid options are: btrfs, directory, lvm, zfs, aufs, overlayfs, loopback, or best. Defaults to `directory`.
* `template_name`: Optional. Defaults to `download`. See `/usr/share/lxc/templates` for more template options.
* `template_distro`: Optional. Defaults to `ubuntu`.
* `template_release`: Optional. Defaults to `trusty`.
* `template_arch`: Optional. Defaults to `amd64`.
* `template_flush_cache`: Optional. Defaults to `false`.
* `template_disable_gpg_validation`: Optional. defaults to `false`.
* `options`: Optional. A set of key/value pairs of extra LXC options. See `lxc.container.conf(5)`.

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
}
```

#### Parameters

* `name`: Required. The name of the container.
* `source`: Required. The source of this clone.
* `backend`: Optional. The storage backend to use. Valid options are: btrfs, directory, lvm, zfs, aufs, overlayfs, loopback, or best. Defaults to `directory`.
* `keep_mac`: Optional. Keep the MAC address(es) of the source. Defaults to `false`.
* `snapshot`: Optional. Whether to clone as a snapshot instead of copy. Defaults to `false`.

* `options`: Optional. A set of key/value pairs of extra LXC options. See `lxc.container.conf(5)`.

#### Exported Parameters

* `address_v4`: The first discovered IPv4 address of the container.
* `address_v6`: The first discovered IPv6 address of the container.
