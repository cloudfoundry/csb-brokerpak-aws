package csbmajorengineversion

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceMajorEngineVersion() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			engineVersionKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			majorVersionKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		ReadContext: resourceMajorEngineVersionRead,
		Description: "Returns major engine version value",
	}
}

func resourceMajorEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	descriptor := meta.(*engineDescriptor)

	engineVersion := d.Get(engineVersionKey).(string)
	majorEngineVersion, err := descriptor.Describe(ctx, engineVersion)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("version")

	tflog.Debug(ctx, "Setting Major DB engine version", map[string]any{
		"major_engine_version": majorEngineVersion,
	})
	if err := d.Set(majorVersionKey, majorEngineVersion); err != nil {
		return diag.FromErr(err)
	}
	return nil

}
