package builder

import (
	"fmt"
	"strings"

	"github.com/asim9115/containerix/internal/types"
)

var shellChars = []string{";", "&&", "||", "`", "$()", "|", ">", "<", "\n", "\r"}

func ValidateBuildRequest(req *types.BuildRequest) error {
	for _, cmd := range []string{req.BuildCommand, req.StartCommand} {
		if len(cmd) > 256 {
			return fmt.Errorf("command too long (max 256 chars)")
		}
		for _, ch := range shellChars {
			if strings.Contains(cmd, ch) {
				return fmt.Errorf("shell metacharacter %q not allowed in command", ch)
			}
		} 
	}
	if req.BuildCommand == "" && req.Language != types.Python && req.Language != types.Static {
        return fmt.Errorf("build_command is required for language %s", req.Language)
	}
	if req.StartCommand == "" {
		return fmt.Errorf("start_command is required")
	}
	return nil
}