packer {
  required_plugins {
    hyperone = {
      version = ">= 2.0.0"
      source  = "github.com/hashicorp/hyperone"
    }
  }
}

variable "project" {
  type    = string
  default = env("HYPERONE_PROJECT")
}

source "hyperone" "demo" {
  chroot_disk            = true
  disk_size              = 10
  network                = "public"
  project                = var.project
  source_image           = "debian"
  vm_type                = "a1.nano"
  chroot_command_wrapper = "sudo {{.Command}}"
  pre_mount_commands = [
    "apt-get update",
    "apt-get install -y parted debootstrap",
    "parted {{.Device}} mklabel msdos mkpart primary 1M 100%% set 1 boot on print",
    "mkfs.ext4 {{.Device}}1"
  ]
  post_mount_commands = [
    "debootstrap --arch amd64 buster {{.MountPath}}"
  ]
}

build {
  sources = ["source.hyperone.demo"]

}
