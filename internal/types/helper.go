package types

import (
	"strconv"
	"strings"
	"fmt"
)

func MemoryToBytes(memory string)(string, error) {
	lower := strings.ToLower(strings.TrimSpace(memory))

	multipliers := []struct{
		suffix string
		mult int64
	}{
        {"g", 1 << 30},
        {"m", 1 << 20},
        {"k", 1 << 10},
	}
	    for _, m := range multipliers {
        if strings.HasSuffix(lower, m.suffix) {
            numStr := lower[:len(lower)-1]
            val, err := strconv.ParseFloat(numStr, 64)
            if err != nil {
                return "", fmt.Errorf("invalid memory value %q: %w", memory, err)
            }
            return strconv.FormatInt(int64(val*float64(m.mult)), 10), nil
        }
    }

	if _, err := strconv.ParseInt(lower, 10, 64); err != nil {
        return "", fmt.Errorf("invalid memory value %q", memory)
    }
    return lower, nil
}