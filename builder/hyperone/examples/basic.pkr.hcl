# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

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
  disk_size    = 10
  network      = "public"
  project      = var.project
  source_image = "debian"
  vm_type      = "a1.nano"
}

build {
  sources = ["source.hyperone.demo"]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get upgrade -y"
    ]
  }

}
