// +build all core resource_group_membership

package azuredevops

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/graph"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that the following sequence of events occurrs without error:
//	(1) TF apply creates resource
//	(2) TF state values are set
//	(3) Group membership exists and can be queried for
// 	(4) TF destroy removes group memberships
//
// Note: This will be uncommented in https://github.com/microsoft/terraform-provider-azuredevops/issues/174
//
func TestAccGroupMembership_CreateAndRemove(t *testing.T) {
	t.Skip("Skipping test TestAccGroupMembership_CreateAndRemove: https://github.com/microsoft/terraform-provider-azuredevops/issues/174")
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	userPrincipalName := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")
	groupName := "Build Administrators"
	tfNode := "azuredevops_group_membership.membership"

	tfStanzaWithMembership := testAccGroupMembershipResource(projectName, groupName, userPrincipalName)
	tfStanzaWithoutMembership := testAccGroupMembershipDependencies(projectName, groupName, userPrincipalName)

	// This test differs from most other acceptance tests in the following ways:
	//	- The second step is the same as the first except it omits the group membershp.
	//	  This lets us test that the membership is removed in isolation of the project being deleted
	//	- There is no CheckDestroy function because that is covered based on the above point
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: TestProviders(),
		Steps: []resource.TestStep{
			{
				// add the group membership
				Config: tfStanzaWithMembership,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "group"),
					// this attribute specifies the number of members in the resource state. the
					// syntax is how terraform maps complex types into a flattened map.
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
					testAccVerifyGroupMembershipMatchesState(),
				),
			}, {
				// remove the group membership
				Config: tfStanzaWithoutMembership,
				Check:  testAccVerifyGroupMembershipMatchesState(),
			},
		},
	})
}

// Verifies that the group membership in AzDO matches the group membership specified by the state
func testAccVerifyGroupMembershipMatchesState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		memberDescriptor := s.RootModule().Outputs["user_descriptor"].Value.(string)
		groupDescriptor := s.RootModule().Outputs["group_descriptor"].Value.(string)
		_, expectingMembership := s.RootModule().Resources["azuredevops_group_membership.membership"]

		// The sleep here is to take into account some propegation delay that can happen with Group Membership APIs.
		// If we want to go inspect the behavior of the service after a Terraform Apply, we'll need to wait a little bit
		// before making the API call.
		//
		// Note: some thought was put behind keeping the time.sleep here vs in the provider implementation. After consideration,
		// I decided to keep it here. Moving to the provider would (1) provide no functional benefit to the end user, (2) increase
		// complexity and (3) be inconsistent with the UI and CLI behavior for the same operation.
		time.Sleep(5 * time.Second)
		memberships, err := getMembersOfGroup(groupDescriptor)
		if err != nil {
			return err
		}

		if !expectingMembership && len(*memberships) == 0 {
			return nil
		}

		if !expectingMembership && len(*memberships) > 0 {
			return fmt.Errorf("unexpectedly found group members: %+v", memberships)
		}

		if expectingMembership && len(*memberships) == 0 {
			return fmt.Errorf("unexpectedly did not find memberships")
		}

		actualMemberDescriptor := *(*memberships)[0].MemberDescriptor
		if !strings.EqualFold(strings.ToLower(actualMemberDescriptor), strings.ToLower(memberDescriptor)) {
			return fmt.Errorf("expected member with descriptor %s but member had descriptor %s", memberDescriptor, actualMemberDescriptor)
		}

		return nil
	}
}

// call AzDO API to query for group members
func getMembersOfGroup(groupDescriptor string) (*[]graph.GraphMembership, error) {
	clients := TestProvider().Meta().(*config.AggregatedClient)
	return clients.GraphClient.ListMemberships(clients.Ctx, graph.ListMembershipsArgs{
		SubjectDescriptor: &groupDescriptor,
		Direction:         &graph.GraphTraversalDirectionValues.Down,
		Depth:             converter.Int(1),
	})
}

// full terraform stanza to standup a group membership
func testAccGroupMembershipResource(projectName, groupName, userPrincipalName string) string {
	membershipDependenciesStanza := testAccGroupMembershipDependencies(projectName, groupName, userPrincipalName)
	membershipStanza := `
resource "azuredevops_group_membership" "membership" {
	group = data.azuredevops_group.group.descriptor
	members = [azuredevops_user_entitlement.user.descriptor]
}`

	return membershipDependenciesStanza + "\n" + membershipStanza
}

// all the dependencies needed to configure a group membership
func testAccGroupMembershipDependencies(projectName, groupName, userPrincipalName string) string {
	return fmt.Sprintf(`
resource "azuredevops_project" "project" {
	project_name = "%s"
}
data "azuredevops_group" "group" {
	project_id = azuredevops_project.project.id
	name       = "%s"
}
resource "azuredevops_user_entitlement" "user" {
	principal_name       = "%s"
	account_license_type = "express"
}

output "group_descriptor" {
	value = data.azuredevops_group.group.descriptor
}
output "user_descriptor" {
	value = azuredevops_user_entitlement.user.descriptor
}
`, projectName, groupName, userPrincipalName)
}
