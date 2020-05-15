package internal

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataProject() *schema.Resource {
	baseSchema := ResourceProject()
	return &schema.Resource{
		Read:   baseSchema.Read,
		Schema: baseSchema.Schema,
	}
}
