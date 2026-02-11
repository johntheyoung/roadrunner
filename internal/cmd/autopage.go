package cmd

import (
	"fmt"

	"github.com/johntheyoung/roadrunner/internal/errfmt"
)

const (
	defaultAutoPageMaxItems = 500
	maxAutoPageItems        = 5000
)

func resolveAutoPageLimit(all bool, maxItems int) (int, error) {
	return resolveAutoPageLimitNamed(all, maxItems, "--all", "--max-items")
}

func resolveAutoPageLimitNamed(all bool, maxItems int, allFlagName, maxItemsFlagName string) (int, error) {
	if !all {
		if maxItems != 0 {
			return 0, errfmt.UsageError("%s requires %s", maxItemsFlagName, allFlagName)
		}
		return 0, nil
	}

	if maxItems == 0 {
		return defaultAutoPageMaxItems, nil
	}
	if maxItems < 1 || maxItems > maxAutoPageItems {
		return 0, errfmt.UsageError("invalid %s %d (expected 1-%d)", maxItemsFlagName, maxItems, maxAutoPageItems)
	}
	return maxItems, nil
}

func nextSearchCursor(direction, oldestCursor, newestCursor string) string {
	if direction == "after" {
		if newestCursor != "" {
			return newestCursor
		}
		return oldestCursor
	}
	if oldestCursor != "" {
		return oldestCursor
	}
	return newestCursor
}

func limitReached(count, limit int) bool {
	return limit > 0 && count >= limit
}

func validateAttachmentSize(width, height *float64) error {
	widthSet := width != nil
	heightSet := height != nil
	if widthSet != heightSet {
		return errfmt.UsageError("both --attachment-width and --attachment-height are required when setting attachment size")
	}
	if width != nil && *width <= 0 {
		return errfmt.UsageError("invalid --attachment-width %g (must be > 0)", *width)
	}
	if height != nil && *height <= 0 {
		return errfmt.UsageError("invalid --attachment-height %g (must be > 0)", *height)
	}
	return nil
}

func hasAttachmentOverrides(fileName, mimeType, attachmentType string, duration, width, height *float64) bool {
	return fileName != "" || mimeType != "" || attachmentType != "" || duration != nil || width != nil || height != nil
}

func attachmentWarning(attachmentUploadID string, overrides bool) error {
	if attachmentUploadID == "" && overrides {
		return errfmt.UsageError("attachment overrides require --attachment-upload-id")
	}
	return nil
}

func validateMaxParticipantCount(v int) error {
	if v == -1 {
		return nil
	}
	if v < 0 || v > 500 {
		return errfmt.UsageError("invalid --max-participant-count %d (expected -1 or 0-500)", v)
	}
	return nil
}

func autoPageStoppedMessage(limit int) string {
	return autoPageStoppedMessageNamed(limit, "--max-items")
}

func autoPageStoppedMessageNamed(limit int, maxItemsFlagName string) string {
	return fmt.Sprintf("Stopped after %s=%d. Narrow filters or increase %s to fetch more.", maxItemsFlagName, limit, maxItemsFlagName)
}
