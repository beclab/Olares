package entranceids

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	locks "github.com/beclab/api/pkg/entranceids"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const markerNamespace = "os-system"

// Run owns the common renewable write lease for the whole upgrade. The
// immutable journal is also the 0600 backup, containing potentially private
// domain/certificate settings. It must be retained across retries.
func Run(ctx context.Context, cs versioned.Interface, kube kubernetes.Interface, journal string) error {
	ctx, release, err := locks.AcquireWriteLock(ctx, kube)
	if err != nil {
		return err
	}
	defer release()
	return RunLocked(ctx, cs, kube, journal)
}

// RunLocked requires the caller to own the entrance-domain lease.
func RunLocked(ctx context.Context, cs versioned.Interface, kube kubernetes.Interface, journal string) error {
	marker, err := kube.CoreV1().ConfigMaps(markerNamespace).Get(ctx, Migration, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		marker = nil
	}
	term, err := cs.SysV1alpha1().Terminuses().Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if term.UID == "" || term.Spec.LocalName == "" {
		return fmt.Errorf("Terminus UID and spec.localName are required")
	}
	if marker != nil {
		if marker.Data["clusterUID"] != string(term.UID) || marker.Data["originalUser"] != term.Spec.LocalName {
			return fmt.Errorf("entrance migration marker identity mismatch")
		}
		if marker.Data["status"] == "complete" {
			return nil
		}
		if marker.Data["status"] != "planned" {
			return fmt.Errorf("unexpected entrance migration marker state")
		}
	}
	var p Plan
	raw, err := os.ReadFile(journal)
	if err == nil {
		if err = json.Unmarshal(raw, &p); err != nil {
			return fmt.Errorf("invalid saved migration plan: %w", err)
		}
	} else if os.IsNotExist(err) {
		if marker != nil {
			return fmt.Errorf("saved migration journal is missing; restore its backup before retrying")
		}
		list, err := cs.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		built, err := Build(list.Items, term.Spec.LocalName, term.UID)
		if err != nil {
			return err
		}
		p = *built
		raw, err = json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		if err = saveExclusive(journal, raw); err != nil {
			return err
		}
	} else {
		return err
	}
	if p.Migration != Migration || p.ClusterUID != term.UID || p.OriginalUser != term.Spec.LocalName {
		return fmt.Errorf("saved entrance migration plan identity mismatch")
	}
	digest := sha256.Sum256(raw)
	checksum := hex.EncodeToString(digest[:])
	if marker == nil {
		marker, err = kube.CoreV1().ConfigMaps(markerNamespace).Create(ctx, &core.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: Migration, Namespace: markerNamespace}, Data: map[string]string{"status": "planned", "clusterUID": string(term.UID), "originalUser": term.Spec.LocalName, "planSHA256": checksum, "journal": journal}}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if marker.Data["planSHA256"] != checksum {
		return fmt.Errorf("saved migration journal checksum mismatch")
	}
	for _, change := range p.Changes {
		current, err := cs.AppV1alpha1().Applications().Get(ctx, change.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if current.UID != change.UID {
			return fmt.Errorf("Application %s was replaced after planning", change.Name)
		}
		if !equalSpec(current.Spec, change.After) {
			if !equalSpec(current.Spec, change.Before) {
				return fmt.Errorf("Application %s drifted from the saved upgrade snapshot", change.Name)
			}
			candidate := current.DeepCopy()
			candidate.Spec = change.After
			saved, err := cs.AppV1alpha1().Applications().Update(ctx, candidate, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			if !equalSpec(saved.Spec, change.After) {
				return fmt.Errorf("Application %s did not retain entrance IDs; check the CRD schema", change.Name)
			}
		}
		receipt := filepath.Join(journal+".progress", change.Name)
		if _, err = os.Stat(receipt); os.IsNotExist(err) {
			if err = saveExclusive(receipt, []byte(string(change.UID))); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	marker.Data["status"] = "complete"
	_, err = kube.CoreV1().ConfigMaps(markerNamespace).Update(ctx, marker, metav1.UpdateOptions{})
	return err
}
func equalSpec(a, b interface{}) bool {
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	var x, y interface{}
	_ = json.Unmarshal(aa, &x)
	_ = json.Unmarshal(bb, &y)
	return reflect.DeepEqual(x, y)
}
func saveExclusive(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".entrance-upgrade-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	// Atomic publication, refusing to overwrite an existing backup/plan.
	if err = os.Link(tmp, path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
