package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type AttributeGetter interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}
