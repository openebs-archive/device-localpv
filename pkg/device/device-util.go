package device

/*
Copyright 2017 The Kubernetes Authors.

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

import (
	"fmt"
	"github.com/openebs/lib-csi/pkg/common/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"math"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
)

// Partition Commands
const (
	PartitionDiskID    = "fdisk -l /dev/%s"
	PartitionDiskList  = "lsblk -b"
	PartitionPrintFree = "parted /dev/%s unit b print free --script"
	PartitionPrint     = "parted /dev/%s unit b print --script"
	PartitionCreate    = "parted /dev/%s mkpart %s %dMiB %dMiB --script"
	PartitionDelete    = "parted /dev/%s rm %d --script"
	PartitionWipeFS    = "wipefs --force -a %s"
)

type partUsed struct {
	DiskName string
	PartNum  uint32
}
type partFree struct {
	DiskName string
	StartMiB uint64
	EndMiB   uint64
	SizeMiB  uint64
}

type diskDetail struct {
	DiskName string
	Size     uint64
}

// CreateVolume Todo
func CreateVolume(vol *apis.DeviceVolume) error {
	//func CreatePartition(diskName string, partitionName string, size int) error {
	diskMetaName := vol.Spec.DevName
	partitionName := vol.Name[4:]

	capacityBytes, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
	if err != nil {
		klog.Warning("error parsing vol.Spec.Capacity. Skipping CreateVolume", err)
		return err
	}
	capacityMiB := uint64(math.Floor(float64(capacityBytes) / (1024 * 1024)))

	pList, err := getAllPartsUsed(diskMetaName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return err
	}
	if len(pList) > 0 {
		klog.Infof("Partition %s already exist, Skipping creation", partitionName)
		// Making Volume creation Idempotent
		return nil
	}
	disk, start, err := findBestPart(diskMetaName, capacityMiB)
	if err != nil {
		klog.Errorf("findBestPart Failed")
		return err
	}
	return wipefsAndCreatePart(disk, start, partitionName, capacityMiB, diskMetaName)
}

// DeletePart Todo
func wipefsAndCreatePart(disk string, start uint64, partitionName string, size uint64, diskMetaName string) error {
	klog.Infof("Creating Partition %s %s", partitionName, diskMetaName)
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionCreate, disk, partitionName, start, start+size), " "))
	if err != nil {
		klog.Errorf("Create Partition failed %s", err)
		return err
	}

	pList, err := getAllPartsUsed(diskMetaName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return err
	}

	err = wipeFsPartition(pList[0].DiskName, pList[0].PartNum)
	if err != nil {
		klog.Infof("Deleting partition %d on disk %s because wipefs failed", pList[0].PartNum, pList[0].DiskName)
		err1 := deletePartition(pList[0].DiskName, pList[0].PartNum)
		if err1 != nil {
			klog.Errorf("could not delete partition %d on disk %s, created during CreateVolume(). Error: %s", pList[0].PartNum, pList[0].DiskName, err1)
		}
		// the error will be returned irrespective of the return value of delete partition,
		// as create partition has failed.
		return err
	}
	return nil
}

// getAllPartsFree Todo
func getAllPartsFree(diskName string) ([]partFree, error) {
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []partFree
	for _, disk := range diskList {
		tmpList, err := getPartsFree(disk.DiskName, diskName)
		if err != nil {
			klog.Infof("GetPart Error, %s", disk.DiskName)
			continue
		}
		pList = append(pList, tmpList...)
	}
	return pList, nil
}

func findBestPart(diskName string, partSize uint64) (string, uint64, error) {
	pList, err := getAllPartsFree(diskName)
	if err != nil {
		klog.Errorln("Device LocalPV: GetAllPartsFree error")
		return "", 0, err
	}

	if len(pList) > 0 {
		sort.Slice(pList, func(i, j int) bool {
			// "<" Ascending order
			return pList[i].SizeMiB < pList[j].SizeMiB
		})
		for _, tmp := range pList {
			if tmp.SizeMiB > partSize {
				return tmp.DiskName, tmp.StartMiB, nil
			}
		}
	}
	klog.Errorln("Device LocalPV: Free space for partition is not found")
	return "", 0, err

}

// GetAllPartsUsed Todo
func getAllPartsUsed(diskMetaName string, partitionName string) ([]partUsed, error) {
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []partUsed
	for _, disk := range diskList {
		tmpList, err := GetPartitionList(disk.DiskName, diskMetaName, false)
		if err != nil {
			klog.Infof("GetPart Error, %+v", disk)
			continue
		}
		for _, tmp := range tmpList {
			if tmp[len(tmp)-1] == partitionName {
				partNum, _ := strconv.ParseInt(tmp[0], 10, 32)
				pList = append(pList, partUsed{disk.DiskName, uint32(partNum)})
			}
		}
	}
	return pList, nil

}

// DestroyVolume Todo
func DestroyVolume(vol *apis.DeviceVolume) error {
	diskMetaName := vol.Spec.DevName
	partitionName := vol.Name[4:]
	pList, err := getAllPartsUsed(diskMetaName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return err
	}
	if len(pList) > 1 {
		klog.Errorf("More than one partition of same name %s\n", partitionName)
		return errors.New("More than one partition of same name")
	}
	if len(pList) == 0 {
		klog.Infof("%s Partition not found, Skipping Deletion\n", partitionName)
		return nil
	}
	return wipefsAndDeletePart(pList[0].DiskName, pList[0].PartNum)

}

func wipefsAndDeletePart(disk string, partNum uint32) error {
	err := wipeFsPartition(disk, partNum)
	if err != nil {
		return err
	}
	return deletePartition(disk, partNum)
}

// deletes the given partition from the disk
func deletePartition(disk string, partNum uint32) error {
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDelete, disk, partNum), " "))
	if err != nil {
		klog.Errorf("Delete Partition failed for disk: %s, partition: %d . Error: %s", disk, partNum, err)
	}
	return err
}

// performs a force wipefs on the given partition
func wipeFsPartition(disk string, partNum uint32) error {
	klog.Infof("Running WipeFS for disk: %s, partition %d", disk, partNum)
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionWipeFS, getPartitionPath(disk, partNum)), " "))
	if err != nil {
		klog.Errorf("WipeFS failed for disk: %s, partition: %d . Error: %s", disk, partNum, err)
	}
	return err
}

// GetVolumeDevPath Todo
func GetVolumeDevPath(vol *apis.DeviceVolume) (string, error) {
	diskMetaName := vol.Spec.DevName
	partitionName := vol.Name[4:]
	pList, err := getAllPartsUsed(diskMetaName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return "", err
	}
	if len(pList) > 1 {
		klog.Errorf("More than one partition of same name %s\n", partitionName)
		return "", errors.New("More than one partition of same name")
	}
	if len(pList) == 0 {
		klog.Errorf("%s Partition not found\n", partitionName)
		return "", errors.New("Partition not found")
	}

	return getPartitionPath(pList[0].DiskName, pList[0].PartNum), nil
}

// RunCommand Todo
func RunCommand(cList []string) (string, error) {
	cmd := exec.Command(cList[0], cList[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("Device LocalPV: could not Run command %+v\n", cList)
		return "", err
	}
	return string(out), nil
}

// GetPartitionList Todo
func GetPartitionList(diskName string, diskMetaName string, free bool) ([][]string, error) {
	var command string
	if free {
		command = PartitionPrintFree
	} else {
		command = PartitionPrint
	}
	out, err := RunCommand(strings.Split(fmt.Sprintf(command, diskName), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not get parts error: %s %v\n", string(out), err)
		return nil, err
	}
	sli := strings.Split(string(out), "\n")
	start := false
	var result [][]string
	for _, value := range sli {
		tmp := strings.Fields(value)
		if len(tmp) == 0 {
			continue
		}
		if tmp[0] == "Partition" && tmp[1] == "Table" && tmp[2] != "gpt" {
			klog.Infof("Disk: %s Not an GPT Partitioned Disk", diskName)
			return nil, errors.New("Wrong Partition type")
		}
		if !start {
			if tmp[0] == "Number" {
				start = true
			}
			continue
		}
		devRegex, err := regexp.Compile(diskMetaName)
		if err != nil {
			klog.Infof("Disk: Regex compile failure %s, %+v", diskMetaName, err)
			return nil, err
		}
		if diskMetaName != "" && tmp[0] == "1" && !devRegex.MatchString(tmp[len(tmp)-1]) {
			klog.Infof("Disk: DiskName not correct")
			return nil, errors.New("Wrong DiskMetaName")
		}
		result = append(result, tmp)
	}
	return result, nil
}

func getPartsFree(diskName string, diskMetaName string) ([]partFree, error) {
	var pList []partFree
	tmpList, err := GetPartitionList(diskName, diskMetaName, true)
	if err != nil {
		klog.Infof("GetPart Error, %s %s", diskName, diskMetaName)
		return nil, errors.New("GetPartitionList Error")
	}
	for _, tmp := range tmpList {
		if tmp[3] == "Free" {
			beginBytes, _ := strconv.ParseUint(string(tmp[0][:len(tmp[0])-1]), 10, 64)
			endBytes, _ := strconv.ParseUint(string(tmp[1][:len(tmp[1])-1]), 10, 64)
			beginMib := math.Ceil(float64(beginBytes) / 1024 / 1024)
			endMib := math.Floor(float64(endBytes) / 1024 / 1024)
			size := uint64(0)
			if endMib > beginMib {
				size = uint64(endMib - beginMib)
			}
			pList = append(pList, partFree{diskName, uint64(beginMib), uint64(endMib), size})
		}
	}
	return pList, nil
}

// GetFreeCapacity Todo
func GetFreeCapacity(diskName string) (uint64, error) {
	pList, err := getPartsFree(diskName, "")
	if err != nil {
		klog.Errorln("Device LocalPV: GetAllPartsFree error")
		return 0, err
	}
	if len(pList) > 0 {
		sort.Slice(pList, func(i, j int) bool {
			// ">" Descending order
			return pList[i].SizeMiB > pList[j].SizeMiB
		})
		return pList[0].SizeMiB, nil
	}
	klog.Errorln("Device LocalPV: Free space for partition is not found")
	return 0, err
}

// GetDiskList Todo
func getDiskList() ([]diskDetail, error) {
	var result []diskDetail
	out, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDiskList), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not list disk error: %s %s", string(out), err)
		return nil, err
	}
	sli := strings.Split(string(out), "\n")
	for _, value := range sli {
		tmp := strings.Fields(value)
		if len(tmp) == 0 {
			continue
		}
		// loop is added here for testing purposes
		if tmp[5] == "disk" || tmp[5] == "loop" {
			tsize, _ := strconv.ParseUint(tmp[3], 10, 64)
			result = append(result, diskDetail{tmp[0], tsize})
		}
	}
	return result, nil
}

func getDiskIdentifier(disk string) (string, error) {
	out, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDiskID, disk), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not list disk error: %s %s", string(out), err)
		return "", err
	}
	sli := strings.Split(string(out), "\n")
	for _, value := range sli {
		tmp := strings.Fields(value)
		if len(tmp) == 0 {
			continue
		}
		if tmp[0] == "Disklabel" && tmp[1] == "type:" && tmp[2] != "gpt" {
			return "", errors.New("Not an GPT disk")
		}
		if tmp[0] == "Disk" && tmp[1] == "identifier:" {
			return tmp[2], nil
		}
	}
	return "", errors.New("Invalid Disk")
}

func getDiskMetaName(diskName string) (string, error) {
	tmpList, err := GetPartitionList(diskName, "", false)
	if err != nil {
		klog.Infof("GetPart Error, %s", diskName)
		return "", err
	}
	for _, tmp := range tmpList {
		if tmp[0] == "1" {
			// DiskMetaPartition will not contain Flags, Filesystem, hence the number of columns would be 5
			// and the Name will not contain special characters
			last := tmp[len(tmp)-1]
			if len(tmp) == 5 && regexp.MustCompile(`^[a-zA-Z0-9_.-]*$`).MatchString(last) && last != "ext4" {
				return last, nil
			}
		}
	}
	return "", errors.New("Meta Partition not found")

}

// GetDiskDetails Todo
func GetDiskDetails() ([]apis.Device, error) {
	var result []apis.Device
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("Device LocalPV: could not list disk error: %+v", err)
		return nil, err
	}
	for _, diskIter := range diskList {
		metaName, err := getDiskMetaName(diskIter.DiskName)
		if err != nil {
			klog.Errorf("Device LocalPV: getDiskMetaName Failed %s", diskIter.DiskName)
			continue
		}
		id, err := getDiskIdentifier(diskIter.DiskName)
		if err != nil {
			klog.Errorf("Device LocalPV: getDiskIdentifier Failed %s", diskIter.DiskName)
			continue
		}
		free, err := GetFreeCapacity(diskIter.DiskName)
		if err != nil {
			klog.Errorf("Device LocalPV: GetFreeCapacity Failed %s", diskIter.DiskName)
			continue
		}
		result = append(result, apis.Device{metaName, id,
			*resource.NewQuantity(int64(diskIter.Size), resource.DecimalSI),
			*resource.NewQuantity(int64(free*1024*1024), resource.DecimalSI)})
	}

	klog.Infof("%+v", result)
	return result, nil
}

// getPartitionPath gets the partition path from disk name and partition number.
func getPartitionPath(diskName string, partNum uint32) string {
	r := regexp.MustCompile(".+[0-9]+$")
	// if the disk name ends in a number, then partition will be of the format /dev/nvme0n1p1
	if r.MatchString(diskName) {
		return fmt.Sprintf("/dev/%sp%d", diskName, partNum)
	} else {
		return fmt.Sprintf("/dev/%s%d", diskName, partNum)
	}
}
