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

func resourceUserDutyPreference() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserDutyPreferenceMethodAll, false),
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
			"before_min": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{0, 15, 30, 60, 180, 360, 720, 1440}),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.UserDutyPreferenceTypeAll, false),
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
		CreateContext: resourceUserDutyPreferenceCreate,
		ReadContext:   resourceUserDutyPreferenceRead,
		UpdateContext: resourceUserDutyPreferenceUpdate,
		DeleteContext: resourceUserDutyPreferenceDelete,
		Exists:        resourceUserDutyPreferenceExists,
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

func buildUserDutyPreference(d *schema.ResourceData) (*ilert.UserDutyPreference, *int64, error) {
	method := d.Get("method").(string)
	beforeMin := int64(d.Get("before_min").(int))
	dutyType := d.Get("type").(string)

	preference := &ilert.UserDutyPreference{
		Method:    method,
		BeforeMin: beforeMin,
		Type:      dutyType,
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

func resourceUserDutyPreferenceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserDutyPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user duty preference error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user duty preference type %s on user id %d", preference.Type, *userId)

	result := &ilert.CreateUserDutyPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUserDutyPreference(&ilert.CreateUserDutyPreferenceInput{UserDutyPreference: preference, UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user duty preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user duty preference to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a user duty preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user duty preference error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UserDutyPreference == nil {
		log.Printf("[ERROR] Creating ilert user duty preference error: empty response")
		return diag.Errorf("user duty preference response is empty")
	}

	d.SetId(strconv.FormatInt(result.UserDutyPreference.ID, 10))

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(*userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return resourceUserDutyPreferenceRead(ctx, d, m)
}

func resourceUserDutyPreferenceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	log.Printf("[DEBUG] Reading user duty preference %s from user %d", d.Id(), userId)
	result := &ilert.GetUserDutyPreferenceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUserDutyPreference(&ilert.GetUserDutyPreferenceInput{UserID: ilert.Int64(userId), UserDutyPreferenceID: ilert.Int64(preferenceId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user duty preference %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user duty preference with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user duty preference with id %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user duty preference error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.UserDutyPreference == nil {
		log.Printf("[ERROR] Reading ilert user duty preference error: empty response")
		return diag.Errorf("user duty preference response is empty")
	}

	d.Set("method", result.UserDutyPreference.Method)

	contact, err := flattenUserContactShort(result.UserDutyPreference.Contact)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact", contact); err != nil {
		return diag.Errorf("error setting contact: %s", err)
	}

	d.Set("before_min", result.UserDutyPreference.BeforeMin)
	d.Set("type", result.UserDutyPreference.Type)

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return nil
}

func resourceUserDutyPreferenceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	preference, userId, err := buildUserDutyPreference(d)
	if err != nil {
		log.Printf("[ERROR] Building user duty preference error %s", err.Error())
		return diag.FromErr(err)
	}

	preferenceId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user duty preference: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUserDutyPreference(&ilert.UpdateUserDutyPreferenceInput{UserDutyPreference: preference, UserDutyPreferenceID: ilert.Int64(preferenceId), UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user duty preference with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update a user duty preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user duty preference error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserDutyPreferenceRead(ctx, d, m)
}

func resourceUserDutyPreferenceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	log.Printf("[DEBUG] Deleting user duty preference: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUserDutyPreference(&ilert.DeleteUserDutyPreferenceInput{UserDutyPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user duty preference with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a user duty preference with id %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user duty preference error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserDutyPreferenceExists(d *schema.ResourceData, m interface{}) (bool, error) {
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
	log.Printf("[DEBUG] Reading user duty preference: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUserDutyPreference(&ilert.GetUserDutyPreferenceInput{UserDutyPreferenceID: ilert.Int64(preferenceId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user duty preference error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user duty preference to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user duty preference with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert user duty preference error: %s", err.Error())
		return false, err
	}
	return result, nil
}
