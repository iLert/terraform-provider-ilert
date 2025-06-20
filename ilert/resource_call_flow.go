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

func resourceCallFlow() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.CallFlowLanguageAll, false),
			},
			"assigned_number": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"phone_number": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"number": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"team": {
				Type:     schema.TypeList,
				Optional: true,
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
			"root_node": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     resourceCallFlowNode(),
			},
		},
		CreateContext: resourceCallFlowCreate,
		ReadContext:   resourceCallFlowRead,
		UpdateContext: resourceCallFlowUpdate,
		DeleteContext: resourceCallFlowDelete,
		Exists:        resourceCallFlowExists,
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

func resourceCallFlowNode() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"node_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeTypeAll, false),
			},
			"metadata": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text_message": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_audio_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ai_voice_model": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataAIVoiceModelAll, false),
						},
						"enabled_options": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"language": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataLanguageAll, false),
						},
						"var_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"var_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"codes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"label": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"support_hours_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"hold_audio_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"targets": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataCallTargetTypeAll, false),
									},
								},
							},
						},
						"call_style": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataCallStyleAll, false),
						},
						"alert_source_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"call_timeout_sec": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"blacklist": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"intents": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:          schema.TypeString,
										Optional:      true,
										ValidateFunc:  validation.StringInSlice(ilert.CallFlowNodeMetadataIntentTypeAll, false),
										ConflictsWith: []string{"intents.label", "intents.description", "intents.examples"},
									},
									"label": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"intents.type"},
									},
									"description": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"intents.type"},
									},
									"examples": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"intents.type"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"gathers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataGatherTypeAll, false),
									},
									"label": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"var_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataGatherVarTypeAll, false),
									},
									"required": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"question": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"enrichment": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"information_types": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataEnrichmentInformationTypeAll, false),
										},
									},
									"sources": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataEnrichmentSourceTypeAll, false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceCallFlowBranch() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"branch_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.CallFlowBranchTypeAll, false),
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceCallFlowNode(),
			},
		},
	}
}

func buildCallFlow(d *schema.ResourceData) (*ilert.CallFlow, error) {
	name := d.Get("name").(string)
	integrationType := d.Get("integration_type").(string)

	CallFlow := &ilert.CallFlow{
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
		CallFlow.Teams = tms
	}

	if val, ok := d.GetOk("github"); ok && integrationType == ilert.CallFlowIntegrationType.GitHub {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.CallFlowGitHubParams{}
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
			CallFlow.Params = params
		}
	}

	if val, ok := d.GetOk("gitlab"); ok && integrationType == ilert.CallFlowIntegrationType.GitLab {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.CallFlowGitLabParams{}
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
			CallFlow.Params = params
		}
	}

	return CallFlow, nil
}

func resourceCallFlowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	CallFlow, err := buildCallFlow(d)
	if err != nil {
		log.Printf("[ERROR] Building Call Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating Call Flow %s", CallFlow.Name)

	result := &ilert.CreateCallFlowOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.CreateCallFlow(&ilert.CreateCallFlowInput{CallFlow: CallFlow, Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert Call Flow error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Call Flow to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a Call Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert Call Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.CallFlow == nil {
		log.Printf("[ERROR] Creating ilert Call Flow error: empty response")
		return diag.Errorf("Call Flow response is empty")
	}

	d.SetId(strconv.FormatInt(result.CallFlow.ID, 10))

	return resourceCallFlowRead(ctx, d, m)
}

func resourceCallFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	CallFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Call Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading Call Flow: %s", d.Id())
	result := &ilert.GetCallFlowOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.GetCallFlow(&ilert.GetCallFlowInput{CallFlowID: ilert.Int64(CallFlowID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing Call Flow %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Call Flow with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an Call Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert Call Flow error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.CallFlow == nil {
		log.Printf("[ERROR] Reading ilert Call Flow error: empty response")
		return diag.Errorf("Call Flow response is empty")
	}

	d.Set("name", result.CallFlow.Name)
	d.Set("integration_type", result.CallFlow.IntegrationType)
	d.Set("integration_key", result.CallFlow.IntegrationKey)

	teams, err := flattenTeamShortList(result.CallFlow.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	d.Set("created_at", result.CallFlow.CreatedAt)
	d.Set("updated_at", result.CallFlow.UpdatedAt)
	d.Set("integration_url", result.CallFlow.IntegrationUrl)

	if result.CallFlow.IntegrationType == ilert.CallFlowIntegrationType.GitHub {
		d.Set("github", []interface{}{
			map[string]interface{}{
				"branch_filter": result.CallFlow.Params.BranchFilters,
				"event_filter":  result.CallFlow.Params.EventFilters,
			},
		})
	}

	if result.CallFlow.IntegrationType == ilert.CallFlowIntegrationType.GitLab {
		d.Set("gitlab", []interface{}{
			map[string]interface{}{
				"branch_filter": result.CallFlow.Params.BranchFilters,
				"event_filter":  result.CallFlow.Params.EventFilters,
			},
		})
	}

	return nil
}

func resourceCallFlowUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	CallFlow, err := buildCallFlow(d)
	if err != nil {
		log.Printf("[ERROR] Building Call Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	// API expects integration key to be always set, even if not allowed to be set by user
	CallFlow.IntegrationKey = d.Get("integration_key").(string)

	CallFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Call Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating Call Flow: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateCallFlow(&ilert.UpdateCallFlowInput{CallFlow: CallFlow, CallFlowID: ilert.Int64(CallFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Call Flow with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an Call Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert Call Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceCallFlowRead(ctx, d, m)
}

func resourceCallFlowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	CallFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Call Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting Call Flow: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteCallFlow(&ilert.DeleteCallFlowInput{CallFlowID: ilert.Int64(CallFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Call Flow with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an Call Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert Call Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceCallFlowExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	CallFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Call Flow id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading Call Flow: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetCallFlow(&ilert.GetCallFlowInput{CallFlowID: ilert.Int64(CallFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert Call Flow error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Call Flow to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a Call Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert Call Flow error: %s", err.Error())
		return false, err
	}
	return result, nil
}
