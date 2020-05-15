package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops"
)

func TestProvider() *schema.Provider {
	return azuredevops.Provider()
}

func TestProviders() map[string]terraform.ResourceProvider {
	return map[string]terraform.ResourceProvider{
		"azuredevops": TestProvider(),
	}
}
