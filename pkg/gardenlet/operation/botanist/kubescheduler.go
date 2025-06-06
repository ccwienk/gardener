// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package botanist

import (
	"github.com/gardener/gardener/imagevector"
	"github.com/gardener/gardener/pkg/component"
	kubescheduler "github.com/gardener/gardener/pkg/component/kubernetes/scheduler"
	imagevectorutils "github.com/gardener/gardener/pkg/utils/imagevector"
)

// DefaultKubeScheduler returns a deployer for the kube-scheduler.
func (b *Botanist) DefaultKubeScheduler() (component.DeployWaiter, error) {
	image, err := imagevector.Containers().FindImage(imagevector.ContainerImageNameKubeScheduler, imagevectorutils.RuntimeVersion(b.SeedVersion()), imagevectorutils.TargetVersion(b.ShootVersion()))
	if err != nil {
		return nil, err
	}

	replicas := b.Shoot.GetReplicas(1)
	if b.Shoot.RunsControlPlane() {
		replicas = 0
	}

	return kubescheduler.New(
		b.SeedClientSet.Client(),
		b.Shoot.ControlPlaneNamespace,
		b.SecretsManager,
		image.String(),
		replicas,
		b.Shoot.GetInfo().Spec.Kubernetes.KubeScheduler,
	), nil
}
