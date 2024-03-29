Type: `hyperone`
Artifact BuilderId: `hyperone.builder`

The `hyperone` Packer builder is able to create new images on the [HyperOne
platform](http://www.hyperone.com/). The builder takes a source image, runs
any provisioning necessary on the image after launching it, then creates a
reusable image.

The builder does _not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder.

### Required:

- `disk_size` (float) - Size of the created disk, in GiB.

- `project` (string) - The id or name of the project. This field is required
  only if using session tokens. It should be skipped when using service
  account authentication.

- `source_image` (string) - ID or name of the image to launch server from.

- `token` (string) - The authentication token used to access your account.
  This can be either a session token or a service account token.
  If not defined, the builder will attempt to find it in the following order:

  - In `HYPERONE_TOKEN` environment variable.
  - In `~/.h1-cli/conf.json` config file used by [h1-cli](https://github.com/hyperonecom/h1-cli).
  - By using SSH authentication if `token_login` variable has been set.

- `vm_type` (string) - ID or name of the type this server should be created with.

### Optional:

- `api_url` (string) - Custom API endpoint URL, compatible with HyperOne.
  It can also be specified via environment variable `HYPERONE_API_URL`.

- `disk_name` (string) - The name of the created disk.

- `disk_type` (string) - The type of the created disk. Defaults to `ssd`.

- `image_description` (string) - The description of the resulting image.

- `image_name` (string) - The name of the resulting image. Defaults to
  `packer-{{timestamp}}`
  (see [configuration templates](/packer/docs/templates/legacy_json_templates/engine) for more info).

- `image_service` (string) - The service of the resulting image.

- `image_tags` (map of key/value strings) - Key/value pair tags to
  add to the created image.

- `network` (string) - The ID of the network to attach to the created server.

- `private_ip` (string) - The ID of the private IP within chosen `network`
  that should be assigned to the created server.

- `public_ip` (string) - The ID of the public IP that should be assigned to
  the created server. If `network` is chosen, the public IP will be associated
  with server's private IP.

- `public_netadp_service` (string) - Custom service of public network adapter.
  Can be useful when using custom `api_url`. Defaults to `public`.

- `ssh_keys` (array of strings) - List of SSH keys by name or id to be added
  to the server on launch.

- `state_timeout` (string) - Timeout for waiting on the API to complete
  a request. Defaults to 5m.

- `token_login` (string) - Login (an e-mail) on HyperOne platform. Set this
  if you want to fetch the token by SSH authentication.

- `user_data` (string) - User data to launch with the server. Packer will not
  automatically wait for a user script to finish before shutting down the
  instance, this must be handled in a provisioner.

- `vm_name` (string) - The name of the created server.

- `vm_tags` (map of key/value strings) - Key/value pair tags to
  add to the created server.

## Chroot disk

### Required:

- `chroot_disk` (bool) - Set to `true` to enable chroot disk build.

- `pre_mount_commands` (array of strings) - A series of commands to execute
  before mounting the chroot. This should include any partitioning and
  filesystem creation commands. The path to the device is provided by
  `{{.Device}}`.

### Optional:

- `chroot_command_wrapper` (string) - How to run shell commands. This defaults
  to `{{.Command}}`. This may be useful to set if you want to set
  environment variables or run commands with `sudo`.

- `chroot_copy_files` (array of strings) - Paths to files on the running VM
  that will be copied into the chroot environment before provisioning.
  Defaults to `/etc/resolv.conf` so that DNS lookups work.

- `chroot_device` (string) - The path of chroot device. Defaults an attempt is
  made to identify it based on the attach location.

- `chroot_disk_size` (float) - The size of the chroot disk in GiB. Defaults
  to `disk_size`.

- `chroot_disk_type` (string) - The type of the chroot disk. Defaults to
  `disk_type`.

- `chroot_mount_path` (string) - The path on which the device will be mounted.

- `chroot_mounts` (array of strings) - A list of devices to mount into the
  chroot environment. This is a list of 3-element tuples, in order:

  - The filesystem type. If this is "bind", then Packer will properly bind the
    filesystem to another mount point.

  - The source device.

  - The mount directory.

- `mount_options` (array of tuples) - Options to supply the `mount` command
  when mounting devices. Each option will be prefixed with `-o` and supplied
  to the `mount` command.

- `mount_partition` (string) - The partition number containing the / partition.
  By default this is the first partition of the volume (for example, sdb1).

- `post_mount_commands` (array of strings) - As `pre_mount_commands`, but the
  commands are executed after mounting the root device and before the extra
  mount and copy steps. The device and mount path are provided by
  `{{.Device}}` and `{{.MountPath}}`.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
token.

```json
{
  "type": "hyperone",
  "token": "YOUR_AUTH_TOKEN",
  "source_image": "ubuntu-18.04",
  "vm_type": "a1.nano",
  "disk_size": 10
}
```

## Chroot Example

```json
{
  "type": "hyperone",
  "token": "YOUR_AUTH_TOKEN",
  "source_image": "ubuntu-18.04",
  "vm_type": "a1.nano",
  "disk_size": 10,
  "chroot_disk": true,
  "pre_mount_commands": [
    "apt-get update",
    "apt-get install debootstrap",
    "debootstrap --arch amd64 bionic {{.MountPath}}"
  ]
}
```

## HCL Example

```hcl
variable "token" {
  type = string
}

variable "project" {
  type = string
}

source "hyperone" "new-syntax" {
  token = var.token
  project = var.project
  source_image = "debian"
  disk_size = 10
  vm_type = "a1.nano"
  image_name = "packerbats-hcl-{{timestamp}}"
  image_tags = {
      key="value"
  }
}

build {
  sources = [
    "source.hyperone.new-syntax"
  ]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get upgrade -y"
    ]
  }
}
```
