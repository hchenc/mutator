package crypto

import (
	"crypto/sha1"
	"fmt"
	"io"
)

func GenerateSHA(data string) (string, error) {
	hasher := sha1.New()
	if _, err := io.WriteString(hasher, data); err != nil {
		return "", err
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha), nil
}
