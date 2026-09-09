package ilert

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v3"
)

// A service dependency is an edge of the service dependency graph, and the API gives it
// its own id and its own endpoints: the service create and update payloads do not accept
// dependencies at all. It is modelled as its own resource for the same reason as the
// alert action / alert source attachment, and because the only bulk write the API offers
// replaces every dependency of a service at once, which would clobber edges created
// outside this configuration.
//
// The API has no update for a single edge, so every field forces a new resource. An edge
// is cheap to recreate: deleting and re-adding it changes nothing about either service.
// The bulk replace can change a type or a note in place, but it rewrites every dependency
// of the service at once, which would drop edges this resource does not know about.
func resourceServiceDependency() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"service_id": {
				// the service that depends on the target service
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"target_service_id": {
				// the service that is being depended upon
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"type": {
				// the strength of the dependency. Absent from the published API spec,
				// but a real enum the API validates and defaults to HARD.
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.ServiceDependencyTypeAll, false),
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"invalid_after": {
				// Read-only: the API documents it as writable but ignores it on create
				// and on the bulk replace, in every date format, and always returns null.
				// Declaring it as optional would make every plan want to recreate the
				// dependency forever, since the value could never read back.
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceServiceDependencyCreate,
		ReadContext:   resourceServiceDependencyRead,
		DeleteContext: resourceServiceDependencyDelete,
		Exists:        resourceServiceDependencyExists,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceDependencyImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceServiceDependencyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	serviceID := int64(d.Get("service_id").(int))
	dependency := &ilert.ServiceDependency{
		TargetServiceID: int64(d.Get("target_service_id").(int)),
	}
	if val, ok := d.GetOk("type"); ok {
		dependency.Type = val.(string)
	}
	if val, ok := d.GetOk("notes"); ok {
		dependency.Notes = val.(string)
	}

	log.Printf("[INFO] Creating dependency of service %d on service %d", serviceID, dependency.TargetServiceID)

	result := &ilert.CreateServiceDependencyOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateServiceDependency(&ilert.CreateServiceDependencyInput{ServiceID: ilert.Int64(serviceID), Dependency: dependency})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert service dependency error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service dependency to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert service dependency error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Dependency == nil {
		log.Printf("[ERROR] Creating ilert service dependency error: empty response")
		return diag.Errorf("service dependency response is empty")
	}

	d.SetId(fmt.Sprintf("%d/%d", serviceID, result.Dependency.ID))

	return resourceServiceDependencyRead(ctx, d, m)
}

func resourceServiceDependencyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	serviceID, edgeID, err := parseServiceDependencyID(d.Id())
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	log.Printf("[DEBUG] Reading service dependency: %s", d.Id())
	result := &ilert.GetServiceDependencyOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetServiceDependency(&ilert.GetServiceDependencyInput{ServiceID: ilert.Int64(serviceID), EdgeID: ilert.Int64(edgeID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing service dependency %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service dependency with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a service dependency with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert service dependency error: %s", err.Error())
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		return nil
	}

	if result == nil || result.Dependency == nil {
		log.Printf("[ERROR] Reading ilert service dependency error: empty response")
		return diag.Errorf("service dependency response is empty")
	}

	d.Set("service_id", serviceID)
	// The endpoint reports the edge from the perspective of the service in the path, but
	// only fills sourceServiceId in some responses, so the path id is what state keeps.
	d.Set("target_service_id", result.Dependency.TargetServiceID)
	d.Set("type", result.Dependency.Type)
	d.Set("notes", result.Dependency.Notes)
	d.Set("invalid_after", result.Dependency.InvalidAfter)
	d.Set("created_at", result.Dependency.CreatedAt)
	d.Set("updated_at", result.Dependency.UpdatedAt)

	return nil
}

func resourceServiceDependencyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	serviceID, edgeID, err := parseServiceDependencyID(d.Id())
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	log.Printf("[DEBUG] Deleting service dependency: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.DeleteServiceDependency(&ilert.DeleteServiceDependencyInput{ServiceID: ilert.Int64(serviceID), EdgeID: ilert.Int64(edgeID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				// already gone, deleting it is what we wanted anyway
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service dependency with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a service dependency with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert service dependency error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceServiceDependencyExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	serviceID, edgeID, err := parseServiceDependencyID(d.Id())
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}

	log.Printf("[DEBUG] Checking service dependency exists: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetServiceDependency(&ilert.GetServiceDependencyInput{ServiceID: ilert.Int64(serviceID), EdgeID: ilert.Int64(edgeID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert service dependency error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service dependency to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a service dependency with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert service dependency error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func resourceServiceDependencyImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	serviceID, _, err := parseServiceDependencyID(d.Id())
	if err != nil {
		return nil, err
	}
	d.Set("service_id", serviceID)
	return []*schema.ResourceData{d}, nil
}

// parseServiceDependencyID splits the "<service id>/<edge id>" resource id. The edge id
// is the id of the dependency itself, not the id of either service.
func parseServiceDependencyID(id string) (serviceID int64, edgeID int64, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, fmt.Errorf("expected an id of the form <service id>/<dependency id>, got %q", id)
	}

	serviceID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("expected a numeric service id in %q: %s", id, err.Error())
	}

	edgeID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("expected a numeric dependency id in %q: %s", id, err.Error())
	}

	return serviceID, edgeID, nil
}
