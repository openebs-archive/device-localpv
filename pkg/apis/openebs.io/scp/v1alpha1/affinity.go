package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Affinity is a group of affinity scheduling rules
type Affinity struct {

	// VolumeAffinity helps to place a volume in the same topology(storage cohort)
	// where old volumes are present, if old volume not found then it can go to any cohort.
	// Old volumes are identified using LabelSelector. Using this we can ensure that if any
	// pod is requesting for 2 volumes then 2 volumes will be scheduled in the same storage
	// cohort topology. ie - if app-0 replica of sts app is requesting for 2 volumes data-0
	// and log-0, both the volumes will be scheduled in the same storage cohort topology.
	// All volumeAffinity requirements are ANDed together.
	// +optional
	VolumeAffinity []VolumeAffinityTerm `json:"volumeAffinity,omitempty"`

	// VolumeAntiAffinity helps to place a volume in the different topology(storage cohort)
	// where old volumes are present, if old volume not found the it can go to any cohort.
	// Old volumes are identified using LabelSelector. Using this we can ensure that replicas
	// of an application will have volume in different storage cohort topology . ie - if app-0
	// and app-1 replica of sts app is requesting for 2 volumes data-0 and data-1,
	// both the volumes will be scheduled in the different storage cohort topology.
	// All volumeAntiAffinity requirements are ANDed together.
	// +optional
	VolumeAntiAffinity []VolumeAffinityTerm `json:"volumeAntiAffinity,omitempty"`

	// Using CohortAffinity we can schedule a volume in a group of storage cohort.
	// All cohortAffinity requirements are ANDed together.
	// +optional
	CohortAffinity []metav1.LabelSelector `json:"cohortAffinity,omitempty"`

	// Using CohortAffinity we can avoid scheduling a volume in a group of storage cohort.
	// All cohortAntiAffinity requirements are ANDed together.
	// +optional
	CohortAntiAffinity []metav1.LabelSelector `json:"cohortAntiAffinity,omitempty"`

	// TopologySpreadConstraints describes how a group of volumes ought to spread
	// across storage cohort topology. Scheduler will schedule volumes in a way which
	// abides by the constraints. All TopologySpreadConstraints are ANDed.
	// +optional
	TopologySpreadConstraint []corev1.TopologySpreadConstraint `json:"topologySpreadConstraint,omitempty"`
}

// VolumeAffinityTerm defines a rule that new volume should be
// co-located (affinity) or not co-located (anti-affinity) with old volumes,
// where co-located is defined as running on a cohort whose value of
// the label with key <topologyKey> matches that of any cohort on which
// old volume is placed and it is identified through labelSelector i.e matchLabels and matchExpressions.
// The result of matchLabels and matchExpressions are ANDed.
type VolumeAffinityTerm struct {
	// TopologyKey is the key of cohort labels. StorageCohort that have a label with this key
	// and identical values are considered to be in the same topology.
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`

	// A label query over a set of resources, in this case storage volume.
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}
