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

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	budget := d.Get("budget").(int)

	payload := map[string]interface{}{
		"name":   name,
		"budget": budget,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("http://%s:%s/api/keys", client.host, client.port)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusCreated {
		return diag.Errorf("failed to create API key")
	}

	var response map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(response["id"]))
	d.Set("key_value", response["key_value"])
	d.Set("created_at", response["created_at"])
	
	return diags
}

func resourceApiKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	var diags diag.Diagnostics

	url := fmt.Sprintf("http://%s:%s/api/keys/%s", client.host, client.port, d.Id())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusNotFound {
		d.SetId("")
		return diags
	}

	var response map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", response["name"])
	d.Set("budget", response["budget"])
	d.Set("costs", response["costs"])
	d.Set("created_at", response["created_at"])
	d.Set("last_used_at", response["last_used_at"])

	return diags
}

func resourceApiKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*apiClient)

	if d.HasChanges("name", "budget") {
		name := d.Get("name").(string)
		budget := d.Get("budget").(int)

		payload := map[string]interface{}{
			"name":   name,
			"budget": budget,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return diag.FromErr(err)
		}

		url := fmt.Sprintf("http://%s:%s/api/keys/%s", client.host, client.port, d.Id())
		req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonPayload)))
		if err != nil {
			return diag.FromErr(err)
		}

		req.Header.Set("Content-Type", "application/json")

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

	url := fmt.Sprintf("http://%s:%s/api/keys/%s", client.host, client.port, d.Id())
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

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
