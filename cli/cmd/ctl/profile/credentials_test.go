package profile

import (
	"testing"
	"time"

	"github.com/beclab/Olares/cli/internal/keychain/keychainfake"
	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// staticProfileLister returns a fixed set of olaresIds for List() tests.
type staticProfileLister []string

func (s staticProfileLister) ListOlaresIDs() ([]string, error) { return []string(s), nil }

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
		newOlaresID   string
		switchCurrent bool
		want          expect
	}{
		{
			name:          "first profile, switch=true: becomes current, previous untouched",
			newOlaresID:   "alice@olares.com",
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
			newOlaresID:   "alice@olares.com",
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
			newOlaresID:   "bob@olares.com",
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
			newOlaresID:   "bob@olares.com",
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
			newOlaresID:   "alice@olares.com",
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
			newOlaresID:   "alice@olares.com",
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
			store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{tc.newOlaresID})

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

			res, err := persistTokenAndProfile(cfg, store,
				profileWrite{flags: commonCredFlags{olaresID: tc.newOlaresID}}, tok(), tc.switchCurrent)
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
			if got, _ := store.Get(tc.newOlaresID); got == nil {
				t.Errorf("token for %q not persisted", tc.newOlaresID)
			}
		})
	}
}

// TestPersistTokenAndProfile_PreservesConcurrentWrites is the regression for
// the stale pre-auth snapshot: login reads config, then spends an unbounded
// amount of time on the location probe, the password prompt and a TOTP code.
// Whatever another command cached in that window (role, backend version,
// cluster context) must survive the upsert — persisting a struct captured
// before authentication silently rolled all of it back.
func TestPersistTokenAndProfile_PreservesConcurrentWrites(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	const id = "alice@olares.com"

	// State login sees when it loads config, before authenticating.
	seed := &cliconfig.MultiProfileConfig{
		Profiles:       []cliconfig.ProfileConfig{{OlaresID: id, LocalURLPrefix: "dev."}},
		CurrentProfile: id,
	}
	if err := cliconfig.SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	flags := commonCredFlags{olaresID: id}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	profile, err := ensureProfileWritable(cfg, auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{id}), flags, time.Now())
	if err != nil {
		t.Fatalf("ensureProfileWritable: %v", err)
	}
	if profile.LocalURLPrefix != "dev." {
		t.Fatalf("re-login without --local-url-prefix should keep the existing one, got %q", profile.LocalURLPrefix)
	}

	// Another command lands while login is still waiting on a TOTP code.
	if err := cliconfig.UpdateLocked(func(c *cliconfig.MultiProfileConfig) error {
		p := c.FindByOlaresID(id)
		p.OwnerRole = "admin"
		p.WhoamiRefreshedAt = 4242
		p.BackendVersion = "1.13.0"
		p.ClusterContextRefreshedAt = 999
		p.LocationUnreachableAt = 555
		return nil
	}); err != nil {
		t.Fatalf("concurrent write: %v", err)
	}

	store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{id})
	res, err := persistTokenAndProfile(cfg, store, profileWrite{
		flags:    flags,
		location: olares.LocationHost,
		probedAt: 777,
	}, &auth.Token{AccessToken: "AT-1", RefreshToken: "RT-1"}, true)
	if err != nil {
		t.Fatalf("persistTokenAndProfile: %v", err)
	}

	persisted, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p := persisted.FindByOlaresID(id)
	if p == nil {
		t.Fatal("profile vanished")
	}
	for _, c := range []struct {
		field string
		got   any
		want  any
	}{
		{"OwnerRole", p.OwnerRole, "admin"},
		{"WhoamiRefreshedAt", p.WhoamiRefreshedAt, int64(4242)},
		{"BackendVersion", p.BackendVersion, "1.13.0"},
		{"ClusterContextRefreshedAt", p.ClusterContextRefreshedAt, int64(999)},
		// Flag-driven fields the user didn't repeat are still carried over.
		{"LocalURLPrefix", p.LocalURLPrefix, "dev."},
		// Login owns these two, so they must reflect this run's probe.
		{"Location", p.Location, string(olares.LocationHost)},
		{"LocationProbedAt", p.LocationProbedAt, int64(777)},
		// We just reached the instance, so a cooldown stamped meanwhile is stale.
		{"LocationUnreachableAt", p.LocationUnreachableAt, int64(0)},
	} {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.field, c.got, c.want)
		}
	}

	// Callers must be able to build on what actually landed, not their own
	// pre-auth copy.
	if res.Profile.OwnerRole != "admin" || res.Profile.Location != string(olares.LocationHost) {
		t.Errorf("res.Profile = %+v, want the persisted record", res.Profile)
	}
}
