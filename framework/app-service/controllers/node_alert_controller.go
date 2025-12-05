package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bytetrade.io/web3os/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NodePressureType represents the type of node pressure
type NodePressureType string

const (
	MemoryPressure NodePressureType = "MemoryPressure"
	DiskPressure   NodePressureType = "DiskPressure"
	PIDPressure    NodePressureType = "PIDPressure"
)

// NodeAlertMessage represents the message structure for node alerts
type alertPayload struct {
	NodeName     string           `json:"nodeName"`
	PressureType NodePressureType `json:"pressureType"`
	Timestamp    time.Time        `json:"timestamp"`
	Message      string           `json:"message"`
	Status       bool             `json:"status"`
}

type NodeAlertEvent struct {
	Topic   NodePressureType `json:"topic"`
	Payload alertPayload     `json:"payload"`
}

// NodeAlertController reconciles a Node object
type NodeAlertController struct {
	client.Client
	KubeConfig *rest.Config
	// lastAlertTime tracks the last time an alert was sent for each pressure type
	lastAlertTime map[string]time.Time
	// lastPressureState tracks the last known pressure state for each node and pressure type
	lastPressureState map[string]bool
	mutex             sync.RWMutex
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeAlertController) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("node-alert-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              r,
	})
	if err != nil {
		klog.Errorf("node-alert-controller setup failed %v", err)
		return fmt.Errorf("node-alert-controller setup failed %w", err)
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&corev1.Node{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, node *corev1.Node) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name: node.GetName(),
				}}}
			}),
		predicate.TypedFuncs[*corev1.Node]{
			CreateFunc: func(e event.TypedCreateEvent[*corev1.Node]) bool {
				return true
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*corev1.Node]) bool {
				return true
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*corev1.Node]) bool {
				return false
			},
		},
	))

	if err != nil {
		klog.Errorf("node-alert-controller add watch failed %v", err)
		return fmt.Errorf("add watch failed %w", err)
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *NodeAlertController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("start reconcile node %s", req.Name)

	if r.lastAlertTime == nil {
		r.lastAlertTime = make(map[string]time.Time)
	}
	if r.lastPressureState == nil {
		r.lastPressureState = make(map[string]bool)
	}

	node := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		klog.Errorf("failed to get node %s: %v", req.Name, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.checkNodePressure(node)
	if err != nil {
		klog.Errorf("check node pressure failed %v", err)
		return ctrl.Result{}, err
	}
	klog.Infof("finished reconcile node %s", req.Name)
	return ctrl.Result{}, nil
}

// checkNodePressure checks for various pressure conditions on the node
func (r *NodeAlertController) checkNodePressure(node *corev1.Node) error {
	pressureTypes := []NodePressureType{MemoryPressure, DiskPressure, PIDPressure}

	for _, pressureType := range pressureTypes {
		err := r.checkPressureStateChange(node, pressureType)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkPressureStateChange checks for pressure state changes and sends alerts accordingly
func (r *NodeAlertController) checkPressureStateChange(node *corev1.Node, pressureType NodePressureType) error {
	currentPressure := false
	conditionMessage := ""

	for _, condition := range node.Status.Conditions {
		var conditionType corev1.NodeConditionType
		switch pressureType {
		case MemoryPressure:
			conditionType = corev1.NodeMemoryPressure
		case DiskPressure:
			conditionType = corev1.NodeDiskPressure
		case PIDPressure:
			conditionType = corev1.NodePIDPressure
		}

		if condition.Type == conditionType {
			currentPressure = condition.Status == corev1.ConditionTrue
			conditionMessage = condition.Message
			break
		}
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("%s-%s", node.Name, pressureType)
	lastPressure, _ := r.lastPressureState[key]
	if lastPressure != currentPressure {
		if currentPressure {
			// from available to pressure
			err := r.sendNodeAlert(node.Name, pressureType, conditionMessage, true)
			if err != nil {
				klog.Errorf("failed to publish available to pressure, type: %s, err: %v", pressureType, err)
				return err
			}
		} else {
			// from pressure to available
			err := r.sendNodeAlert(node.Name, pressureType, conditionMessage, false)
			if err != nil {
				klog.Errorf("failed to publish pressure to available, type: %s, err: %v", pressureType, err)
				return err
			}
		}
	} else if currentPressure {
		// pressure persists
		if r.shouldSendAlert(node.Name, pressureType) {
			err := r.sendNodeAlert(node.Name, pressureType, conditionMessage, true)
			if err != nil {
				klog.Errorf("failed to publish persists pressure, type: %s, err: %v", pressureType, err)
				return err
			}
		}
	}
	r.lastPressureState[key] = currentPressure
	return nil
}

// shouldSendAlert checks if enough time has passed since the last alert for this pressure type
func (r *NodeAlertController) shouldSendAlert(nodeName string, pressureType NodePressureType) bool {
	key := fmt.Sprintf("%s-%s", nodeName, pressureType)
	lastTime, exists := r.lastAlertTime[key]
	if !exists {
		return true
	}

	// Check if 60 minutes has passed since the last alert
	return time.Since(lastTime) >= 60*time.Minute
}

// sendNodeAlertUnlocked sends an alert message to NATS
func (r *NodeAlertController) sendNodeAlert(nodeName string, pressureType NodePressureType, message string, isPressure bool) error {
	key := fmt.Sprintf("%s-%s", nodeName, pressureType)

	status := false
	if isPressure {
		status = true
	}

	data := NodeAlertEvent{
		Topic: pressureType,
		Payload: alertPayload{
			NodeName:     nodeName,
			PressureType: pressureType,
			Timestamp:    time.Now(),
			Message:      message,
			Status:       status,
		},
	}

	if err := r.publishToNats("os.notification", data); err != nil {
		klog.Errorf("failed to publish node alert to NATS: %v", err)
		return err
	} else {
		if isPressure {
			klog.Infof("successfully published node pressure alert for %s: %s", nodeName, pressureType)
		} else {
			klog.Infof("successfully published node pressure recovery for %s: %s", nodeName, pressureType)
		}
	}
	if isPressure {
		r.lastAlertTime[key] = time.Now()
	}
	return nil
}

// publishToNats publishes a message to the specified NATS subject
func (r *NodeAlertController) publishToNats(subject string, data interface{}) error {
	return utils.PublishToNats(subject, data)
}
