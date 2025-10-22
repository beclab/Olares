package user

import (
	"fmt"
	"log"
	"strings"

	"github.com/beclab/Olares/cli/pkg/wizard"
	"github.com/spf13/cobra"
)

type activateUserOptions struct {
	Mnemonic string
	BflUrl   string
	VaultUrl string
	Password string
	OlaresId string

	Location     string
	Language     string
	EnableTunnel bool
	Host         string
	Jws          string
}

func NewCmdActivateUser() *cobra.Command {
	o := &activateUserOptions{}
	cmd := &cobra.Command{
		Use:   "activate {Olares ID (e.g., user@example.com)}",
		Short: "activate a new user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.OlaresId = args[0]
			if err := o.Validate(); err != nil {
				log.Fatal(err)
			}
			if err := o.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func (o *activateUserOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Mnemonic, "mnemonic", "", "12-word mnemonic phrase, required for activation")
	cmd.Flags().StringVar(&o.BflUrl, "bfl", "", "Bfl URL (e.g., https://example.com, default: http://127.0.0.1:30180)")
	cmd.Flags().StringVar(&o.VaultUrl, "vault", "", "Vault URL (e.g., https://example.com, default: http://127.0.0.1:30181)")
	cmd.Flags().StringVarP(&o.Password, "password", "p", "", "OS password for authentication, required for activation")
	cmd.Flags().StringVar(&o.Location, "location", "Asia/Shanghai", "Timezone location (default: Asia/Shanghai)")
	cmd.Flags().StringVar(&o.Language, "language", "en-US", "System language (default: en-US)")
	cmd.Flags().BoolVar(&o.EnableTunnel, "enable-tunnel", false, "Enable tunnel mode (default: false)")
	cmd.Flags().StringVar(&o.Host, "host", "", "FRP host (only used when tunnel is enabled)")
	cmd.Flags().StringVar(&o.Jws, "jws", "", "FRP JWS token (only used when tunnel is enabled)")
}

func (o *activateUserOptions) Validate() error {
	if o.OlaresId == "" {
		return fmt.Errorf("Olares ID is required")
	}
	if o.Password == "" {
		return fmt.Errorf("Password is required")
	}
	if o.Mnemonic == "" {
		return fmt.Errorf("Mnemonic is required")
	}
	return nil
}

func (c *activateUserOptions) Run() error {
	log.Println("=== TermiPass CLI - User Bind Terminus ===")

	localName := c.OlaresId
	if strings.Contains(c.OlaresId, "@") {
		localName = strings.Split(c.OlaresId, "@")[0]
	}

	log.Printf("Parameters:")
	log.Printf("  BflUrl: %s", c.BflUrl)
	log.Printf("  VaultUrl: %s", c.VaultUrl)
	log.Printf("  Terminus Name: %s", c.OlaresId)
	log.Printf("  Local Name: %s", localName)

	log.Printf("Initializing global stores with mnemonic...")
	err := wizard.InitializeGlobalStores(c.Mnemonic, c.OlaresId)
	if err != nil {
		return fmt.Errorf("failed to initialize global stores: %v", err)
	}

	err = wizard.UserBindTerminus(c.Mnemonic, c.BflUrl, c.VaultUrl, c.Password, c.OlaresId, localName)
	if err != nil {
		return fmt.Errorf("user bind failed: %v", err)
	}

	log.Printf("✅ Vault activation completed successfully!")
	log.Printf("🚀 Starting system activation wizard...")

	wizardConfig := wizard.CustomWizardConfig(c.Location, c.Language, c.EnableTunnel, c.Host, c.Jws, c.Password, c.Password)

	log.Printf("Wizard configuration:")
	log.Printf("  Location: %s", wizardConfig.System.Location)
	log.Printf("  Language: %s", wizardConfig.System.Language)
	log.Printf("  Enable Tunnel: %t", c.EnableTunnel)
	if c.EnableTunnel && wizardConfig.System.FRP != nil {
		log.Printf("  FRP Host: %s", wizardConfig.System.FRP.Host)
		log.Printf("  FRP JWS: %s", wizardConfig.System.FRP.Jws)
	}

	err = wizard.RunActivationWizard(c.BflUrl, "", wizardConfig)
	if err != nil {
		return fmt.Errorf("activation wizard failed: %v", err)
	}

	log.Printf("🎉 Complete Terminus activation finished successfully!")
	return nil
}
