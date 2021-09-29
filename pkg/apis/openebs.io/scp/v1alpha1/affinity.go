package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Affinity is a group of affinity scheduling rules
type Affinity struct {

	// Describes volume affinity scheduling rules
	// +optional
	VolumeAffinity []VolumeAffinityTerm `json:"volumeAffinity,omitempty"`

	// Describes volume anti-affinity scheduling rules
	// +optional
	VolumeAntiAffinity []VolumeAffinityTerm `json:"volumeAntiAffinity,omitempty"`

	// Describes cohort affinity scheduling rules
	// +optional
	CohortAffinity []metav1.LabelSelector `json:"cohortAffinity,omitempty"`

	// Describes cohort anti-affinity scheduling rules
	// +optional
	CohortAntiAffinity []metav1.LabelSelector `json:"cohortAntiAffinity,omitempty"`

	// TopologySpreadConstraint specifies how to spread matching volumes among the given topology.
	// +optional
	TopologySpreadConstraint []v1.TopologySpreadConstraint `json:"topologySpreadConstraint,omitempty"`
}

// VolumeAffinityTerm specifies affinity requirements for a StorageVolume
type VolumeAffinityTerm struct {
	// TopologyKey is the key of cohort labels. StorageCohort that have a label with this key
	// and identical values are considered to be in the same topology.
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`

	// A label query over a set of resources, in this case storage volume.
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}
