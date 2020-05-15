// +build all resource_variable_group

package internal

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/taskagent"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/stretchr/testify/require"
)

var testVarGroupProjectID = uuid.New().String()

// This definition matches the overall structure of what a configured git repository would
// look like. Note that the ID and Name attributes match -- this is the service-side behavior
// when configuring a GitHub repo.
var testVariableGroup = taskagent.VariableGroup{
	Id:          converter.Int(100),
	Name:        converter.String("Name"),
	Description: converter.String("This is a test variable group."),
	Variables: &map[string]taskagent.VariableValue{
		"var1": {
			Value:    converter.String("value1"),
			IsSecret: converter.Bool(false),
		},
	},
}
var resourceRefType = "variablegroup"
var testDefinitionResource = build.DefinitionResourceReference{
	Type:       &resourceRefType,
	Authorized: converter.Bool(true),
	Name:       testVariableGroup.Name,
	Id:         converter.String("100"),
}

/**
 * Begin unit tests
 */
// verifies that the flatten/expand round trip yields the same build definition
func TestAzureDevOpsVariableGroup_ExpandFlatten_Roundtrip(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceVariableGroup().Schema, nil)
	flattenVariableGroup(resourceData, &testVariableGroup, &testVarGroupProjectID)
	var testArrayDefinitionResourceReference []build.DefinitionResourceReference
	testArrayDefinitionResourceReference = append(testArrayDefinitionResourceReference, testDefinitionResource)
	flattenAllowAccess(resourceData, &testArrayDefinitionResourceReference)

	variableGroupParams, projectID := expandVariableGroupParameters(resourceData)
	definitionResourceReferenceArgs := expandDefinitionResourceAuth(resourceData, &testVariableGroup)

	require.Equal(t, *testVariableGroup.Name, *variableGroupParams.Name)
	require.Equal(t, *testVariableGroup.Description, *variableGroupParams.Description)
	require.Equal(t, *testVariableGroup.Variables, *variableGroupParams.Variables)
	require.Equal(t, testVarGroupProjectID, *projectID)

	require.Equal(t, testDefinitionResource.Authorized, definitionResourceReferenceArgs[0].Authorized)
	require.Equal(t, testDefinitionResource.Id, definitionResourceReferenceArgs[0].Id)
}
