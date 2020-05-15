// +build all core resource_group

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"testing"
)

func TestAccGroupResource_CreateAndUpdate(t *testing.T) {
	t.Skip("Skipping test TestAccGroupResource_CreateAndUpdate: broken graph implementation in Go Azure DevOps REST API")

	//projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	//groupName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	//
	//resource.Test(t, resource.TestCase{
	//	PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
	//	Providers:    TestProviders(),
	//	CheckDestroy: testAccGroupCheckDestroy,
	//	Steps: []resource.TestStep{
	//		{
	//			Config: testhelper.TestAccGroupResource("mygroup", projectName, groupName),
	//			Check: resource.ComposeTestCheckFunc(
	//				testAccCheckGroupResourceExists("mygroup", groupName),
	//				resource.TestCheckResourceAttrSet("azuredevops_group.mygroup", "scope"),
	//				resource.TestCheckResourceAttr("azuredevops_group.mygroup", "display_name", groupName),
	//			),
	//		},
	//		{
	//			// Resource Acceptance Testing https://www.terraform.io/docs/extend/resources/import.html#resource-acceptance-testing-implementation
	//			ResourceName:      "azuredevops_group.mygroup",
	//			ImportState:       true,
	//			ImportStateVerify: true,
	//		},
	//	},
	//})
}

//
//func testAccCheckGroupResourceExists(resourceName, expectedName string) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		varGroup, ok := s.RootModule().Resources[fmt.Sprintf("azuredevops_group.%s", resourceName)]
//		if !ok {
//			return fmt.Errorf("Did not find a group resource with name %s in the TF state", resourceName)
//		}
//
//		getGroupArgs := graph.GetGroupArgs{
//			GroupDescriptor: converter.String(varGroup.Primary.Attributes["display_name"]),
//		}
//		clients := TestProvider().Meta().(*config.AggregatedClient)
//		group, err := clients.GraphClient.GetGroup(clients.Ctx, getGroupArgs)
//		if err != nil {
//			return err
//		}
//		if group == nil {
//			return fmt.Errorf("Group with Name=%s does not exit", varGroup.Primary.Attributes["display_name"])
//		}
//		if *group.DisplayName != expectedName {
//			return fmt.Errorf("Group has Name=%s, but expected %s", *group.DisplayName, expectedName)
//		}
//
//		return nil
//	}
//}
//
//func testAccGroupCheckDestroy(s *terraform.State) error {
//	clients := TestProvider().Meta().(*config.AggregatedClient)
//
//	// verify that every project referenced in the state does not exist in AzDO
//	for _, resource := range s.RootModule().Resources {
//		if resource.Type != "azuredevops_group" {
//			continue
//		}
//
//		id := resource.Primary.ID
//
//		getGroupArgs := graph.GetGroupArgs{
//			GroupDescriptor: converter.String(id),
//		}
//		group, err := clients.GraphClient.GetGroup(clients.Ctx, getGroupArgs)
//		if err != nil {
//			return err
//		}
//		if group.Descriptor != nil {
//			return fmt.Errorf("Group with ID %s should not exist in scope %s", id, resource.Primary.Attributes["scope"])
//		}
//	}
//
//	return nil
//}
