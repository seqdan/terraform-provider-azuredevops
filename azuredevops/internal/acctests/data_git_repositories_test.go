// +build all core data_git_repositories

package azuredevops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that the following sequence of events occurrs without error:
//	(1) TF can create a project
//	(2) A data source is added to the configuration, and that data source can find the created project
func TestAccAzureTfsGitRepositories_DataSource(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfConfigStep1 := testhelper.TestAccAzureGitRepoResource(projectName, gitRepoName, "Clean")
	tfConfigStep2 := fmt.Sprintf("%s\n%s", tfConfigStep1, testhelper.TestAccProjectGitRepositories(projectName, gitRepoName))

	tfNode := "data.azuredevops_git_repositories.repositories"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: TestProviders(),
		Steps: []resource.TestStep{
			{
				Config: tfConfigStep1,
			}, {
				Config: tfConfigStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfNode, "repositories.0.name", gitRepoName),
					resource.TestCheckResourceAttr(tfNode, "repositories.0.default_branch", "refs/heads/master"),
				),
			},
		},
	})
}
