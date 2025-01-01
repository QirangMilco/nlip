package tokenUtils

import (
    "crypto/rand"
    "encoding/base64"
)

func GenerateSecureToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}