package crypto

import (
	"bytes"
	"encoding/base64"
	v1 "k8s.io/api/core/v1"
	"sort"
	"strings"
)

// ConvertToEnvVarName converts configMap/secret name into a usable env var
// removing any special chars with '_' and transforming text to upper case
func ConvertToEnvVarName(name string) string {
	var buffer bytes.Buffer
	upper := strings.ToUpper(name)
	lastCharValid := false
	for i := 0; i < len(upper); i++ {
		ch := upper[i]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			buffer.WriteString(string(ch))
			lastCharValid = true
		} else {
			if lastCharValid {
				buffer.WriteString("_")
			}
			lastCharValid = false
		}
	}
	return buffer.String()
}

func GetSHAfromConfigmap(configmap *v1.ConfigMap) (string, error) {
	var values []string
	for k, v := range configmap.Data {
		values = append(values, k+"="+v)
	}
	for k, v := range configmap.BinaryData {
		values = append(values, k+"="+base64.StdEncoding.EncodeToString(v))
	}
	sort.Strings(values)
	return GenerateSHA(strings.Join(values, ";"))
}

func GetSHAfromSecret(data map[string][]byte) (string, error) {
	var values []string
	for k, v := range data {
		values = append(values, k+"="+string(v[:]))
	}
	sort.Strings(values)
	return GenerateSHA(strings.Join(values, ";"))
}
