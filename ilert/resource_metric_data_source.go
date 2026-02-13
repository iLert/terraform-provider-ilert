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

func resourceMetricDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.MetricDataSourceTypeAll, false),
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
			"metadata": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:        schema.TypeString,
							Description: "Used for Datadog",
							Optional:    true,
						},
						"api_key": {
							Type:        schema.TypeString,
							Description: "Used for Datadog",
							Optional:    true,
						},
						"application_key": {
							Type:        schema.TypeString,
							Description: "Used for Datadog",
							Optional:    true,
						},
						"auth_type": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
						"basic_user": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
						"basic_pass": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
						"header_key": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
						"header_value": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
						"url": {
							Type:        schema.TypeString,
							Description: "Used for Prometheus",
							Optional:    true,
						},
					},
				},
			},
		},
		CreateContext: resourceMetricDataSourceCreate,
		ReadContext:   resourceMetricDataSourceRead,
		UpdateContext: resourceMetricDataSourceUpdate,
		DeleteContext: resourceMetricDataSourceDelete,
		Exists:        resourceMetricDataSourceExists,
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

func buildMetricDataSource(d *schema.ResourceData) (*ilert.MetricDataSource, error) {
	name := d.Get("name").(string)
	dataSourceType := d.Get("type").(string)

	metricDataSource := &ilert.MetricDataSource{
		Name: name,
		Type: dataSourceType,
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]any)
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		metricDataSource.Teams = tms
	}

	if val, ok := d.GetOk("metadata"); ok {
		vL := val.([]any)
		v := vL[0].(map[string]any)
		mt := ilert.MetricDataSourceMetadata{}
		if v["region"] != nil && v["region"].(string) != "" {
			mt.Region = v["region"].(string)
		}
		if v["api_key"] != nil && v["api_key"].(string) != "" {
			mt.ApiKey = v["api_key"].(string)
		}
		if v["application_key"] != nil && v["application_key"].(string) != "" {
			mt.ApplicationKey = v["application_key"].(string)
		}
		if v["auth_type"] != nil && v["auth_type"].(string) != "" {
			mt.AuthType = v["auth_type"].(string)
		}
		if v["basic_user"] != nil && v["basic_user"].(string) != "" {
			mt.BasicUser = v["basic_user"].(string)
		}
		if v["basic_pass"] != nil && v["basic_pass"].(string) != "" {
			mt.BasicPass = v["basic_pass"].(string)
		}
		if v["header_key"] != nil && v["header_key"].(string) != "" {
			mt.HeaderKey = v["header_key"].(string)
		}
		if v["header_value"] != nil && v["header_value"].(string) != "" {
			mt.HeaderValue = v["header_value"].(string)
		}
		if v["url"] != nil && v["url"].(string) != "" {
			mt.Url = v["url"].(string)
		}

		metricDataSource.Metadata = &mt
	}

	return metricDataSource, nil
}

func resourceMetricDataSourceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricDataSource, err := buildMetricDataSource(d)
	if err != nil {
		log.Printf("[ERROR] Building metric data source error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating metric data source %s", metricDataSource.Name)

	result := &ilert.CreateMetricDataSourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateMetricDataSource(&ilert.CreateMetricDataSourceInput{MetricDataSource: metricDataSource})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert metric data source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert metric data source error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.MetricDataSource == nil {
		log.Printf("[ERROR] Creating ilert metric data source error: empty response")
		return diag.Errorf("metric data source response is empty")
	}

	d.SetId(strconv.FormatInt(result.MetricDataSource.ID, 10))

	return resourceMetricDataSourceRead(ctx, d, m)
}

func resourceMetricDataSourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricDataSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric data source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading metric data source: %s", d.Id())
	result := &ilert.GetMetricDataSourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetMetricDataSource(&ilert.GetMetricDataSourceInput{MetricDataSourceID: ilert.Int64(metricDataSourceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing metric data source %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an metric data source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert metric data source error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.MetricDataSource == nil {
		log.Printf("[ERROR] Reading ilert metric data source error: empty response")
		return diag.Errorf("metric data source response is empty")
	}

	err = transformMetricDataSourceResource(result.MetricDataSource, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMetricDataSourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricdatasource, err := buildMetricDataSource(d)
	if err != nil {
		log.Printf("[ERROR] Building metric data source error %s", err.Error())
		return diag.FromErr(err)
	}

	metricdatasourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric data source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating metric data source: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateMetricDataSource(&ilert.UpdateMetricDataSourceInput{MetricDataSource: metricdatasource, MetricDataSourceID: ilert.Int64(metricdatasourceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an metric data source with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert metric data source error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceMetricDataSourceRead(ctx, d, m)
}

func resourceMetricDataSourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	metricdatasourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric data source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting metric data source: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteMetricDataSource(&ilert.DeleteMetricDataSourceInput{MetricDataSourceID: ilert.Int64(metricdatasourceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an metric data source with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert metric data source error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceMetricDataSourceExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	metricdatasourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse metric data source id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading metric data source: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetMetricDataSource(&ilert.GetMetricDataSourceInput{MetricDataSourceID: ilert.Int64(metricdatasourceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert metric data source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a metric data source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert metric data source error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformMetricDataSourceResource(metricDataSource *ilert.MetricDataSource, d *schema.ResourceData) error {
	d.Set("name", metricDataSource.Name)
	d.Set("type", metricDataSource.Type)

	teams, err := flattenTeamShortList(metricDataSource.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	metadata, err := flattenProviderMetadata(metricDataSource.Metadata)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening metadata: %s", err.Error())
	}
	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("[ERROR] Error setting metadata: %s", err.Error())
	}

	return nil
}

func flattenProviderMetadata(metadata *ilert.MetricDataSourceMetadata) ([]any, error) {
	if metadata == nil {
		return make([]any, 0), nil
	}
	results := make([]any, 0)
	result := make(map[string]any, 0)
	if metadata.Region != "" {
		result["region"] = metadata.Region
	}
	if metadata.ApiKey != "" {
		result["api_key"] = metadata.ApiKey
	}
	if metadata.ApplicationKey != "" {
		result["application_key"] = metadata.ApplicationKey
	}
	if metadata.AuthType != "" {
		result["auth_type"] = metadata.AuthType
	}
	if metadata.BasicUser != "" {
		result["basic_user"] = metadata.BasicUser
	}
	if metadata.BasicPass != "" {
		result["basic_pass"] = metadata.BasicPass
	}
	if metadata.HeaderKey != "" {
		result["header_key"] = metadata.HeaderKey
	}
	if metadata.HeaderValue != "" {
		result["header_value"] = metadata.HeaderValue
	}
	if metadata.Url != "" {
		result["url"] = metadata.Url
	}
	results = append(results, result)

	return results, nil
}
