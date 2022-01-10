// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package crypto

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	testEncryptionKey         = "passphrasewhichneedstobe32bytes!"
	testEncryptionFallbackKey = "anotherpassphrasewhichis32bytes!"
)

func TestSetEncryptionKeys(t *testing.T) {
	SetAESEncryptionKey("key")
	assert.Equal(t, "key", encryptionKey)

	SetAESEncryptionFallbackKey("fallback_key")
	assert.Equal(t, "fallback_key", encryptionFallbackKey)
}

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       string
		Key         string
		FallbackKey string
		Err         error
	}{
		{
			Name: "ok, no keys",
		},
		{
			Name:  "ok, with encryption key",
			Value: "my data",
			Key:   testEncryptionKey,
		},
		{
			Name:        "ok, with encryption and fallback key",
			Value:       "my data",
			Key:         testEncryptionKey,
			FallbackKey: testEncryptionFallbackKey,
		},
		{
			Name: "ko, wrong key",
			Key:  "dummy",
			Err:  errors.New("crypto/aes: invalid key size 5"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			key := tc.Key
			if tc.FallbackKey != "" {
				key = tc.FallbackKey
			}
			SetAESEncryptionKey(key)
			out, err := AESEncrypt(tc.Value)
			if tc.Err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)

			SetAESEncryptionKey(tc.Key)
			SetAESEncryptionFallbackKey(tc.FallbackKey)
			decrypted, err := AESDecrypt(out)
			assert.NoError(t, err)
			assert.Equal(t, tc.Value, decrypted)
		})
	}
}

func TestDecryptErrCipherWrongKey(t *testing.T) {
	SetAESEncryptionKey(testEncryptionKey)
	out, _ := AESEncrypt("value")

	SetAESEncryptionKey("dummy")
	SetAESEncryptionFallbackKey("")
	_, err := AESDecrypt([]byte(out))
	assert.Error(t, err)
	assert.EqualError(t, err, "unable to decrypt the data: crypto/aes: invalid key size 5")
}

func TestDecryptErrCipherTextTooShort(t *testing.T) {
	SetAESEncryptionKey(testEncryptionKey)
	out, err := AESDecrypt([]byte("a"))
	assert.Equal(t, "", out)
	assert.EqualError(t, err, ErrCipherTextTooShort.Error())
}
