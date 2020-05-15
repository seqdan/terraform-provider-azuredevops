// +build all resource_agentpool

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that the following sequence of events occurrs without error:
//	(1) TF apply creates agent pool
//	(2) TF state values are set
//	(3) Agent pool can be queried by ID and has expected name
//  (4) TF apply updates agent pool with new name
//  (5) Agent pool can be queried by ID and has expected name
// 	(6) TF destroy deletes agent pool
//	(7) Agent pool can no longer be queried by ID
func TestAccAzureDevOpsAgentPool_CreateAndUpdate(t *testing.T) {
	poolNameFirst := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	poolNameSecond := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfNode := "azuredevops_agent_pool.pool"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccAgentPoolCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccAgentPoolResource(poolNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolNameFirst),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
					testAccCheckAgentPoolResourceExists(poolNameFirst),
				),
			},
			{
				Config: testhelper.TestAccAgentPoolResource(poolNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolNameSecond),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
					testAccCheckAgentPoolResourceExists(poolNameSecond),
				),
			},
			{
				// Resource Acceptance Testing https://www.terraform.io/docs/extend/resources/import.html#resource-acceptance-testing-implementation
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Given the name of an AzDO project, this will return a function that will check whether
// or not the project (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckAgentPoolResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources["azuredevops_agent_pool.pool"]
		if !ok {
			return fmt.Errorf("Did not find a agent pool in the TF state")
		}

		clients := TestProvider().Meta().(*config.AggregatedClient)
		id, err := strconv.Atoi(resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parse ID error, ID:  %v !. Error= %v", resource.Primary.ID, err)
		}

		project, agentPoolErr := AzureAgentPoolRead(clients, id)

		if agentPoolErr != nil {
			return fmt.Errorf("Agent Pool with ID=%d cannot be found!. Error=%v", id, err)
		}

		if *project.Name != expectedName {
			return fmt.Errorf("Agent Pool with ID=%d has Name=%s, but expected Name=%s", id, *project.Name, expectedName)
		}

		return nil
	}
}

// verifies that agent pool referenced in the state is destroyed. This will be invoked
// *after* terrafform destroys the resource but *before* the state is wiped clean.
func testAccAgentPoolCheckDestroy(s *terraform.State) error {
	clients := TestProvider().Meta().(*config.AggregatedClient)

	// verify that every agent pool referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_agent_pool" {
			continue
		}

		id, err := strconv.Atoi(resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Agent Pool ID=%d cannot be parsed!. Error=%v", id, err)
		}

		// indicates the agent pool still exists - this should fail the test
		if _, err := AzureAgentPoolRead(clients, id); err == nil {
			return fmt.Errorf("Agent Pool ID %d should not exist", id)
		}
	}

	return nil
}
