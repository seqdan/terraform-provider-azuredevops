// +build all resource_resource_authorization

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that resource authorization can be created, updated and deleted
func TestAccResourceAuthorization_CRUD(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	serviceEndpointName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourcesHCL := testhelper.TestAccServiceEndpointGitHubResource(projectName, serviceEndpointName)
	authedHCL := testhelper.TestAccResourceAuthorization("azuredevops_serviceendpoint_github.serviceendpoint.id", true)
	unAuthedHCL := testhelper.TestAccResourceAuthorization("azuredevops_serviceendpoint_github.serviceendpoint.id", false)

	tfAuthNode := "azuredevops_resource_authorization.auth"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: TestProviders(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", resourcesHCL, authedHCL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfAuthNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfAuthNode, "resource_id"),
					resource.TestCheckResourceAttr(tfAuthNode, "authorized", "true"),
				),
			}, {
				Config: fmt.Sprintf("%s\n%s", resourcesHCL, unAuthedHCL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfAuthNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfAuthNode, "resource_id"),
					resource.TestCheckResourceAttr(tfAuthNode, "authorized", "false"),
				),
			},
		},
	})
}
