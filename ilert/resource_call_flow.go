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

var callFlowDepth = 10

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
				Elem:     resourceCallFlowRoot(callFlowDepth),
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

func resourceCallFlowRoot(depth int) *schema.Resource {
	if depth <= 0 {
		return resourceCallFlowNodeNoBranches()
	}

	return resourceCallFlowNode(depth - 1)
}

func resourceCallFlowNode(depth int) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
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
				Elem:     resourceCallFlowNodeMetadata(),
			},
			"branches": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceCallFlowBranch(depth),
			},
		},
	}
}

func resourceCallFlowNodeNoBranches() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
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
				Elem:     resourceCallFlowNodeMetadata(),
			},
		},
	}
}

func resourceCallFlowNodeMetadata() *schema.Resource {
	return &schema.Resource{
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ilert.CallFlowNodeMetadataIntentTypeAll, false),
						},
						"label": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"examples": {
							Type:     schema.TypeList,
							Optional: true,
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
							Type:     schema.TypeList,
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
	}
}

func resourceCallFlowBranch(depth int) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
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
				Optional: true,
				MaxItems: 1,
				Elem:     resourceCallFlowRoot(depth),
			},
		},
	}
}

func buildCallFlow(d *schema.ResourceData) (*ilert.CallFlow, error) {
	name := d.Get("name").(string)
	language := d.Get("language").(string)

	callFlow := &ilert.CallFlow{
		Name:     name,
		Language: language,
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
		callFlow.Teams = tms
	}

	if val, ok := d.GetOk("root_node"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 && vL[0] != nil {
			rn := vL[0].(map[string]interface{})
			node, err := buildCallFlowNodeFromMap(rn)
			if err != nil {
				return nil, err
			}
			callFlow.RootNode = node
		}
	}

	return callFlow, nil
}

func buildCallFlowNodeFromMap(rn map[string]interface{}) (*ilert.CallFlowNode, error) {
	node := &ilert.CallFlowNode{}
	if v, ok := rn["id"].(int); ok && v > 0 {
		node.ID = int64(v)
	}
	if s, ok := rn["node_type"].(string); ok && s != "" {
		node.NodeType = s
	}
	if s, ok := rn["name"].(string); ok && s != "" {
		node.Name = s
	}
	if mvL, ok := rn["metadata"].([]interface{}); ok && len(mvL) > 0 && mvL[0] != nil {
		mv := mvL[0].(map[string]interface{})
		md := &ilert.CallFlowNodeMetadata{}
		if s, ok := mv["text_message"].(string); ok && s != "" {
			md.TextMessage = s
		}
		if s, ok := mv["custom_audio_url"].(string); ok && s != "" {
			md.CustomAudioUrl = s
		}
		if s, ok := mv["ai_voice_model"].(string); ok && s != "" {
			md.AIVoiceModel = s
		}
		if v, ok := mv["enabled_options"].([]interface{}); ok && len(v) > 0 {
			sL := make([]string, 0, len(v))
			for _, it := range v {
				if s, ok := it.(string); ok && s != "" {
					sL = append(sL, s)
				}
			}
			md.EnabledOptions = sL
		}
		if s, ok := mv["language"].(string); ok && s != "" {
			md.Language = s
		}
		if s, ok := mv["var_key"].(string); ok && s != "" {
			md.VarKey = s
		}
		if s, ok := mv["var_value"].(string); ok && s != "" {
			md.VarValue = s
		}
		if v, ok := mv["codes"].([]interface{}); ok && len(v) > 0 {
			codes := make([]ilert.CallFlowNodeMetadataCode, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				cv := it.(map[string]interface{})
				code := ilert.CallFlowNodeMetadataCode{}
				if lbl, ok := cv["label"].(string); ok && lbl != "" {
					code.Label = lbl
				}
				if c, ok := cv["code"].(int); ok && c != 0 {
					code.Code = int64(c)
				}
				codes = append(codes, code)
			}
			md.Codes = codes
		}
		if v, ok := mv["support_hours_id"].(int); ok && v > 0 {
			md.SupportHoursId = int64(v)
		}
		if s, ok := mv["hold_audio_url"].(string); ok && s != "" {
			md.HoldAudioUrl = s
		}
		if v, ok := mv["targets"].([]interface{}); ok && len(v) > 0 {
			targets := make([]ilert.CallFlowNodeMetadataCallTarget, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				tv := it.(map[string]interface{})
				t := ilert.CallFlowNodeMetadataCallTarget{}
				if s, ok := tv["target"].(string); ok && s != "" {
					t.Target = s
				}
				if s, ok := tv["type"].(string); ok && s != "" {
					t.Type = s
				}
				targets = append(targets, t)
			}
			md.Targets = targets
		}
		if s, ok := mv["call_style"].(string); ok && s != "" {
			md.CallStyle = s
		}
		if v, ok := mv["alert_source_id"].(int); ok && v > 0 {
			md.AlertSourceId = int64(v)
		}
		if v, ok := mv["retries"].(int); ok && v != 0 {
			md.Retries = int64(v)
		}
		if v, ok := mv["call_timeout_sec"].(int); ok && v != 0 {
			md.CallTimeoutSec = int64(v)
		}
		if v, ok := mv["blacklist"].([]interface{}); ok && len(v) > 0 {
			sL := make([]string, 0, len(v))
			for _, it := range v {
				if s, ok := it.(string); ok && s != "" {
					sL = append(sL, s)
				}
			}
			md.Blacklist = sL
		}
		if v, ok := mv["intents"].([]interface{}); ok && len(v) > 0 {
			intents := make([]ilert.CallFlowNodeMetadataIntent, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				iv := it.(map[string]interface{})
				in := ilert.CallFlowNodeMetadataIntent{}
				if s, ok := iv["type"].(string); ok && s != "" {
					in.Type = s
				}
				if s, ok := iv["label"].(string); ok && s != "" {
					in.Label = s
				}
				if s, ok := iv["description"].(string); ok && s != "" {
					in.Description = s
				}
				if arr, ok := iv["examples"].([]interface{}); ok && len(arr) > 0 {
					ex := make([]string, 0, len(arr))
					for _, e := range arr {
						if s, ok := e.(string); ok && s != "" {
							ex = append(ex, s)
						}
					}
					in.Examples = ex
				}
				intents = append(intents, in)
			}
			md.Intents = intents
		}
		if v, ok := mv["gathers"].([]interface{}); ok && len(v) > 0 {
			gathers := make([]ilert.CallFlowNodeMetadataGather, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				gv := it.(map[string]interface{})
				g := ilert.CallFlowNodeMetadataGather{}
				if s, ok := gv["type"].(string); ok && s != "" {
					g.Type = s
				}
				if s, ok := gv["label"].(string); ok && s != "" {
					g.Label = s
				}
				if s, ok := gv["var_type"].(string); ok && s != "" {
					g.VarType = s
				}
				if b, ok := gv["required"].(bool); ok {
					g.Required = b
				}
				if s, ok := gv["question"].(string); ok && s != "" {
					g.Question = s
				}
				gathers = append(gathers, g)
			}
			md.Gathers = gathers
		}
		if v, ok := mv["enrichment"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			ev := v[0].(map[string]interface{})
			enr := &ilert.CallFlowNodeMetadataEnrichment{}
			if b, ok := ev["enabled"].(bool); ok {
				enr.Enabled = b
			}
			if m, ok := ev["information_types"].(map[string]interface{}); ok && len(m) > 0 {
				infos := make([]string, 0, len(m))
				for _, val := range m {
					if s, ok := val.(string); ok && s != "" {
						infos = append(infos, s)
					}
				}
				enr.InformationTypes = infos
			}
			if m, ok := ev["sources"].(map[string]interface{}); ok && len(m) > 0 {
				srcs := make([]ilert.CallFlowNodeMetadataEnrichmentSource, 0, len(m))
				for _, val := range m {
					if val == nil {
						continue
					}
					sv := val.(map[string]interface{})
					src := ilert.CallFlowNodeMetadataEnrichmentSource{}
					if id, ok := sv["id"].(int); ok && id > 0 {
						src.ID = int64(id)
					}
					if t, ok := sv["type"].(string); ok && t != "" {
						src.Type = t
					}
					srcs = append(srcs, src)
				}
				enr.Sources = srcs
			}
			md.Enrichment = enr
		}
		node.Metadata = md
	}

	if br, ok := rn["branches"].([]interface{}); ok && len(br) > 0 {
		branches := make([]ilert.CallFlowBranch, 0, len(br))
		for _, be := range br {
			if be == nil {
				continue
			}
			bv := be.(map[string]interface{})
			b := ilert.CallFlowBranch{}
			if v, ok := bv["id"].(int); ok && v > 0 {
				b.ID = int64(v)
			}
			if s, ok := bv["branch_type"].(string); ok && s != "" {
				b.BranchType = s
			}
			if s, ok := bv["condition"].(string); ok && s != "" {
				b.Condition = s
			}
			if tvL, ok := bv["target"].([]interface{}); ok && len(tvL) > 0 && tvL[0] != nil {
				tv := tvL[0].(map[string]interface{})
				tn, err := buildCallFlowNodeFromMap(tv)
				if err != nil {
					return nil, err
				}
				b.Target = tn
			}
			branches = append(branches, b)
		}
		node.Branches = branches
	}
	return node, nil
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
		r, err := client.CreateCallFlow(&ilert.CreateCallFlowInput{CallFlow: CallFlow})
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
		r, err := client.GetCallFlow(&ilert.GetCallFlowInput{CallFlowID: ilert.Int64(CallFlowID)})
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
	d.Set("language", result.CallFlow.Language)

	teams, err := flattenTeamShortList(result.CallFlow.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	if result.CallFlow.AssignedNumber != nil {
		assigned := make(map[string]interface{})
		assigned["id"] = result.CallFlow.AssignedNumber.ID
		assigned["name"] = result.CallFlow.AssignedNumber.Name
		if result.CallFlow.AssignedNumber.PhoneNumber != nil {
			assigned["phone_number"] = []interface{}{
				map[string]interface{}{
					"region_code": result.CallFlow.AssignedNumber.PhoneNumber.RegionCode,
					"number":      result.CallFlow.AssignedNumber.PhoneNumber.Number,
				},
			}
		} else {
			assigned["phone_number"] = []interface{}{}
		}
		d.Set("assigned_number", []interface{}{assigned})
	} else {
		d.Set("assigned_number", []interface{}{})
	}

	rn, err := flattenCallFlowNodeOutput(result.CallFlow.RootNode)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("root_node", rn); err != nil {
		return diag.Errorf("error setting root_node: %s", err)
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

func flattenCallFlowNodeOutput(node *ilert.CallFlowNodeOutput) ([]interface{}, error) {
	if node == nil {
		return make([]interface{}, 0), nil
	}

	result := make(map[string]interface{})
	if node.ID != 0 {
		result["id"] = int(node.ID)
	}
	result["node_type"] = node.NodeType

	if node.Name != "" {
		result["name"] = node.Name
	}

	mds, err := flattenCallFlowNodeMetadata(node.Metadata)
	if err != nil {
		return nil, err
	}
	if len(mds) > 0 {
		result["metadata"] = mds
	} else {
		result["metadata"] = []interface{}{}
	}

	if len(node.Branches) > 0 {
		branches := make([]interface{}, 0, len(node.Branches))
		for _, b := range node.Branches {
			bm := make(map[string]interface{})
			if b.ID != 0 {
				bm["id"] = int(b.ID)
			}
			if b.BranchType != "" {
				bm["branch_type"] = b.BranchType
			}
			if b.Condition != "" {
				bm["condition"] = b.Condition
			}
			tn, err := flattenCallFlowNode(&b.Target)
			if err != nil {
				return nil, err
			}
			bm["target"] = tn
			branches = append(branches, bm)
		}
		result["branches"] = branches
	}

	return []interface{}{result}, nil
}

func flattenCallFlowNode(node **ilert.CallFlowNode) ([]interface{}, error) {
	if node == nil || *node == nil {
		return []interface{}{}, nil
	}
	n := *node
	result := make(map[string]interface{})
	if n.ID != 0 {
		result["id"] = int(n.ID)
	}
	result["node_type"] = n.NodeType
	if n.Name != "" {
		result["name"] = n.Name
	}
	if n.Metadata != nil {
		if mdMap, ok := n.Metadata.(map[string]interface{}); ok {

			md := &ilert.CallFlowNodeMetadata{}

			if v, ok := mdMap["textMessage"].(string); ok && v != "" {
				md.TextMessage = v
			}
			if v, ok := mdMap["customAudioUrl"].(string); ok && v != "" {
				md.CustomAudioUrl = v
			}
			if v, ok := mdMap["aiVoiceModel"].(string); ok && v != "" {
				md.AIVoiceModel = v
			}
			if v, ok := mdMap["language"].(string); ok && v != "" {
				md.Language = v
			}
			if v, ok := mdMap["varKey"].(string); ok && v != "" {
				md.VarKey = v
			}
			if v, ok := mdMap["varValue"].(string); ok && v != "" {
				md.VarValue = v
			}
			if v, ok := mdMap["holdAudioUrl"].(string); ok && v != "" {
				md.HoldAudioUrl = v
			}
			if v, ok := mdMap["callStyle"].(string); ok && v != "" {
				md.CallStyle = v
			}
			if v, ok := mdMap["supportHoursId"].(float64); ok && v > 0 {
				md.SupportHoursId = int64(v)
			}
			if v, ok := mdMap["alertSourceId"].(float64); ok && v > 0 {
				md.AlertSourceId = int64(v)
			}
			if v, ok := mdMap["retries"].(float64); ok && v != 0 {
				md.Retries = int64(v)
			}
			if v, ok := mdMap["callTimeoutSec"].(float64); ok && v != 0 {
				md.CallTimeoutSec = int64(v)
			}

			if v, ok := mdMap["codes"].([]interface{}); ok && len(v) > 0 {
				codes := make([]ilert.CallFlowNodeMetadataCode, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					cv := it.(map[string]interface{})
					code := ilert.CallFlowNodeMetadataCode{}
					if lbl, ok := cv["label"].(string); ok && lbl != "" {
						code.Label = lbl
					}
					if c, ok := cv["code"].(float64); ok && c != 0 {
						code.Code = int64(c)
					}
					codes = append(codes, code)
				}
				md.Codes = codes
			}

			if v, ok := mdMap["targets"].([]interface{}); ok && len(v) > 0 {
				targets := make([]ilert.CallFlowNodeMetadataCallTarget, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					tv := it.(map[string]interface{})
					t := ilert.CallFlowNodeMetadataCallTarget{}
					if s, ok := tv["target"].(string); ok && s != "" {
						t.Target = s
					}
					if s, ok := tv["type"].(string); ok && s != "" {
						t.Type = s
					}
					targets = append(targets, t)
				}
				md.Targets = targets
			}

			if v, ok := mdMap["intents"].([]interface{}); ok && len(v) > 0 {
				intents := make([]ilert.CallFlowNodeMetadataIntent, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					iv := it.(map[string]interface{})
					in := ilert.CallFlowNodeMetadataIntent{}
					if s, ok := iv["type"].(string); ok && s != "" {
						in.Type = s
					}
					if s, ok := iv["label"].(string); ok && s != "" {
						in.Label = s
					}
					if s, ok := iv["description"].(string); ok && s != "" {
						in.Description = s
					}
					if arr, ok := iv["examples"].([]interface{}); ok && len(arr) > 0 {
						ex := make([]string, 0, len(arr))
						for _, e := range arr {
							if s, ok := e.(string); ok && s != "" {
								ex = append(ex, s)
							}
						}
						in.Examples = ex
					}
					intents = append(intents, in)
				}
				md.Intents = intents
			}

			if v, ok := mdMap["gathers"].([]interface{}); ok && len(v) > 0 {
				gathers := make([]ilert.CallFlowNodeMetadataGather, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					gv := it.(map[string]interface{})
					g := ilert.CallFlowNodeMetadataGather{}
					if s, ok := gv["type"].(string); ok && s != "" {
						g.Type = s
					}
					if s, ok := gv["label"].(string); ok && s != "" {
						g.Label = s
					}
					if s, ok := gv["varType"].(string); ok && s != "" {
						g.VarType = s
					}
					if b, ok := gv["required"].(bool); ok {
						g.Required = b
					}
					if s, ok := gv["question"].(string); ok && s != "" {
						g.Question = s
					}
					gathers = append(gathers, g)
				}
				md.Gathers = gathers
			}

			if v, ok := mdMap["enrichment"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				ev := v[0].(map[string]interface{})
				enr := &ilert.CallFlowNodeMetadataEnrichment{}
				if b, ok := ev["enabled"].(bool); ok {
					enr.Enabled = b
				}
				if m, ok := ev["informationTypes"].(map[string]interface{}); ok && len(m) > 0 {
					infos := make([]string, 0, len(m))
					for _, val := range m {
						if s, ok := val.(string); ok && s != "" {
							infos = append(infos, s)
						}
					}
					enr.InformationTypes = infos
				}
				if m, ok := ev["sources"].(map[string]interface{}); ok && len(m) > 0 {
					srcs := make([]ilert.CallFlowNodeMetadataEnrichmentSource, 0, len(m))
					for _, val := range m {
						if val == nil {
							continue
						}
						sv := val.(map[string]interface{})
						src := ilert.CallFlowNodeMetadataEnrichmentSource{}
						if id, ok := sv["id"].(float64); ok && id > 0 {
							src.ID = int64(id)
						}
						if t, ok := sv["type"].(string); ok && t != "" {
							src.Type = t
						}
						srcs = append(srcs, src)
					}
					enr.Sources = srcs
				}
				md.Enrichment = enr
			}

			if v, ok := mdMap["enabledOptions"].([]interface{}); ok && len(v) > 0 {
				opts := make([]string, 0, len(v))
				for _, opt := range v {
					if s, ok := opt.(string); ok && s != "" {
						opts = append(opts, s)
					}
				}
				md.EnabledOptions = opts
			}
			if v, ok := mdMap["blacklist"].([]interface{}); ok && len(v) > 0 {
				bl := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok && s != "" {
						bl = append(bl, s)
					}
				}
				md.Blacklist = bl
			}

			mds, err := flattenCallFlowNodeMetadata(md)
			if err != nil {
				return nil, err
			}
			result["metadata"] = mds
		}
	}
	if len(n.Branches) > 0 {
		branches := make([]interface{}, 0, len(n.Branches))
		for _, b := range n.Branches {
			bm := make(map[string]interface{})
			if b.ID != 0 {
				bm["id"] = int(b.ID)
			}
			if b.BranchType != "" {
				bm["branch_type"] = b.BranchType
			}
			if b.Condition != "" {
				bm["condition"] = b.Condition
			}
			tn, err := flattenCallFlowNode(&b.Target)
			if err != nil {
				return nil, err
			}
			bm["target"] = tn
			branches = append(branches, bm)
		}
		result["branches"] = branches
	}
	return []interface{}{result}, nil
}

func flattenCallFlowNodeMetadata(md *ilert.CallFlowNodeMetadata) ([]interface{}, error) {
	if md == nil {
		return make([]interface{}, 0), nil
	}

	result := make(map[string]interface{})

	if md.TextMessage != "" {
		result["text_message"] = md.TextMessage
	}
	if md.CustomAudioUrl != "" {
		result["custom_audio_url"] = md.CustomAudioUrl
	}
	if md.AIVoiceModel != "" {
		result["ai_voice_model"] = md.AIVoiceModel
	}
	if len(md.EnabledOptions) > 0 {
		opts := make([]interface{}, 0, len(md.EnabledOptions))
		for _, s := range md.EnabledOptions {
			if s != "" {
				opts = append(opts, s)
			}
		}
		result["enabled_options"] = opts
	}
	if md.Language != "" {
		result["language"] = md.Language
	}
	if md.VarKey != "" {
		result["var_key"] = md.VarKey
	}
	if md.VarValue != "" {
		result["var_value"] = md.VarValue
	}
	if len(md.Codes) > 0 {
		codes := make([]interface{}, 0, len(md.Codes))
		for _, c := range md.Codes {
			m := make(map[string]interface{})
			if c.Code != 0 {
				m["code"] = int(c.Code)
			}
			if c.Label != "" {
				m["label"] = c.Label
			}
			codes = append(codes, m)
		}
		result["codes"] = codes
	}
	if md.SupportHoursId != 0 {
		result["support_hours_id"] = int(md.SupportHoursId)
	}
	if md.HoldAudioUrl != "" {
		result["hold_audio_url"] = md.HoldAudioUrl
	}
	if len(md.Targets) > 0 {
		targets := make([]interface{}, 0, len(md.Targets))
		for _, t := range md.Targets {
			m := make(map[string]interface{})
			if t.Target != "" {
				m["target"] = t.Target
			}
			if t.Type != "" {
				m["type"] = t.Type
			}
			targets = append(targets, m)
		}
		result["targets"] = targets
	}
	if md.CallStyle != "" {
		result["call_style"] = md.CallStyle
	}
	if md.AlertSourceId != 0 {
		result["alert_source_id"] = int(md.AlertSourceId)
	}
	if md.Retries != 0 {
		result["retries"] = int(md.Retries)
	}
	if md.CallTimeoutSec != 0 {
		result["call_timeout_sec"] = int(md.CallTimeoutSec)
	}
	if len(md.Blacklist) > 0 {
		bl := make([]interface{}, 0, len(md.Blacklist))
		for _, s := range md.Blacklist {
			if s != "" {
				bl = append(bl, s)
			}
		}
		result["blacklist"] = bl
	}
	if len(md.Intents) > 0 {
		intents := make([]interface{}, 0, len(md.Intents))
		for _, in := range md.Intents {
			m := make(map[string]interface{})
			if in.Type != "" {
				m["type"] = in.Type
			}
			if in.Label != "" {
				m["label"] = in.Label
			}
			if in.Description != "" {
				m["description"] = in.Description
			}
			if len(in.Examples) > 0 {
				ex := make([]interface{}, 0, len(in.Examples))
				for _, s := range in.Examples {
					if s != "" {
						ex = append(ex, s)
					}
				}
				m["examples"] = ex
			}
			intents = append(intents, m)
		}
		result["intents"] = intents
	}
	if len(md.Gathers) > 0 {
		gathers := make([]interface{}, 0, len(md.Gathers))
		for _, g := range md.Gathers {
			m := make(map[string]interface{})
			if g.Type != "" {
				m["type"] = g.Type
			}
			if g.Label != "" {
				m["label"] = g.Label
			}
			if g.VarType != "" {
				m["var_type"] = g.VarType
			}
			if g.Required {
				m["required"] = g.Required
			}
			if g.Question != "" {
				m["question"] = g.Question
			}
			gathers = append(gathers, m)
		}
		result["gathers"] = gathers
	}
	if md.Enrichment != nil {
		em := make(map[string]interface{})
		em["enabled"] = md.Enrichment.Enabled
		if len(md.Enrichment.InformationTypes) > 0 {
			itMap := make(map[string]interface{}, len(md.Enrichment.InformationTypes))
			for _, v := range md.Enrichment.InformationTypes {
				if v != "" {
					itMap[v] = v
				}
			}
			em["information_types"] = itMap
		}
		if len(md.Enrichment.Sources) > 0 {
			srcMap := make(map[string]interface{}, len(md.Enrichment.Sources))
			for _, s := range md.Enrichment.Sources {
				srcMap[strconv.FormatInt(s.ID, 10)] = map[string]interface{}{
					"id":   int(s.ID),
					"type": s.Type,
				}
			}
			em["sources"] = srcMap
		}
		result["enrichment"] = []interface{}{em}
	}

	return []interface{}{result}, nil
}
