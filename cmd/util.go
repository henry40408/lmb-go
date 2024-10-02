package cmd

import (
	"context"
	"fmt"
	"time"
)

func setupTimeoutContext(timeout string) (context.Context, context.CancelFunc, error) {
	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid timeout format: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), parsedTimeout)
	return ctx, cancel, nil
}
