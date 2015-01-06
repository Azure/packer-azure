package utils

import (
	"fmt"
	"time"
)

func BuildContainerName() string {
	return fmt.Sprintf("packer-provision-%s", RandomString("abcdefghijklmnopqrstuvwxyz0123456789", 10))
}
