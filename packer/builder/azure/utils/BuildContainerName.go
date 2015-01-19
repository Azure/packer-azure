package utils

import (
	"fmt"
)

func BuildContainerName() string {
	return fmt.Sprintf("packer-provision-%s", RandomString("abcdefghijklmnopqrstuvwxyz0123456789", 10))
}
