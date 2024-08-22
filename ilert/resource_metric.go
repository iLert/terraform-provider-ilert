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

func resourceMetric() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aggregation_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.MetricAggregationTypeAll, false),
			},
			"display_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.MetricDisplayTypeAll, false),
			},
			"interpolate_gaps": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"lock_y_axis_max": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"lock_y_axis_min": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"mouse_over_decimal": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"show_values_on_mouse_over": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"team": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"unit_label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metadata": {
				Type:         schema.TypeList,
				MinItems:     1,
				MaxItems:     1,
				Optional:     true,
				RequiredWith: []string{"data_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:        schema.TypeString,
							Description: "Used for Datadog, Prometheus",
							Required:    true,
						},
					},
				},
			},
			"data_source": {
				Type:         schema.TypeList,
				MinItems:     1,
				MaxItems:     1,
				Optional:     true,
				RequiredWith: []string{"metadata"},
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
		CreateContext: resourceMetricCreate,
		ReadContext:   resourceMetricRead,
		UpdateContext: resourceMetricUpdate,
		DeleteContext: resourceMetricDelete,
		Exists:        resourceMetricExists,
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

func buildMetric(d *schema.ResourceData) (*ilert.Metric, error) {
	name := d.Get("name").(string)
	aggregationType := d.Get("aggregation_type").(string)
	displayType := d.Get("display_type").(string)

	metric := &ilert.Metric{
		Name:            name,
		AggregationType: aggregationType,
		DisplayType:     displayType,
	}

	if val, ok := d.GetOk("description"); ok {
		metric.Description = val.(string)
	}

	if val, ok := d.GetOk("interpolate_gaps"); ok {
		metric.InterpolateGaps = val.(bool)
	}

	if val, ok := d.GetOk("lock_y_axis_max"); ok {
		metric.LockYAxisMax = val.(float64)
	}

	if val, ok := d.GetOk("lock_y_axis_min"); ok {
		metric.LockYAxisMin = val.(float64)
	}

	if val, ok := d.GetOk("mouse_over_decimal"); ok {
		metric.MouseOverDecimal = float64(val.(int))
	}

	if val, ok := d.GetOk("show_values_on_mouse_over"); ok {
		metric.ShowValuesOnMouseOver = val.(bool)
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]interface{})
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		metric.Teams = tms
	}

	if val, ok := d.GetOk("unit_label"); ok {
		metric.UnitLabel = val.(string)
	}

	if val, ok := d.GetOk("metadata"); ok {
		vL := val.([]interface{})
		v := vL[0].(map[string]interface{})
		mt := &ilert.MetricProviderMetadata{}
		if v["query"] != nil && v["query"].(string) != "" {
			mt.Query = v["query"].(string)
		}
		metric.Metadata = mt
	}

	if val, ok := d.GetOk("data_source"); ok {
		vL := val.([]interface{})
		v := vL[0].(map[string]interface{})
		ds := &ilert.MetricDataSource{}
		if v["id"] != nil && v["id"].(int) > 0 {
			ds.ID = int64(v["id"].(int))
		}
		metric.DataSource = ds
	}

	return metric, nil
}

func resourceMetricCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	metric, err := buildMetric(d)
	if err != nil {
		log.Printf("[ERROR] Building metric error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating metric %s", metric.Name)

	result := &ilert.CreateMetricOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateMetric(&ilert.CreateMetricInput{Metric: metric})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert metric error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a metric with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert metric error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Metric == nil {
		log.Printf("[ERROR] Creating ilert metric error: empty response")
		return diag.Errorf("metric response is empty")
	}

	d.SetId(strconv.FormatInt(result.Metric.ID, 10))

	return resourceMetricRead(ctx, d, m)
}

func resourceMetricRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading metric: %s", d.Id())
	result := &ilert.GetMetricOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetMetric(&ilert.GetMetricInput{MetricID: ilert.Int64(metricID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing metric %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an metric with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert metric error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Metric == nil {
		log.Printf("[ERROR] Reading ilert metric error: empty response")
		return diag.Errorf("metric response is empty")
	}

	d.Set("name", result.Metric.Name)
	d.Set("aggregation_type", result.Metric.AggregationType)
	d.Set("display_type", result.Metric.DisplayType)
	d.Set("interpolate_gaps", result.Metric.InterpolateGaps)
	d.Set("lock_y_axis_max", result.Metric.LockYAxisMax)
	d.Set("lock_y_axis_min", result.Metric.LockYAxisMin)
	d.Set("mouse_over_decimal", int(result.Metric.MouseOverDecimal))
	d.Set("show_values_on_mouse_over", result.Metric.ShowValuesOnMouseOver)

	teams, err := flattenTeamShortList(result.Metric.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	d.Set("unit_label", result.Metric.UnitLabel)

	if result.Metric.Metadata != nil {
		d.Set("metadata", []interface{}{map[string]interface{}{
			"query": result.Metric.Metadata.Query,
		}})
	}

	if result.Metric.DataSource != nil {
		d.Set("data_source", []interface{}{map[string]interface{}{
			"id": int(result.Metric.DataSource.ID),
		}})
	}

	return nil
}

func resourceMetricUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	metric, err := buildMetric(d)
	if err != nil {
		log.Printf("[ERROR] Building metric error %s", err.Error())
		return diag.FromErr(err)
	}

	metricID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating metric: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateMetric(&ilert.UpdateMetricInput{Metric: metric, MetricID: ilert.Int64(metricID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an metric with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert metric error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceMetricRead(ctx, d, m)
}

func resourceMetricDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting metric: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteMetric(&ilert.DeleteMetricInput{MetricID: ilert.Int64(metricID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an metric with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert metric error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceMetricExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	metricID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading metric: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetMetric(&ilert.GetMetricInput{MetricID: ilert.Int64(metricID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert metric error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a metric with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert metric error: %s", err.Error())
		return false, err
	}
	return result, nil
}
