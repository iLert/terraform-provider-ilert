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
	"ilert_alert_action": {
		factory: func() any { return &ilert.AlertActionOutput{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformAlertActionResource(e.(*ilert.AlertActionOutput), d)
		},
	},
	"ilert_alert_source": {
		factory: func() any { return &ilert.AlertSource{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformAlertSourceResource(e.(*ilert.AlertSource), d)
		},
	},
	"ilert_connector": {
		factory: func() any { return &ilert.ConnectorOutput{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformConnectorResource(e.(*ilert.ConnectorOutput), d)
		},
	},
	"ilert_deployment_pipeline": {
		factory: func() any { return &ilert.DeploymentPipelineOutput{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformDeploymentPipelineResource(e.(*ilert.DeploymentPipelineOutput), d)
		},
	},
	"ilert_escalation_policy": {
		factory: func() any { return &ilert.EscalationPolicy{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformEscalationPolicyResource(e.(*ilert.EscalationPolicy), d)
		},
	},
	"ilert_heartbeat_monitor": {
		factory: func() any { return &ilert.HeartbeatMonitor{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformHeartbeatMonitorResource(e.(*ilert.HeartbeatMonitor), d)
		},
	},
	"ilert_incident_template": {
		factory: func() any { return &ilert.IncidentTemplate{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformIncidentTemplateResource(e.(*ilert.IncidentTemplate), d)
		},
	},
	"ilert_metric": {
		factory: func() any { return &ilert.Metric{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformMetricResource(e.(*ilert.Metric), d)
		},
	},
	"ilert_metric_data_source": {
		factory: func() any { return &ilert.MetricDataSource{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformMetricDataSourceResource(e.(*ilert.MetricDataSource), d)
		},
	},
	"ilert_schedule": {
		factory: func() any { return &ilert.Schedule{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformScheduleResource(e.(*ilert.Schedule), d)
		},
	},
	"ilert_service": {
		factory: func() any { return &ilert.Service{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformServiceResource(e.(*ilert.Service), d)
		},
	},
	"ilert_support_hour": {
		factory: func() any { return &ilert.SupportHour{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformSupportHourResource(e.(*ilert.SupportHour), d)
		},
	},
	"ilert_team": {
		factory: func() any { return &ilert.Team{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformTeamResource(e.(*ilert.Team), d)
		},
	},
	"ilert_user": {
		factory: func() any { return &ilert.User{} },
		transformer: func(e any, d *schema.ResourceData) error {
			return transformUserResource(e.(*ilert.User), d)
		},
	},
}

func getResourceType(resourceType string) string {
	switch resourceType {
	case "ALERT_ACTION":
		return "ilert_alert_action"
	case "ALERT_SOURCE":
		return "ilert_alert_source"
	case "ALERT_ACTION_CONNECTOR":
		return "ilert_connector"
	case "DEPLOYMENT_PIPELINE":
		return "ilert_deployment_pipeline"
	case "ESCALATION_POLICY":
		return "ilert_escalation_policy"
	case "HEARTBEAT_MONITOR":
		return "ilert_heartbeat_monitor"
	case "INCIDENT_TEMPLATE":
		return "ilert_incident_template"
	case "METRIC":
		return "ilert_metric"
	case "METRIC_PROVIDER":
		return "ilert_metric_data_source"
	case "ONCALL_SCHEDULE":
		return "ilert_schedule"
	case "SERVICE":
		return "ilert_service"
	case "SUPPORT_HOUR":
		return "ilert_support_hour"
	case "TEAM":
		return "ilert_team"
	case "USER":
		return "ilert_user"
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
