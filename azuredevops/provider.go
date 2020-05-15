package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	azdo "github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
)

// Provider - The top level Azure DevOps Provider definition.
func Provider() *schema.Provider {
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"azuredevops_resource_authorization":     azdo.ResourceResourceAuthorization(),
			"azuredevops_build_definition":           azdo.ResourceBuildDefinition(),
			"azuredevops_project":                    azdo.ResourceProject(),
			"azuredevops_variable_group":             azdo.ResourceVariableGroup(),
			"azuredevops_serviceendpoint_azurerm":    azdo.ResourceServiceEndpointAzureRM(),
			"azuredevops_serviceendpoint_bitbucket":  azdo.ResourceServiceEndpointBitBucket(),
			"azuredevops_serviceendpoint_dockerhub":  azdo.ResourceServiceEndpointDockerHub(),
			"azuredevops_serviceendpoint_github":     azdo.ResourceServiceEndpointGitHub(),
			"azuredevops_serviceendpoint_kubernetes": azdo.ResourceServiceEndpointKubernetes(),
			"azuredevops_git_repository":             azdo.ResourceGitRepository(),
			"azuredevops_user_entitlement":           azdo.ResourceUserEntitlement(),
			"azuredevops_group_membership":           azdo.ResourceGroupMembership(),
			"azuredevops_agent_pool":                 azdo.ResourceAzureAgentPool(),
			"azuredevops_group":                      azdo.ResourceGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"azuredevops_group":            azdo.DataGroup(),
			"azuredevops_project":          azdo.DataProject(),
			"azuredevops_projects":         azdo.DataProjects(),
			"azuredevops_git_repositories": azdo.DataGitRepositories(),
			"azuredevops_users":            azdo.DataUsers(),
		},
		Schema: map[string]*schema.Schema{
			"org_service_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZDO_ORG_SERVICE_URL", nil),
				Description: "The url of the Azure DevOps instance which should be used.",
			},
			"personal_access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZDO_PERSONAL_ACCESS_TOKEN", nil),
				Description: "The personal access token which should be used.",
				Sensitive:   true,
			},
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}

		client, err := config.GetAzdoClient(d.Get("personal_access_token").(string), d.Get("org_service_url").(string), terraformVersion)

		return client, err
	}
}
