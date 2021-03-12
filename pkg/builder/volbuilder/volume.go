/*
Copyright 2019 The OpenEBS Authors

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

package volbuilder

import (
	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
	"github.com/openebs/lib-csi/pkg/common/errors"
)

// Builder is the builder object for DeviceVolume
type Builder struct {
	volume *DeviceVolume
	errs   []error
}

// DeviceVolume is a wrapper over
// DeviceVolume API instance
type DeviceVolume struct {
	// DeviceVolume object
	Object *apis.DeviceVolume
}

// From returns a new instance of
// device volume
func From(vol *apis.DeviceVolume) *DeviceVolume {
	return &DeviceVolume{
		Object: vol,
	}
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		volume: &DeviceVolume{
			Object: &apis.DeviceVolume{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(volume *apis.DeviceVolume) *Builder {
	if volume == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume"),
		)
		return b
	}
	return &Builder{
		volume: &DeviceVolume{
			Object: volume,
		},
	}
}

// WithNamespace sets the namespace of  DeviceVolume
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build device volume object: missing namespace",
			),
		)
		return b
	}
	b.volume.Object.Namespace = namespace
	return b
}

// WithName sets the name of DeviceVolume
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build device volume object: missing name",
			),
		)
		return b
	}
	b.volume.Object.Name = name
	return b
}

// WithCapacity sets the Capacity of device volume by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	if capacity == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build device volume object: missing capacity",
			),
		)
		return b
	}
	b.volume.Object.Spec.Capacity = capacity
	return b
}

// WithOwnerNode sets owner node for the DeviceVolume where the volume should be provisioned
func (b *Builder) WithOwnerNode(host string) *Builder {
	b.volume.Object.Spec.OwnerNodeID = host
	return b
}

// WithVolumeStatus sets DeviceVolume status
func (b *Builder) WithVolumeStatus(status string) *Builder {
	b.volume.Object.Status.State = status
	return b
}

// WithNodeName sets NodeID for creating the volume
func (b *Builder) WithNodeName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build device volume object: missing node name",
			),
		)
		return b
	}
	b.volume.Object.Spec.OwnerNodeID = name
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.volume.Object.Labels == nil {
		b.volume.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.volume.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer sets Finalizer name creating the volume
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.volume.Object.Finalizers = append(b.volume.Object.Finalizers, finalizer...)
	return b
}

// Build returns DeviceVolume API object
func (b *Builder) Build() (*apis.DeviceVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
