package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func newCreateOnlyPredicate(f func(e event.CreateEvent) bool) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			if f != nil {
				return f(createEvent)
			}
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
	}
}
