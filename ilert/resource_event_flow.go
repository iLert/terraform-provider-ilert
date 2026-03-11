package ilert

import (
	"context"
	"encoding/json"
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

var eventFlowDepth = 50

func resourceEventFlow() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
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
				Elem:     resourceEventFlowRoot(eventFlowDepth),
			},
		},
		CreateContext: resourceEventFlowCreate,
		ReadContext:   resourceEventFlowRead,
		UpdateContext: resourceEventFlowUpdate,
		DeleteContext: resourceEventFlowDelete,
		Exists:        resourceEventFlowExists,
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

func resourceEventFlowRoot(depth int) *schema.Resource {
	if depth <= 0 {
		return resourceEventFlowNodeNoBranches()
	}

	return resourceEventFlowNode(depth - 1)
}

func resourceEventFlowNode(depth int) *schema.Resource {
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
				ValidateFunc: validation.StringInSlice(ilert.EventFlowNodeTypeAll, false),
			},
			"metadata": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem:     resourceEventFlowNodeMetadata(),
			},
			"branches": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceEventFlowBranch(depth),
			},
		},
	}
}

func resourceEventFlowNodeNoBranches() *schema.Resource {
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
				ValidateFunc: validation.StringInSlice(ilert.EventFlowNodeTypeAll, false),
			},
			"metadata": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem:     resourceEventFlowNodeMetadata(),
			},
		},
	}
}

func resourceEventFlowNodeMetadata() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"var_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"var_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"support_hours_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"alert_source_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"overwrite_priority": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ilert.EventFlowNodeMetadataOverwritePriorityAll, false),
			},
			"escalation_policy_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"definitions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"conditions": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"wait_for_duration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"wait_start_support_hours_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"wait_end_support_hours_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ilert.EventFlowNodeRuleOperatorAll, false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mapping": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"default": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"properties": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"items": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceEventFlowBranch(depth int) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"branch_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.EventFlowBranchTypeAll, false),
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceEventFlowRoot(depth),
			},
		},
	}
}

func buildEventFlow(d *schema.ResourceData) (*ilert.EventFlow, error) {
	name := d.Get("name").(string)

	eventFlow := &ilert.EventFlow{
		Name: name,
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]any)
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		eventFlow.Teams = tms
	}

	if val, ok := d.GetOk("root_node"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 && vL[0] != nil {
			rn := vL[0].(map[string]any)
			node, err := buildEventFlowNodeFromMap(rn)
			if err != nil {
				return nil, err
			}
			eventFlow.RootNode = node
		}
	}

	return eventFlow, nil
}

func buildEventFlowNodeFromMap(rn map[string]any) (*ilert.EventFlowNode, error) {
	node := &ilert.EventFlowNode{}
	if v, ok := rn["id"].(int); ok && v > 0 {
		node.ID = int64(v)
	}
	if s, ok := rn["node_type"].(string); ok && s != "" {
		node.NodeType = s
	}
	if s, ok := rn["name"].(string); ok && s != "" {
		node.Name = s
	}
	if mvL, ok := rn["metadata"].([]any); ok && len(mvL) > 0 && mvL[0] != nil {
		mv := mvL[0].(map[string]any)
		md := &ilert.EventFlowNodeMetadata{}
		if s, ok := mv["var_key"].(string); ok && s != "" {
			md.VarKey = s
		}
		if s, ok := mv["var_value"].(string); ok && s != "" {
			md.VarValue = s
		}
		if v, ok := mv["support_hours_id"].(int); ok && v > 0 {
			vInt64 := int64(v)
			md.SupportHoursID = &vInt64
		}
		if v, ok := mv["alert_source_id"].(int); ok && v > 0 {
			vInt64 := int64(v)
			md.AlertSourceID = &vInt64
		}
		if s, ok := mv["overwrite_priority"].(string); ok && s != "" {
			md.OverwritePriority = s
		}
		if v, ok := mv["escalation_policy_id"].(int); ok && v > 0 {
			vInt64 := int64(v)
			md.EscalationPolicyID = &vInt64
		}
		if v, ok := mv["definitions"].([]any); ok && len(v) > 0 {
			definitions := make([]ilert.EventFlowNodeDefinition, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				dv := it.(map[string]any)
				definition := ilert.EventFlowNodeDefinition{}
				if s, ok := dv["branch_name"].(string); ok && s != "" {
					definition.BranchName = s
				}
				if s, ok := dv["conditions"].(string); ok {
					definition.Conditions = s
				}
				definitions = append(definitions, definition)
			}
			md.Definitions = definitions
		}
		if s, ok := mv["wait_for_duration"].(string); ok && s != "" {
			md.WaitForDuration = s
		}
		if v, ok := mv["wait_start_support_hours_id"].(int); ok && v > 0 {
			vInt64 := int64(v)
			md.WaitStartSupportHoursID = &vInt64
		}
		if v, ok := mv["wait_end_support_hours_id"].(int); ok && v > 0 {
			vInt64 := int64(v)
			md.WaitEndSupportHoursID = &vInt64
		}
		if s, ok := mv["condition"].(string); ok && s != "" {
			md.Condition = s
		}
		if v, ok := mv["rules"].([]any); ok && len(v) > 0 {
			rules := make([]ilert.EventFlowNodeRuleMetadata, 0, len(v))
			for _, it := range v {
				if it == nil {
					continue
				}
				rv := it.(map[string]any)
				rule := ilert.EventFlowNodeRuleMetadata{}
				if s, ok := rv["name"].(string); ok && s != "" {
					rule.Name = s
				}
				if s, ok := rv["target"].(string); ok && s != "" {
					rule.Target = s
				}
				if s, ok := rv["operator"].(string); ok && s != "" {
					rule.Operator = s
				}
				if s, ok := rv["value"].(string); ok && s != "" {
					rule.Value = s
				}
				if s, ok := rv["source"].(string); ok && s != "" {
					rule.Source = s
				}
				if m, ok := rv["mapping"].(map[string]any); ok && len(m) > 0 {
					mapping := make(map[string]*string, len(m))
					for k, val := range m {
						if s, ok := val.(string); ok {
							sVal := s
							mapping[k] = &sVal
						}
					}
					rule.Mapping = mapping
				}
				if s, ok := rv["default"].(string); ok && s != "" {
					rule.Default = s
				}
				if m, ok := rv["properties"].(map[string]any); ok && len(m) > 0 {
					properties := make(map[string]*string, len(m))
					for k, val := range m {
						if s, ok := val.(string); ok {
							sVal := s
							properties[k] = &sVal
						}
					}
					rule.Properties = properties
				}
				if v, ok := rv["items"].([]any); ok && len(v) > 0 {
					items := make([]map[string]*string, 0, len(v))
					for _, item := range v {
						if item == nil {
							continue
						}
						itemMap, ok := item.(map[string]any)
						if !ok {
							continue
						}
						newItem := make(map[string]*string, len(itemMap))
						for k, val := range itemMap {
							if s, ok := val.(string); ok {
								sVal := s
								newItem[k] = &sVal
							}
						}
						items = append(items, newItem)
					}
					rule.Items = items
				}
				rules = append(rules, rule)
			}
			md.Rules = rules
		}
		node.Metadata = md
	}

	if br, ok := rn["branches"].([]any); ok && len(br) > 0 {
		branches := make([]ilert.EventFlowBranch, 0, len(br))
		for _, be := range br {
			if be == nil {
				continue
			}
			bv := be.(map[string]any)
			b := ilert.EventFlowBranch{}
			if v, ok := bv["id"].(int); ok && v > 0 {
				b.ID = int64(v)
			}
			if s, ok := bv["branch_type"].(string); ok && s != "" {
				b.BranchType = s
			}
			if s, ok := bv["condition"].(string); ok && s != "" {
				b.Condition = s
			}
			if tvL, ok := bv["target"].([]any); ok && len(tvL) > 0 && tvL[0] != nil {
				tv := tvL[0].(map[string]any)
				tn, err := buildEventFlowNodeFromMap(tv)
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

func resourceEventFlowCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	EventFlow, err := buildEventFlow(d)
	if err != nil {
		log.Printf("[ERROR] Building Event Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating Event Flow %s", EventFlow.Name)

	result := &ilert.CreateEventFlowOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateEventFlow(&ilert.CreateEventFlowInput{EventFlow: EventFlow})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert Event Flow error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Event Flow to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a Event Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert Event Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EventFlow == nil {
		log.Printf("[ERROR] Creating ilert Event Flow error: empty response")
		return diag.Errorf("Event Flow response is empty")
	}

	d.SetId(strconv.FormatInt(result.EventFlow.ID, 10))

	return resourceEventFlowRead(ctx, d, m)
}

func resourceEventFlowRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	EventFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Event Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading Event Flow: %s", d.Id())
	result := &ilert.GetEventFlowOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetEventFlow(&ilert.GetEventFlowInput{EventFlowID: ilert.Int64(EventFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing Event Flow %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Event Flow with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an Event Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert Event Flow error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EventFlow == nil {
		log.Printf("[ERROR] Reading ilert Event Flow error: empty response")
		return diag.Errorf("Event Flow response is empty")
	}

	if err := transformEventFlowResource(result.EventFlow, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceEventFlowUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	EventFlow, err := buildEventFlow(d)
	if err != nil {
		log.Printf("[ERROR] Building Event Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	EventFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Event Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating Event Flow: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateEventFlow(&ilert.UpdateEventFlowInput{EventFlow: EventFlow, EventFlowID: ilert.Int64(EventFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Event Flow with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an Event Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert Event Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceEventFlowRead(ctx, d, m)
}

func resourceEventFlowDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	EventFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Event Flow id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting Event Flow: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteEventFlow(&ilert.DeleteEventFlowInput{EventFlowID: ilert.Int64(EventFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Event Flow with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an Event Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert Event Flow error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceEventFlowExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	EventFlowID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse Event Flow id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading Event Flow: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetEventFlow(&ilert.GetEventFlowInput{EventFlowID: ilert.Int64(EventFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert Event Flow error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for Event Flow to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an Event Flow with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert Event Flow error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformEventFlowResource(eventFlow *ilert.EventFlowOutput, d *schema.ResourceData) error {
	d.Set("name", eventFlow.Name)

	teams, err := flattenTeamShortList(eventFlow.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	rootNode, err := flattenEventFlowNodeOutput(eventFlow.RootNode)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening root node: %s", err.Error())
	}
	if err := d.Set("root_node", rootNode); err != nil {
		return fmt.Errorf("[ERROR] Error setting root node: %s", err.Error())
	}

	return nil
}

func flattenEventFlowNodeOutput(node *ilert.EventFlowNodeOutput) ([]any, error) {
	if node == nil {
		return make([]any, 0), nil
	}

	result := make(map[string]any)
	if node.ID != 0 {
		result["id"] = int(node.ID)
	}
	result["node_type"] = node.NodeType

	if node.Name != "" {
		result["name"] = node.Name
	}

	mds, err := flattenEventFlowNodeMetadata(node.Metadata)
	if err != nil {
		return nil, err
	}
	if len(mds) > 0 {
		result["metadata"] = mds
	} else {
		result["metadata"] = []any{}
	}

	if len(node.Branches) > 0 {
		branches := make([]any, 0, len(node.Branches))
		for _, b := range node.Branches {
			bm := make(map[string]any)
			if b.ID != 0 {
				bm["id"] = int(b.ID)
			}
			if b.BranchType != "" {
				bm["branch_type"] = b.BranchType
			}
			if b.Condition != "" {
				bm["condition"] = b.Condition
			}
			tn, err := flattenEventFlowNode(&b.Target)
			if err != nil {
				return nil, err
			}
			bm["target"] = tn
			branches = append(branches, bm)
		}
		result["branches"] = branches
	}

	return []any{result}, nil
}

func flattenEventFlowNode(node **ilert.EventFlowNode) ([]any, error) {
	if node == nil || *node == nil {
		return []any{}, nil
	}
	n := *node
	result := make(map[string]any)
	if n.ID != 0 {
		result["id"] = int(n.ID)
	}
	result["node_type"] = n.NodeType
	if n.Name != "" {
		result["name"] = n.Name
	}
	if n.Metadata != nil {
		if mdMap, ok := n.Metadata.(map[string]any); ok {
			md := &ilert.EventFlowNodeMetadata{}

			if v, ok := mdMap["varKey"].(string); ok && v != "" {
				md.VarKey = v
			}
			if v, ok := mdMap["varValue"].(string); ok && v != "" {
				md.VarValue = v
			}
			if v, ok := mdMap["supportHoursId"].(float64); ok && v > 0 {
				vInt64 := int64(v)
				md.SupportHoursID = &vInt64
			}
			if v, ok := mdMap["alertSourceId"].(float64); ok && v > 0 {
				vInt64 := int64(v)
				md.AlertSourceID = &vInt64
			}
			if v, ok := mdMap["overwritePriority"].(string); ok && v != "" {
				md.OverwritePriority = v
			}
			if v, ok := mdMap["escalationPolicyId"].(float64); ok && v > 0 {
				vInt64 := int64(v)
				md.EscalationPolicyID = &vInt64
			}
			if v, ok := mdMap["definitions"].([]any); ok && len(v) > 0 {
				definitions := make([]ilert.EventFlowNodeDefinition, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					dv := it.(map[string]any)
					definition := ilert.EventFlowNodeDefinition{}
					if s, ok := dv["branchName"].(string); ok && s != "" {
						definition.BranchName = s
					}
					if s, ok := dv["conditions"].(string); ok {
						definition.Conditions = s
					}
					definitions = append(definitions, definition)
				}
				md.Definitions = definitions
			}
			if v, ok := mdMap["waitForDuration"].(string); ok && v != "" {
				md.WaitForDuration = v
			}
			if v, ok := mdMap["waitStartSupportHoursId"].(float64); ok && v > 0 {
				vInt64 := int64(v)
				md.WaitStartSupportHoursID = &vInt64
			}
			if v, ok := mdMap["waitEndSupportHoursId"].(float64); ok && v > 0 {
				vInt64 := int64(v)
				md.WaitEndSupportHoursID = &vInt64
			}
			if v, ok := mdMap["condition"].(string); ok && v != "" {
				md.Condition = v
			}
			if v, ok := mdMap["rules"].([]any); ok && len(v) > 0 {
				rules := make([]ilert.EventFlowNodeRuleMetadata, 0, len(v))
				for _, it := range v {
					if it == nil {
						continue
					}
					rv := it.(map[string]any)
					rule := ilert.EventFlowNodeRuleMetadata{}
					if s, ok := rv["name"].(string); ok && s != "" {
						rule.Name = s
					}
					if s, ok := rv["target"].(string); ok && s != "" {
						rule.Target = s
					}
					if s, ok := rv["operator"].(string); ok && s != "" {
						rule.Operator = s
					}
					if val, ok := rv["value"]; ok && val != nil {
						rule.Value = flattenEventFlowAnyToString(val)
					}
					if s, ok := rv["source"].(string); ok && s != "" {
						rule.Source = s
					}
					if m, ok := rv["mapping"].(map[string]any); ok && len(m) > 0 {
						mapping := make(map[string]*string, len(m))
						for k, val := range m {
							if s, ok := val.(string); ok {
								sVal := s
								mapping[k] = &sVal
							}
						}
						rule.Mapping = mapping
					}
					if val, ok := rv["default"]; ok && val != nil {
						rule.Default = flattenEventFlowAnyToString(val)
					}
					if m, ok := rv["properties"].(map[string]any); ok && len(m) > 0 {
						properties := make(map[string]*string, len(m))
						for k, val := range m {
							if s, ok := val.(string); ok {
								sVal := s
								properties[k] = &sVal
							}
						}
						rule.Properties = properties
					}
					if v, ok := rv["items"].([]any); ok && len(v) > 0 {
						items := make([]map[string]*string, 0, len(v))
						for _, item := range v {
							if item == nil {
								continue
							}
							itemMap, ok := item.(map[string]any)
							if !ok {
								continue
							}
							newItem := make(map[string]*string, len(itemMap))
							for k, val := range itemMap {
								if s, ok := val.(string); ok {
									sVal := s
									newItem[k] = &sVal
								}
							}
							items = append(items, newItem)
						}
						rule.Items = items
					}
					rules = append(rules, rule)
				}
				md.Rules = rules
			}

			mds, err := flattenEventFlowNodeMetadata(md)
			if err != nil {
				return nil, err
			}
			result["metadata"] = mds
		} else if md, ok := n.Metadata.(*ilert.EventFlowNodeMetadata); ok && md != nil {
			mds, err := flattenEventFlowNodeMetadata(md)
			if err != nil {
				return nil, err
			}
			result["metadata"] = mds
		}
	}
	if len(n.Branches) > 0 {
		branches := make([]any, 0, len(n.Branches))
		for _, b := range n.Branches {
			bm := make(map[string]any)
			if b.ID != 0 {
				bm["id"] = int(b.ID)
			}
			if b.BranchType != "" {
				bm["branch_type"] = b.BranchType
			}
			if b.Condition != "" {
				bm["condition"] = b.Condition
			}
			tn, err := flattenEventFlowNode(&b.Target)
			if err != nil {
				return nil, err
			}
			bm["target"] = tn
			branches = append(branches, bm)
		}
		result["branches"] = branches
	}
	return []any{result}, nil
}

func flattenEventFlowNodeMetadata(md *ilert.EventFlowNodeMetadata) ([]any, error) {
	if md == nil {
		return make([]any, 0), nil
	}

	result := make(map[string]any)

	if md.VarKey != "" {
		result["var_key"] = md.VarKey
	}
	if md.VarValue != "" {
		result["var_value"] = md.VarValue
	}
	if md.SupportHoursID != nil && *md.SupportHoursID > 0 {
		result["support_hours_id"] = int(*md.SupportHoursID)
	}
	if md.AlertSourceID != nil && *md.AlertSourceID > 0 {
		result["alert_source_id"] = int(*md.AlertSourceID)
	}
	if md.OverwritePriority != "" {
		result["overwrite_priority"] = md.OverwritePriority
	}
	if md.EscalationPolicyID != nil && *md.EscalationPolicyID > 0 {
		result["escalation_policy_id"] = int(*md.EscalationPolicyID)
	}
	if len(md.Definitions) > 0 {
		definitions := make([]any, 0, len(md.Definitions))
		for _, definition := range md.Definitions {
			m := make(map[string]any)
			if definition.BranchName != "" {
				m["branch_name"] = definition.BranchName
			}
			m["conditions"] = definition.Conditions
			definitions = append(definitions, m)
		}
		result["definitions"] = definitions
	}
	if md.WaitForDuration != "" {
		result["wait_for_duration"] = md.WaitForDuration
	}
	if md.WaitStartSupportHoursID != nil && *md.WaitStartSupportHoursID > 0 {
		result["wait_start_support_hours_id"] = int(*md.WaitStartSupportHoursID)
	}
	if md.WaitEndSupportHoursID != nil && *md.WaitEndSupportHoursID > 0 {
		result["wait_end_support_hours_id"] = int(*md.WaitEndSupportHoursID)
	}
	if md.Condition != "" {
		result["condition"] = md.Condition
	}
	if len(md.Rules) > 0 {
		rules := make([]any, 0, len(md.Rules))
		for _, rule := range md.Rules {
			m := make(map[string]any)
			if rule.Name != "" {
				m["name"] = rule.Name
			}
			if rule.Target != "" {
				m["target"] = rule.Target
			}
			if rule.Operator != "" {
				m["operator"] = rule.Operator
			}
			if value := flattenEventFlowAnyToString(rule.Value); value != "" {
				m["value"] = value
			}
			if rule.Source != "" {
				m["source"] = rule.Source
			}
			if len(rule.Mapping) > 0 {
				mapping := make(map[string]any, len(rule.Mapping))
				for key, value := range rule.Mapping {
					if value != nil {
						mapping[key] = *value
					}
				}
				m["mapping"] = mapping
			}
			if defaultValue := flattenEventFlowAnyToString(rule.Default); defaultValue != "" {
				m["default"] = defaultValue
			}
			if len(rule.Properties) > 0 {
				properties := make(map[string]any, len(rule.Properties))
				for key, value := range rule.Properties {
					if value != nil {
						properties[key] = *value
					}
				}
				m["properties"] = properties
			}
			if len(rule.Items) > 0 {
				items := make([]any, 0, len(rule.Items))
				for _, item := range rule.Items {
					itemMap := make(map[string]any, len(item))
					for key, value := range item {
						if value != nil {
							itemMap[key] = *value
						}
					}
					items = append(items, itemMap)
				}
				m["items"] = items
			}
			rules = append(rules, m)
		}
		result["rules"] = rules
	}

	return []any{result}, nil
}

func flattenEventFlowAnyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
