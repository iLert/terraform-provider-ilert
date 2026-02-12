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

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ilert.TeamVisibility.Public,
				ValidateFunc: validation.StringInSlice(ilert.TeamVisibilityAll, false),
			},
			"member": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ilert.TeamMemberRoles.Responder,
							ValidateFunc: validation.StringInSlice(ilert.TeamMemberRolesAll, false),
						},
					},
				},
			},
		},
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Exists:        resourceTeamExists,
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

func buildTeam(d *schema.ResourceData) (*ilert.Team, error) {
	name := d.Get("name").(string)
	visibility := d.Get("visibility").(string)

	team := &ilert.Team{
		Name:       name,
		Visibility: visibility,
	}

	members := make([]ilert.TeamMember, 0)
	if val, ok := d.GetOk("member"); ok {
		vL := val.([]any)
		for _, m := range vL {
			v := m.(map[string]any)
			ep := ilert.TeamMember{
				Role: v["role"].(string),
			}
			if v["user"] != nil && v["user"].(string) != "" {
				userID, err := strconv.ParseInt(v["user"].(string), 10, 64)
				if err != nil {
					log.Printf("[ERROR] Could not parse user id %s", err.Error())
					return nil, unconvertibleIDErr(v["user"].(string), err)
				}
				ep.User = ilert.User{
					ID: userID,
				}
			}
			members = append(members, ep)
		}
	}
	team.Members = members

	return team, nil
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	team, err := buildTeam(d)
	if err != nil {
		log.Printf("[ERROR] Building team error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating team %s", team.Name)

	result := &ilert.CreateTeamOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateTeam(&ilert.CreateTeamInput{Team: team})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert team error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Team == nil {
		log.Printf("[ERROR] Creating ilert team error: empty response")
		return diag.Errorf("team response is empty")
	}

	d.SetId(strconv.FormatInt(result.Team.ID, 10))

	return resourceTeamRead(ctx, d, m)
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading team: %s", d.Id())
	result := &ilert.GetTeamOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetTeam(&ilert.GetTeamInput{TeamID: ilert.Int64(teamID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing team %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert team error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an team with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert team error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Team == nil {
		log.Printf("[ERROR] Reading ilert team error: empty response")
		return diag.Errorf("team response is empty")
	}

	err = transformTeamResource(result.Team, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	team, err := buildTeam(d)
	if err != nil {
		log.Printf("[ERROR] Building team error %s", err.Error())
		return diag.FromErr(err)
	}

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating team: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateTeam(&ilert.UpdateTeamInput{Team: team, TeamID: ilert.Int64(teamID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an team with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert team error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceTeamRead(ctx, d, m)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting team: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteTeam(&ilert.DeleteTeamInput{TeamID: ilert.Int64(teamID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an team with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert team error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceTeamExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading team: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetTeam(&ilert.GetTeamInput{TeamID: ilert.Int64(teamID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert team error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a team with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert team error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformTeamResource(team *ilert.Team, d *schema.ResourceData) error {
	d.Set("name", team.Name)
	d.Set("visibility", team.Visibility)

	members, err := flattenMembersListSorted(team.Members, d)
	if err != nil {
		return err
	}
	if err := d.Set("member", members); err != nil {
		return fmt.Errorf("error setting members: %s", err)
	}

	return nil
}

func flattenMembersListSorted(list []ilert.TeamMember, d *schema.ResourceData) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}

	configMembers := d.Get("member")
	if configMembers == nil {
		return make([]any, 0), nil
	}

	configMembersList, ok := configMembers.([]any)
	if !ok {
		return make([]any, 0), nil
	}

	serverMembersMap := make(map[string]ilert.TeamMember)
	for _, member := range list {
		if member.User.ID > 0 {
			userID := strconv.FormatInt(member.User.ID, 10)
			serverMembersMap[userID] = member
		}
	}

	results := make([]any, 0)
	for _, configMember := range configMembersList {
		if configMemberMap, ok := configMember.(map[string]any); ok {
			if userID, ok := configMemberMap["user"].(string); ok {
				if serverMember, exists := serverMembersMap[userID]; exists {
					result := make(map[string]any)
					result["role"] = serverMember.Role
					result["user"] = userID
					results = append(results, result)
					delete(serverMembersMap, userID)
				}
			}
		}
	}

	for _, serverMember := range serverMembersMap {
		result := make(map[string]any)
		result["role"] = serverMember.Role
		if serverMember.User.ID > 0 {
			result["user"] = strconv.FormatInt(serverMember.User.ID, 10)
		}
		results = append(results, result)
	}

	return results, nil
}
