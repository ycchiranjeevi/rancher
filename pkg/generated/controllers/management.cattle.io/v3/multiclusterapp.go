/*
Copyright 2023 Rancher Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package v3

import (
	"context"
	"time"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// MultiClusterAppController interface for managing MultiClusterApp resources.
type MultiClusterAppController interface {
	generic.ControllerInterface[*v3.MultiClusterApp, *v3.MultiClusterAppList]
}

// MultiClusterAppClient interface for managing MultiClusterApp resources in Kubernetes.
type MultiClusterAppClient interface {
	generic.ClientInterface[*v3.MultiClusterApp, *v3.MultiClusterAppList]
}

// MultiClusterAppCache interface for retrieving MultiClusterApp resources in memory.
type MultiClusterAppCache interface {
	generic.CacheInterface[*v3.MultiClusterApp]
}

type MultiClusterAppStatusHandler func(obj *v3.MultiClusterApp, status v3.MultiClusterAppStatus) (v3.MultiClusterAppStatus, error)

type MultiClusterAppGeneratingHandler func(obj *v3.MultiClusterApp, status v3.MultiClusterAppStatus) ([]runtime.Object, v3.MultiClusterAppStatus, error)

func RegisterMultiClusterAppStatusHandler(ctx context.Context, controller MultiClusterAppController, condition condition.Cond, name string, handler MultiClusterAppStatusHandler) {
	statusHandler := &multiClusterAppStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

func RegisterMultiClusterAppGeneratingHandler(ctx context.Context, controller MultiClusterAppController, apply apply.Apply,
	condition condition.Cond, name string, handler MultiClusterAppGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &multiClusterAppGeneratingHandler{
		MultiClusterAppGeneratingHandler: handler,
		apply:                            apply,
		name:                             name,
		gvk:                              controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterMultiClusterAppStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type multiClusterAppStatusHandler struct {
	client    MultiClusterAppClient
	condition condition.Cond
	handler   MultiClusterAppStatusHandler
}

func (a *multiClusterAppStatusHandler) sync(key string, obj *v3.MultiClusterApp) (*v3.MultiClusterApp, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type multiClusterAppGeneratingHandler struct {
	MultiClusterAppGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *multiClusterAppGeneratingHandler) Remove(key string, obj *v3.MultiClusterApp) (*v3.MultiClusterApp, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v3.MultiClusterApp{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *multiClusterAppGeneratingHandler) Handle(obj *v3.MultiClusterApp, status v3.MultiClusterAppStatus) (v3.MultiClusterAppStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.MultiClusterAppGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
