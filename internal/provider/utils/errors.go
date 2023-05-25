package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func DiagnosticsToErr(diagnostics diag.Diagnostics) error {
	for _, d := range diagnostics {
		if d.Severity() != diag.SeverityError {
			continue
		}
		return fmt.Errorf("%s: %s", d.Summary(), d.Detail())
	}
	return nil
}
