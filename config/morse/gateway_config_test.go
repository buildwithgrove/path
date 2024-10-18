package morse

import (
	"testing"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/stretchr/testify/require"

	morseRelayer "github.com/buildwithgrove/path/relayer/morse"
)

func Test_GetSignedAAT(t *testing.T) {
	tests := []struct {
		name   string
		config MorseGatewayConfig
		appID  string
		want   provider.PocketAAT
		ok     bool
	}{
		{
			name: "should return valid PocketAAT for existing appID",
			config: MorseGatewayConfig{
				SignedAATs: map[string]SignedAAT{
					"af929e588bb37d8e6bbc8cb25ba4b4d9383f9238": {
						ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
						ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
						ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
					},
				},
			},
			appID: "af929e588bb37d8e6bbc8cb25ba4b4d9383f9238",
			want: provider.PocketAAT{
				Version:      "0.0.1",
				ClientPubKey: "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				AppPubKey:    "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				Signature:    "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			ok: true,
		},
		{
			name: "should return false for non-existing appID",
			config: MorseGatewayConfig{
				SignedAATs: map[string]SignedAAT{
					"af929e588bb37d8e6bbc8cb25ba4b4d9383f9238": {
						ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
						ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
						ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
					},
				},
			},
			appID: "who_am_i_tho",
			want:  provider.PocketAAT{},
			ok:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got, ok := test.config.GetSignedAAT(test.appID)
			c.Equal(test.ok, ok)
			if ok {
				c.Equal(test.want, got)
			}
		})
	}
}

func Test_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MorseGatewayConfig
		wantErr bool
	}{
		{
			name: "should pass with valid config",
			config: MorseGatewayConfig{
				FullNodeConfig: morseRelayer.FullNodeConfig{
					URL:             "https://full-node-url.io",
					RelaySigningKey: "05d126124d35fd7c645b78bf3128b989d03fa2c38cd69a81742b0dedbf9ca05aab35ab6f5137076136d0ef926a37fb3ac70249c3b0266b95d4b5db85a11fef8e",
					HttpConfig:      morseRelayer.HttpConfig{Retries: 3, Timeout: 5000000000},
					RequestConfig:   provider.RequestConfigOpts{Retries: 3},
				},
				SignedAATs: map[string]SignedAAT{
					"af929e588bb37d8e6bbc8cb25ba4b4d9383f9238": {
						ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
						ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
						ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should fail with invalid application ID",
			config: MorseGatewayConfig{
				FullNodeConfig: morseRelayer.FullNodeConfig{
					URL:             "https://full-node-url.io",
					RelaySigningKey: "05d126124d35fd7c645b78bf3128b989d03fa2c38cd69a81742b0dedbf9ca05aab35ab6f5137076136d0ef926a37fb3ac70249c3b0266b95d4b5db85a11fef8e",
					HttpConfig:      morseRelayer.HttpConfig{Retries: 3, Timeout: 5000000000},
					RequestConfig:   provider.RequestConfigOpts{Retries: 3},
				},
				SignedAATs: map[string]SignedAAT{
					"invalid_app_id": {
						ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
						ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
						ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should fail with invalid full node URL",
			config: MorseGatewayConfig{
				FullNodeConfig: morseRelayer.FullNodeConfig{
					URL:             "invalid-url",
					RelaySigningKey: "05d126124d35fd7c645b78bf3128b989d03fa2c38cd69a81742b0dedbf9ca05aab35ab6f5137076136d0ef926a37fb3ac70249c3b0266b95d4b5db85a11fef8e",
					HttpConfig:      morseRelayer.HttpConfig{Retries: 3, Timeout: 5000000000},
					RequestConfig:   provider.RequestConfigOpts{Retries: 3},
				},
				SignedAATs: map[string]SignedAAT{
					"af929e588bb37d8e6bbc8cb25ba4b4d9383f9238": {
						ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
						ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
						ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			err := test.config.Validate()
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
			}
		})
	}
}
