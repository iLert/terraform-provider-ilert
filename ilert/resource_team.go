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
				ValidateFunc: validateStringValueFunc(ilert.TeamVisibilityAll),
			},
			"member": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
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
							ValidateFunc: validateStringValueFunc(ilert.TeamMemberRolesAll),
						},
					},
				},
			},
		},
		Create: resourceTeamCreate,
		Read:   resourceTeamRead,
		Update: resourceTeamUpdate,
		Delete: resourceTeamDelete,
		Exists: resourceTeamExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

	if val, ok := d.GetOk("member"); ok {
		vL := val.([]interface{})
		nps := make([]ilert.TeamMember, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
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
			nps = append(nps, ep)
		}
		team.Members = nps
	}

	return team, nil
}

func resourceTeamCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	team, err := buildTeam(d)
	if err != nil {
		log.Printf("[ERROR] Building team error %s", err.Error())
		return err
	}

	log.Printf("[INFO] Creating team %s", team.Name)

	result, err := client.CreateTeam(&ilert.CreateTeamInput{Team: team})
	if err != nil {
		log.Printf("[ERROR] Creating iLert team error %s", err.Error())
		return err
	}

	d.SetId(strconv.FormatInt(result.Team.ID, 10))

	return resourceTeamRead(d, m)
}

func resourceTeamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading team: %s", d.Id())
	result, err := client.GetTeam(&ilert.GetTeamInput{TeamID: ilert.Int64(teamID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			log.Printf("[WARN] Removing team %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an team with ID %s", d.Id())
	}

	d.Set("name", result.Team.Name)
	d.Set("visibility", result.Team.Visibility)

	members, err := flattenMembersList(result.Team.Members)
	if err != nil {
		return err
	}
	if err := d.Set("member", members); err != nil {
		return fmt.Errorf("error setting members: %s", err)
	}

	return nil
}

func resourceTeamUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	team, err := buildTeam(d)
	if err != nil {
		log.Printf("[ERROR] Building team error %s", err.Error())
		return err
	}

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Updating team: %s", d.Id())
	_, err = client.UpdateTeam(&ilert.UpdateTeamInput{Team: team, TeamID: ilert.Int64(teamID)})
	if err != nil {
		log.Printf("[ERROR] Updating iLert team error %s", err.Error())
		return err
	}
	return resourceTeamRead(d, m)
}

func resourceTeamDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Deleting team: %s", d.Id())
	_, err = client.DeleteTeam(&ilert.DeleteTeamInput{TeamID: ilert.Int64(teamID)})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceTeamExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	teamID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse team id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading team: %s", d.Id())
	_, err = client.GetTeam(&ilert.GetTeamInput{TeamID: ilert.Int64(teamID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func flattenMembersList(list []ilert.TeamMember) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["role"] = item.Role
		if item.User.ID > 0 {
			result["user"] = strconv.FormatInt(item.User.ID, 10)
		}
		results = append(results, result)
	}

	return results, nil
}
