package ilert

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go"
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
			"ilert_alert_source":      dataSourceAlertSource(),
			"ilert_escalation_policy": dataSourceEscalationPolicy(),
			"ilert_user":              dataSourceUser(),
			"ilert_schedule":          dataSourceSchedule(),
			"ilert_uptime_monitor":    dataSourceUptimeMonitor(),
			"ilert_connection":        dataSourceConnection(),
			"ilert_connector":         dataSourceConnector(),
			"ilert_team":              dataSourceTeam(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ilert_alert_source":      resourceAlertSource(),
			"ilert_user":              resourceUser(),
			"ilert_escalation_policy": resourceEscalationPolicy(),
			"ilert_uptime_monitor":    resourceUptimeMonitor(),
			"ilert_connection":        resourceConnection(),
			"ilert_connector":         resourceConnector(),
			"ilert_team":              resourceTeam(),
		},
	}
	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}
	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
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
		return nil, errors.New("Api token or basic credentials are required")
	}
	return client, nil
}
