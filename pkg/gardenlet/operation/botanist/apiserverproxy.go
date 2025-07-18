// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package botanist

import (
	"context"

	"github.com/gardener/gardener/imagevector"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"github.com/gardener/gardener/pkg/component/networking/apiserverproxy"
	"github.com/gardener/gardener/pkg/features"
	imagevectorutils "github.com/gardener/gardener/pkg/utils/imagevector"
)

// DefaultAPIServerProxy returns a deployer for the apiserver-proxy.
func (b *Botanist) DefaultAPIServerProxy() (apiserverproxy.Interface, error) {
	image, err := imagevector.Containers().FindImage(imagevector.ContainerImageNameEnvoyProxy, imagevectorutils.RuntimeVersion(b.ShootVersion()), imagevectorutils.TargetVersion(b.ShootVersion()))
	if err != nil {
		return nil, err
	}

	sidecarImage, err := imagevector.Containers().FindImage(imagevector.ContainerImageNameApiserverProxySidecar, imagevectorutils.RuntimeVersion(b.ShootVersion()), imagevectorutils.TargetVersion(b.ShootVersion()))
	if err != nil {
		return nil, err
	}

	var (
		// we don't want to use AUTO for single-stack clusters as it causes an unnecessary failed lookup
		// ref https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#enum-config-cluster-v3-cluster-dnslookupfamily
		dnsLookupFamily = "V4_ONLY"
	)

	if gardencorev1beta1.IsIPv6SingleStack(b.Shoot.GetInfo().Spec.Networking.IPFamilies) {
		dnsLookupFamily = "V6_ONLY"
	}

	values := apiserverproxy.Values{
		Image:               image.String(),
		SidecarImage:        sidecarImage.String(),
		ProxySeedServerHost: b.outOfClusterAPIServerFQDN(),
		DNSLookupFamily:     dnsLookupFamily,
		IstioTLSTermination: features.DefaultFeatureGate.Enabled(features.IstioTLSTermination) && v1beta1helper.IsShootIstioTLSTerminationEnabled(b.Shoot.GetInfo()),
	}

	return apiserverproxy.New(b.SeedClientSet.Client(), b.Shoot.ControlPlaneNamespace, b.SecretsManager, values), nil
}

// DeployAPIServerProxy deploys the apiserver-proxy.
func (b *Botanist) DeployAPIServerProxy(ctx context.Context) error {
	if !b.ShootUsesDNS() {
		return b.Shoot.Components.SystemComponents.APIServerProxy.Destroy(ctx)
	}

	b.Shoot.Components.SystemComponents.APIServerProxy.SetAdvertiseIPAddress(b.APIServerClusterIP)

	if err := b.Shoot.Components.SystemComponents.APIServerProxy.Deploy(ctx); err != nil {
		return err
	}

	// TODO(Wieneo): Remove this after Gardener v1.123
	return b.Shoot.UpdateInfoStatus(ctx, b.GardenClient, true, false, func(shoot *gardencorev1beta1.Shoot) error {
		shoot.Status.Constraints = v1beta1helper.RemoveConditions(shoot.Status.Constraints, gardencorev1beta1.ShootAPIServerProxyUsesHTTPProxy)
		return nil
	})
}
