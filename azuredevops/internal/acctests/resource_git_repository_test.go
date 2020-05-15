// +build all core resource_git_repository

package azuredevops

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that the following sequence of events occurrs without error:
//	(1) TF apply creates resource
//	(2) TF state values are set
//	(3) resource can be queried by ID and has expected name
// 	(4) TF destroy deletes resource
//	(5) resource can no longer be queried by ID
func TestAccAzureGitRepo_CreateAndUpdate(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoNameFirst := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoNameSecond := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfRepoNode := "azuredevops_git_repository.gitrepo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccAzureGitRepoCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccAzureGitRepoResource(projectName, gitRepoNameFirst, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoNameFirst),
					testAccCheckAzureGitRepoResourceExists(gitRepoNameFirst),
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "size"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "ssh_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
				),
			},
			{
				Config: testhelper.TestAccAzureGitRepoResource(projectName, gitRepoNameSecond, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoNameSecond),
					testAccCheckAzureGitRepoResourceExists(gitRepoNameSecond),
					resource.TestCheckResourceAttrSet(tfRepoNode, "is_fork"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "remote_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "size"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "ssh_url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "url"),
					resource.TestCheckResourceAttrSet(tfRepoNode, "web_url"),
				),
			},
		},
	})
}

// Given the name of an AzDO git repository, this will return a function that will check whether
// or not the definition (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckAzureGitRepoResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients := TestProvider().Meta().(*config.AggregatedClient)

		gitRepo, ok := s.RootModule().Resources["azuredevops_git_repository.gitrepo"]
		if !ok {
			return fmt.Errorf("Did not find a repo definition in the TF state")
		}

		repoID := gitRepo.Primary.ID
		projectID := gitRepo.Primary.Attributes["project_id"]

		repo, err := gitRepositoryRead(clients, repoID, "", projectID)
		if err != nil {
			return err
		}

		if *repo.Name != expectedName {
			return fmt.Errorf("AzDO Git Repository has Name=%s, but expected Name=%s", *repo.Name, expectedName)
		}

		return nil
	}
}

func testAccAzureGitRepoCheckDestroy(s *terraform.State) error {
	clients := TestProvider().Meta().(*config.AggregatedClient)

	// verify that every repository referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_git_repository" {
			continue
		}

		repoID := resource.Primary.ID
		projectID := resource.Primary.Attributes["project_id"]

		// indicates the git repository still exists - this should fail the test
		if _, err := gitRepositoryRead(clients, repoID, "", projectID); err == nil {
			return fmt.Errorf("repository with ID %s should not exist", repoID)
		}
	}

	return nil
}

// Verifies that a newly created repo with init_type of "Clean" has the expected
// master branch available
func TestAccAzureGitRepo_RepoInitialization_Clean(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfRepoNode := "azuredevops_git_repository.gitrepo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccAzureGitRepoCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccAzureGitRepoResource(projectName, gitRepoName, "Clean"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					testAccCheckAzureGitRepoResourceExists(gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
				),
			},
		},
	})
}

// Verifies that a newly created repo with init_type of "Uninitialized" does NOT
// have a master branch established
func TestAccAzureGitRepo_RepoInitialization_Uninitialized(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfRepoNode := "azuredevops_git_repository.gitrepo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccAzureGitRepoCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccAzureGitRepoResource(projectName, gitRepoName, "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureGitRepoResourceExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", ""),
				),
			},
		},
	})
}

// Verifies that a newly forked repo does NOT return an empty branch_name
func TestAccAzureGitRepo_RepoFork_BranchNotEmpty(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitForkedRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfRepoNode := "azuredevops_git_repository.gitrepo"
	tfForkedRepoNode := "azuredevops_git_repository.gitforkedrepo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccAzureGitRepoCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccAzureForkedGitRepoResource(projectName, gitRepoName, gitForkedRepoName, "Clean", "Uninitialized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureGitRepoResourceExists(gitRepoName),
					resource.TestCheckResourceAttrSet(tfRepoNode, "project_id"),
					resource.TestCheckResourceAttr(tfRepoNode, "name", gitRepoName),
					resource.TestCheckResourceAttr(tfRepoNode, "default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttr(tfForkedRepoNode, "name", gitForkedRepoName),
					resource.TestCheckResourceAttr(tfForkedRepoNode, "default_branch", "refs/heads/master"),
				),
			},
		},
	})
}
