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

func resourceStatusPageGroup() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"status_page": {
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
		CreateContext: resourceStatusPageGroupCreate,
		ReadContext:   resourceStatusPageGroupRead,
		UpdateContext: resourceStatusPageGroupUpdate,
		DeleteContext: resourceStatusPageGroupDelete,
		Exists:        resourceStatusPageGroupExists,
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

func buildStatusPageGroup(d *schema.ResourceData) (*ilert.StatusPageGroup, *int64, error) {
	name := d.Get("name").(string)

	StatusPageGroup := &ilert.StatusPageGroup{
		Name: name,
	}

	spL := d.Get("status_page").([]any)
	StatusPageID := int64(-1)
	if len(spL) > 0 && spL[0] != nil {
		sp := spL[0].(map[string]any)
		id := int64(sp["id"].(int))
		StatusPageID = id
	}

	return StatusPageGroup, &StatusPageID, nil
}

func resourceStatusPageGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageGroup, statusPageID, err := buildStatusPageGroup(d)
	if err != nil {
		log.Printf("[ERROR] Building status page group error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating status page group %s on status page with ID %d", statusPageGroup.Name, statusPageID)

	result := &ilert.CreateStatusPageGroupOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateStatusPageGroup(&ilert.CreateStatusPageGroupInput{StatusPageGroup: statusPageGroup, StatusPageID: statusPageID})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a status page group with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Creating ilert status page group error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.StatusPageGroup == nil {
		log.Printf("[ERROR] Creating ilert status page group error: empty response")
		return diag.Errorf("status page group response is empty")
	}

	d.SetId(strconv.FormatInt(result.StatusPageGroup.ID, 10))

	sp := make([]any, 0)
	s := make(map[string]any, 0)
	s["id"] = int(*statusPageID)
	sp = append(sp, s)
	d.Set("status_page", sp)

	return resourceStatusPageGroupRead(ctx, d, m)
}

func resourceStatusPageGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page group id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	spL := d.Get("status_page").([]any)
	statusPageID := int64(-1)
	if len(spL) > 0 && spL[0] != nil {
		sp := spL[0].(map[string]any)
		id := int64(sp["id"].(int))
		statusPageID = id
	}
	log.Printf("[DEBUG] Reading status page group id %s from status page id %d", d.Id(), statusPageID)
	result := &ilert.GetStatusPageGroupOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetStatusPageGroup(&ilert.GetStatusPageGroupInput{StatusPageGroupID: ilert.Int64(statusPageGroupID), StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing status page group %s from state because it no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert status page group error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an status page group with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert status page group error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.StatusPageGroup == nil {
		log.Printf("[ERROR] Reading ilert status page group error: empty response")
		return diag.Errorf("status page group response is empty")
	}

	err = transformStatusPageGroupResource(result.StatusPageGroup, statusPageID, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStatusPageGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageGroup, statusPageID, err := buildStatusPageGroup(d)
	if err != nil {
		log.Printf("[ERROR] Building status page group error %s", err.Error())
		return diag.FromErr(err)
	}

	statusPageGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page group id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating status page group: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateStatusPageGroup(&ilert.UpdateStatusPageGroupInput{StatusPageGroup: statusPageGroup, StatusPageGroupID: ilert.Int64(statusPageGroupID), StatusPageID: statusPageID})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an status page group with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert status page group error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceStatusPageGroupRead(ctx, d, m)
}

func resourceStatusPageGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page group id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	spL := d.Get("status_page").([]any)
	statusPageID := int64(-1)
	if len(spL) > 0 && spL[0] != nil {
		sp := spL[0].(map[string]any)
		id := int64(sp["id"].(int))
		statusPageID = id
	}
	log.Printf("[DEBUG] Deleting status page group: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteStatusPageGroup(&ilert.DeleteStatusPageGroupInput{StatusPageGroupID: ilert.Int64(statusPageGroupID), StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a status page group with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert status page group error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceStatusPageGroupExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	statusPageGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page group id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	spL := d.Get("status_page").([]any)
	statusPageID := int64(-1)
	if len(spL) > 0 && spL[0] != nil {
		sp := spL[0].(map[string]any)
		id := int64(sp["id"].(int))
		statusPageID = id
	}
	log.Printf("[DEBUG] Reading status page group: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetStatusPageGroup(&ilert.GetStatusPageGroupInput{StatusPageGroupID: ilert.Int64(statusPageGroupID), StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert status page group error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a status page group with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert status page group error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformStatusPageGroupResource(statusPageGroup *ilert.StatusPageGroup, statusPageID int64, d *schema.ResourceData) error {
	d.Set("name", statusPageGroup.Name)

	sp := make([]any, 0)
	s := make(map[string]any, 0)
	s["id"] = int(statusPageID)
	sp = append(sp, s)
	d.Set("status_page", sp)

	return nil
}
