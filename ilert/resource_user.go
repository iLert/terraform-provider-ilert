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

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mobile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						"number": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"landline": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_code": {
							Type:     schema.TypeString,
							Required: true,
						},
						"number": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Europe/Berlin",
			},
			"position": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"department": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"language": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "en",
				ValidateFunc: validation.StringInSlice([]string{
					"en",
					"de",
				}, false),
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "USER",
				ValidateFunc: validation.StringInSlice([]string{
					"ADMIN",
					"USER",
					"RESPONDER",
					"STAKEHOLDER",
					"GUEST",
				}, false),
			},
		},
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Exists:        resourceUserExists,
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

func buildUser(d *schema.ResourceData) (*ilert.User, error) {
	email := d.Get("email").(string)
	username := d.Get("username").(string)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)

	user := &ilert.User{
		Email:     email,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
	}

	if val, ok := d.GetOk("mobile"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 && vL[0] != nil {
			mobile := &ilert.Phone{}
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				if code, ok := v["region_code"].(string); ok && code != "" {
					mobile.RegionCode = code
				}
				if number, ok := v["number"].(string); ok && number != "" {
					mobile.Number = number
				}
			}
			if mobile.RegionCode != "" && mobile.Number != "" {
				user.Mobile = mobile
			} else {
				user.Mobile = nil
			}
		}
	}

	if val, ok := d.GetOk("landline"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 && vL[0] != nil {
			landline := &ilert.Phone{}
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				if code, ok := v["region_code"].(string); ok && code != "" {
					landline.RegionCode = code
				}
				if number, ok := v["number"].(string); ok && number != "" {
					landline.Number = number
				}
			}
			if landline.RegionCode != "" && landline.Number != "" {
				user.Landline = landline
			} else {
				user.Landline = nil
			}
		}
	}

	if val, ok := d.GetOk("timezone"); ok {
		user.Timezone = val.(string)
	}

	if val, ok := d.GetOk("position"); ok {
		user.Position = val.(string)
	}

	if val, ok := d.GetOk("department"); ok {
		user.Department = val.(string)
	}

	if val, ok := d.GetOk("language"); ok {
		user.Language = val.(string)
	}

	if val, ok := d.GetOk("role"); ok {
		user.Role = val.(string)
	}

	return user, nil
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	user, err := buildUser(d)
	if err != nil {
		log.Printf("[ERROR] Building user error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user %s", user.Username)

	result := &ilert.CreateUserOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUser(&ilert.CreateUserInput{User: user})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.User == nil {
		log.Printf("[ERROR] Creating ilert user error: empty response ")
		return diag.Errorf("user response is empty")
	}

	d.SetId(strconv.FormatInt(result.User.ID, 10))

	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	userID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading user: %s", d.Id())

	result := &ilert.GetUserOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUser(&ilert.GetUserInput{UserID: ilert.Int64(userID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.User == nil {
		log.Printf("[ERROR] Reading ilert user error: empty response ")
		return diag.Errorf("user response is empty")
	}

	d.Set("username", result.User.Username)
	d.Set("first_name", result.User.FirstName)
	d.Set("last_name", result.User.LastName)
	d.Set("email", result.User.Email)
	d.Set("timezone", result.User.Timezone)
	d.Set("position", result.User.Position)
	d.Set("department", result.User.Department)
	d.Set("language", result.User.Language)
	d.Set("role", result.User.Role)

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	user, err := buildUser(d)
	if err != nil {
		log.Printf("[ERROR] Building user error %s", err.Error())
		return diag.FromErr(err)
	}

	userID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUser(&ilert.UpdateUserInput{User: user, UserID: ilert.Int64(userID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an user with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	userID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting user: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUser(&ilert.DeleteUserInput{UserID: ilert.Int64(userID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an user with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	userID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading user: %s", d.Id())

	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUser(&ilert.GetUserInput{UserID: ilert.Int64(userID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = true
		return nil
	})

	if err != nil {
		return false, err
	}
	return result, nil
}
