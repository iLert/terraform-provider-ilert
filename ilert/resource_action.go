package ilert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v4"
)

// Actions are early access. The API merges an update onto the action, and ilert-go
// sends every field the configuration leaves empty empty, so the action ends up as
// configured. The exception is config: some templates write values into it, so an
// update sends the config as it was read with the configuration applied, and the
// state only holds the keys the configuration sets, see actionConfigToSend and
// flattenActionConfig.
func resourceAction() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"template_key": {
				// the API cannot remove a template from an action, and a different
				// template brings a different config and credential
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`), "a template key has up to 80 letters, digits, '-' and '_'"),
			},
			"template_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"entity_kind": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.ActionEntityKindAll, false),
			},
			"dispatch_mode": {
				// the API refuses to change it after creation
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ilert.ActionDispatchMode.Automatic,
				ValidateFunc: validation.StringInSlice(ilert.ActionDispatchModeAll, false),
			},
			"trigger_events": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ilert.ActionTriggerEventsAll, false),
				},
			},
			"trigger_event_delay": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(actionDelayableEvents, false),
						},
						"delay_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(30, 172800),
						},
					},
				},
			},
			"condition_icl": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONObject,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"steps": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONArray,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"field_mapping": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONArray,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"inbound": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJSONObject,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"credential_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bidirectional": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"team": {
				// TypeSet, not TypeList: the API returns teams sorted by id
				Type:     schema.TypeSet,
				Optional: true,
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"integration_url": {
				// carries the key that authenticates the callbacks of a bidirectional
				// action
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"brand": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
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
		CustomizeDiff: validateActionDiff,
		CreateContext: resourceActionCreate,
		ReadContext:   resourceActionRead,
		UpdateContext: resourceActionUpdate,
		DeleteContext: resourceActionDelete,
		Exists:        resourceActionExists,
		Importer: &schema.ResourceImporter{
			StateContext: resourceActionImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

// actionDelayableEvents are the trigger events a delay can be set for
var actionDelayableEvents = []string{"alert.not_resolved", "incident.not_resolved", "alert.escalation_ended"}

// validateActionDiff fails the plan of an action while actions are not in
// early_access_enabled, then checks what the API would reject or store differently:
// an action needs a template or its own steps, its trigger events and delays have to
// be of its entity kind, a not_resolved event needs a delay, an event takes one delay,
// and the API raises a not_resolved delay below 60 seconds to 60 and drops an
// escalation_ended delay below 30
func validateActionDiff(_ context.Context, d *schema.ResourceDiff, m any) error {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return err
	}

	if d.NewValueKnown("template_key") && d.NewValueKnown("steps") &&
		d.Get("template_key").(string) == "" && d.Get("steps").(string) == "" {
		return errors.New("an action needs a template_key, or its own steps")
	}

	kind := d.Get("entity_kind").(string)
	if !d.NewValueKnown("entity_kind") || kind == "" {
		return nil
	}
	prefix := strings.ToLower(kind) + "."
	if d.NewValueKnown("trigger_events") {
		for _, event := range d.Get("trigger_events").(*schema.Set).List() {
			if e, _ := event.(string); e != "" && !strings.HasPrefix(e, prefix) {
				return fmt.Errorf("trigger_events: %s is not an event of an %s action, only %s* events are", e, kind, prefix)
			}
		}
	}
	if !d.NewValueKnown("trigger_event_delay") {
		return nil
	}
	delayed := map[string]bool{}
	for _, delay := range d.Get("trigger_event_delay").([]any) {
		v, ok := delay.(map[string]any)
		if !ok {
			continue
		}
		event, _ := v["event"].(string)
		sec, _ := v["delay_sec"].(int)
		if event == "" {
			continue
		}
		if !strings.HasPrefix(event, prefix) {
			return fmt.Errorf("trigger_event_delay: %s is not an event of an %s action", event, kind)
		}
		if delayed[event] {
			return fmt.Errorf("trigger_event_delay: %s is delayed twice, an event takes one delay", event)
		}
		delayed[event] = true
		switch {
		case strings.HasSuffix(event, ".not_resolved") && sec < 60:
			return fmt.Errorf("trigger_event_delay: %s takes 60 to 172800 seconds, got %d", event, sec)
		case event == "alert.escalation_ended" && sec > 7200:
			return fmt.Errorf("trigger_event_delay: alert.escalation_ended takes 30 to 7200 seconds, got %d", sec)
		}
	}
	if d.NewValueKnown("trigger_events") {
		for _, event := range d.Get("trigger_events").(*schema.Set).List() {
			if e, _ := event.(string); strings.HasSuffix(e, ".not_resolved") && !delayed[e] {
				return fmt.Errorf("trigger_events: %s is raised once an alert or incident is still open after a delay, so it needs a trigger_event_delay", e)
			}
		}
	}
	return nil
}

// buildAction writes the configuration onto an action. The config is the one to
// send, see actionConfigToSend.
func buildAction(d *schema.ResourceData, config json.RawMessage) (*ilert.Action, error) {
	action := &ilert.Action{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Enabled:       ilert.Bool(d.Get("enabled").(bool)),
		TemplateKey:   d.Get("template_key").(string),
		EntityKind:    d.Get("entity_kind").(string),
		DispatchMode:  d.Get("dispatch_mode").(string),
		ConditionICL:  d.Get("condition_icl").(string),
		Config:        config,
		Bidirectional: d.Get("bidirectional").(bool),
	}

	if version := d.Get("template_version").(int); version > 0 {
		action.TemplateVersion = &version
	}

	for _, event := range d.Get("trigger_events").(*schema.Set).List() {
		action.TriggerEvents = append(action.TriggerEvents, event.(string))
	}

	for _, delay := range d.Get("trigger_event_delay").([]any) {
		if v, ok := delay.(map[string]any); ok {
			action.TriggerEventDelays = append(action.TriggerEventDelays, ilert.ActionTriggerEventDelay{
				Event:    v["event"].(string),
				DelaySec: v["delay_sec"].(int),
			})
		}
	}

	for key, field := range map[string]*json.RawMessage{"steps": &action.Steps, "field_mapping": &action.FieldMapping, "inbound": &action.Inbound} {
		if value := d.Get(key).(string); value != "" {
			*field = json.RawMessage(value)
		}
	}

	if id := d.Get("credential_id").(string); id != "" {
		credentialID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse the credential_id %q: %s", id, err.Error())
		}
		action.CredentialID = &credentialID
	}

	for _, m := range d.Get("team").(*schema.Set).List() {
		v := m.(map[string]any)
		team := ilert.TeamShort{ID: int64(v["id"].(int))}
		if name, _ := v["name"].(string); name != "" {
			team.Name = name
		}
		action.Teams = append(action.Teams, team)
	}

	if labels := d.Get("labels").(map[string]any); len(labels) > 0 {
		action.Labels = make(map[string]string, len(labels))
		for key, value := range labels {
			action.Labels[key] = value.(string)
		}
	}

	return action, nil
}

// actionConfigToSend is the config an update sends: the config as it was read, with
// the keys the configuration removed taken out and its own keys applied. The API
// replaces the config with what an update sends, so the values a template wrote into
// it go back as they were.
func actionConfigToSend(read json.RawMessage, d *schema.ResourceData) (json.RawMessage, error) {
	old, configured := d.GetChange("config")
	return mergeActionConfig(read, old.(string), configured.(string))
}

// mergeActionConfig applies the configured config onto the one read, and takes out
// the keys the state held that the configuration no longer sets
func mergeActionConfig(read json.RawMessage, old, configured string) (json.RawMessage, error) {
	config, err := jsonObject(read)
	if err != nil {
		return nil, fmt.Errorf("could not read the config of the action: %s", err.Error())
	}
	removed, err := jsonObject(json.RawMessage(old))
	if err != nil {
		return nil, err
	}
	set, err := jsonObject(json.RawMessage(configured))
	if err != nil {
		return nil, err
	}
	for key := range removed {
		if _, ok := set[key]; !ok {
			delete(config, key)
		}
	}
	for key, value := range set {
		config[key] = value
	}
	if len(config) == 0 {
		return nil, nil
	}
	return json.Marshal(config)
}

// flattenActionConfig is the config as the state holds it: only the keys the
// configuration sets, so the values a template writes into it do not show as a
// change. None after an import, which has no configuration to take the keys from,
// and all for the IaC exporter, which reads the action without an id.
func flattenActionConfig(read json.RawMessage, d *schema.ResourceData) (string, error) {
	config, err := jsonObject(read)
	if err != nil {
		return "", fmt.Errorf("could not read the config of the action: %s", err.Error())
	}
	if d.Id() == "" {
		if len(config) == 0 {
			return "", nil
		}
		encoded, err := json.Marshal(config)
		return string(encoded), err
	}

	managed := d.Get("config").(string)
	if managed == "" {
		return "", nil
	}
	keys, err := jsonObject(json.RawMessage(managed))
	if err != nil {
		return "", err
	}
	result := make(map[string]json.RawMessage, len(keys))
	for key := range keys {
		if value, ok := config[key]; ok {
			result[key] = value
		}
	}
	encoded, err := json.Marshal(result)
	return string(encoded), err
}

func resourceActionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	var config json.RawMessage
	if value := d.Get("config").(string); value != "" {
		config = json.RawMessage(value)
	}
	action, err := buildAction(d, config)
	if err != nil {
		log.Printf("[ERROR] Building action error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating action %s", action.Name)

	result := &ilert.CreateActionOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateAction(&ilert.CreateActionInput{Action: action})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert action error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for action to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(actionsUnavailable(err))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert action error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Action == nil {
		log.Printf("[ERROR] Creating ilert action error: empty response")
		return diag.Errorf("action response is empty")
	}

	d.SetId(strconv.FormatInt(result.Action.ID, 10))

	return resourceActionRead(ctx, d, m)
}

// fetchAction reads the action with retries. A missing action is returned as the
// *ilert.NotFoundAPIError of the API, and an account without actions as an error.
func fetchAction(ctx context.Context, client *ilert.Client, id string, timeout time.Duration) (*ilert.Action, error) {
	actionID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse action id %s", err.Error())
		return nil, unconvertibleIDErr(id, err)
	}

	var action *ilert.Action
	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		r, err := client.GetAction(&ilert.GetActionInput{ActionID: ilert.Int64(actionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for action with id '%s' to be read, error: %s", id, err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		action = r.Action
		return nil
	})
	if err != nil {
		return nil, actionsUnavailable(err)
	}
	if action == nil {
		return nil, errors.New("action response is empty")
	}
	return action, nil
}

// actionsUnavailable turns the 404 of an account without actions into an error that
// says so, instead of the not found that would remove the action from the state
func actionsUnavailable(err error) error {
	var notFound *ilert.NotFoundAPIError
	if errors.As(err, &notFound) && strings.Contains(notFound.Message, "not available") {
		return fmt.Errorf("actions are not available for this ilert account, they are enabled account by account while they are in early access: %s", notFound.Message)
	}
	return err
}

// assertNativeAction refuses the mirror of an alert action: it cannot be updated,
// and deleting it deletes the alert action
func assertNativeAction(action *ilert.Action) error {
	if action.Origin == ilert.ActionOrigin.LegacyRef {
		return fmt.Errorf("action %d mirrors the alert action %d and cannot be managed as an ilert_action, deleting it would delete the alert action: manage it with ilert_alert_action instead", action.ID, action.LegacyAlertActionID)
	}
	return nil
}

func resourceActionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	log.Printf("[DEBUG] Reading action: %s", d.Id())
	action, err := fetchAction(ctx, client, d.Id(), d.Timeout(schema.TimeoutRead))
	if err != nil {
		var notFound *ilert.NotFoundAPIError
		if errors.As(err, &notFound) {
			log.Printf("[WARN] Removing action %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		log.Printf("[ERROR] Reading ilert action error: %s", err.Error())
		return diag.Errorf("could not read an action with ID %s, error: %s", d.Id(), err.Error())
	}
	if err := assertNativeAction(action); err != nil {
		return diag.FromErr(err)
	}

	if err := transformActionResource(action, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceActionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	actionID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse action id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	read, err := fetchAction(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		log.Printf("[ERROR] Reading ilert action before updating it error: %s", err.Error())
		return diag.Errorf("could not read the action with ID %s to update it, error: %s", d.Id(), err.Error())
	}
	if err := assertNativeAction(read); err != nil {
		return diag.FromErr(err)
	}

	config, err := actionConfigToSend(read.Config, d)
	if err != nil {
		return diag.FromErr(err)
	}
	action, err := buildAction(d, config)
	if err != nil {
		log.Printf("[ERROR] Building action error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Updating action: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateAction(&ilert.UpdateActionInput{Action: action, ActionID: ilert.Int64(actionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for action with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an action with ID %s, error: %s", d.Id(), actionsUnavailable(err).Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Updating ilert action error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceActionRead(ctx, d, m)
}

func resourceActionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	actionID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse action id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting action: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteAction(&ilert.DeleteActionInput{ActionID: ilert.Int64(actionID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok && actionsUnavailable(err) == err {
				// already gone, deleted in the web app for example
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for action with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an action with ID %s, error: %s", d.Id(), actionsUnavailable(err).Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert action error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceActionExists(d *schema.ResourceData, m any) (bool, error) {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return false, err
	}

	client := m.(*ilert.Client)

	_, err := fetchAction(context.Background(), client, d.Id(), d.Timeout(schema.TimeoutRead))
	if err != nil {
		var notFound *ilert.NotFoundAPIError
		if errors.As(err, &notFound) {
			return false, nil
		}
		log.Printf("[ERROR] Reading ilert action error: %s", err.Error())
		return false, err
	}
	return true, nil
}

// resourceActionImport refuses the mirror of an alert action. The config is not
// captured: its keys are managed once the configuration sets them, so the first
// apply cannot take out a value a template wrote into it.
func resourceActionImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	if err := requireEarlyAccess(m, "ilert_action"); err != nil {
		return nil, err
	}

	action, err := fetchAction(ctx, m.(*ilert.Client), d.Id(), d.Timeout(schema.TimeoutRead))
	if err != nil {
		return nil, fmt.Errorf("could not import the action with ID %s, error: %s", d.Id(), err.Error())
	}
	if err := assertNativeAction(action); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func transformActionResource(action *ilert.Action, d *schema.ResourceData) error {
	d.Set("name", action.Name)
	d.Set("description", action.Description)
	d.Set("enabled", action.Enabled == nil || *action.Enabled)
	d.Set("template_key", action.TemplateKey)
	version := 0
	if action.TemplateVersion != nil {
		version = *action.TemplateVersion
	}
	d.Set("template_version", version)
	d.Set("entity_kind", action.EntityKind)
	d.Set("dispatch_mode", action.DispatchMode)
	d.Set("condition_icl", action.ConditionICL)
	d.Set("bidirectional", action.Bidirectional)
	d.Set("integration_url", action.IntegrationURL)
	d.Set("brand", action.Brand)
	d.Set("status", action.Status)
	d.Set("created_at", action.CreatedAt)
	d.Set("updated_at", action.UpdatedAt)

	if err := d.Set("trigger_events", action.TriggerEvents); err != nil {
		return fmt.Errorf("[ERROR] Error setting trigger events: %s", err.Error())
	}
	delays := make([]any, 0, len(action.TriggerEventDelays))
	for _, delay := range action.TriggerEventDelays {
		delays = append(delays, map[string]any{"event": delay.Event, "delay_sec": delay.DelaySec})
	}
	if err := d.Set("trigger_event_delay", delays); err != nil {
		return fmt.Errorf("[ERROR] Error setting trigger event delays: %s", err.Error())
	}

	config, err := flattenActionConfig(action.Config, d)
	if err != nil {
		return err
	}
	d.Set("config", config)
	for key, value := range map[string]json.RawMessage{"steps": action.Steps, "field_mapping": action.FieldMapping, "inbound": action.Inbound} {
		d.Set(key, jsonString(value))
	}

	credentialID := ""
	if action.CredentialID != nil {
		credentialID = strconv.FormatInt(*action.CredentialID, 10)
	}
	d.Set("credential_id", credentialID)

	teams, err := flattenTeamShortList(action.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}
	if err := d.Set("labels", action.Labels); err != nil {
		return fmt.Errorf("[ERROR] Error setting labels: %s", err.Error())
	}

	return nil
}

// jsonString is a JSON value as the state holds it, empty for null or an empty
// object or array, which the API treats as not set
func jsonString(value json.RawMessage) string {
	s := strings.TrimSpace(string(value))
	switch s {
	case "", "null", "{}", "[]":
		return ""
	}
	return s
}

// jsonObject decodes a JSON object, an empty one for an empty value or null
func jsonObject(value json.RawMessage) (map[string]json.RawMessage, error) {
	result := map[string]json.RawMessage{}
	if s := strings.TrimSpace(string(value)); s == "" || s == "null" {
		return result, nil
	}
	if err := json.Unmarshal(value, &result); err != nil {
		return nil, fmt.Errorf("not a JSON object: %s", err.Error())
	}
	return result, nil
}

func validateJSONObject(v any, key string) ([]string, []error) {
	var value map[string]any
	if err := json.Unmarshal([]byte(v.(string)), &value); err != nil {
		return nil, []error{fmt.Errorf("%s must be a JSON object, such as jsonencode({ ... }): %s", key, err.Error())}
	}
	return nil, nil
}

func validateJSONArray(v any, key string) ([]string, []error) {
	var value []any
	if err := json.Unmarshal([]byte(v.(string)), &value); err != nil {
		return nil, []error{fmt.Errorf("%s must be a JSON array, such as jsonencode([ ... ]): %s", key, err.Error())}
	}
	return nil, nil
}

// suppressEquivalentJSON ignores a change in the formatting or key order of a JSON
// value, and treats an empty object or array like no value
func suppressEquivalentJSON(_, old, new string, _ *schema.ResourceData) bool {
	old, new = jsonString(json.RawMessage(old)), jsonString(json.RawMessage(new))
	if old == new {
		return true
	}
	if old == "" || new == "" {
		return false
	}
	var o, n any
	if json.Unmarshal([]byte(old), &o) != nil || json.Unmarshal([]byte(new), &n) != nil {
		return false
	}
	return reflect.DeepEqual(o, n)
}
