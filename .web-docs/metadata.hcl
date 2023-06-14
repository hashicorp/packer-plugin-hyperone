# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "HyperOne"
  description = "The HyperOne plugin can be used with HashiCorp Packer to create custom images on HyperOne."
  identifier = "packer/hashicorp/hyperone"
  component {
    type = "builder"
    name = "HyperOne"
    slug = "hyperone"
  }
}
