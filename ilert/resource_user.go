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
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
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
			"region": {
				Type:     schema.TypeString,
				Optional: true,
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
			"shift_color": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"send_no_invitation": {
				Type:     schema.TypeBool,
				Optional: true,
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
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)
	email := d.Get("email").(string)

	user := &ilert.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
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

	if val, ok := d.GetOk("region"); ok {
		user.Region = val.(string)
	}

	if val, ok := d.GetOk("role"); ok {
		user.Role = val.(string)
	}

	if val, ok := d.GetOk("shift_color"); ok {
		user.ShiftColor = val.(string)
	}

	return user, nil
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	user, err := buildUser(d)
	if err != nil {
		log.Printf("[ERROR] Building user error %s", err.Error())
		return diag.FromErr(err)
	}

	input := &ilert.CreateUserInput{User: user}
	if val, ok := d.GetOk("send_no_invitation"); ok {
		input.SendNoInvitation = Bool(val.(bool))
	}

	log.Printf("[INFO] Creating user %s", user.Username)

	result := &ilert.CreateUserOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUser(input)
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
		log.Printf("[ERROR] Creating ilert user error: empty response")
		return diag.Errorf("user response is empty")
	}

	d.SetId(strconv.FormatInt(result.User.ID, 10))

	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.User == nil {
		log.Printf("[ERROR] Reading ilert user error: empty response")
		return diag.Errorf("user response is empty")
	}

	err = transformUserResource(result.User, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update a user with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for user with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a user with id %s, error: %s", d.Id(), err.Error()))
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

func resourceUserExists(d *schema.ResourceData, m any) (bool, error) {
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
			return resource.NonRetryableError(fmt.Errorf("could not read a user with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformUserResource(user *ilert.User, d *schema.ResourceData) error {
	d.Set("first_name", user.FirstName)
	d.Set("last_name", user.LastName)
	d.Set("username", user.Username)
	d.Set("email", user.Email)
	d.Set("timezone", user.Timezone)
	d.Set("position", user.Position)
	d.Set("department", user.Department)
	d.Set("language", user.Language)
	d.Set("region", user.Region)
	d.Set("role", user.Role)
	d.Set("shift_color", user.ShiftColor)

	return nil
}
