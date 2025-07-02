package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiKeyCreate,
		ReadContext:   resourceApiKeyRead,
		UpdateContext: resourceApiKeyUpdate,
		DeleteContext: resourceApiKeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"budget": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"costs": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_used_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApiKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	name := d.Get("name").(string)
	budget := d.Get("budget").(int)

	// Check if the API key already exists
	url := fmt.Sprintf("%s/keyView", client.host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	var existingKeys []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&existingKeys); err != nil {
		return diag.FromErr(err)
	}

	// Check if key with same name exists
	for _, key := range existingKeys {
		if key["name"].(string) == name {
			// Key exists, import it into our state
			keyID := fmt.Sprint(int(key["id"].(float64)))
			d.SetId(keyID)

			// If budget has changed, update it
			existingBudget := int(key["budget"].(float64))
			if budget != existingBudget {
				if err := updateKeyBudget(client, keyID, float64(budget)); err != nil {
					return diag.FromErr(err)
				}
			}

			// Read the state to ensure it's up to date
			return resourceApiKeyRead(ctx, d, m)
		}
	}

	// Create new key if it doesn't exist
	payload := map[string]interface{}{
		"name":   name,
		"budget": float64(budget),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	url = fmt.Sprintf("%s/key", client.host)
	req, err = http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err = http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusCreated {
		return diag.Errorf("failed to create API key, status: %d", r.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return diag.FromErr(err)
	}
	// Set the ID and key_value immediately after creation
	d.SetId(fmt.Sprint(int(response["id"].(float64))))
	if keyValue, ok := response["key"].(string); ok {
		if err := d.Set("key_value", keyValue); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the resource to ensure all other state is fully populated
	return resourceApiKeyRead(ctx, d, m)
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	var diags diag.Diagnostics

	// First try to get the key from /keyView endpoint
	url := fmt.Sprintf("%s/keyView", client.host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	var keys []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
		return diag.FromErr(err)
	}

	// Find the key with matching ID
	var foundKey map[string]interface{}
	currentID := d.Id()
	for _, key := range keys {
		if fmt.Sprint(int(key["id"].(float64))) == currentID {
			foundKey = key
			break
		}
	}

	if foundKey == nil {
		// Key not found, remove it from state
		d.SetId("")
		return diags
	}

	// Set all known fields from the found key
	if err := d.Set("name", foundKey["name"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("budget", int(foundKey["budget"].(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("costs", foundKey["costs"].(float64)); err != nil {
		return diag.FromErr(err)
	}
	if createdAt, ok := foundKey["created_at"].(string); ok {
		if err := d.Set("created_at", createdAt); err != nil {
			return diag.FromErr(err)
		}
	}
	if lastUsedAt, ok := foundKey["last_used_at"].(string); ok {
		if err := d.Set("last_used_at", lastUsedAt); err != nil {
			return diag.FromErr(err)
		}
	}

	// Get the full key information including key_value from /key endpoint
	url = fmt.Sprintf("%s/key", client.host)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err = http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	var fullKeys []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fullKeys); err != nil {
		return diag.FromErr(err)
	}

	// Find the matching key with the sensitive data
	for _, key := range fullKeys {
		if fmt.Sprint(int(key["id"].(float64))) == currentID {
			if keyValue, ok := key["key"].(string); ok {
				if err := d.Set("key_value", keyValue); err != nil {
					return diag.FromErr(err)
				}
			}
			break
		}
	}

	return diags
}

func resourceApiKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	if d.HasChanges("name", "budget") {
		budget := d.Get("budget").(int)

		payload := map[string]interface{}{
			"budget": float64(budget),
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return diag.FromErr(err)
		}

		url := fmt.Sprintf("%s/key?id=%s", client.host, d.Id())
		req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonPayload)))
		if err != nil {
			return diag.FromErr(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return diag.FromErr(err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			return diag.Errorf("failed to update API key")
		}
	}

	return resourceApiKeyRead(ctx, d, m)
}

func resourceApiKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	var diags diag.Diagnostics

	url := fmt.Sprintf("%s/key?id=%s", client.host, d.Id())
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return diag.Errorf("failed to delete API key")
	}

	d.SetId("")

	return diags
}

func checkKeyExists(ctx context.Context, client *apiClient, name string) (bool, string, error) {
	url := fmt.Sprintf("%s/keyView", client.host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer r.Body.Close()

	var response []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return false, "", err
	}

	for _, key := range response {
		if key["name"].(string) == name {
			return true, fmt.Sprint(key["id"]), nil
		}
	}

	return false, "", nil
}

// Helper function to update key budget
func updateKeyBudget(client *apiClient, keyID string, budget float64) error {
	payload := map[string]interface{}{
		"budget": budget,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/key?id=%s", client.host, keyID)
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update API key budget")
	}

	return nil
}
