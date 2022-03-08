// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package certificatesigningrequest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	ctrlruntimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap"
	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap/keys"
	"github.com/gardener/gardener/pkg/controllerutils"
)

const (
	// ControllerName is the name of this controller.
	ControllerName = "csr-autoapprove"
)

// Controller controls CertificateSigningRequests.
type Controller struct {
	reconciler     reconcile.Reconciler
	hasSyncedFuncs []cache.InformerSynced

	log                    logr.Logger
	csrQueue               workqueue.RateLimitingInterface
	workerCh               chan int
	numberOfRunningWorkers int
}

// NewCSRController takes a Kubernetes client for the Garden clusters <k8sGardenClient>, a struct
// holding information about the acting Gardener, a <kubeInformerFactory>, and a <recorder> for
// event recording. It creates a new CSR controller.
func NewCSRController(ctx context.Context, log logr.Logger, clientMap clientmap.ClientMap) (*Controller, error) {
	log = log.WithName(ControllerName)

	gardenClient, err := clientMap.GetClient(ctx, keys.ForGarden())
	if err != nil {
		return nil, err
	}

	var (
		csrInformer            ctrlruntimecache.Informer
		certificatesAPIVersion = "v1"
	)

	csrInformer, err = gardenClient.Cache().GetInformer(ctx, &certificatesv1.CertificateSigningRequest{})
	if err != nil {
		if !meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("failed to get CSR Informer: %w", err)
		}

		// fallback to v1beta1
		csrInformer, err = gardenClient.Cache().GetInformer(ctx, &certificatesv1beta1.CertificateSigningRequest{})
		if err != nil {
			return nil, fmt.Errorf("failed to get CSR Informer: %w", err)
		}
		certificatesAPIVersion = "v1beta1"
	}

	csrController := &Controller{
		reconciler: NewCSRReconciler(gardenClient, certificatesAPIVersion),

		log:      log,
		csrQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CertificateSigningRequest"),
		workerCh: make(chan int),
	}

	csrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    csrController.csrAdd,
		UpdateFunc: csrController.csrUpdate,
	})

	csrController.hasSyncedFuncs = append(csrController.hasSyncedFuncs, csrInformer.HasSynced)

	return csrController, nil
}

// Run runs the Controller until the given context <ctx> is alive.
func (c *Controller) Run(ctx context.Context, workers int) {
	var waitGroup sync.WaitGroup

	if !cache.WaitForCacheSync(ctx.Done(), c.hasSyncedFuncs...) {
		c.log.Error(wait.ErrWaitTimeout, "Timed out waiting for caches to sync")
		return
	}

	// Count number of running workers.
	go func() {
		for res := range c.workerCh {
			c.numberOfRunningWorkers += res
		}
	}()

	c.log.Info("CertificateSigningRequest controller initialized")

	for i := 0; i < workers; i++ {
		controllerutils.CreateWorker(ctx, c.csrQueue, "CertificateSigningRequest", c.reconciler, &waitGroup, c.workerCh, controllerutils.WithLogger(c.log))
	}

	// Shutdown handling
	<-ctx.Done()
	c.csrQueue.ShutDown()

	for {
		if c.csrQueue.Len() == 0 && c.numberOfRunningWorkers == 0 {
			c.log.V(1).Info("No running CertificateSigningRequest worker and no items left in the queues. Terminating CertificateSigningRequest controller")
			break
		}
		c.log.V(1).Info("Waiting for CertificateSigningRequest workers to finish", "numberOfRunningWorkers", c.numberOfRunningWorkers, "queueLength", c.csrQueue.Len())
		time.Sleep(5 * time.Second)
	}

	waitGroup.Wait()
}