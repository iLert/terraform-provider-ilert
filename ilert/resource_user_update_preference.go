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

func resourceUserUpdatePreference() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserUpdatePreferenceMethodAll, false),
			},
			"contact": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Either UserEmailContact or UserPhoneNumberContact",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserUpdatePreferenceTypeAll, false),
			},
			"user": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
		CreateContext: resourceUserUpdatePreferenceCreate,
		ReadContext:   resourceUserUpdatePreferenceRead,
		UpdateContext: resourceUserUpdatePreferenceUpdate,
		DeleteContext: resourceUserUpdatePreferenceDelete,
		Exists:        resourceUserUpdatePreferenceExists,
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

func buildUserUpdatePreference(d *schema.ResourceData) (*ilert.UserUpdatePreference, *int64, error) {
	method := d.Get("method").(string)
	updateType := d.Get("type").(string)

	preference := &ilert.UserUpdatePreference{
		Method: method,
		Type:   updateType,
	}

	user := d.Get("user").([]any)
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]any)
		id := int64(usr["id"].(int))
		userId = id
	}

	if val, ok := d.GetOk("contact"); ok {
		if preference.Method == "PUSH" {
			return nil, nil, fmt.Errorf("[ERROR] Field 'contact' must not be set when method is 'PUSH'")
		}
		contactList := val.([]any)
		contact := &ilert.UserContactShort{}
		if len(contactList) > 0 && contactList[0] != nil {
			cnt := contactList[0].(map[string]any)
			contact.ID = int64(cnt["id"].(int))
		}
		preference.Contact = contact
	} else {
		if preference.Method != "PUSH" {
			return nil, nil, fmt.Errorf("[ERROR] Field 'contact' must be set when method is 'EMAIL', 'SMS', 'VOICE', 'WHATSAPP' or 'TELEGRAM'")
		}
	}

	return preference, ilert.Int64(userId), nil
}

func resourceUserUpdatePreferenceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserUpdatePreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user update preference error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user update preference on user id %d", *userId)

	result := &ilert.CreateUserUpdatePreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUserUpdatePreference(&ilert.CreateUserUpdatePreferenceInput{UserUpdatePreference: preference, UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user update preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user update preference to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a user update preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user update preference error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UserUpdatePreference == nil {
		log.Printf("[ERROR] Creating ilert user update preference error: empty response")
		return diag.Errorf("user update preference response is empty")
	}

	d.SetId(strconv.FormatInt(result.UserUpdatePreference.ID, 10))

	usr := make([]any, 0)
	u := make(map[string]any, 0)
	u["id"] = int(*userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return resourceUserUpdatePreferenceRead(ctx, d, m)
}

func resourceUserUpdatePreferenceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	user := d.Get("user").([]any)
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]any)
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Reading user update preference %s from user %d", d.Id(), userId)
	result := &ilert.GetUserUpdatePreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUserUpdatePreference(&ilert.GetUserUpdatePreferenceInput{UserID: ilert.Int64(userId), UserUpdatePreferenceID: ilert.Int64(preferenceId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user update preference %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user update preference with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user update preference with id %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user update preference error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.UserUpdatePreference == nil {
		log.Printf("[ERROR] Reading ilert user update preference error: empty response")
		return diag.Errorf("user update preference response is empty")
	}

	d.Set("method", result.UserUpdatePreference.Method)
	d.Set("type", result.UserUpdatePreference.Type)

	contact, err := flattenUserContactShort(result.UserUpdatePreference.Contact)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact", contact); err != nil {
		return diag.Errorf("error setting contact: %s", err)
	}

	usr := make([]any, 0)
	u := make(map[string]any, 0)
	u["id"] = int(userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return nil
}

func resourceUserUpdatePreferenceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserUpdatePreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user update preference error %s", err.Error())
		return diag.FromErr(err)
	}

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user update preference: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUserUpdatePreference(&ilert.UpdateUserUpdatePreferenceInput{UserUpdatePreference: preference, UserUpdatePreferenceID: ilert.Int64(preferenceId), UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user update preference with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update a user update preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user update preference error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserUpdatePreferenceRead(ctx, d, m)
}

func resourceUserUpdatePreferenceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	user := d.Get("user").([]any)
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]any)
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Deleting user update preference: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUserUpdatePreference(&ilert.DeleteUserUpdatePreferenceInput{UserUpdatePreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user update preference with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a user update preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user update preference error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserUpdatePreferenceExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}
	user := d.Get("user").([]any)
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]any)
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Reading user update preference: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUserUpdatePreference(&ilert.GetUserUpdatePreferenceInput{UserUpdatePreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user update preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user update preference to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user update preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user update preference error: %s", err.Error())
		return false, err
	}
	return result, nil
}
