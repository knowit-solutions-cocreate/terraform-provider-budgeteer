package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("BUDGETEER_HOST", nil),
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("BUDGETEER_API_KEY", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"budgeteer_api_key": resourceApiKey(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

type apiClient struct {
	host string
	port string
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	port := d.Get("port").(string)

	var diags diag.Diagnostics

	client := &apiClient{
		host: host,
		port: port,
	}

	return client, diags
}
