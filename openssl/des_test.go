package openssl

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DesECBEncrypt(t *testing.T) {
	src := []byte("123456")

	// DES-ECB, PKCS7_PADDING
	key := []byte("12345123")
	dst, err := DesECBEncrypt(src, key, PKCS7_PADDING)
	assert.NoError(t, err)
	t.Log(base64.StdEncoding.EncodeToString(dst))
	assert.Equal(t, base64.StdEncoding.EncodeToString(dst), "RJK5Sd4AS44=")
}

func Test_DesECBDecrypt(t *testing.T) {
	src, err := base64.StdEncoding.DecodeString("RJK5Sd4AS44=")
	assert.NoError(t, err)

	// DES-ECB, PKCS7_PADDING
	key := []byte("12345123")
	dst, err := DesECBDecrypt(src, key, PKCS7_PADDING)
	assert.NoError(t, err)
	t.Log(string(dst))
	assert.Equal(t, dst, []byte("123456"))
}

func Test_DesCBCEncrypt(t *testing.T) {
	src := []byte("123456")
	iv := []byte("67890678")
	// DES-ECB, PKCS7_PADDING
	key := []byte("12345123")
	dst, err := DesCBCEncrypt(src, key, iv, PKCS7_PADDING)
	assert.NoError(t, err)
	t.Log(base64.StdEncoding.EncodeToString(dst))
	assert.Equal(t, base64.StdEncoding.EncodeToString(dst), "fPHNaq8PdWA=")
}

func Test_DesCBCDecrypt(t *testing.T) {
	src, err := base64.StdEncoding.DecodeString("fPHNaq8PdWA=")
	assert.NoError(t, err)

	iv := []byte("67890678")

	// DES-ECB, PKCS7_PADDING
	key := []byte("12345123")
	dst, err := DesCBCDecrypt(src, key, iv, PKCS7_PADDING)
	assert.NoError(t, err)
	t.Log(string(dst))
	assert.Equal(t, dst, []byte("123456"))
}

// go test -v -run Test_DesECBEncryptAndDecrypt
func Test_DesECBEncryptAndDecrypt(t *testing.T) {
	type args struct {
		plainText []byte
		key       []byte
		padding   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal_1",
			args: args{
				plainText: []byte("48dbe70d217358644dedca69483bf6c0"),
				key:       []byte("3f0ddf1c"),
				padding:   ZEROS_PADDING,
			},
			wantErr: false,
		},
		{
			name: "normal_2",
			args: args{
				plainText: []byte("123344"),
				key:       []byte("3f0ddf1c"),
				padding:   ZEROS_PADDING,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DesECBEncrypt(tt.args.plainText, tt.args.key, tt.args.padding)
			if tt.wantErr {
				if err == nil {
					t.Errorf("DesECBEncrypt() error=%v, wantErr=%v", err, tt.wantErr)
				}
				return
			}

			t.Log("got:", base64.StdEncoding.EncodeToString(got))
			plainText, err := DesECBDecrypt([]byte(got), tt.args.key, tt.args.padding)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(plainText, tt.args.plainText) {
				t.Errorf("want=%s, got=%s", string(tt.args.plainText), string(plainText))
			}
		})
	}

	sampleAppSecretEncrypt := "Lr58fjeirOEVNpsvZw2aPAamQSPKIMw+iLccEvP2RiY="
	src, _ := base64.StdEncoding.DecodeString(sampleAppSecretEncrypt)
	txt, err := DesECBDecrypt(src, []byte("3f0ddf1c"), ZEROS_PADDING)
	t.Logf("DesECBDecrypt(\"%s\")=%s, err=%+v\n", sampleAppSecretEncrypt, string(txt), err)
}
