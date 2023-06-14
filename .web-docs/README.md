The HyperOne Plugin is able to create new images on the HyperOne platform.

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    hyperone = {
      source  = "github.com/hashicorp/hyperone"
      version = "~> 1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
packer plugins install github.com/hashicorp/hyperone
```

### Components

#### Builders

- [hyperone](packer/integrations/hashicorp/hyperone/latest/components/builder/hyperone.mdx) - The hyperone builder takes a source image, runs any
provisioning necessary on the image after launching it, then creates a reusable image.

### Authentication

HyperOne supports several authentication methods, which are all supported by
this builder.

#### User session

If using user session, set the `token` field to your authentication token.
The `project` field is required when using this method.

```json
{
  "token": "YOUR TOKEN",
  "project": "YOUR_PROJECT"
}
```

#### User session by SSH key

If you've added an SSH key as a credential to your user account and the
private key is added to the ssh-agent on your local machine, you can
authenticate by setting just the platform login (your e-mail address):

```json
{
  "token_login": "your.user@example.com"
}
```

#### h1 CLI

If you're using [h1-cli](https://github.com/hyperonecom/h1-cli) on your local
machine, HyperOne builder can use your credentials saved in a config file.

All you have to do is login within the tool:

```shell-session
$ h1 login --username your.user@example.com
```

You don't have to set `token` or `project` fields at all using this method.

#### Service account

Using `h1`, you can create a new token associated with chosen project.

```shell-session
$ h1 project token add --name packer-builder --project PROJECT_ID
```

Set the `token` field to the generated token or save it in the `HYPERONE_TOKEN`
environment variable. You don't have to set the `project` option using this
method.

```json
{
  "token": "YOUR TOKEN"
}
```
