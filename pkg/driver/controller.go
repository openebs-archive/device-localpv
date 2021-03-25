/*
Copyright © 2019 The OpenEBS Authors

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

package driver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	"github.com/openebs/lib-csi/pkg/common/errors"
	"github.com/openebs/lib-csi/pkg/common/helpers"
	schd "github.com/openebs/lib-csi/pkg/scheduler"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
	"github.com/openebs/device-localpv/pkg/builder/volbuilder"
	"github.com/openebs/device-localpv/pkg/device"
	clientset "github.com/openebs/device-localpv/pkg/generated/clientset/internalclientset"
	informers "github.com/openebs/device-localpv/pkg/generated/informer/externalversions"
	csipayload "github.com/openebs/device-localpv/pkg/response"
)

// size constants
const (
	MB = 1000 * 1000
	GB = 1000 * 1000 * 1000
	Mi = 1024 * 1024
	Gi = 1024 * 1024 * 1024
)

// controller is the server implementation
// for CSI Controller
type controller struct {
	driver       *CSIDriver
	capabilities []*csi.ControllerServiceCapability

	indexedLabel string

	k8sNodeInformer    cache.SharedIndexInformer
	deviceNodeInformer cache.SharedIndexInformer
}

// NewController returns a new instance
// of CSI controller
func NewController(d *CSIDriver) csi.ControllerServer {
	ctrl := &controller{
		driver:       d,
		capabilities: newControllerCapabilities(),
	}

	if err := ctrl.init(); err != nil {
		klog.Fatalf("init controller: %v", err)
	}

	return ctrl
}

// SupportedVolumeCapabilityAccessModes contains the list of supported access
// modes for the volume
var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	&csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

// getRoundedCapacity rounds the capacity on 1024 base
func getRoundedCapacity(size int64) int64 {

	/*
	 * volblocksize and recordsize must be power of 2 from 512B to 1M
	 * so keeping the size in the form of Gi or Mi should be
	 * sufficient to make volsize multiple of volblocksize/recordsize.
	 */
	if size > Gi {
		return ((size + Gi - 1) / Gi) * Gi
	}

	// Keeping minimum allocatable size as 1Mi (1024 * 1024)
	return ((size + Mi - 1) / Mi) * Mi
}

// waitForDeviceVolume waits for completion of any processing of device volume.
// It returns the final status of device volume along with a boolean denoting
// whether it should be rescheduled on some other device name or node.
// In case volume ends up in failed state and rescheduling is required,
// func is also deleting the device volume resource, so that it can be
// re provisioned on some other node.
func waitForDeviceVolume(ctx context.Context,
	vol *apis.DeviceVolume) (*apis.DeviceVolume, bool, error) {
	var reschedule bool // tracks if rescheduling is required or not.
	var err error
	if vol.Status.State == device.DeviceStatusPending {
		if vol, err = device.WaitForDeviceVolumeProcessed(ctx, vol.GetName()); err != nil {
			return nil, false, err
		}
	}
	// if device volume is ready, return the provisioned node.
	if vol.Status.State == device.DeviceStatusReady {
		return vol, false, nil
	}

	// Now it must be in failed state if not above. See if we need
	// to reschedule the device volume.
	var errMsg string
	if volErr := vol.Status.Error; volErr != nil {
		errMsg = volErr.Message
		reschedule = true
	} else {
		errMsg = fmt.Sprintf("failed devicevol must have error set")
	}

	if reschedule {
		// if rescheduling is required, we can deleted the existing device volume object,
		// so that it can be recreated.
		if err = device.DeleteVolume(vol.GetName()); err != nil {
			return nil, false, status.Errorf(codes.Aborted,
				"failed to delete volume %v: %v", vol.GetName(), err)
		}
		if err = device.WaitForDeviceVolumeDestroy(ctx, vol.GetName()); err != nil {
			return nil, false, status.Errorf(codes.Aborted,
				"failed to delete volume %v: %v", vol.GetName(), err)
		}
		return vol, true, status.Error(codes.ResourceExhausted, errMsg)
	}

	return vol, false, status.Error(codes.Aborted, errMsg)
}

func (cs *controller) init() error {
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return errors.Wrapf(err, "failed to build kubeconfig")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to build k8s clientset")
	}

	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to build openebs clientset")
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, 0)
	openebsInformerfactory := informers.NewSharedInformerFactoryWithOptions(openebsClient,
		0, informers.WithNamespace(device.DeviceNamespace))

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cs.k8sNodeInformer = kubeInformerFactory.Core().V1().Nodes().Informer()
	cs.deviceNodeInformer = openebsInformerfactory.Local().V1alpha1().DeviceNodes().Informer()

	if err = cs.deviceNodeInformer.AddIndexers(map[string]cache.IndexFunc{
		LabelIndexName(cs.indexedLabel): LabelIndexFunc(cs.indexedLabel),
	}); err != nil {
		return errors.Wrapf(err, "failed to add index on label %v", cs.indexedLabel)
	}

	go cs.k8sNodeInformer.Run(stopCh)
	go cs.deviceNodeInformer.Run(stopCh)

	// wait for all the caches to be populated.
	klog.Info("waiting for k8s & device node informer caches to be synced")
	cache.WaitForCacheSync(stopCh,
		cs.k8sNodeInformer.HasSynced,
		cs.deviceNodeInformer.HasSynced)
	klog.Info("synced k8s & device node informer caches")
	return nil
}

// CreateDeviceVolume create new device volume for csi volume request
func CreateDeviceVolume(ctx context.Context, req *csi.CreateVolumeRequest,
	params *VolumeParams) (*apis.DeviceVolume, error) {
	volName := strings.ToLower(req.GetName())
	capacity := strconv.FormatInt(getRoundedCapacity(
		req.GetCapacityRange().RequiredBytes), 10)

	vol, err := device.GetDeviceVolume(volName)
	if err != nil {
		if !k8serror.IsNotFound(err) {
			return nil, status.Errorf(codes.Aborted,
				"failed get device volume %v: %v", volName, err.Error())
		}
		vol, err = nil, nil
	}

	if vol != nil {
		if vol.DeletionTimestamp != nil {
			if err = device.WaitForDeviceVolumeDestroy(ctx, volName); err != nil {
				return nil, err
			}
		} else {
			if vol.Spec.Capacity != capacity {
				return nil, status.Errorf(codes.AlreadyExists,
					"volume %s already present", volName)
			}
			var reschedule bool
			vol, reschedule, err = waitForDeviceVolume(ctx, vol)
			// If the device volume becomes ready or we can't reschedule failed volume,
			// return the err.
			if err == nil || !reschedule {
				return vol, err
			}
		}
	}

	nmap, err := getNodeMap(params.Scheduler, params.DeviceName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get node map failed : %s", err.Error())
	}

	// run the scheduler
	selected := schd.Scheduler(req, nmap)

	if len(selected) == 0 {
		return nil, status.Error(codes.Internal, "scheduler failed, not able to select a node to create the PV")
	}

	owner := selected[0]
	klog.Infof("scheduling the volume %s/%s on node %s", params.DeviceName, volName, owner)

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithCapacity(capacity).
		WithDeviceName(params.DeviceName).
		WithOwnerNode(owner).
		WithVolumeStatus(device.DeviceStatusPending).Build()

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	vol, err = device.ProvisionVolume(volObj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "not able to provision the volume %s", err.Error())
	}
	vol, _, err = waitForDeviceVolume(ctx, vol)
	return vol, err
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	var err error

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	params, err := NewVolumeParams(req.GetParameters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument,
			"failed to parse csi volume params: %v", err)
	}

	volName := strings.ToLower(req.GetName())
	size := getRoundedCapacity(req.GetCapacityRange().GetRequiredBytes())
	contentSource := req.GetVolumeContentSource()

	var vol *apis.DeviceVolume
	if contentSource != nil && contentSource.GetVolume() != nil {
		return nil, status.Error(codes.Unimplemented, "")
	}

	vol, err = CreateDeviceVolume(ctx, req, params)

	if err != nil {
		return nil, err
	}

	topology := map[string]string{device.DeviceTopologyKey: vol.Spec.OwnerNodeID}
	cntx := map[string]string{device.DeviceNameKey: params.DeviceName}

	return csipayload.NewCreateVolumeResponseBuilder().
		WithName(volName).
		WithCapacity(size).
		WithTopology(topology).
		WithContext(cntx).
		WithContentSource(contentSource).
		Build(), nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	klog.Infof("received request to delete volume {%s}", req.VolumeId)

	var (
		err error
	)

	if err = cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := strings.ToLower(req.GetVolumeId())

	// verify if the volume has already been deleted
	vol, err := device.GetDeviceVolume(volumeID)
	if vol != nil && vol.DeletionTimestamp != nil {
		goto deleteResponse
	}

	if err != nil {
		if k8serror.IsNotFound(err) {
			goto deleteResponse
		}
		return nil, errors.Wrapf(
			err,
			"failed to get volume for {%s}",
			volumeID,
		)
	}

	// Delete the corresponding Device Volume CR
	err = device.DeleteVolume(volumeID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			volumeID,
		)
	}

deleteResponse:
	return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
}

func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range SupportedVolumeCapabilityAccessModes {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}

	foundAll := true
	for _, c := range volCaps {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}

// TODO Implementation will be taken up later

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeID := strings.ToLower(req.GetVolumeId())
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volCaps := req.GetVolumeCapabilities()
	if len(volCaps) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not provided")
	}

	if _, err := device.GetDeviceVolume(volumeID); err != nil {
		return nil, status.Errorf(codes.NotFound, "Get volume failed err %s", err.Error())
	}

	var confirmed *csi.ValidateVolumeCapabilitiesResponse_Confirmed
	if isValidVolumeCapabilities(volCaps) {
		confirmed = &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: volCaps}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: confirmed,
	}, nil
}

// ControllerGetCapabilities fetches controller capabilities
//
// This implements csi.ControllerServer
func (cs *controller) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.capabilities,
	}

	return resp, nil
}

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot deletes given snapshot
//
// This implements csi.ControllerServer
func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots lists all snapshots for the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume removes a previously
// attached volume from the given node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume attaches given volume
// at the specified node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {

	var segments map[string]string
	if topology := req.GetAccessibleTopology(); topology != nil {
		segments = topology.Segments
	}
	nodeNames, err := cs.filterNodesByTopology(segments)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	deviceNodeCache := cs.deviceNodeInformer.GetIndexer()
	params := req.GetParameters()
	deviceParam := helpers.GetInsensitiveParameter(&params, "deviceName")

	var availableCapacity int64
	for _, nodeName := range nodeNames {
		v, exists, err := deviceNodeCache.GetByKey(device.DeviceNamespace + "/" + nodeName)
		if err != nil {
			klog.Warning("unexpected error after querying the deviceNode informer cache")
			continue
		}
		if !exists {
			continue
		}
		deviceNode := v.(*apis.DeviceNode)
		// rather than summing all free capacity, we are calculating maximum
		// partition size that gets fit in given device.
		// See https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1472-storage-capacity-tracking#available-capacity-vs-maximum-volume-size &
		// https://github.com/container-storage-interface/spec/issues/432 for more details
		for _, device := range deviceNode.Devices {
			if device.Name != deviceParam {
				continue
			}
			freeCapacity := device.Free.Value()
			if availableCapacity < freeCapacity {
				availableCapacity = freeCapacity
			}
		}
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: availableCapacity,
	}, nil
}

func (cs *controller) filterNodesByTopology(segments map[string]string) ([]string, error) {
	nodesCache := cs.k8sNodeInformer.GetIndexer()
	if len(segments) == 0 {
		return nodesCache.ListKeys(), nil
	}

	filterNodes := func(vs []interface{}) ([]string, error) {
		var names []string
		selector := labels.SelectorFromSet(segments)
		for _, v := range vs {
			meta, err := apimeta.Accessor(v)
			if err != nil {
				return nil, err
			}
			if selector.Matches(labels.Set(meta.GetLabels())) {
				names = append(names, meta.GetName())
			}
		}
		return names, nil
	}

	// first see if we need to filter the informer cache by indexed label,
	// so that we don't need to iterate over all the nodes for performance
	// reasons in large cluster.
	indexName := LabelIndexName(cs.indexedLabel)
	if _, ok := nodesCache.GetIndexers()[indexName]; !ok {
		// run through all the nodes in case indexer doesn't exists.
		return filterNodes(nodesCache.List())
	}

	if segValue, ok := segments[cs.indexedLabel]; ok {
		vs, err := nodesCache.ByIndex(indexName, segValue)
		if err != nil {
			return nil, errors.Wrapf(err, "query indexed store indexName=%v indexKey=%v",
				indexName, segValue)
		}
		return filterNodes(vs)
	}
	return filterNodes(nodesCache.List())
}

// ListVolumes lists all the volumes
//
// This implements csi.ControllerServer
func (cs *controller) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// validateCapabilities validates if provided capabilities
// are supported by this driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {

	for _, cap := range caps {
		if !IsSupportedVolumeCapabilityAccessMode(cap.AccessMode.Mode) {
			return false
		}
	}
	return true
}

func (cs *controller) validateDeleteVolumeReq(req *csi.DeleteVolumeRequest) error {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle delete volume request: missing volume id",
		)
	}

	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s} : validation failed",
			volumeID,
		)
	}
	return nil
}

// IsSupportedVolumeCapabilityAccessMode valides the requested access mode
func IsSupportedVolumeCapabilityAccessMode(
	accessMode csi.VolumeCapability_AccessMode_Mode,
) bool {

	for _, access := range SupportedVolumeCapabilityAccessModes {
		if accessMode == access.Mode {
			return true
		}
	}
	return false
}

// newControllerCapabilities returns a list
// of this controller's capabilities
func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(
		cap csi.ControllerServiceCapability_RPC_Type,
	) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

// validateRequest validates if the requested service is
// supported by the driver
func (cs *controller) validateRequest(
	c csi.ControllerServiceCapability_RPC_Type,
) error {

	for _, cap := range cs.capabilities {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}

	return status.Error(
		codes.InvalidArgument,
		fmt.Sprintf("failed to validate request: {%s} is not supported", c),
	)
}

func (cs *controller) validateVolumeCreateReq(req *csi.CreateVolumeRequest) error {
	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle create volume request for {%s}",
			req.GetName(),
		)
	}

	if req.GetName() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume name",
		)
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume capabilities",
		)
	}
	return nil
}

// LabelIndexName add prefix for label index.
func LabelIndexName(label string) string {
	return "l:" + label
}

// LabelIndexFunc defines index values for given label.
func LabelIndexFunc(label string) cache.IndexFunc {
	return func(obj interface{}) ([]string, error) {
		meta, err := apimeta.Accessor(obj)
		if err != nil {
			return nil, fmt.Errorf(
				"k8s api object type (%T) doesn't implements metav1.Object interface: %v", obj, err)
		}
		var vs []string
		if v, ok := meta.GetLabels()[label]; ok {
			vs = append(vs, v)
		}
		return vs, nil
	}
}
