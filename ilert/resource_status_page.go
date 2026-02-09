package ilert

import (
	"context"
	"errors"
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

func resourceStatusPage() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subdomain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_css": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Don't use this field yet.",
			},
			"favicon_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"logo_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"visibility": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.StatusPageVisibilityAll, false),
			},
			"hidden_from_search": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"show_subscribe_action": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"show_incident_history_option": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"page_title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"page_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"page_layout": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ilert.StatusPageLayout.SingleColumn,
				ValidateFunc: validation.StringInSlice(ilert.StatusPageLayoutAll, false),
			},
			"logo_redirect_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"activated": {
				Type:     schema.TypeBool,
				Computed: true,
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
			"service": {
				Type:     schema.TypeList,
				Required: true,
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
			"ip_whitelist": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"account_wide_view": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"structure": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Description: "Please make sure to follow the instructions for setting a structure with Terraform in the status page group example.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"element": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ilert.StatusPageElementTypeAll, false),
									},
									"options": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(ilert.StatusPageElementOptionAll, false),
										},
									},
									"child": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{ilert.StatusPageElementTypeAll[0]}, false),
												},
												"options": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ilert.StatusPageElementOptionAll, false),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"theme_mode": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(ilert.StatusPageAppearanceAll, false),
				Description:   "Please use `appearance` instead.",
				ConflictsWith: []string{"appearance"},
			},
			"appearance": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(ilert.StatusPageAppearanceAll, false),
				ConflictsWith: []string{"theme_mode"},
			},
			"email_whitelist": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"announcement": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"announcement_on_page": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"announcement_in_widget": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"metric": {
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
		},
		CreateContext: resourceStatusPageCreate,
		ReadContext:   resourceStatusPageRead,
		UpdateContext: resourceStatusPageUpdate,
		DeleteContext: resourceStatusPageDelete,
		Exists:        resourceStatusPageExists,
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

func buildStatusPage(d *schema.ResourceData) (*ilert.StatusPage, error) {
	name := d.Get("name").(string)
	visibility := d.Get("visibility").(string)
	subdomain := d.Get("subdomain").(string)

	statusPage := &ilert.StatusPage{
		Name:       name,
		Visibility: visibility,
		Subdomain:  subdomain,
	}

	if val, ok := d.GetOk("domain"); ok {
		statusPage.Domain = val.(string)
	}

	if val, ok := d.GetOk("timezone"); ok {
		statusPage.Timezone = val.(string)
	}

	if val, ok := d.GetOk("custom_css"); ok {
		statusPage.CustomCss = val.(string)
	}

	if val, ok := d.GetOk("favicon_url"); ok {
		statusPage.FaviconUrl = val.(string)
	}

	if val, ok := d.GetOk("logo_url"); ok {
		statusPage.LogoUrl = val.(string)
	}

	if val, ok := d.GetOk("hidden_from_search"); ok {
		statusPage.HiddenFromSearch = val.(bool)
	}

	if val, ok := d.GetOk("show_subscribe_action"); ok {
		statusPage.ShowSubscribeAction = val.(bool)
	}

	if val, ok := d.GetOk("show_incident_history_option"); ok {
		statusPage.ShowIncidentHistoryOption = val.(bool)
	}

	if val, ok := d.GetOk("page_title"); ok {
		statusPage.PageTitle = val.(string)
	}

	if val, ok := d.GetOk("page_description"); ok {
		statusPage.PageDescription = val.(string)
	}

	if val, ok := d.GetOk("page_layout"); ok {
		statusPage.PageLayout = val.(string)
	}

	if val, ok := d.GetOk("logo_redirect_url"); ok {
		statusPage.LogoRedirectUrl = val.(string)
	}

	if val, ok := d.GetOk("activated"); ok {
		statusPage.Activated = val.(bool)
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
		statusPage.Teams = tms
	}

	if val, ok := d.GetOk("service"); ok {
		vL := val.([]any)
		svc := make([]ilert.Service, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			sv := ilert.Service{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				sv.Name = v["name"].(string)
			}
			svc = append(svc, sv)
		}
		statusPage.Services = svc
	}

	if val, ok := d.GetOk("ip_whitelist"); ok {
		if statusPage.Visibility == ilert.StatusPageVisibility.Private {
			vL := val.([]any)
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			statusPage.IpWhitelist = sL
		} else {
			return nil, fmt.Errorf("[ERROR] Can't set ip whitelist, as it is only allowed on private status pages")
		}
	}

	if val, ok := d.GetOk("account_wide_view"); ok {
		statusPage.AccountWideView = val.(bool)
	}

	if val, ok := d.GetOk("structure"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 && vL[0] != nil {
			if v, ok := vL[0].(map[string]any); ok && len(v) > 0 {
				st := ilert.StatusPageStructure{}

				elm := make([]ilert.StatusPageElement, 0)
				eL := v["element"].([]any)
				for _, e := range eL {
					v := e.(map[string]any)
					eid, err := strconv.ParseInt(v["id"].(string), 10, 64)
					if err != nil {
						log.Printf("[ERROR] Could not parse status page element id %s", err.Error())
						return nil, unconvertibleIDErr(v["id"].(string), err)
					}
					el := ilert.StatusPageElement{
						ID:   eid,
						Type: v["type"].(string),
					}
					if len(v["options"].([]any)) > 0 {
						optionsList := v["options"].([]any)
						options := make([]string, 0)
						for _, value := range optionsList {
							option := value.(string)
							options = append(options, option)
						}
						el.Options = options
					}
					if v["child"] != nil {
						cdr := make([]ilert.StatusPageElement, 0)
						cL := v["child"].([]any)
						for _, c := range cL {
							v := c.(map[string]any)
							cid, err := strconv.ParseInt(v["id"].(string), 10, 64)
							if err != nil {
								log.Printf("[ERROR] Could not parse status page child id %s", err.Error())
								return nil, unconvertibleIDErr(v["id"].(string), err)
							}
							ch := ilert.StatusPageElement{
								ID:   cid,
								Type: v["type"].(string),
							}
							if len(v["options"].([]any)) > 0 {
								optionsList := v["options"].([]any)
								options := make([]string, 0)
								for _, value := range optionsList {
									option := value.(string)
									options = append(options, option)
								}
								ch.Options = options
							}
							if v["child"] != nil && ch.Type == "SERVICE" {
								err = errors.New("[ERROR] Could not set child, as no children are allowed on type service")
								log.Print(err.Error())
								return nil, err
							}
							cdr = append(cdr, ch)
						}
						el.Children = cdr
					}
					elm = append(elm, el)
				}
				st.Elements = elm
				statusPage.Structure = &st
			}
		}
	}

	if val, ok := d.GetOk("theme_mode"); ok {
		statusPage.Appearance = val.(string)
	}

	if val, ok := d.GetOk("appearance"); ok {
		statusPage.Appearance = val.(string)
	}

	if val, ok := d.GetOk("email_whitelist"); ok {
		if statusPage.Visibility == ilert.StatusPageVisibility.Private {
			vL := val.([]any)
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			statusPage.EmailWhitelist = sL
		} else {
			return nil, fmt.Errorf("[ERROR] Can't set email whitelist, as it is only allowed on private status pages")
		}
	}

	if val, ok := d.GetOk("announcement"); ok {
		statusPage.Announcement = val.(string)
	}

	if val, ok := d.GetOk("announcement_on_page"); ok {
		statusPage.AnnouncementOnPage = val.(bool)
	}

	if val, ok := d.GetOk("announcement_in_widget"); ok {
		statusPage.AnnouncementInWidget = val.(bool)
	}

	if val, ok := d.GetOk("metric"); ok {
		vL := val.([]any)
		mts := make([]ilert.Metric, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			mt := ilert.Metric{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				mt.Name = v["name"].(string)
			}
			mts = append(mts, mt)
		}
		statusPage.Metrics = mts
	}

	return statusPage, nil
}

func resourceStatusPageCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPage, err := buildStatusPage(d)
	if err != nil {
		log.Printf("[ERROR] Building status page error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating status page %s", statusPage.Name)

	result := &ilert.CreateStatusPageOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateStatusPage(&ilert.CreateStatusPageInput{StatusPage: statusPage})
		if err != nil {
			log.Printf("[DEBUG] Creating status page error occurred: %s", err.Error())
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert status page error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a status page with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert status page error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.StatusPage == nil {
		log.Printf("[ERROR] Creating ilert status page error: empty response")
		return diag.Errorf("status page response is empty")
	}

	d.SetId(strconv.FormatInt(result.StatusPage.ID, 10))

	return resourceStatusPageRead(ctx, d, m)
}

func resourceStatusPageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading status page: %s", d.Id())
	result := &ilert.GetStatusPageOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetStatusPage(&ilert.GetStatusPageInput{StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing status page %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an status page with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert status page error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.StatusPage == nil {
		log.Printf("[ERROR] Reading ilert status page error: empty response")
		return diag.Errorf("status page response is empty")
	}

	d.Set("name", result.StatusPage.Name)
	d.Set("domain", result.StatusPage.Domain)
	d.Set("subdomain", result.StatusPage.Subdomain)

	if val, ok := d.GetOk("timezone"); ok && val != nil && val.(string) != "" {
		d.Set("timezone", result.StatusPage.Timezone)
	}

	d.Set("custom_css", result.StatusPage.CustomCss)
	d.Set("favicon_url", result.StatusPage.FaviconUrl)
	d.Set("logo_url", result.StatusPage.LogoUrl)
	d.Set("visibility", result.StatusPage.Visibility)

	if val, ok := d.GetOk("hidden_from_search"); ok && val != nil {
		d.Set("hidden_from_search", result.StatusPage.HiddenFromSearch)
	}

	d.Set("show_subscribe_action", result.StatusPage.ShowSubscribeAction)
	d.Set("show_incident_history_option", result.StatusPage.ShowIncidentHistoryOption)
	d.Set("page_title", result.StatusPage.PageTitle)
	d.Set("page_description", result.StatusPage.PageDescription)
	d.Set("page_layout", result.StatusPage.PageLayout)
	d.Set("logo_redirect_url", result.StatusPage.LogoRedirectUrl)
	d.Set("activated", result.StatusPage.Activated)

	teams, err := flattenTeamShortList(result.StatusPage.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	services, err := flattenServicesList(result.StatusPage.Services, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service", services); err != nil {
		return diag.Errorf("error setting services: %s", err)
	}

	if val, ok := d.GetOk("ip_whitelist"); ok && val.([]any) != nil && len(val.([]any)) > 0 {
		d.Set("ip_whitelist", result.StatusPage.IpWhitelist)
	}

	d.Set("account_wide_view", result.StatusPage.AccountWideView)

	structure, err := flattenStatusPageStructure(result.StatusPage.Structure)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("structure", structure); err != nil {
		return diag.Errorf("error setting structure: %s", err)
	}

	if _, ok := d.GetOk("theme_mode"); ok {
		d.Set("theme_mode", result.StatusPage.Appearance)
	}

	if _, ok := d.GetOk("appearance"); ok {
		d.Set("appearance", result.StatusPage.Appearance)
	}

	if val, ok := d.GetOk("email_whitelist"); ok && val.([]any) != nil && len(val.([]any)) > 0 {
		d.Set("email_whitelist", result.StatusPage.EmailWhitelist)
	}

	if _, ok := d.GetOk("announcement"); ok {
		d.Set("announcement", result.StatusPage.Announcement)
	}

	if _, ok := d.GetOk("announcement_on_page"); ok {
		d.Set("announcement_on_page", result.StatusPage.AnnouncementOnPage)
	}

	if _, ok := d.GetOk("announcement_in_widget"); ok {
		d.Set("announcement_in_widget", result.StatusPage.AnnouncementInWidget)
	}

	metrics, err := flattenMetricsList(result.StatusPage.Metrics, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metric", metrics); err != nil {
		return diag.Errorf("error setting metrics: %s", err)
	}

	return nil
}

func resourceStatusPageUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPage, err := buildStatusPage(d)
	if err != nil {
		log.Printf("[ERROR] Building status page error %s", err.Error())
		return diag.FromErr(err)
	}

	statusPageID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating status page: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateStatusPage(&ilert.UpdateStatusPageInput{StatusPage: statusPage, StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an status page with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert status page error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceStatusPageRead(ctx, d, m)
}

func resourceStatusPageDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	statusPageID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting status page: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteStatusPage(&ilert.DeleteStatusPageInput{StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an status page with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert status page error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceStatusPageExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	statusPageID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse status page id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading status page: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetStatusPage(&ilert.GetStatusPageInput{StatusPageID: ilert.Int64(statusPageID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert status page error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a status page with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert status page error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func flattenServicesList(list []ilert.Service, d *schema.ResourceData) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}
	results := make([]any, 0)
	if val, ok := d.GetOk("service"); ok && val != nil {
		vL := val.([]any)
		for i, item := range list {
			if vL != nil && i < len(vL) && vL[i] != nil {
				result := make(map[string]any)
				v := vL[i].(map[string]any)
				result["id"] = item.ID
				if item.Name != "" && v["name"] != nil && v["name"].(string) != "" {
					result["name"] = item.Name
				}
				results = append(results, result)
			}
		}
	}
	return results, nil
}

func flattenMetricsList(list []ilert.Metric, d *schema.ResourceData) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}
	results := make([]any, 0)
	if val, ok := d.GetOk("metric"); ok && val != nil {
		vL := val.([]any)
		for i, item := range list {
			if vL != nil && i < len(vL) && vL[i] != nil {
				result := make(map[string]any)
				v := vL[i].(map[string]any)
				result["id"] = item.ID
				if item.Name != "" && v["name"] != nil && v["name"].(string) != "" {
					result["name"] = item.Name
				}
				results = append(results, result)
			}
		}
	}
	return results, nil
}

func flattenStatusPageStructure(structure *ilert.StatusPageStructure) ([]any, error) {
	if structure == nil {
		return make([]any, 0), nil
	}
	structureResult := make([]any, 0)
	results := make(map[string]any, 0)
	if len(structure.Elements) > 0 {
		elm := make([]any, 0)
		for _, e := range structure.Elements {
			el := make(map[string]any)
			el["id"] = strconv.FormatInt(e.ID, 10)
			el["type"] = e.Type

			if len(e.Options) > 0 {
				options := make([]any, 0)
				for _, option := range e.Options {
					options = append(options, option)
				}
				el["options"] = options
			}

			if len(e.Children) > 0 {
				chd := make([]any, 0)
				for _, c := range e.Children {
					ch := make(map[string]any)
					ch["id"] = strconv.FormatInt(c.ID, 10)
					ch["type"] = c.Type

					if len(c.Options) > 0 {
						options := make([]any, 0)
						for _, option := range c.Options {
							options = append(options, option)
						}
						ch["options"] = options
					}

					chd = append(chd, ch)
				}
				el["child"] = chd
			}
			elm = append(elm, el)
		}
		results["element"] = elm
	}
	structureResult = append(structureResult, results)

	return structureResult, nil
}
