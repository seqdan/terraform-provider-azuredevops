// +build all resource_serviceendpoint_azurerm

package azuredevops

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// validates that an apply followed by another apply (i.e., resource update) will be reflected in AzDO and the
// underlying terraform state.
func TestAccAzureDevOpsServiceEndpointAzureRm_CreateAndUpdate(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	serviceEndpointNameFirst := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	serviceEndpointNameSecond := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	tfSvcEpNode := "azuredevops_serviceendpoint_azurerm.serviceendpointrm"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccServiceEndpointAzureRMCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccServiceEndpointAzureRMResource(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceEndpointAzureRMResourceExists(serviceEndpointNameFirst),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "credentials.0.serviceprincipalid"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "credentials.0.serviceprincipalkey_hash"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "credentials.0.serviceprincipalkey", ""),
				),
			}, {
				Config: testhelper.TestAccServiceEndpointAzureRMResource(projectName, serviceEndpointNameSecond),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceEndpointAzureRMResourceExists(serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "credentials.0.serviceprincipalid"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "credentials.0.serviceprincipalkey_hash"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "credentials.0.serviceprincipalkey", ""),
				),
			},
			{
				Config: testhelper.TestAccServiceEndpointAzureRMAutomaticResource(projectName, serviceEndpointNameSecond),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceEndpointAzureRMResourceExists(serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckNoResourceAttr(tfSvcEpNode, "credentials.0"),
				),
			},
			{
				Config: testhelper.TestAccServiceEndpointAzureRMAutomaticResource(projectName, serviceEndpointNameSecond),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceEndpointAzureRMResourceExists(serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckNoResourceAttr(tfSvcEpNode, "credentials.0"),
				),
			},
		},
	})
}

// Given the name of an AzDO service endpoint, this will return a function that will check whether
// or not the resource (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckServiceEndpointAzureRMResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serviceEndpointDef, ok := s.RootModule().Resources["azuredevops_serviceendpoint_azurerm.serviceendpointrm"]
		if !ok {
			return fmt.Errorf("Did not find a service endpoint in the TF state")
		}

		serviceEndpoint, err := getServiceEndpointAzureRMFromResource(serviceEndpointDef)
		if err != nil {
			return err
		}

		if *serviceEndpoint.Name != expectedName {
			return fmt.Errorf("Service Endpoint has Name=%s, but expected Name=%s", *serviceEndpoint.Name, expectedName)
		}

		return nil
	}
}

// verifies that all service endpoints referenced in the state are destroyed. This will be invoked
// *after* terrafform destroys the resource but *before* the state is wiped clean.
func testAccServiceEndpointAzureRMCheckDestroy(s *terraform.State) error {
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_serviceendpoint_azurerm" {
			continue
		}

		// indicates the service endpoint still exists - this should fail the test
		if _, err := getServiceEndpointAzureRMFromResource(resource); err == nil {
			return fmt.Errorf("Unexpectedly found a service endpoint that should be deleted")
		}
	}

	return nil
}

// given a resource from the state, return a service endpoint (and error)
func getServiceEndpointAzureRMFromResource(resource *terraform.ResourceState) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpointDefID, err := uuid.Parse(resource.Primary.ID)
	if err != nil {
		return nil, err
	}

	projectID := resource.Primary.Attributes["project_id"]
	clients := TestProvider().Meta().(*config.AggregatedClient)
	return clients.ServiceEndpointClient.GetServiceEndpointDetails(clients.Ctx, serviceendpoint.GetServiceEndpointDetailsArgs{
		Project:    &projectID,
		EndpointId: &serviceEndpointDefID,
	})
}
