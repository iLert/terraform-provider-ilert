package ilert

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/iLert/ilert-go"
)

// Provider represents the provider interface
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ILERT_ENDPOINT", ""),
			},
			"api_token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ILERT_API_TOKEN", ""),
				ConflictsWith: []string{"organization", "username", "password"},
			},
			"organization": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ILERT_ORGANIZATION", ""),
				ConflictsWith: []string{"api_token"},
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ILERT_USERNAME", ""),
				ConflictsWith: []string{"api_token"},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("ILERT_PASSWORD", ""),
				ConflictsWith: []string{"api_token"},
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ilert_alert_source": dataSourceAlertSource(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ilert_alert_source": resourceAlertSource(),
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
		ilert.WithUserAgent(fmt.Sprintf("(%s %s) Terraform/%s", runtime.GOOS, runtime.GOARCH, terraformVersion))(client)
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
