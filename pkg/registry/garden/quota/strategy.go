// Copyright 2018 The Gardener Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quota

import (
	"github.com/gardener/gardener/pkg/api"
	"github.com/gardener/gardener/pkg/apis/garden"
	"github.com/gardener/gardener/pkg/apis/garden/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage/names"
)

type quotaStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy defines the storage strategy for Quotas.
var Strategy = quotaStrategy{api.Scheme, names.SimpleNameGenerator}

func (quotaStrategy) NamespaceScoped() bool {
	return true
}

func (quotaStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	_ = obj.(*garden.Quota)
}

func (quotaStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	quota := obj.(*garden.Quota)
	return validation.ValidateQuota(quota)
}

func (quotaStrategy) Canonicalize(obj runtime.Object) {
}

func (quotaStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (quotaStrategy) PrepareForUpdate(ctx genericapirequest.Context, newObj, oldObj runtime.Object) {
	_ = oldObj.(*garden.Quota)
	_ = newObj.(*garden.Quota)
}

func (quotaStrategy) ValidateUpdate(ctx genericapirequest.Context, newObj, oldObj runtime.Object) field.ErrorList {
	oldQuota, newQuota := oldObj.(*garden.Quota), newObj.(*garden.Quota)
	return validation.ValidateQuotaUpdate(newQuota, oldQuota)
}

func (quotaStrategy) AllowUnconditionalUpdate() bool {
	return true
}

type quotaStatusStrategy struct {
	quotaStrategy
}

// StatusStrategy defines the storage strategy for the status subresource of Quotas.
var StatusStrategy = quotaStatusStrategy{Strategy}

func (quotaStatusStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	newQuota := obj.(*garden.Quota)
	oldQuota := old.(*garden.Quota)
	newQuota.Spec = oldQuota.Spec
}

func (quotaStatusStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateQuotaStatusUpdate(obj.(*garden.Quota), old.(*garden.Quota))
}
