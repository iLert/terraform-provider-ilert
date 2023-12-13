package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v3"
)

func resourceConnector() *schema.Resource {
	// include only type that schema supports
	connectorTypesAll := removeStringsFromSlice(ilert.ConnectorTypesAll, ilert.ConnectorTypes.Email, ilert.ConnectorTypes.MicrosoftTeams, ilert.ConnectorTypes.MicrosoftTeamsBot, ilert.ConnectorTypes.ZoomChat, ilert.ConnectorTypes.ZoomMeeting, ilert.ConnectorTypes.Webex, ilert.ConnectorTypes.Slack, ilert.ConnectorTypes.Webhook, ilert.ConnectorTypes.Zapier, ilert.ConnectorTypes.DingTalkAction, ilert.ConnectorTypes.AutomationRule)
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
				ValidateFunc: validation.StringInSlice(ilert.ConnectorTypesAll, false),
			},
			"datadog": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Datadog),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Jira),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.MicrosoftTeams),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.ServiceNow),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Zendesk),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Discord),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Github),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Topdesk),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.AWSLambda),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.AzureFAAS),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.GoogleFAAS),
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
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Sysdig),
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
			"autotask": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Autotask),
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
			"mattermost": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Mattermost),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"zammad": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.Zammad),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
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
			"status_page_io": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.StatusPageIO),
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
			"dingtalk": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				ConflictsWith: removeStringsFromSlice(connectorTypesAll, ilert.ConnectorTypes.DingTalk),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret": {
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
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
		Exists:        resourceConnectorExists,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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

	if val, ok := d.GetOk("autotask"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsAutotask{
				URL:      v["url"].(string),
				Email:    v["email"].(string),
				Password: v["password"].(string),
			}
		}
	}

	if val, ok := d.GetOk("mattermost"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsMattermost{
				URL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zammad"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsZammad{
				URL:    v["url"].(string),
				APIKey: v["api_key"].(string),
			}
		}
	}

	if val, ok := d.GetOk("status_page_io"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsStatusPageIO{
				APIKey: v["api_key"].(string),
			}
		}
	}

	if val, ok := d.GetOk("dingtalk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connector.Params = &ilert.ConnectorParamsDingTalk{
				URL:    v["url"].(string),
				Secret: v["secret"].(string),
			}
		}
	}

	return connector, nil
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connector, err := buildConnector(d)
	if err != nil {
		log.Printf("[ERROR] Building connector error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating connector %s", connector.Name)

	result := &ilert.CreateConnectorOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateConnector(&ilert.CreateConnectorInput{Connector: connector})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert connector error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert connector error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Connector == nil {
		log.Printf("[ERROR] Creating ilert connector error: empty response")
		return diag.Errorf("connector response is empty")
	}

	d.SetId(result.Connector.ID)

	return resourceConnectorRead(ctx, d, m)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Reading connector: %s", d.Id())

	result := &ilert.GetConnectorOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetConnector(&ilert.GetConnectorInput{ConnectorID: ilert.String(connectorID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing connector %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an connector with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert connector error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Connector == nil {
		log.Printf("[ERROR] Reading ilert connector error: empty response")
		return diag.Errorf("connector response is empty")
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
	case ilert.ConnectorTypes.Autotask:
		d.Set("autotask", []interface{}{
			map[string]interface{}{
				"url":      result.Connector.Params.URL,
				"email":    result.Connector.Params.Email,
				"password": result.Connector.Params.Password,
			},
		})
	case ilert.ConnectorTypes.Mattermost:
		d.Set("mattermost", []interface{}{
			map[string]interface{}{
				"url": result.Connector.Params.URL,
			},
		})
	case ilert.ConnectorTypes.Zammad:
		d.Set("zammad", []interface{}{
			map[string]interface{}{
				"url":     result.Connector.Params.URL,
				"api_key": result.Connector.Params.APIKey,
			},
		})
	case ilert.ConnectorTypes.StatusPageIO:
		d.Set("status_page_io", []interface{}{
			map[string]interface{}{
				"api_key": result.Connector.Params.APIKey,
			},
		})
	case ilert.ConnectorTypes.DingTalk:
		d.Set("dingtalk", []interface{}{
			map[string]interface{}{
				"url":    result.Connector.Params.URL,
				"secret": result.Connector.Params.Secret,
			},
		})
	}

	return nil
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connector, err := buildConnector(d)
	if err != nil {
		log.Printf("[ERROR] Building connector error %s", err.Error())
		return diag.FromErr(err)
	}

	connectorID := d.Id()
	log.Printf("[DEBUG] Updating connector: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateConnector(&ilert.UpdateConnectorInput{Connector: connector, ConnectorID: ilert.String(connectorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an connector with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert connector error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceConnectorRead(ctx, d, m)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Deleting connector: %s", d.Id())
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.DeleteConnector(&ilert.DeleteConnectorInput{ConnectorID: ilert.String(connectorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an connector with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert connector error %s", err.Error())
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceConnectorExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	connectorID := d.Id()
	log.Printf("[DEBUG] Reading connector: %s", d.Id())
	ctx := context.Background()
	result := false
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetConnector(&ilert.GetConnectorInput{ConnectorID: ilert.String(connectorID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert connector error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a connector with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert connector error: %s", err.Error())
		return false, err
	}
	return result, nil
}
