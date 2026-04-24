package profile

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// TestPersistTokenAndProfile_Switching exercises the auto-switch contract
// added by the "login auto switch profile" plan. The behavior matrix lives
// in docs/notes/olares-cli-auth-profile-config.md §7.3; this table mirrors
// the verification checklist from the plan.
func TestPersistTokenAndProfile_Switching(t *testing.T) {
	tok := func() *auth.Token {
		return &auth.Token{AccessToken: "ignored", RefreshToken: "ignored"}
	}

	type expect struct {
		current  string
		previous string
		switched bool
		prevPtr  string // res.PreviousCurrent
	}

	cases := []struct {
		name          string
		seedProfiles  []cliconfig.ProfileConfig
		seedCurrent   string
		seedPrevious  string
		newProfile    cliconfig.ProfileConfig
		switchCurrent bool
		want          expect
	}{
		{
			name:          "first profile, switch=true: becomes current, previous untouched",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "alice@olares.com"},
			switchCurrent: true,
			want: expect{
				current:  "alice@olares.com",
				previous: "",
				switched: true,
				prevPtr:  "", // there was no prior current to demote
			},
		},
		{
			name:          "first profile, --no-switch: still becomes current (bootstrap fallback)",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "alice@olares.com"},
			switchCurrent: false,
			want: expect{
				current:  "alice@olares.com",
				previous: "",
				switched: true,
				prevPtr:  "",
			},
		},
		{
			name:          "different profile, switch=true: old current → previous, new is current",
			seedProfiles:  []cliconfig.ProfileConfig{{OlaresID: "alice@olares.com"}},
			seedCurrent:   "alice@olares.com",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "bob@olares.com"},
			switchCurrent: true,
			want: expect{
				current:  "bob@olares.com",
				previous: "alice@olares.com",
				switched: true,
				prevPtr:  "alice@olares.com",
			},
		},
		{
			name:          "different profile, --no-switch: current/previous untouched",
			seedProfiles:  []cliconfig.ProfileConfig{{OlaresID: "alice@olares.com"}},
			seedCurrent:   "alice@olares.com",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "bob@olares.com"},
			switchCurrent: false,
			want: expect{
				current:  "alice@olares.com",
				previous: "",
				switched: false,
			},
		},
		{
			name:          "same-account re-login, switch=true: no-op, no switched signal",
			seedProfiles:  []cliconfig.ProfileConfig{{OlaresID: "alice@olares.com"}},
			seedCurrent:   "alice@olares.com",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "alice@olares.com"},
			switchCurrent: true,
			want: expect{
				current:  "alice@olares.com",
				previous: "",
				switched: false,
			},
		},
		{
			name: "same-account re-login preserves PreviousProfile that was already set",
			seedProfiles: []cliconfig.ProfileConfig{
				{OlaresID: "alice@olares.com"},
				{OlaresID: "bob@olares.com"},
			},
			seedCurrent:   "alice@olares.com",
			seedPrevious:  "bob@olares.com",
			newProfile:    cliconfig.ProfileConfig{OlaresID: "alice@olares.com"},
			switchCurrent: true,
			want: expect{
				current:  "alice@olares.com",
				previous: "bob@olares.com", // untouched
				switched: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OLARES_CLI_HOME", t.TempDir())
			path, err := cliconfig.TokensFile()
			if err != nil {
				t.Fatalf("TokensFile: %v", err)
			}
			store := auth.NewFileStoreAt(path)

			cfg := &cliconfig.MultiProfileConfig{
				Profiles:        append([]cliconfig.ProfileConfig(nil), tc.seedProfiles...),
				CurrentProfile:  tc.seedCurrent,
				PreviousProfile: tc.seedPrevious,
			}
			if len(tc.seedProfiles) > 0 {
				if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
					t.Fatalf("seed save config: %v", err)
				}
			}

			res, err := persistTokenAndProfile(cfg, store, tc.newProfile, tok(), tc.switchCurrent)
			if err != nil {
				t.Fatalf("persistTokenAndProfile: %v", err)
			}

			if cfg.CurrentProfile != tc.want.current {
				t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, tc.want.current)
			}
			if cfg.PreviousProfile != tc.want.previous {
				t.Errorf("PreviousProfile = %q, want %q", cfg.PreviousProfile, tc.want.previous)
			}
			if res.Switched != tc.want.switched {
				t.Errorf("res.Switched = %v, want %v", res.Switched, tc.want.switched)
			}
			if res.PreviousCurrent != tc.want.prevPtr {
				t.Errorf("res.PreviousCurrent = %q, want %q", res.PreviousCurrent, tc.want.prevPtr)
			}

			// Cross-check on-disk state matches in-memory state, since the
			// helper is supposed to have flushed via SaveMultiProfileConfig.
			persisted, err := cliconfig.LoadMultiProfileConfig()
			if err != nil {
				t.Fatalf("reload config: %v", err)
			}
			if persisted.CurrentProfile != cfg.CurrentProfile {
				t.Errorf("on-disk CurrentProfile = %q, want %q", persisted.CurrentProfile, cfg.CurrentProfile)
			}
			if persisted.PreviousProfile != cfg.PreviousProfile {
				t.Errorf("on-disk PreviousProfile = %q, want %q", persisted.PreviousProfile, cfg.PreviousProfile)
			}
			if got, _ := store.Get(tc.newProfile.OlaresID); got == nil {
				t.Errorf("token for %q not persisted", tc.newProfile.OlaresID)
			}
		})
	}
}
