package ilert

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v4"
)

// Agents are early access. The API replaces an agent with what an update sends,
// so an update reads the agent first and changes only what the configuration
// manages, see applyAgentConfig.
func resourceAgent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ilert.AgentStatus.Enabled,
				ValidateFunc: validation.StringInSlice(ilert.AgentStatusAll, false),
			},
			"avatar_url": {
				// computed so an avatar set in the web app is kept while the
				// configuration does not set one
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"team": {
				// TypeSet, not TypeList: the API returns teams sorted by id, so an
				// ordered list produces a permanent diff whenever the config order
				// differs from that.
				Type:     schema.TypeSet,
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
			"mcp_server": {
				// TypeList: the API keeps the servers in the order they were sent
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"brand": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"credential_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9]+$`), "must be the id of a credential"),
						},
					},
				},
			},
			"mcp_tool": {
				// TypeList: the API keeps the tools in the order they were sent
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"input_overwrite": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      ilert.AgentMcpToolInputType.String,
										ValidateFunc: validation.StringInSlice(ilert.AgentMcpToolInputTypeAll, false),
									},
									"values": {
										// an overwrite without values has no effect, the
										// agent decides on the input as without one
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
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
		CreateContext: resourceAgentCreate,
		ReadContext:   resourceAgentRead,
		UpdateContext: resourceAgentUpdate,
		DeleteContext: resourceAgentDelete,
		Exists:        resourceAgentExists,
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

// applyAgentConfig writes the configuration onto agent: an empty one on create, the
// agent as the API returned it on update. The API replaces an agent with what an
// update sends and clears most of what it leaves out, so everything Terraform does not
// manage has to go back as it was read: the fields the resource has no argument for,
// such as the additional instructions and the RCA policy, an avatar the configuration
// does not set, and the teams, MCP servers and tools of a configuration that never
// declared them. Those are managed once declared, the way teams are on every resource.
func applyAgentConfig(agent *ilert.Agent, d *schema.ResourceData) error {
	agent.Name = d.Get("name").(string)
	agent.Status = d.Get("status").(string)

	if d.HasChange("avatar_url") {
		agent.AvatarURL = d.Get("avatar_url").(string)
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.(*schema.Set).List()
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
		agent.Teams = tms
	} else if d.HasChange("team") {
		// All team blocks removed. The API only clears teams on an explicit empty
		// array, an omitted or null field leaves them untouched.
		agent.Teams = []ilert.TeamShort{}
	}

	if val, ok := d.GetOk("mcp_server"); ok {
		servers, err := buildAgentMcpServers(val.([]any))
		if err != nil {
			return err
		}
		agent.McpServers = servers
	} else if d.HasChange("mcp_server") {
		agent.McpServers = []ilert.AgentMcpServer{}
	}

	if val, ok := d.GetOk("mcp_tool"); ok {
		agent.McpTools = buildAgentMcpTools(val.([]any))
	} else if d.HasChange("mcp_tool") {
		agent.McpTools = []ilert.AgentMcpTool{}
	}

	return nil
}

func buildAgentMcpServers(vL []any) ([]ilert.AgentMcpServer, error) {
	servers := make([]ilert.AgentMcpServer, 0, len(vL))
	for _, m := range vL {
		v, ok := m.(map[string]any)
		if !ok {
			continue
		}
		server := ilert.AgentMcpServer{
			Name:        v["name"].(string),
			Active:      v["active"].(bool),
			DisplayName: v["display_name"].(string),
			Brand:       v["brand"].(string),
		}
		if id := v["credential_id"].(string); id != "" {
			credentialID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("could not parse the credential_id %q of the mcp_server %q: %s", id, server.Name, err.Error())
			}
			server.Credentials = &ilert.AgentMcpServerCredentials{ID: credentialID}
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func buildAgentMcpTools(vL []any) []ilert.AgentMcpTool {
	tools := make([]ilert.AgentMcpTool, 0, len(vL))
	for _, m := range vL {
		v, ok := m.(map[string]any)
		if !ok {
			continue
		}
		tool := ilert.AgentMcpTool{
			Server:          v["server"].(string),
			Name:            v["name"].(string),
			Active:          v["active"].(bool),
			InputOverwrites: make([]ilert.AgentMcpToolInputOverwrite, 0),
		}
		overwrites, _ := v["input_overwrite"].([]any)
		for _, o := range overwrites {
			ov, ok := o.(map[string]any)
			if !ok {
				continue
			}
			values := make(ilert.AgentMcpToolInputValues, 0)
			vals, _ := ov["values"].([]any)
			for _, value := range vals {
				s, _ := value.(string)
				values = append(values, s)
			}
			tool.InputOverwrites = append(tool.InputOverwrites, ilert.AgentMcpToolInputOverwrite{
				Name:  ov["name"].(string),
				Type:  ov["type"].(string),
				Value: values,
			})
		}
		tools = append(tools, tool)
	}
	return tools
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	agent := &ilert.Agent{}
	if err := applyAgentConfig(agent, d); err != nil {
		log.Printf("[ERROR] Building agent error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating agent %s", agent.Name)

	result := &ilert.CreateAgentOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateAgent(&ilert.CreateAgentInput{Agent: agent})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert agent error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert agent error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Agent == nil {
		log.Printf("[ERROR] Creating ilert agent error: empty response")
		return diag.Errorf("agent response is empty")
	}

	d.SetId(strconv.FormatInt(result.Agent.ID, 10))

	return resourceAgentRead(ctx, d, m)
}

// fetchAgent reads the agent with retries. A missing agent is returned as the
// *ilert.NotFoundAPIError of the API.
func fetchAgent(ctx context.Context, client *ilert.Client, d *schema.ResourceData, timeout time.Duration) (*ilert.Agent, error) {
	agentID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse agent id %s", err.Error())
		return nil, unconvertibleIDErr(d.Id(), err)
	}

	var agent *ilert.Agent
	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		r, err := client.GetAgent(&ilert.GetAgentInput{AgentID: ilert.Int64(agentID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		agent = r.Agent
		return nil
	})
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent response is empty")
	}
	return agent, nil
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	log.Printf("[DEBUG] Reading agent: %s", d.Id())
	agent, err := fetchAgent(ctx, client, d, d.Timeout(schema.TimeoutRead))
	if err != nil {
		var notFound *ilert.NotFoundAPIError
		if errors.As(err, &notFound) {
			log.Printf("[WARN] Removing agent %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		log.Printf("[ERROR] Reading ilert agent error: %s", err.Error())
		return diag.Errorf("could not read an agent with ID %s, error: %s", d.Id(), err.Error())
	}

	err = transformAgentResource(agent, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	agentID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse agent id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	agent, err := fetchAgent(ctx, client, d, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		log.Printf("[ERROR] Reading ilert agent before updating it error: %s", err.Error())
		return diag.Errorf("could not read the agent with ID %s to update it, error: %s", d.Id(), err.Error())
	}

	if err := applyAgentConfig(agent, d); err != nil {
		log.Printf("[ERROR] Building agent error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Updating agent: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateAgent(&ilert.UpdateAgentInput{Agent: agent, AgentID: ilert.Int64(agentID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an agent with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert agent error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceAgentRead(ctx, d, m)
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	agentID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse agent id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting agent: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteAgent(&ilert.DeleteAgentInput{AgentID: ilert.Int64(agentID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				// already gone, deleted in the web app for example
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an agent with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert agent error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceAgentExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	agentID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse agent id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading agent: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetAgent(&ilert.GetAgentInput{AgentID: ilert.Int64(agentID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert agent error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an agent with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert agent error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformAgentResource(agent *ilert.Agent, d *schema.ResourceData) error {
	d.Set("name", agent.Name)
	d.Set("status", agent.Status)
	d.Set("avatar_url", agent.AvatarURL)
	d.Set("created_at", agent.CreatedAt)
	d.Set("updated_at", agent.UpdatedAt)

	teams, err := flattenTeamShortList(agent.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	servers := make([]any, 0)
	if agentManagesList(d, "mcp_server") {
		servers = flattenAgentMcpServers(agent.McpServers)
	}
	if err := d.Set("mcp_server", servers); err != nil {
		return fmt.Errorf("[ERROR] Error setting mcp servers: %s", err.Error())
	}

	tools := make([]any, 0)
	if agentManagesList(d, "mcp_tool") {
		tools = flattenAgentMcpTools(agent.McpTools)
	}
	if err := d.Set("mcp_tool", tools); err != nil {
		return fmt.Errorf("[ERROR] Error setting mcp tools: %s", err.Error())
	}

	return nil
}

// agentManagesList reports whether the MCP servers or tools of an agent are read into
// the state: once the configuration declared a block of them, like teams. Without an
// id the agent is read without a configuration, for the IaC exporter, which takes all.
func agentManagesList(d *schema.ResourceData, key string) bool {
	if _, ok := d.GetOk(key); ok {
		return true
	}
	return d.Id() == ""
}

func flattenAgentMcpServers(servers []ilert.AgentMcpServer) []any {
	results := make([]any, 0, len(servers))
	for _, server := range servers {
		result := map[string]any{
			"name":          server.Name,
			"active":        server.Active,
			"display_name":  server.DisplayName,
			"brand":         server.Brand,
			"credential_id": "",
		}
		if server.Credentials != nil {
			result["credential_id"] = strconv.FormatInt(server.Credentials.ID, 10)
		}
		results = append(results, result)
	}
	return results
}

func flattenAgentMcpTools(tools []ilert.AgentMcpTool) []any {
	results := make([]any, 0, len(tools))
	for _, tool := range tools {
		overwrites := make([]any, 0, len(tool.InputOverwrites))
		for _, overwrite := range tool.InputOverwrites {
			if len(overwrite.Value) == 0 {
				// no effect, the agent decides on the input as without an overwrite.
				// The web app writes these, the configuration cannot hold them.
				continue
			}
			values := make([]any, 0, len(overwrite.Value))
			for _, value := range overwrite.Value {
				values = append(values, value)
			}
			overwrites = append(overwrites, map[string]any{
				"name":   overwrite.Name,
				"type":   overwrite.Type,
				"values": values,
			})
		}
		results = append(results, map[string]any{
			"server":          tool.Server,
			"name":            tool.Name,
			"active":          tool.Active,
			"input_overwrite": overwrites,
		})
	}
	return results
}
