package wizard

import (
	"fmt"
	"log"
	"time"
)

// Synchronize fetches AuthInfo, Account, Orgs and Vaults from the server,
// merges any local changes back in, and persists everything to the local
// Storage. Mirrors App.synchronize in apps/packages/sdk/src/core/app.ts.
//
// Requires the AppState to already have an unlocked account (call
// Account.Unlock before invoking this).
func (a *App) Synchronize() error {
	if a.State == nil {
		return fmt.Errorf("app has no AppState")
	}
	unlocked := a.State.Unlocked()
	if unlocked == nil {
		return fmt.Errorf("synchronize requires an unlocked account")
	}

	// 1. AuthInfo
	if info, err := a.API.GetAuthInfo(); err != nil {
		log.Printf("synchronize: getAuthInfo failed: %v", err)
	} else {
		a.State.AuthInfo = info
	}

	// 2. Refresh Account
	freshAccount, err := a.API.GetAccount()
	if err != nil {
		return fmt.Errorf("synchronize: getAccount failed: %w", err)
	}
	unlocked.Account = freshAccount
	a.State.SetAccount(freshAccount)

	// 3. Orgs
	a.State.Orgs = a.State.Orgs[:0]
	for _, info := range freshAccount.Orgs {
		org, err := a.API.GetOrg(info.ID)
		if err != nil {
			log.Printf("synchronize: getOrg %s failed: %v", info.ID, err)
			continue
		}
		a.State.Orgs = append(a.State.Orgs, *org)
	}

	// 4. Vaults
	if err := a.syncVaults(unlocked); err != nil {
		return fmt.Errorf("synchronize: syncVaults failed: %w", err)
	}

	a.State.LastSync = time.Now().UTC().Format(time.RFC3339Nano)
	if err := a.State.Save(); err != nil {
		log.Printf("synchronize: failed to persist app state: %v", err)
	}
	return nil
}

// syncVaults pulls the main vault and every shared vault available
// through the user's orgs, merges in any local changes, and pushes
// changes back. Errors on individual vaults are logged but do not stop
// the overall sync.
func (a *App) syncVaults(unlocked *UnlockedAccount) error {
	// Build the ordered list of (vaultID, isMain) tuples to sync.
	type vaultSpec struct {
		ID     string
		IsMain bool
	}
	var specs []vaultSpec
	if id := unlocked.Account.MainVault.ID; id != "" {
		specs = append(specs, vaultSpec{ID: id, IsMain: true})
	}
	for _, org := range a.State.Orgs {
		for _, v := range org.Vaults {
			specs = append(specs, vaultSpec{ID: v.ID})
		}
	}

	for _, spec := range specs {
		if err := a.syncVault(unlocked, spec.ID); err != nil {
			log.Printf("synchronize: vault %s failed: %v", spec.ID, err)
		}
	}
	return nil
}

// syncVault is the per-vault flow: fetch → unlock → merge with local →
// commit + push if there were local changes → store.
func (a *App) syncVault(unlocked *UnlockedAccount, id string) error {
	remote, err := a.API.GetVault(id)
	if err != nil {
		return fmt.Errorf("getVault: %w", err)
	}
	if err := remote.Unlock(unlocked); err != nil {
		return fmt.Errorf("unlock remote vault: %w", err)
	}

	local := a.State.GetVault(id)
	if local != nil && local.items != nil && local.items.HasChanges() {
		// Local has unsynced edits — unlock the local copy too so we can
		// merge then push back.
		if local.aesKey == nil {
			if err := local.Unlock(unlocked); err != nil {
				log.Printf("syncVault: failed to unlock local %s, dropping local: %v", id, err)
				local = nil
			}
		}
		if local != nil {
			local.Merge(remote)
			if err := local.Commit(); err != nil {
				return fmt.Errorf("commit merged vault: %w", err)
			}
			pushed, err := a.API.UpdateVault(*local)
			if err != nil {
				return fmt.Errorf("updateVault: %w", err)
			}
			pushed.aesKey = local.aesKey
			pushed.items = local.items
			if err := pushed.Unlock(unlocked); err != nil {
				return fmt.Errorf("re-unlock pushed vault: %w", err)
			}
			pushed.MarkSynced(time.Now())
			a.State.PutVault(*pushed)
			return nil
		}
	}

	a.State.PutVault(*remote)
	return nil
}

// CreateItem builds a fresh VaultItem, adds it to the given vault,
// commits, pushes to the server, and re-unlocks (so the vault is left
// in a usable state). Mirrors App.createItem → addItems → saveVault →
// syncVault in TS.
func (a *App) CreateItem(params CreateItemParams) (*VaultItem, error) {
	if a.State == nil || a.State.Unlocked() == nil {
		return nil, fmt.Errorf("CreateItem requires an unlocked account")
	}
	if params.Vault == nil {
		return nil, fmt.Errorf("CreateItem requires a vault")
	}
	unlocked := a.State.Unlocked()
	account := unlocked.Account

	now := time.Now().UTC().Format(time.RFC3339Nano)
	itemID := params.ID
	if itemID == "" {
		itemID = generateUUID()
	}
	item := VaultItem{
		ID:        itemID,
		Name:      params.Name,
		Type:      params.Type,
		Icon:      params.Icon,
		Fields:    params.Fields,
		Tags:      append([]string{}, params.Tags...),
		Updated:   now,
		UpdatedBy: account.ID,
	}

	vault := params.Vault
	log.Printf("CreateItem: enter id=%s revision=%s accessors=%d encryptedData=%d aesKey?=%t items?=%t itemCount=%d hasChanges=%t",
		vault.ID, vault.Revision, len(vault.Accessors), len(vault.EncryptedData),
		vault.aesKey != nil, vault.items != nil, vaultItemCount(vault), vaultHasChanges(vault))

	if vault.aesKey == nil {
		log.Printf("CreateItem: vault aesKey nil, calling Unlock to bootstrap")
		if err := vault.Unlock(unlocked); err != nil {
			return nil, fmt.Errorf("unlock vault: %w", err)
		}
		log.Printf("CreateItem: post-Unlock accessors=%d encryptedData=%d aesKey?=%t",
			len(vault.Accessors), len(vault.EncryptedData), vault.aesKey != nil)
	}

	vault.AddItems(item)
	{
		fieldType := ""
		fieldValueLen := 0
		if len(item.Fields) > 0 {
			fieldType = string(item.Fields[0].Type)
			fieldValueLen = len(item.Fields[0].Value)
		}
		log.Printf("CreateItem: post-AddItems itemCount=%d added id=%s type=%d fields=%d firstFieldType=%s firstFieldValueLen=%d",
			vaultItemCount(vault), item.ID, item.Type, len(item.Fields), fieldType, fieldValueLen)
	}

	if err := vault.Commit(); err != nil {
		return nil, fmt.Errorf("commit vault: %w", err)
	}
	log.Printf("CreateItem: post-Commit encryptedData=%d iv?=%t aad?=%t",
		len(vault.EncryptedData),
		vault.EncryptionParams.IV != "", vault.EncryptionParams.AdditionalData != "")

	log.Printf("CreateItem: pre-UpdateVault id=%s revision=%s encryptedData=%d accessors=%d",
		vault.ID, vault.Revision, len(vault.EncryptedData), len(vault.Accessors))
	pushed, err := a.API.UpdateVault(*vault)
	if err != nil {
		log.Printf("CreateItem: API.UpdateVault failed: %+v", err)
		return nil, fmt.Errorf("updateVault: %w", err)
	}
	log.Printf("CreateItem: post-UpdateVault id=%s revision=%s encryptedData=%d accessors=%d",
		pushed.ID, pushed.Revision, len(pushed.EncryptedData), len(pushed.Accessors))

	pushed.aesKey = vault.aesKey
	pushed.items = vault.items
	if err := pushed.Unlock(unlocked); err != nil {
		return nil, fmt.Errorf("re-unlock pushed vault: %w", err)
	}
	pushed.MarkSynced(time.Now())
	a.State.PutVault(*pushed)
	if err := a.State.Save(); err != nil {
		log.Printf("CreateItem: persist app state failed: %v", err)
	}
	return &item, nil
}

// CreateItemParams is the Go counterpart of TS CreateItemParams.
type CreateItemParams struct {
	ID     string // optional; if empty, a new UUID is generated
	Name   string
	Vault  *Vault
	Fields []Field
	Tags   []string
	Icon   string
	Type   VaultType
}

// vaultItemCount returns the number of items currently held in the
// vault's in-memory collection (0 if not yet unlocked).
func vaultItemCount(v *Vault) int {
	if v == nil || v.items == nil {
		return 0
	}
	return len(v.items.Items)
}

// vaultHasChanges reports whether the vault's in-memory items collection
// has any pending local changes.
func vaultHasChanges(v *Vault) bool {
	if v == nil || v.items == nil {
		return false
	}
	return v.items.HasChanges()
}

// MainVault returns the (possibly nil) Vault stored locally that
// corresponds to account.mainVault.id.
func (a *App) MainVault() *Vault {
	if a.State == nil {
		return nil
	}
	acc := a.State.GetAccount()
	if acc == nil || acc.MainVault.ID == "" {
		return nil
	}
	return a.State.GetVault(acc.MainVault.ID)
}
