package ilert

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

// Provider represents the provider interface
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_ENDPOINT", ""),
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_API_TOKEN", ""),
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_ORGANIZATION", ""),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_USERNAME", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_PASSWORD", ""),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ilert_alert_action":       dataSourceAlertAction(),
			"ilert_alert_source":       dataSourceAlertSource(),
			"ilert_connection":         dataSourceConnection(),
			"ilert_connector":          dataSourceConnector(),
			"ilert_escalation_policy":  dataSourceEscalationPolicy(),
			"ilert_incident_template":  dataSourceIncidentTemplate(),
			"ilert_metric":             dataSourceMetric(),
			"ilert_metric_data_source": dataSourceMetricDataSource(),
			"ilert_schedule":           dataSourceSchedule(),
			"ilert_service":            dataSourceService(),
			"ilert_status_page":        dataSourceStatusPage(),
			"ilert_status_page_group":  dataSourceStatusPageGroup(),
			"ilert_team":               dataSourceTeam(),
			"ilert_uptime_monitor":     dataSourceUptimeMonitor(),
			"ilert_user":               dataSourceUser(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ilert_alert_action":              resourceAlertAction(),
			"ilert_alert_source":              resourceAlertSource(),
			"ilert_automation_rule":           resourceAutomationRule(),
			"ilert_connection":                resourceConnection(),
			"ilert_connector":                 resourceConnector(),
			"ilert_escalation_policy":         resourceEscalationPolicy(),
			"ilert_incident_template":         resourceIncidentTemplate(),
			"ilert_metric":                    resourceMetric(),
			"ilert_metric_data_source":        resourceMetricDataSource(),
			"ilert_schedule":                  resourceSchedule(),
			"ilert_service":                   resourceService(),
			"ilert_status_page":               resourceStatusPage(),
			"ilert_status_page_group":         resourceStatusPageGroup(),
			"ilert_team":                      resourceTeam(),
			"ilert_uptime_monitor":            resourceUptimeMonitor(),
			"ilert_user":                      resourceUser(),
			"ilert_user_email_contact":        resourceUserEmailContact(),
			"ilert_user_phone_number_contact": resourceUserPhoneNumberContact(),
		},
	}
	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}
	return p
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	endpoint := d.Get("endpoint").(string)
	apiToken := d.Get("api_token").(string)
	organization := d.Get("organization").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	client := ilert.NewClient()
	if endpoint != "" {
		ilert.WithAPIEndpoint(endpoint)(client)
	}
	if terraformVersion != "" {
		ilert.WithUserAgent(fmt.Sprintf("terraform/%s-%s-%s", terraformVersion, runtime.GOOS, runtime.GOARCH))(client)
	}
	if apiToken != "" {
		ilert.WithAPIToken(apiToken)(client)
	} else if organization != "" && username != "" && password != "" {
		ilert.WithBasicAuth(organization, username, password)(client)
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Api token or basic credentials are required",
			Detail:   "Unable to create ilert client with the given token or basic credentials, either the token or basic credentials are empty or invalid",
		})
		return nil, diags
	}
	return client, nil
}
