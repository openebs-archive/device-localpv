/*
Copyright 2021 The OpenEBS Authors

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

package nodebuilder

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
)

// Builder is the builder object for DeviceNode
type Builder struct {
	node *DeviceNode
	errs []error
}

// DeviceNode is a wrapper over
// DeviceNode API instance
type DeviceNode struct {
	// LVMVolume object
	Object *apis.DeviceNode
}

// From returns a new instance of
// device volume
func From(node *apis.DeviceNode) *DeviceNode {
	return &DeviceNode{
		Object: node,
	}
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		node: &DeviceNode{
			Object: &apis.DeviceNode{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(node *apis.DeviceNode) *Builder {
	if node == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build device node object: nil node"),
		)
		return b
	}
	return &Builder{
		node: &DeviceNode{
			Object: node,
		},
	}
}

// WithNamespace sets the namespace of DeviceNode
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build device node object: missing namespace",
			),
		)
		return b
	}
	b.node.Object.Namespace = namespace
	return b
}

// WithName sets the name of DeviceNode
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm node object: missing name",
			),
		)
		return b
	}
	b.node.Object.Name = name
	return b
}

// WithDevices sets the devices of DeviceNode
func (b *Builder) WithDevices(devices []apis.Device) *Builder {
	b.node.Object.Devices = devices
	return b
}

// WithOwnerReferences sets the owner references of DeviceNode
func (b *Builder) WithOwnerReferences(ownerRefs ...metav1.OwnerReference) *Builder {
	b.node.Object.OwnerReferences = ownerRefs
	return b
}

// Build returns DeviceNode API object
func (b *Builder) Build() (*apis.DeviceNode, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.node.Object, nil
}
