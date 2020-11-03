package ilert

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringValueFunc(ilert.ConnectorTypesAll),
			},
			"datadog": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"jira",
					"microsoft_teams",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"jira": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"email": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"microsoft_teams": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"servicenow": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"zendesk": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"email": {
							Type:     schema.TypeString,
							Required: true,
						},
						"api_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"discord": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"microsoft_teams",
					"servicenow",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"github": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"topdesk": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"aws_lambda": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"azure_faas",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"azure_faas": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"google_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"google_faas": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"sysdig",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"sysdig": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"microsoft_teams",
					"jira",
					"servicenow",
					"zendesk",
					"discord",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Create: resourceConnectorCreate,
		Read:   resourceConnectorRead,
		Update: resourceConnectorUpdate,
		Delete: resourceConnectorDelete,
		Exists: resourceConnectorExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func buildConnector(d *schema.ResourceData) (*ilert.Connector, error) {
	name := d.Get("name").(string)
	connectorType := d.Get("type").(string)

	connector := &ilert.Connector{
		Name: name,
		Type: connectorType,
	}

	if val, ok := d.GetOk("datadog"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsDatadog{
				APIKey: v["api_key"].(string),
			}
		}
	}

	if val, ok := d.GetOk("jira"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsJira{
				URL:      v["url"].(string),
				Email:    v["email"].(string),
				Password: v["password"].(string),
			}
		}
	}

	if val, ok := d.GetOk("microsoft_teams"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsMicrosoftTeams{
				URL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("servicenow"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsServiceNow{
				URL:      v["url"].(string),
				Username: v["username"].(string),
				Password: v["password"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zendesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsZendesk{
				URL:    v["url"].(string),
				Email:  v["email"].(string),
				APIKey: v["api_key"].(string),
			}
		}
	}

	if val, ok := d.GetOk("discord"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsDiscord{
				URL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("github"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsGithub{
				APIKey: v["api_key"].(string),
			}
		}
	}

	if val, ok := d.GetOk("topdesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsTopdesk{
				URL:      v["url"].(string),
				Username: v["username"].(string),
				Password: v["password"].(string),
			}
		}
	}

	if val, ok := d.GetOk("aws_lambda"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsAWSLambda{
				Authorization: v["authorization"].(string),
			}
		}
	}

	if val, ok := d.GetOk("azure_faas"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsAzureFunction{
				Authorization: v["authorization"].(string),
			}
		}
	}

	if val, ok := d.GetOk("google_faas"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsGoogleFunction{
				Authorization: v["authorization"].(string),
			}
		}
	}

	if val, ok := d.GetOk("sysdig"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsSysdig{
				APIKey: v["api_key"].(string),
			}
		}
	}

	return connector, nil
}

func resourceConnectorCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connector, err := buildConnector(d)
	if err != nil {
		log.Printf("[ERROR] Building connector error %s", err.Error())
		return err
	}

	log.Printf("[INFO] Creating connector %s", connector.Name)

	result, err := client.CreateConnector(&ilert.CreateConnectorInput{Connector: connector})
	if err != nil {
		log.Printf("[ERROR] Creating iLert connector error %s", err.Error())
		return err
	}

	d.SetId(result.Connector.ID)

	return resourceConnectorRead(d, m)
}

func resourceConnectorRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Reading connector: %s", d.Id())
	result, err := client.GetConnector(&ilert.GetConnectorInput{ConnectorID: ilert.String(connectorID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			log.Printf("[WARN] Removing connector %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an connector with ID %s", d.Id())
	}

	d.Set("name", result.Connector.Name)
	d.Set("type", result.Connector.Type)
	d.Set("created_at", result.Connector.CreatedAt)
	d.Set("updated_at", result.Connector.UpdatedAt)

	switch result.Connector.Type {
	case ilert.ConnectorTypes.Datadog:
		d.Set("datadog", []interface{}{
			map[string]interface{}{
				"api_key": result.Connector.Params.APIKey,
			},
		})
	case ilert.ConnectorTypes.Jira:
		d.Set("jira", []interface{}{
			map[string]interface{}{
				"url":      result.Connector.Params.URL,
				"email":    result.Connector.Params.Email,
				"password": result.Connector.Params.Password,
			},
		})
	case ilert.ConnectorTypes.MicrosoftTeams:
		d.Set("microsoft_teams", []interface{}{
			map[string]interface{}{
				"url": result.Connector.Params.URL,
			},
		})
	case ilert.ConnectorTypes.ServiceNow:
		d.Set("servicenow", []interface{}{
			map[string]interface{}{
				"url":      result.Connector.Params.URL,
				"username": result.Connector.Params.Username,
				"password": result.Connector.Params.Password,
			},
		})
	case ilert.ConnectorTypes.Zendesk:
		d.Set("zendesk", []interface{}{
			map[string]interface{}{
				"url":     result.Connector.Params.URL,
				"email":   result.Connector.Params.Email,
				"api_key": result.Connector.Params.APIKey,
			},
		})
	case ilert.ConnectorTypes.Discord:
		d.Set("discord", []interface{}{
			map[string]interface{}{
				"url": result.Connector.Params.URL,
			},
		})
	case ilert.ConnectorTypes.Github:
		d.Set("github", []interface{}{
			map[string]interface{}{
				"api_key": result.Connector.Params.APIKey,
			},
		})
	case ilert.ConnectorTypes.Topdesk:
		d.Set("topdesk", []interface{}{
			map[string]interface{}{
				"url":      result.Connector.Params.URL,
				"username": result.Connector.Params.Username,
				"password": result.Connector.Params.Password,
			},
		})
	case ilert.ConnectorTypes.AWSLambda,
		ilert.ConnectorTypes.AzureFAAS,
		ilert.ConnectorTypes.GoogleFAAS:
		d.Set("aws_lambda", []interface{}{
			map[string]interface{}{
				"authorization": result.Connector.Params.Authorization,
			},
		})
	case ilert.ConnectorTypes.Sysdig:
		d.Set("sysdig", []interface{}{
			map[string]interface{}{
				"api_key": result.Connector.Params.APIKey,
			},
		})
	}

	return nil
}

func resourceConnectorUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connector, err := buildConnector(d)
	if err != nil {
		log.Printf("[ERROR] Building connector error %s", err.Error())
		return err
	}

	connectorID := d.Id()
	log.Printf("[DEBUG] Updating connector: %s", d.Id())
	_, err = client.UpdateConnector(&ilert.UpdateConnectorInput{Connector: connector, ConnectorID: ilert.String(connectorID)})
	if err != nil {
		log.Printf("[ERROR] Updating iLert connector error %s", err.Error())
		return err
	}
	return resourceConnectorRead(d, m)
}

func resourceConnectorDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Deleting connector: %s", d.Id())
	_, err := client.DeleteConnector(&ilert.DeleteConnectorInput{ConnectorID: ilert.String(connectorID)})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceConnectorExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Reading connector: %s", d.Id())
	_, err := client.GetConnector(&ilert.GetConnectorInput{ConnectorID: ilert.String(connectorID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func flattenConnectorAlertSourceIDList(list []int64) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["id"] = strconv.FormatInt(item, 10)
		results = append(results, result)
	}

	return results, nil
}
