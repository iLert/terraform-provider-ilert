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

func resourceUserSubscriptionPreference() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserSubscriptionPreferenceMethodAll, false),
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
		CreateContext: resourceUserSubscriptionPreferenceCreate,
		ReadContext:   resourceUserSubscriptionPreferenceRead,
		UpdateContext: resourceUserSubscriptionPreferenceUpdate,
		DeleteContext: resourceUserSubscriptionPreferenceDelete,
		Exists:        resourceUserSubscriptionPreferenceExists,
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

func buildUserSubscriptionPreference(d *schema.ResourceData) (*ilert.UserSubscriptionPreference, *int64, error) {
	method := d.Get("method").(string)

	preference := &ilert.UserSubscriptionPreference{
		Method: method,
	}

	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}

	if val, ok := d.GetOk("contact"); ok {
		if preference.Method == "PUSH" {
			return nil, nil, fmt.Errorf("[ERROR] Field 'contact' must not be set when method is 'PUSH'")
		}
		contactList := val.([]interface{})
		contact := &ilert.UserContactShort{}
		if len(contactList) > 0 && contactList[0] != nil {
			cnt := contactList[0].(map[string]interface{})
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

func resourceUserSubscriptionPreferenceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserSubscriptionPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user subscription preference error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user subscription preference on user id %d", *userId)

	result := &ilert.CreateUserSubscriptionPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUserSubscriptionPreference(&ilert.CreateUserSubscriptionPreferenceInput{UserSubscriptionPreference: preference, UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user subscription preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user subscription preference to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a user subscription preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user subscription preference error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UserSubscriptionPreference == nil {
		log.Printf("[ERROR] Creating ilert user subscription preference error: empty response")
		return diag.Errorf("user subscription preference response is empty")
	}

	d.SetId(strconv.FormatInt(result.UserSubscriptionPreference.ID, 10))

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(*userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return resourceUserSubscriptionPreferenceRead(ctx, d, m)
}

func resourceUserSubscriptionPreferenceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Reading user subscription preference %s from user %d", d.Id(), userId)
	result := &ilert.GetUserSubscriptionPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUserSubscriptionPreference(&ilert.GetUserSubscriptionPreferenceInput{UserID: ilert.Int64(userId), UserSubscriptionPreferenceID: ilert.Int64(preferenceId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user subscription preference %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user subscription preference with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user subscription preference with id %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user subscription preference error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.UserSubscriptionPreference == nil {
		log.Printf("[ERROR] Reading ilert user subscription preference error: empty response")
		return diag.Errorf("user subscription preference response is empty")
	}

	d.Set("method", result.UserSubscriptionPreference.Method)

	contact, err := flattenUserContactShort(result.UserSubscriptionPreference.Contact)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact", contact); err != nil {
		return diag.Errorf("error setting contact: %s", err)
	}

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return nil
}

func resourceUserSubscriptionPreferenceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserSubscriptionPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user subscription preference error %s", err.Error())
		return diag.FromErr(err)
	}

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user subscription preference: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUserSubscriptionPreference(&ilert.UpdateUserSubscriptionPreferenceInput{UserSubscriptionPreference: preference, UserSubscriptionPreferenceID: ilert.Int64(preferenceId), UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user subscription preference with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update a user subscription preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user subscription preference error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserSubscriptionPreferenceRead(ctx, d, m)
}

func resourceUserSubscriptionPreferenceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Deleting user subscription preference: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUserSubscriptionPreference(&ilert.DeleteUserSubscriptionPreferenceInput{UserSubscriptionPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user subscription preference with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a user subscription preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user subscription preference error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserSubscriptionPreferenceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}
	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}
	log.Printf("[DEBUG] Reading user subscription preference: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUserSubscriptionPreference(&ilert.GetUserSubscriptionPreferenceInput{UserSubscriptionPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user subscription preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user subscription preference to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user subscription preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user subscription preference error: %s", err.Error())
		return false, err
	}
	return result, nil
}
