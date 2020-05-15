// +build all resource_build_definition

package azuredevops

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/tfhelper"
)

// validates that an apply followed by another apply (i.e., resource update) will be reflected in AzDO and the
// underlying terraform state.
func TestAccAzureDevOpsBuildDefinition_Create_Update_Import(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	gitRepoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	buildDefinitionPathEmpty := `\`
	buildDefinitionNameFirst := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	buildDefinitionNameSecond := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	buildDefinitionNameThird := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	buildDefinitionPathFirst := `\` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	buildDefinitionPathSecond := `\` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	buildDefinitionPathThird := `\` + buildDefinitionNameFirst + `\` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	buildDefinitionPathFourth := `\` + buildDefinitionNameSecond + `\` + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	tfBuildDefNode := "azuredevops_build_definition.build"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccBuildDefinitionCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameFirst, buildDefinitionPathEmpty),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathEmpty),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameSecond, buildDefinitionPathEmpty),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameSecond),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameSecond),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathEmpty),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameFirst, buildDefinitionPathFirst),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathFirst),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameFirst,
					buildDefinitionPathSecond),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathSecond),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameFirst, buildDefinitionPathThird),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathThird),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceGitHub(projectName, buildDefinitionNameFirst, buildDefinitionPathFourth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameFirst),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameFirst),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathFourth),
				),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceTfsGit(projectName, gitRepoName, buildDefinitionNameThird, buildDefinitionPathEmpty),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildDefinitionResourceExists(buildDefinitionNameThird),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "revision"),
					resource.TestCheckResourceAttrSet(tfBuildDefNode, "repository.0.repo_id"),
					resource.TestCheckResourceAttr(tfBuildDefNode, "name", buildDefinitionNameThird),
					resource.TestCheckResourceAttr(tfBuildDefNode, "path", buildDefinitionPathEmpty),
				),
			}, {
				// Resource Acceptance Testing https://www.terraform.io/docs/extend/resources/import.html#resource-acceptance-testing-implementation
				ResourceName:      tfBuildDefNode,
				ImportStateIdFunc: testAccImportStateIDFunc(tfBuildDefNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Verifies a build for Bitbucket can happen. Note: the update/import logic is tested in other tests
func TestAccAzureDevOpsBuildDefinitionBitbucket_Create(t *testing.T) {
	projectName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccBuildDefinitionCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testhelper.TestAccBuildDefinitionResourceBitbucket(projectName, "build-def-name", "\\", ""),
				ExpectError: regexp.MustCompile("bitbucket repositories need a referenced service connection ID"),
			}, {
				Config: testhelper.TestAccBuildDefinitionResourceBitbucket(projectName, "build-def-name", "\\", "some-service-connection"),
				Check:  testAccCheckBuildDefinitionResourceExists("build-def-name"),
			},
		},
	})
}

// Given the name of an AzDO build definition, this will return a function that will check whether
// or not the definition (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckBuildDefinitionResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		buildDef, ok := s.RootModule().Resources["azuredevops_build_definition.build"]
		if !ok {
			return fmt.Errorf("Did not find a build definition in the TF state")
		}

		buildDefinition, err := getBuildDefinitionFromResource(buildDef)
		if err != nil {
			return err
		}

		if *buildDefinition.Name != expectedName {
			return fmt.Errorf("Build Definition has Name=%s, but expected Name=%s", *buildDefinition.Name, expectedName)
		}

		return nil
	}
}

// verifies that all build definitions referenced in the state are destroyed. This will be invoked
// *after* terrafform destroys the resource but *before* the state is wiped clean.
func testAccBuildDefinitionCheckDestroy(s *terraform.State) error {
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_build_definition" {
			continue
		}

		// indicates the build definition still exists - this should fail the test
		if _, err := getBuildDefinitionFromResource(resource); err == nil {
			return fmt.Errorf("Unexpectedly found a build definition that should be deleted")
		}
	}

	return nil
}

// given a resource from the state, return a build definition (and error)
func getBuildDefinitionFromResource(resource *terraform.ResourceState) (*build.BuildDefinition, error) {
	buildDefID, err := strconv.Atoi(resource.Primary.ID)
	if err != nil {
		return nil, err
	}

	projectID := resource.Primary.Attributes["project_id"]
	clients := TestProvider().Meta().(*config.AggregatedClient)
	return clients.BuildClient.GetDefinition(clients.Ctx, build.GetDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &buildDefID,
	})
}

func sortBuildDefinition(b build.BuildDefinition) build.BuildDefinition {
	if b.Triggers == nil {
		return b
	}
	for _, t := range *b.Triggers {
		if m, ok := t.(map[string]interface{}); ok {
			if m2, ok := m["branchFilters"].([]interface{}); ok {
				bf := tfhelper.ExpandStringList(m2)
				sort.Strings(bf)
				m["branchFilters"] = bf
			}
			if m3, ok := m["pathFilters"].([]interface{}); ok {
				pf := tfhelper.ExpandStringList(m3)
				sort.Strings(pf)
				m["pathFilters"] = pf
			}
		}
	}
	return b
}
