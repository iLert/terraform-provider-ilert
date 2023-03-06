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

func resourceUserPhoneNumberContact() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"region_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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
		CreateContext: resourceUserPhoneNumberContactCreate,
		ReadContext:   resourceUserPhoneNumberContactRead,
		UpdateContext: resourceUserPhoneNumberContactUpdate,
		DeleteContext: resourceUserPhoneNumberContactDelete,
		Exists:        resourceUserPhoneNumberContactExists,
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

func buildUserPhoneNumberContact(d *schema.ResourceData) (*ilert.UserPhoneNumberContact, *int64, error) {
	regionCode := d.Get("region_code").(string)
	target := d.Get("target").(string)

	contact := &ilert.UserPhoneNumberContact{
		RegionCode: regionCode,
		Target:     target,
	}

	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}

	return contact, ilert.Int64(userId), nil
}

func resourceUserPhoneNumberContactCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	contact, userId, err := buildUserPhoneNumberContact(d)
	if err != nil {
		log.Printf("[ERROR] Building user phone number contact error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating user phone number contact %s on user id %d", contact.Target, *userId)

	result := &ilert.CreateUserPhoneNumberContactOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUserPhoneNumberContact(&ilert.CreateUserPhoneNumberContactInput{UserPhoneNumberContact: contact, UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert user phone number contact error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user phone number contact to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert user phone number contact error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UserPhoneNumberContact == nil {
		log.Printf("[ERROR] Creating ilert user phone number contact error: empty response ")
		return diag.Errorf("user phone number contact response is empty")
	}

	d.SetId(strconv.FormatInt(result.UserPhoneNumberContact.ID, 10))

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(*userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return resourceUserPhoneNumberContactRead(ctx, d, m)
}

func resourceUserPhoneNumberContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	contactId, err := strconv.ParseInt(d.Id(), 10, 64)
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
	log.Printf("[DEBUG] Reading user phone number contact %s from user %d", d.Id(), userId)
	result := &ilert.GetUserPhoneNumberContactOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUserPhoneNumberContact(&ilert.GetUserPhoneNumberContactInput{UserID: ilert.Int64(userId), UserPhoneNumberContactID: ilert.Int64(contactId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing user phone number contact %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user phone number contact with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user phone number contact with id %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.UserPhoneNumberContact == nil {
		log.Printf("[ERROR] Reading ilert user phone number contact error: empty response ")
		return diag.Errorf("user phone number contact response is empty")
	}

	d.Set("region_code", result.UserPhoneNumberContact.RegionCode)
	d.Set("target", result.UserPhoneNumberContact.Target)
	d.Set("status", result.UserPhoneNumberContact.Status)

	usr := make([]interface{}, 0)
	u := make(map[string]interface{}, 0)
	u["id"] = int(userId)
	usr = append(usr, u)
	d.Set("user", usr)

	return nil
}

func resourceUserPhoneNumberContactUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	contact, userId, err := buildUserPhoneNumberContact(d)
	if err != nil {
		log.Printf("[ERROR] Building user phone number contact error %s", err.Error())
		return diag.FromErr(err)
	}

	contactId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating user phone number contact: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUserPhoneNumberContact(&ilert.UpdateUserPhoneNumberContactInput{UserPhoneNumberContact: contact, UserPhoneNumberContactID: ilert.Int64(contactId), UserID: userId})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user phone number contact with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an user phone number contact with id %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert user phone number contact error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUserPhoneNumberContactRead(ctx, d, m)
}

func resourceUserPhoneNumberContactDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	contactId, err := strconv.ParseInt(d.Id(), 10, 64)
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
	log.Printf("[DEBUG] Deleting user phone number contact: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUserPhoneNumberContact(&ilert.DeleteUserPhoneNumberContactInput{UserPhoneNumberContactID: ilert.Int64(contactId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user phone number contact with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an user phone number contact with id %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert user phone number contact error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUserPhoneNumberContactExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	contactId, err := strconv.ParseInt(d.Id(), 10, 64)
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
	log.Printf("[DEBUG] Reading user phone number contact: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUserPhoneNumberContact(&ilert.GetUserPhoneNumberContactInput{UserPhoneNumberContactID: ilert.Int64(contactId), UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert user phone number contact error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user phone number contact to be read, error: %s", err.Error()))
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
