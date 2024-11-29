package ilert

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v3"
)

func resourceDeploymentPipeline() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"integration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.DeploymentPipelineIntegrationTypeAll, false),
			},
			"integration_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"team": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
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
			"integration_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"github": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"event_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"gitlab": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"event_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		CreateContext: resourceDeploymentPipelineCreate,
		ReadContext:   resourceDeploymentPipelineRead,
		UpdateContext: resourceDeploymentPipelineUpdate,
		DeleteContext: resourceDeploymentPipelineDelete,
		Exists:        resourceDeploymentPipelineExists,
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

func buildDeploymentPipeline(d *schema.ResourceData) (*ilert.DeploymentPipeline, error) {
	name := d.Get("name").(string)
	integrationType := d.Get("integration_type").(string)

	deploymentPipeline := &ilert.DeploymentPipeline{
		Name:            name,
		IntegrationType: integrationType,
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]interface{})
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		deploymentPipeline.Teams = tms
	}

	if val, ok := d.GetOk("github"); ok && integrationType == ilert.DeploymentPipelineIntegrationType.GitHub {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.DeploymentPipelineGitHubParams{}
			if vL, ok := v["branch_filter"].([]interface{}); ok && len(vL) > 0 {
				sL := make([]string, 0)
				for _, m := range vL {
					if v, ok := m.(string); ok && v != "" {
						sL = append(sL, v)
					}
				}
				params.BranchFilters = sL
			}
			if vL, ok := v["event_filter"].([]interface{}); ok && len(vL) > 0 {
				sL := make([]string, 0)
				for _, m := range vL {
					if v, ok := m.(string); ok && v != "" {
						sL = append(sL, v)
					}
				}
				params.EventFilters = sL
			}
			deploymentPipeline.Params = params
		}
	}

	if val, ok := d.GetOk("gitlab"); ok && integrationType == ilert.DeploymentPipelineIntegrationType.GitLab {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.DeploymentPipelineGitLabParams{}
			if vL, ok := v["branch_filter"].([]interface{}); ok && len(vL) > 0 {
				sL := make([]string, 0)
				for _, m := range vL {
					if v, ok := m.(string); ok && v != "" {
						sL = append(sL, v)
					}
				}
				params.BranchFilters = sL
			}
			if vL, ok := v["event_filter"].([]interface{}); ok && len(vL) > 0 {
				sL := make([]string, 0)
				for _, m := range vL {
					if v, ok := m.(string); ok && v != "" {
						sL = append(sL, v)
					}
				}
				params.EventFilters = sL
			}
			deploymentPipeline.Params = params
		}
	}

	return deploymentPipeline, nil
}

func resourceDeploymentPipelineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	deploymentPipeline, err := buildDeploymentPipeline(d)
	if err != nil {
		log.Printf("[ERROR] Building deployment pipeline error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating deployment pipeline %s", deploymentPipeline.Name)

	result := &ilert.CreateDeploymentPipelineOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.CreateDeploymentPipeline(&ilert.CreateDeploymentPipelineInput{DeploymentPipeline: deploymentPipeline, Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert deployment pipeline error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for deployment pipeline to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a deployment pipeline with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert deployment pipeline error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.DeploymentPipeline == nil {
		log.Printf("[ERROR] Creating ilert deployment pipeline error: empty response")
		return diag.Errorf("deployment pipeline response is empty")
	}

	d.SetId(strconv.FormatInt(result.DeploymentPipeline.ID, 10))

	return resourceDeploymentPipelineRead(ctx, d, m)
}

func resourceDeploymentPipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	deploymentPipelineID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse deployment pipeline id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading deployment pipeline: %s", d.Id())
	result := &ilert.GetDeploymentPipelineOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.GetDeploymentPipeline(&ilert.GetDeploymentPipelineInput{DeploymentPipelineID: ilert.Int64(deploymentPipelineID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing deployment pipeline %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for deployment pipeline with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an deployment pipeline with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert deployment pipeline error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.DeploymentPipeline == nil {
		log.Printf("[ERROR] Reading ilert deployment pipeline error: empty response")
		return diag.Errorf("deployment pipeline response is empty")
	}

	d.Set("name", result.DeploymentPipeline.Name)
	d.Set("integration_type", result.DeploymentPipeline.IntegrationType)
	d.Set("integration_key", result.DeploymentPipeline.IntegrationKey)

	teams, err := flattenTeamShortList(result.DeploymentPipeline.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	d.Set("created_at", result.DeploymentPipeline.CreatedAt)
	d.Set("updated_at", result.DeploymentPipeline.UpdatedAt)
	d.Set("integration_url", result.DeploymentPipeline.IntegrationUrl)

	if result.DeploymentPipeline.IntegrationType == ilert.DeploymentPipelineIntegrationType.GitHub {
		d.Set("github", []interface{}{
			map[string]interface{}{
				"branch_filter": result.DeploymentPipeline.Params.BranchFilters,
				"event_filter":  result.DeploymentPipeline.Params.EventFilters,
			},
		})
	}

	if result.DeploymentPipeline.IntegrationType == ilert.DeploymentPipelineIntegrationType.GitLab {
		d.Set("gitlab", []interface{}{
			map[string]interface{}{
				"branch_filter": result.DeploymentPipeline.Params.BranchFilters,
				"event_filter":  result.DeploymentPipeline.Params.EventFilters,
			},
		})
	}

	return nil
}

func resourceDeploymentPipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	deploymentPipeline, err := buildDeploymentPipeline(d)
	if err != nil {
		log.Printf("[ERROR] Building deployment pipeline error %s", err.Error())
		return diag.FromErr(err)
	}

	// API expects integration key to be always set, even if not allowed to be set by user
	deploymentPipeline.IntegrationKey = d.Get("integration_key").(string)

	deploymentPipelineID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse deployment pipeline id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating deployment pipeline: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateDeploymentPipeline(&ilert.UpdateDeploymentPipelineInput{DeploymentPipeline: deploymentPipeline, DeploymentPipelineID: ilert.Int64(deploymentPipelineID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for deployment pipeline with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an deployment pipeline with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert deployment pipeline error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceDeploymentPipelineRead(ctx, d, m)
}

func resourceDeploymentPipelineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	deploymentPipelineID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse deployment pipeline id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting deployment pipeline: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteDeploymentPipeline(&ilert.DeleteDeploymentPipelineInput{DeploymentPipelineID: ilert.Int64(deploymentPipelineID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for deployment pipeline with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an deployment pipeline with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert deployment pipeline error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceDeploymentPipelineExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	deploymentPipelineID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse deployment pipeline id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading deployment pipeline: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetDeploymentPipeline(&ilert.GetDeploymentPipelineInput{DeploymentPipelineID: ilert.Int64(deploymentPipelineID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert deployment pipeline error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for deployment pipeline to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a deployment pipeline with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert deployment pipeline error: %s", err.Error())
		return false, err
	}
	return result, nil
}
