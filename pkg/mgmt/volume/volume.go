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

package volume

import (
	"fmt"
	"time"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
	"github.com/openebs/device-localpv/pkg/device"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// isDeletionCandidate checks if a device volume is a deletion candidate.
func (c *VolController) isDeletionCandidate(Vol *apis.DeviceVolume) bool {
	return Vol.ObjectMeta.DeletionTimestamp != nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *VolController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Vol resource with this namespace/name
	Vol, err := c.VolLister.DeviceVolumes(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("devicevolume '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}
	VolCopy := Vol.DeepCopy()
	err = c.syncVol(VolCopy)
	return err
}

// enqueueVol takes a DeviceVolume resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than DeviceVolume.
func (c *VolController) enqueueVol(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)

}

// syncVol is the function which tries to converge to a desired state for the
// DeviceVolume
func (c *VolController) syncVol(vol *apis.DeviceVolume) error {
	var err error
	// Device Volume should be deleted. Check if deletion timestamp is set
	if c.isDeletionCandidate(vol) {
		err = device.DestroyVolume(vol)
		if err == nil {
			err = device.RemoveVolFinalizer(vol)
		}
		return err
	}

	// if status is not Pending then we are just ignoring the event.
	switch vol.Status.State {
	case device.DeviceStatusFailed:
		klog.Warningf("Skipping retrying device volume provisioning as its already in failed state: %+v", vol.Status.Error)
		return nil
	case device.DeviceStatusReady:
		klog.Info("device volume already provisioned")
		return nil
	}

	// if the status Pending means we will try to create the volume
	if vol.Status.State == device.DeviceStatusPending {
		err = device.CreateVolume(vol)
		if err == nil {
			err = device.UpdateVolInfo(vol, device.DeviceStatusReady)
		} else if custError, ok := err.(*apis.VolumeError); ok && custError.Code == apis.InsufficientCapacity {
			vol.Status.Error = custError
			return device.UpdateVolInfo(vol, device.DeviceStatusFailed)
		}
	}
	return err
}

// addVol is the add event handler for DeviceVolume
func (c *VolController) addVol(obj interface{}) {
	Vol, ok := obj.(*apis.DeviceVolume)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get Vol object %#v", obj))
		return
	}

	if device.NodeID != Vol.Spec.OwnerNodeID {
		return
	}
	klog.Infof("Got add event for Vol %s", Vol.Name)
	c.enqueueVol(Vol)
}

// updateVol is the update event handler for DeviceVolume
func (c *VolController) updateVol(oldObj, newObj interface{}) {

	newVol, ok := newObj.(*apis.DeviceVolume)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get Vol object %#v", newVol))
		return
	}

	if device.NodeID != newVol.Spec.OwnerNodeID {
		return
	}

	if c.isDeletionCandidate(newVol) {
		klog.Infof("Got update event for deleted Vol %s", newVol.Name)
		c.enqueueVol(newVol)
	}
}

// deleteVol is the delete event handler for DeviceVolume
func (c *VolController) deleteVol(obj interface{}) {
	Vol, ok := obj.(*apis.DeviceVolume)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		Vol, ok = tombstone.Obj.(*apis.DeviceVolume)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a devicevolume %#v", obj))
			return
		}
	}

	if device.NodeID != Vol.Spec.OwnerNodeID {
		return
	}

	klog.Infof("Got delete event for Vol %s", Vol.Name)
	c.enqueueVol(Vol)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *VolController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Vol controller")

	// Wait for the k8s caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.VolSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	klog.Info("Starting Vol workers")
	// Launch worker to process Vol resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started Vol workers")
	<-stopCh
	klog.Info("Shutting down Vol workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *VolController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *VolController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Vol resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}
