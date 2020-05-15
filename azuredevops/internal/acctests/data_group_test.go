// +build all core data_group

package azuredevops

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Validates that a configuration containing a project group lookup is able to read the resource correctly.
// Because this is a data source, there are no resources to inspect in AzDO
func TestAccGroupDataSource_Read_HappyPath(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	group := "Build Administrators"
	tfBuildDefNode := "data.azuredevops_group.group"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: TestProviders(),
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccGroupDataSource(projectName, group),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "name"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "descriptor"),
				),
			},
		},
	})
}
