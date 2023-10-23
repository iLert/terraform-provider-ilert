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

func resourceUserAlertPreference() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserAlertPreferenceMethodAll, false),
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
			"delay_min": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 120),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserAlertPreferenceTypeAll, false),
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
		CreateContext: resourceUserAlertPreferenceCreate,
		ReadContext:   resourceUserAlertPreferenceRead,
		UpdateContext: resourceUserAlertPreferenceUpdate,
		DeleteContext: resourceUserAlertPreferenceDelete,
		Exists:        resourceUserAlertPreferenceExists,
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

func buildUserAlertPreference(d *schema.ResourceData) (*ilert.UserAlertPreference, *int64, error) {
	method := d.Get("method").(string)
	delayMin := int64(d.Get("delay_min").(int))
	alertType := d.Get("type").(string)

	preference := &ilert.UserAlertPreference{
		Method:   method,
		DelayMin: delayMin,
		Type:     alertType,
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

func resourceUserAlertPreferenceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserAlertPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user alert preference error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user alert preference type %s on user id %d", preference.Type, *userId)

	result := &ilert.CreateUserAlertPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUserAlertPreference(&ilert.CreateUserAlertPreferenceInput{UserAlertPreference: preference, UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user alert preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user alert preference to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user alert preference error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UserAlertPreference == nil {
		log.Printf("[ERROR] Creating ilert user alert preference error: empty response ")
		return diag.Errorf("user alert preference response is empty")
	}

	d.SetId(strconv.FormatInt(result.UserAlertPreference.ID, 10))

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(*userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return resourceUserAlertPreferenceRead(ctx, d, m)
}

func resourceUserAlertPreferenceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	log.Printf("[DEBUG] Reading user alert preference %s from user %d", d.Id(), userId)
	result := &ilert.GetUserAlertPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUserAlertPreference(&ilert.GetUserAlertPreferenceInput{UserID: ilert.Int64(userId), UserAlertPreferenceID: ilert.Int64(preferenceId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user alert preference %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user alert preference with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user alert preference with id %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.UserAlertPreference == nil {
		log.Printf("[ERROR] Reading ilert user alert preference error: empty response ")
		return diag.Errorf("user alert preference response is empty")
	}

	d.Set("method", result.UserAlertPreference.Method)

	contact, err := flattenUserContactShort(result.UserAlertPreference.Contact)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact", contact); err != nil {
		return diag.Errorf("error setting contact: %s", err)
	}

	d.Set("delay_min", result.UserAlertPreference.DelayMin)
	d.Set("type", result.UserAlertPreference.Type)

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return nil
}

func resourceUserAlertPreferenceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserAlertPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user alert preference error %s", err.Error())
		return diag.FromErr(err)
	}

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user alert preference: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUserAlertPreference(&ilert.UpdateUserAlertPreferenceInput{UserAlertPreference: preference, UserAlertPreferenceID: ilert.Int64(preferenceId), UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user alert preference with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an user alert preference with id %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user alert preference error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserAlertPreferenceRead(ctx, d, m)
}

func resourceUserAlertPreferenceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	log.Printf("[DEBUG] Deleting user alert preference: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUserAlertPreference(&ilert.DeleteUserAlertPreferenceInput{UserAlertPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user alert preference with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an user alert preference with id %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user alert preference error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserAlertPreferenceExists(d *schema.ResourceData, m interface{}) (bool, error) {
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
	log.Printf("[DEBUG] Reading user alert preference: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUserAlertPreference(&ilert.GetUserAlertPreferenceInput{UserAlertPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user alert preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user alert preference to be read, error: %s", err.Error()))
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

func flattenUserContactShort(contact *ilert.UserContactShort) ([]interface{}, error) {
	if contact == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	result := make(map[string]interface{})
	if contact.ID > 0 {
		result["id"] = contact.ID
	}
	results = append(results, result)

	return results, nil
}
