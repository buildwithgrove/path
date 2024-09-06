package morse

import (
	"testing"

	"github.com/pokt-foundation/pocket-go/provider"

	"github.com/stretchr/testify/require"
)

func TestApplication_AAT(t *testing.T) {
	tests := []struct {
		name string
		app  SignedAAT
		want provider.PocketAAT
	}{
		{
			name: "should return PocketAAT representation of the application",
			app: SignedAAT{
				ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			want: provider.PocketAAT{
				ClientPubKey: "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				AppPubKey:    "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				Signature:    "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
				Version:      AATVersion,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got := test.app.AAT()
			c.Equal(test.want, got)
		})
	}
}

func TestApplication_validate(t *testing.T) {
	tests := []struct {
		name    string
		app     SignedAAT
		wantErr bool
	}{
		{
			name: "should pass with valid application",
			app: SignedAAT{
				ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			wantErr: false,
		},
		{
			name: "should fail with invalid application public key",
			app: SignedAAT{
				ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				ApplicationPublicKey: "invalid_public_key",
				ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			wantErr: true,
		},
		{
			name: "should fail with invalid client public key",
			app: SignedAAT{
				ClientPublicKey:      "invalid_client_key",
				ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			wantErr: true,
		},
		{
			name: "should fail with invalid application signature",
			app: SignedAAT{
				ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				ApplicationSignature: "invalid_signature",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			err := test.app.validate()
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
			}
		})
	}
}

func Test_configToPocketAAT(t *testing.T) {
	tests := []struct {
		name string
		app  SignedAAT
		want provider.PocketAAT
	}{
		{
			name: "should convert SignedAAT to PocketAAT",
			app: SignedAAT{
				ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
			},
			want: provider.PocketAAT{
				ClientPubKey: "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
				AppPubKey:    "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
				Signature:    "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
				Version:      AATVersion,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got := configToPocketAAT(test.app)
			c.Equal(test.want, got)
		})
	}
}
