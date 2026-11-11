package ilert

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v4"
)

// Credentials are early access. The API rejects every update of one, so the
// resource has no update: every argument forces a new credential.
func resourceCredential() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.CredentialTypeAll, false),
			},
			"secret_value": {
				// sent once on create, the API never returns it, so the state keeps
				// the configured value
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"username": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateNoSurroundingWhitespace,
			},
			"host_url": {
				// sensitive because the URL of a NONE credential is the secret itself
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validateCredentialHostURL,
			},
			"idp_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCredentialHostURL,
			},
			"idp_scopes": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"idp_client": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"oauth_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.CredentialOAuthMethodAll, false),
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sub_category": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"api_key_header_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"team": {
				// TypeSet, not TypeList: the API returns teams sorted by id, so an
				// ordered list produces a permanent diff whenever the config order
				// differs from that. ForceNew on the set only covers the number of
				// teams, the nested arguments need it too for a team swapped for
				// another to replace the credential.
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
		},
		CustomizeDiff: validateCredentialDiff,
		CreateContext: resourceCredentialCreate,
		ReadContext:   resourceCredentialRead,
		DeleteContext: resourceCredentialDelete,
		Exists:        resourceCredentialExists,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCredentialImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func buildCredential(d *schema.ResourceData) (*ilert.Credential, error) {
	credential := &ilert.Credential{
		Name:             d.Get("name").(string),
		TypeHint:         d.Get("type").(string),
		SecretValue:      d.Get("secret_value").(string),
		Username:         d.Get("username").(string),
		HostURL:          d.Get("host_url").(string),
		IdpURL:           d.Get("idp_url").(string),
		IdpScopes:        d.Get("idp_scopes").(string),
		IdpClient:        d.Get("idp_client").(string),
		Audience:         d.Get("audience").(string),
		OAuthMethod:      d.Get("oauth_method").(string),
		Category:         d.Get("category").(string),
		SubCategory:      d.Get("sub_category").(string),
		APIKeyHeaderName: d.Get("api_key_header_name").(string),
		RoleArn:          d.Get("role_arn").(string),
		Region:           d.Get("region").(string),
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.(*schema.Set).List()
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
		credential.Teams = tms
	}

	return credential, nil
}

func resourceCredentialCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	credential, err := buildCredential(d)
	if err != nil {
		log.Printf("[ERROR] Building credential error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating credential %s", credential.Name)

	result := &ilert.CreateCredentialOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateCredential(&ilert.CreateCredentialInput{Credential: credential})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert credential error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert credential error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Credential == nil {
		log.Printf("[ERROR] Creating ilert credential error: empty response")
		return diag.Errorf("credential response is empty")
	}

	d.SetId(strconv.FormatInt(result.Credential.ID, 10))

	return resourceCredentialRead(ctx, d, m)
}

func resourceCredentialRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	credentialID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse credential id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading credential: %s", d.Id())
	result := &ilert.GetCredentialOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetCredential(&ilert.GetCredentialInput{CredentialID: ilert.Int64(credentialID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing credential %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a credential with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert credential error: %s", err.Error())
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		return nil
	}

	if result == nil || result.Credential == nil {
		log.Printf("[ERROR] Reading ilert credential error: empty response")
		return diag.Errorf("credential response is empty")
	}

	err = transformCredentialResource(result.Credential, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceCredentialDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*ilert.Client)

	credentialID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse credential id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting credential: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteCredential(&ilert.DeleteCredentialInput{CredentialID: ilert.Int64(credentialID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				// already gone, deleted in the web app for example
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a credential with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert credential error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceCredentialExists(d *schema.ResourceData, m any) (bool, error) {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return false, err
	}

	client := m.(*ilert.Client)

	credentialID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse credential id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading credential: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetCredential(&ilert.GetCredentialInput{CredentialID: ilert.Int64(credentialID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert credential error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a credential with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert credential error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformCredentialResource(credential *ilert.Credential, d *schema.ResourceData) error {
	d.Set("name", credential.Name)
	d.Set("type", credential.TypeHint)
	d.Set("username", credential.Username)
	d.Set("host_url", credentialHostURL(credential, d))
	d.Set("idp_url", credential.IdpURL)
	d.Set("idp_scopes", credential.IdpScopes)
	d.Set("idp_client", credential.IdpClient)
	d.Set("audience", credential.Audience)
	d.Set("oauth_method", credential.OAuthMethod)
	d.Set("category", credential.Category)
	d.Set("sub_category", credential.SubCategory)
	d.Set("api_key_header_name", credential.APIKeyHeaderName)
	d.Set("role_arn", credential.RoleArn)
	d.Set("region", credential.Region)

	// secret_value is left as it is: the API never returns the secret, and setting
	// it empty would replace the credential on every plan

	if err := d.Set("team", flattenCredentialTeams(credential.Teams, d)); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	return nil
}

// flattenCredentialTeams reads the teams of a credential into the state, managed once
// declared like on every other resource. Unlike there, a team name is kept as it was
// configured rather than taken from the API: every change replaces a credential, so a
// team renamed in ilert would otherwise replace it on every apply.
func flattenCredentialTeams(teams []ilert.TeamShort, d *schema.ResourceData) []any {
	results := make([]any, 0, len(teams))
	val, declared := d.GetOk("team")
	if !declared {
		if d.Id() != "" {
			return results
		}
		// read without a configuration, for the IaC exporter
		for _, team := range teams {
			result := map[string]any{"id": team.ID}
			if team.Name != "" {
				result["name"] = team.Name
			}
			results = append(results, result)
		}
		return results
	}

	configuredNames := make(map[int64]string)
	for _, m := range val.(*schema.Set).List() {
		if v, ok := m.(map[string]any); ok {
			name, _ := v["name"].(string)
			configuredNames[int64(v["id"].(int))] = name
		}
	}
	for _, team := range teams {
		result := map[string]any{"id": team.ID}
		if name := configuredNames[team.ID]; name != "" {
			result["name"] = name
		}
		results = append(results, result)
	}
	return results
}

// resourceCredentialImport imports a credential with its teams, which other resources
// leave to the configuration: every change replaces a credential, so a team missing
// from the state would replace it on the first apply. The teams are taken with their
// names, as the IaC exporter writes them.
func resourceCredentialImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return nil, err
	}

	client := m.(*ilert.Client)

	credentialID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, unconvertibleIDErr(d.Id(), err)
	}

	var credential *ilert.Credential
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetCredential(&ilert.GetCredentialInput{CredentialID: ilert.Int64(credentialID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential with id '%s' to be imported, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not import a credential with ID %s, error: %s", d.Id(), err.Error()))
		}
		credential = r.Credential
		return nil
	})
	if err != nil {
		return nil, err
	}

	teams := make([]any, 0, len(credential.Teams))
	for _, team := range credential.Teams {
		result := map[string]any{"id": team.ID}
		if team.Name != "" {
			result["name"] = team.Name
		}
		teams = append(teams, result)
	}
	if err := d.Set("team", teams); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// credentialHostURL picks the host url for the state. A NONE credential is created
// from a whole URL: the API keeps its path and query as the secret and returns only
// the origin. The URL in the state is kept while it still has that origin, otherwise
// every plan would replace the credential.
func credentialHostURL(credential *ilert.Credential, d *schema.ResourceData) string {
	if credential.TypeHint != ilert.CredentialType.None {
		return credential.HostURL
	}
	current, _ := d.Get("host_url").(string)
	if current != "" && credentialURLOrigin(current) == credential.HostURL {
		return current
	}
	return credential.HostURL
}

// credentialURLOrigin is the origin the API stores for the URL of a NONE credential:
// the scheme and host in lower case, and the port when the URL names one.
func credentialURLOrigin(value string) string {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		// an IPv6 address keeps its brackets in the origin
		host = "[" + host + "]"
	}
	origin := fmt.Sprintf("%s://%s", strings.ToLower(u.Scheme), host)
	if port, err := strconv.Atoi(u.Port()); err == nil && port > 0 {
		origin = fmt.Sprintf("%s:%d", origin, port)
	}
	return origin
}

// credentialArgumentKeys are the arguments of a credential, any change of which
// replaces it
var credentialArgumentKeys = []string{
	"name", "type", "secret_value", "username", "host_url", "idp_url", "idp_scopes", "idp_client",
	"audience", "oauth_method", "category", "sub_category", "api_key_header_name", "role_arn", "region", "team",
}

// uriCharacters are the ASCII characters a URL may carry as the API parses it. Anything
// else has to be percent-encoded, the API rejects it with a 400.
const uriCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#@!$&'()*+,;=%"

// validateCredentialHostURL checks a URL the way the API parses it, which is stricter
// than Go. The value is never part of a message: the URL of a NONE credential is the
// secret, and Terraform does not hide what an error message says.
func validateCredentialHostURL(v any, k string) ([]string, []error) {
	value, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	if strings.TrimSpace(value) != value {
		return nil, []error{fmt.Errorf("%q must not start or end with whitespace, the API removes it", k)}
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, []error{fmt.Errorf("%q must be an absolute http or https URL", k)}
	}
	if u.User != nil {
		return nil, []error{fmt.Errorf("%q must not carry a user name and password, a credential of type BASIC takes them as username and secret_value", k)}
	}
	if strings.Contains(u.Hostname(), "_") {
		return nil, []error{fmt.Errorf("%q names a host with an underscore, which the API does not accept", k)}
	}
	// brackets are only allowed around an IPv6 host
	rest := value
	if strings.HasPrefix(u.Host, "[") {
		rest = strings.Replace(value, u.Host, "", 1)
	}
	for _, r := range rest {
		if r < 0x80 && !strings.ContainsRune(uriCharacters, r) {
			return nil, []error{fmt.Errorf("%q contains a character the API does not accept in a URL, such as a space, a bracket or a brace: percent-encode it", k)}
		}
	}
	return nil, nil
}

func validateNoSurroundingWhitespace(v any, k string) ([]string, []error) {
	value, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	if strings.TrimSpace(value) != value {
		return nil, []error{fmt.Errorf("%q must not start or end with whitespace, the API removes it", k)}
	}
	return nil, nil
}

// planValues is what validateCredentialArguments reads a plan through, a
// *schema.ResourceDiff in the provider
type planValues interface {
	Id() string
	Get(key string) any
	NewValueKnown(key string) bool
	HasChanges(keys ...string) bool
}

// validateCredentialDiff fails the plan of a credential while ilert_credential is
// not in early_access_enabled, then checks its arguments
func validateCredentialDiff(_ context.Context, d *schema.ResourceDiff, m any) error {
	if err := requireEarlyAccess(m, "ilert_credential"); err != nil {
		return err
	}
	return validateCredentialArguments(d)
}

// validateCredentialArguments rejects at plan time what the API rejects with a 400 on
// create for every account. Every change replaces a credential, and Terraform deletes
// the old one first, so a credential the API rejects would leave none behind.
//
// A credential that exists and is not changed is not created again, so nothing is
// checked for it. That is what lets one be imported without its secret, which the API
// never returns: the configuration leaves secret_value out, and the plan stays empty.
func validateCredentialArguments(d planValues) error {
	if d.Id() != "" && !d.HasChanges(credentialArgumentKeys...) {
		return nil
	}
	if !d.NewValueKnown("type") {
		return nil
	}
	credentialType := d.Get("type").(string)

	// isSet reports whether an argument is known and set, isUnset whether it is
	// known and blank, as the API takes a value of only whitespace for none. An
	// argument whose value is not known yet is neither.
	isSet := func(key string) bool {
		return d.NewValueKnown(key) && strings.TrimSpace(d.Get(key).(string)) != ""
	}
	isUnset := func(key string) bool {
		return d.NewValueKnown(key) && strings.TrimSpace(d.Get(key).(string)) == ""
	}

	if credentialType == ilert.CredentialType.None {
		if isUnset("host_url") {
			return fmt.Errorf("a credential of type NONE takes the whole URL, including the secret part of it, in host_url")
		}
		if isSet("secret_value") {
			return fmt.Errorf("a credential of type NONE takes no secret_value, its secret is the path and query of host_url")
		}
	} else if isUnset("secret_value") {
		return fmt.Errorf("a credential of type %s needs a secret_value", credentialType)
	}

	switch credentialType {
	case ilert.CredentialType.Basic:
		if isUnset("username") {
			return fmt.Errorf("a credential of type BASIC needs a username")
		}
	case ilert.CredentialType.None, ilert.CredentialType.Bearer, ilert.CredentialType.OAuth2, ilert.CredentialType.Unknown:
		if isSet("username") {
			return fmt.Errorf("a credential of type %s takes no username", credentialType)
		}
	case ilert.CredentialType.AWSSigV4:
		if isUnset("username") && isUnset("role_arn") {
			return fmt.Errorf("a credential of type AWS_SIGV4 needs the access key id as username, unless it assumes a role_arn")
		}
	}

	if credentialType == ilert.CredentialType.OAuth2 {
		if isUnset("idp_url") {
			return fmt.Errorf("a credential of type OAUTH2 needs the token endpoint in idp_url")
		}
		if d.NewValueKnown("oauth_method") {
			method := d.Get("oauth_method").(string)
			clientCredentials := method == ilert.CredentialOAuthMethod.ClientSecretBasic || method == ilert.CredentialOAuthMethod.ClientSecretPost
			if clientCredentials && isUnset("idp_client") {
				return fmt.Errorf("a credential of type OAUTH2 with the oauth_method %s needs the client id in idp_client", method)
			}
		}
	}

	return nil
}
