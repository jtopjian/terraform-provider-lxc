# terraform-provider-lxc

This is a Terraform plugin that provides an LXC provider.

## Installation

### Requirements

Make sure you have [Terraform](http://terraform.io) installed and is accessible on your `$PATH`.

Also make sure you have LXC installed and configured on the server you will be running this.

### From Source (only method right now)

1. Install Go and [configure](https://golang.org/doc/code.html) your workspace.
2. Install `godep`:

```shell
$ go get github.com/tools/godep
```

2. Download this repo:

```shell
$ go get github.com/jtopjian/terraform-provider-lxc
```

3. Install the dependencies:

```shell
$ cd $GOPATH/src/github.com/jtopjian/terraform-provider-lxc
$ godep restore
```

4. Compile it:

```shell
$ go build -o terraform-provider-lxc
```

5. Copy it to a directory:

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
  name = "ubuntu_clone"
  source = "${lxc_container.ubuntu.name}"
}
```

Call this file `demo.tf` or anything you want and place it in the same directory as `terraform-provider-lxc`. Then run:

```shell
$ terraform plan
$ terraform apply
```

## Reference

### Provider

`provider "lxc" {}`

#### Parameters

* `lxc_path`: Optional. Explicitly set the path to where containers will be built.

### lxc_container

`resource "lxc_container" "my_container" {}`

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

`resource "lxc_clone" "my_clone" {}`

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
