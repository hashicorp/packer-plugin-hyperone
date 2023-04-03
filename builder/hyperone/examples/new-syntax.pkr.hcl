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
  type = string
}

source "hyperone" "new-syntax" {
  project      = var.project
  network      = "public"
  source_image = "debian"
  disk_size    = 10
  vm_type      = "a1.nano"
  image_name   = "packerbats-hcl-{{timestamp}}"
  image_tags = {
    key = "value"
  }
}

build {
  sources = [
    "source.hyperone.new-syntax"
  ]

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get upgrade -y"
    ]
  }
}