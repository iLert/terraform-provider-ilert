package ilert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

type Transformer func(entity any, d *schema.ResourceData) error

type ResourceInfo struct {
	ResourceType string
	Schema       map[string]*schema.Schema
	Transformer  Transformer
	NewEntity    func() any
}

var resourceRegistry = map[string]struct {
	factory     func() any
	transformer Transformer
}{
	"ilert_service": {
		factory: func() any { return &ilert.Service{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformServiceResource(e.(*ilert.Service), d)
		},
	},
	"ilert_alert_source": {
		factory: func() any { return &ilert.AlertSource{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformAlertSourceResource(e.(*ilert.AlertSource), d)
		},
	},
	"ilert_alert_action": {
		factory: func() any { return &ilert.AlertActionOutput{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformAlertActionResource(e.(*ilert.AlertActionOutput), d)
		},
	},
}

func getResourceType(resourceType string) string {
	switch resourceType {
	case "SERVICE":
		return "ilert_service"
	case "ALERT_SOURCE":
		return "ilert_alert_source"
	case "ALERT_ACTION":
		return "ilert_alert_action"
	}
	return resourceType
}

func GetResourceInfo(resourceType string) (*ResourceInfo, error) {
	provider := Provider()

	resourceType = getResourceType(resourceType)
	resource, ok := provider.ResourcesMap[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", resourceType)
	}

	reg, ok := resourceRegistry[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource %s not registered", resourceType)
	}

	return &ResourceInfo{
		ResourceType: resourceType,
		Schema:       resource.Schema,
		Transformer:  reg.transformer,
		NewEntity:    reg.factory,
	}, nil
}
