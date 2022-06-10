package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	openapi "github.com/hyperonecom/h1-client-go"
)

type Artifact struct {
	imageName string
	imageID   string
	state     multistep.StateBag

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.imageID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Image '%s' created, ID: %s", a.imageName, a.imageID)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	if a.imageID == "" {
		// No image to destroy
		return nil
	}

	client := a.state.Get("client").(*openapi.APIClient)
	config := a.state.Get("config").(*Config)

	refreshToken(a.state) //TODO move to h1-client-go
	_, _, err := client.
		StorageProjectImageApi.
		StorageProjectImageDelete(context.TODO(), config.Project, config.Location, a.imageID).
		Execute()

	if err != nil {
		return err
	}

	return nil
}
