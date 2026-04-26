package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings integration accounts add ...`
//
// Backed by POST /api/account/create on user-service. The SPA's
// stores/settings/integration.ts createAccount() and the
// IntegrationAddInputs.vue / AddAccountDialog.vue flow build the body
// shape this CLI verb emits:
//
//   {
//     "name": <accessKeyID>,
//     "type": "awss3" | "tencent",
//     "raw_data": {
//       "access_token":  <accessKeySecret>,
//       "refresh_token": <accessKeySecret>,
//       "endpoint":      <url>,
//       "bucket":        <name>,        // optional in the SPA
//       "expires_in":    0,
//       "expires_at":    0
//     }
//   }
//
// The SPA's Add-Account dialog only collects {accessKeyID, accessKeySecret,
// endpoint, bucket} for object-storage flows and reuses the same shape
// for both AWS S3 and Tencent COS — that's what we expose here.
//
// The OAuth-style add flows (Google Drive, Dropbox, Olares Space NFT
// binding) need a browser / wallet redirect to obtain the access token,
// so they stay in the SPA — there's no useful one-shot CLI surface.

func newAccountsAddCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add an integration account (object-storage flows)",
		Long: `Add an integration account against user-service's
POST /api/account/create.

Phase 2 ships the *direct* object-storage flows that don't need an
OAuth/wallet redirect:

  add awss3      AWS S3 (or any S3-compatible endpoint)
  add tencent    Tencent COS

OAuth flows (Google Drive, Dropbox, Olares Space) stay in the SPA — the
access tokens they produce are scoped to a browser session and have no
useful one-shot CLI capture.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAccountsAddObjectStorageCommand(f, "awss3", "AWS S3", "AWS S3 (or any S3-compatible endpoint)"))
	cmd.AddCommand(newAccountsAddObjectStorageCommand(f, "tencent", "Tencent COS", "Tencent COS"))
	return cmd
}

// newAccountsAddObjectStorageCommand returns one of the
// `accounts add awss3 ...` / `accounts add tencent ...` subcommands.
// The two share a wire shape and a flag set, so we factor them through
// a single constructor parameterized by AccountType.
func newAccountsAddObjectStorageCommand(f *cmdutil.Factory, accountType, displayName, longBlurb string) *cobra.Command {
	var (
		accessKeyID     string
		accessKeySecret string
		endpoint        string
		bucket          string
	)
	cmd := &cobra.Command{
		Use:   accountType,
		Short: "add a " + displayName + " integration account",
		Long: fmt.Sprintf(`Add a %s integration account.

The flag set mirrors the SPA's Add-Account dialog under
Settings -> Integration > +Add (object storage). All three of
--access-key-id, --access-key-secret, --endpoint are required;
--bucket is optional (the SPA labels it "optional" too).

Example:
  olares-cli settings integration accounts add %s \
    --access-key-id  AKIA... \
    --access-key-secret  XYZ... \
    --endpoint  https://s3.us-east-1.amazonaws.com \
    --bucket    olares-backup
`, longBlurb, accountType),
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runAccountsAddObjectStorage(c.Context(), f, accountType, accessKeyID, accessKeySecret, endpoint, bucket)
		},
	}
	cmd.Flags().StringVar(&accessKeyID, "access-key-id", "", "access-key id (used as the account name)")
	cmd.Flags().StringVar(&accessKeySecret, "access-key-secret", "", "access-key secret")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3-compatible endpoint URL")
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket name (optional)")
	_ = cmd.MarkFlagRequired("access-key-id")
	_ = cmd.MarkFlagRequired("access-key-secret")
	_ = cmd.MarkFlagRequired("endpoint")
	return cmd
}

// objectStorageRawData mirrors the SPA's allAccountValues() output in
// IntegrationAddInputs.vue. Both access_token and refresh_token are
// populated with the secret because the upstream provider treats them
// as one and the same for these direct (non-OAuth) flows.
type objectStorageRawData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int    `json:"expires_at"`
}

type objectStorageAccountReq struct {
	Name    string               `json:"name"`
	Type    string               `json:"type"`
	RawData objectStorageRawData `json:"raw_data"`
}

func runAccountsAddObjectStorage(ctx context.Context, f *cmdutil.Factory, accountType, accessKeyID, accessKeySecret, endpoint, bucket string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	accessKeyID = strings.TrimSpace(accessKeyID)
	accessKeySecret = strings.TrimSpace(accessKeySecret)
	endpoint = strings.TrimSpace(endpoint)
	bucket = strings.TrimSpace(bucket)
	if accessKeyID == "" || accessKeySecret == "" || endpoint == "" {
		return fmt.Errorf("--access-key-id, --access-key-secret and --endpoint are required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	req := objectStorageAccountReq{
		Name: accessKeyID,
		Type: accountType,
		RawData: objectStorageRawData{
			AccessToken:  accessKeySecret,
			RefreshToken: accessKeySecret,
			Endpoint:     endpoint,
			Bucket:       bucket,
		},
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/account/create", req, nil); err != nil {
		return err
	}
	fmt.Printf("Added %s account %q.\n", accountType, accessKeyID)
	return nil
}
