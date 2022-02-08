// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package predicate_test

import (
	. "github.com/gardener/gardener/pkg/resourcemanager/predicate"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("Predicate", func() {
	Describe("#HasOperationAnnotation", func() {
		var (
			createEvent  event.CreateEvent
			updateEvent  event.UpdateEvent
			deleteEvent  event.DeleteEvent
			genericEvent event.GenericEvent
		)

		BeforeEach(func() {
			configMap := &corev1.ConfigMap{}

			createEvent = event.CreateEvent{
				Object: configMap,
			}
			updateEvent = event.UpdateEvent{
				ObjectOld: configMap,
				ObjectNew: configMap,
			}
			deleteEvent = event.DeleteEvent{
				Object: configMap,
			}
			genericEvent = event.GenericEvent{
				Object: configMap,
			}
		})

		DescribeTable("it should match",
			func(value string) {
				annotations := map[string]string{"gardener.cloud/operation": value}
				createEvent.Object.SetAnnotations(annotations)
				updateEvent.ObjectOld.SetAnnotations(annotations)
				updateEvent.ObjectNew.SetAnnotations(annotations)
				deleteEvent.Object.SetAnnotations(annotations)
				genericEvent.Object.SetAnnotations(annotations)

				predicate := HasOperationAnnotation()

				Expect(predicate.Create(createEvent)).To(BeTrue())
				Expect(predicate.Update(updateEvent)).To(BeTrue())
				Expect(predicate.Delete(deleteEvent)).To(BeTrue())
				Expect(predicate.Generic(genericEvent)).To(BeTrue())
			},

			Entry("reconcile", "reconcile"),
			Entry("migrate", "migrate"),
			Entry("restore", "restore"),
		)

		It("should not match", func() {
			predicate := HasOperationAnnotation()

			Expect(predicate.Create(createEvent)).To(BeFalse())
			Expect(predicate.Update(updateEvent)).To(BeFalse())
			Expect(predicate.Delete(deleteEvent)).To(BeTrue())
			Expect(predicate.Generic(genericEvent)).To(BeFalse())
		})
	})
})