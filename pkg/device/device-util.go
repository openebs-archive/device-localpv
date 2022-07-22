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
	"math"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/openebs/lib-csi/pkg/common/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
)

// Partition Commands
const (
	PartitionDiskID    = "fdisk -l /dev/%s"
	PartitionDiskList  = "lsblk -b"
	PartitionPrintFree = "parted /dev/%s unit b print free --script -m"
	PartitionPrint     = "parted /dev/%s unit b print --script -m"
	PartitionCreate    = "parted /dev/%s mkpart %s %dMiB %dMiB --script"
	PartitionDelete    = "parted /dev/%s rm %d --script"
	PartitionWipeFS    = "wipefs --force -a %s"
)

// types of devices which the plugin support
const (
	deviceTypeDisk = "disk"
	deviceTypeLoop = "loop"
)

const (
	partitionTypeGPT    = "gpt"
	metaPartitionNumber = 1
	freeSlotFSType      = "free"
)

// column indices for command outputs
const (
	partedDiskInfoPartTypeIndex = 5

	partedPartInfoPartNumIndex    = 0
	partedPartInfoBeginBytesIndex = 1
	partedPartInfoEndBytesIndex   = 2
	partedPartInfoSizeIndex       = 3
	partedPartInfoFSTypeIndex     = 4
	partedPartInfoPartNameIndex   = 5
	partedPartInfoFlagSetIndex    = 6

	partedOutputDefaultNoOfColumns       = 7
	partedOutputFreePartitionNoOfColumns = 5

	lsblkDevTypeIndex = 5
	lsblkSizeIndex    = 3
	lsblkNameIndex    = 0
)

// partedOutput is the partition output as produced from parted command
// It can be either of the below 2, depending on whether free slots are printed or not
// BYT;
// /dev/sdd:17179869184B:scsi:512:512:gpt:VMware Virtual disk:;
// 1:1048576B:10485759B:9437184B::test-device:;
//
//
// BYT;
// /dev/sdd:17179869184B:scsi:512:512:gpt:VMware Virtual disk:;
// 1:17408B:1048575B:1031168B:free;
// 1:1048576B:10485759B:9437184B::test-device:;
// 1:10485760B:17179852287B:17169366528B:free;
type partedOutput struct {
	partNum    uint32
	beginBytes uint64
	endBytes   uint64
	// size in bytes
	size uint64
	// if fsType is "free", partNum, partName and flags fields are invalid
	fsType   string
	partName string
	flags    string
}

// PartUsed represents disk partition created by device plugin.
type PartUsed struct {
	DiskPath string
	PartNum  uint32

	// Name denotes name of the partition.
	Name string

	// DevicePath denotes path of partition device file.
	DevicePath string

	// Total size of the partition in bytes.
	Size uint64
}

// GetPVName returns the related persistent volume name.
func (p *PartUsed) GetPVName() string {
	return fmt.Sprintf("pvc-%v", p.Name)
}

type partFree struct {
	DiskName string
	StartMiB uint64
	EndMiB   uint64
	SizeMiB  uint64
}

type diskDetail struct {
	DiskPath   string
	Size       uint64
	deviceType string
}

// CreateVolume creates a partition on the disk with partition name as the pv name
// and size as pv size.
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
		return &apis.VolumeError{
			Code:    apis.InsufficientCapacity,
			Message: err.Error(),
		}
	}
	return createPartAndWipeFS(disk, start, partitionName, capacityMiB, diskMetaName)
}

// createPartAndWipeFS creates a partition at the provided start address
// and perform a wipefs operation on the created partition.
func createPartAndWipeFS(disk string, start uint64, partitionName string, size uint64, diskMetaName string) error {
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

	if len(pList) == 0 {
		return fmt.Errorf("could not find created partition %s", partitionName)
	}

	err = wipeFsPartition(pList[0].DiskPath, pList[0].PartNum)
	if err != nil {
		klog.Infof("Deleting partition %d on disk %s because wipefs failed", pList[0].PartNum, pList[0].DiskPath)
		err1 := deletePartition(pList[0].DiskPath, pList[0].PartNum)
		if err1 != nil {
			klog.Errorf("could not delete partition %d on disk %s, created during CreateVolume(). Error: %s", pList[0].PartNum, pList[0].DiskPath, err1)
		}
		// the error will be returned irrespective of the return value of delete partition,
		// as create partition has failed.
		return err
	}
	return nil
}

// getAllPartsFree lists all the free slots on the disk with the provided
// disk meta name
func getAllPartsFree(diskMetaName string) ([]partFree, error) {
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []partFree
	for _, disk := range diskList {
		tmpList, err := getPartsFree(disk.DiskPath, diskMetaName)
		if err != nil {
			klog.V(4).Infof("GetPart Error, %s", disk.DiskPath)
			continue
		}
		pList = append(pList, tmpList...)
	}
	return pList, nil
}

// findBestPart returns the disk path and start address which fits for creating a
// new partition of size partSize
func findBestPart(diskMetaName string, partSize uint64) (string, uint64, error) {
	pList, err := getAllPartsFree(diskMetaName)
	if err != nil {
		klog.Errorln("Device LocalPV: GetAllPartsFree error")
		return "", 0, err
	}

	if len(pList) == 0 {
		return "", 0, fmt.Errorf("could not find any free slots on disk name: %s", diskMetaName)
	}

	if len(pList) > 0 {
		sort.Slice(pList, func(i, j int) bool {
			// "<" Ascending order
			return pList[i].SizeMiB < pList[j].SizeMiB
		})
		for _, tmp := range pList {
			if tmp.SizeMiB >= partSize {
				return tmp.DiskName, tmp.StartMiB, nil
			}
		}
	}
	klog.Errorln("Device LocalPV: Free space for partition is not found")
	return "", 0, fmt.Errorf("free space of %dMiB not found on disk name: %s", partSize, diskMetaName)

}

// getAllPartsUsed returns the list of all partitions that are in use on the disk
// with given disk meta name
func getAllPartsUsed(diskMetaName string, partitionName string) ([]PartUsed, error) {
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []PartUsed
	for _, disk := range diskList {
		tmpList, err := GetPartitionList(disk.DiskPath, diskMetaName, false)
		if err != nil {
			klog.V(4).Infof("GetPart Error, %+v", disk)
			continue
		}
		for _, tmp := range tmpList {
			if tmp.partName == partitionName {
				partUsed, err := parsePartUsed(disk.DiskPath, tmp)
				if err != nil {
					return pList, err
				}
				pList = append(pList, partUsed)
			}
		}
	}
	return pList, nil
}

// parsePartUsed converts the partedOutput to PartUsed struct
func parsePartUsed(diskPath string, row partedOutput) (PartUsed, error) {
	p := PartUsed{DiskPath: diskPath}

	p.PartNum = row.partNum
	p.Name = row.partName
	p.DevicePath = getPartitionPath(diskPath, p.PartNum)
	p.Size = row.size
	return p, nil
}

// parsePartFree converts the partedOutput to partFree struct
func parsePartFree(row partedOutput) partFree {
	beginMib := math.Ceil(float64(row.beginBytes) / 1024 / 1024)
	endMib := math.Floor(float64(row.endBytes) / 1024 / 1024)
	sizeMib := uint64(0)
	if endMib > beginMib {
		// calculate the part free size, needs increase the difference by 1
		sizeMib = uint64(math.Floor(float64(row.endBytes-row.beginBytes+1) / 1024 / 1024))
	}

	return partFree{
		StartMiB: uint64(beginMib),
		EndMiB:   uint64(endMib),
		SizeMiB:  sizeMib,
	}
}

// DestroyVolume gets the partition corresponding to a DeviceVolume resource, wipes
// the partition and delete the partition from the disk.
func DestroyVolume(vol *apis.DeviceVolume) error {
	diskMetaName := vol.Spec.DevName
	partitionName := getPartitionName(vol.Name)
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
	return wipeFSAndDeletePart(pList[0].DiskPath, pList[0].PartNum)

}

// wipeFSAndDeletePart performs a wipefs operation on the partition and then
// deletes the partition from the disk
func wipeFSAndDeletePart(disk string, partNum uint32) error {
	err := wipeFsPartition(disk, partNum)
	if err != nil {
		return err
	}
	return deletePartition(disk, partNum)
}

// deletePartition deletes the partition from the disk
func deletePartition(disk string, partNum uint32) error {
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDelete, disk, partNum), " "))
	if err != nil {
		klog.Errorf("Delete Partition failed for disk: %s, partition: %d . Error: %s", disk, partNum, err)
	}
	return err
}

// wipeFsPartition performs a force wipefs on the given partition
func wipeFsPartition(disk string, partNum uint32) error {
	klog.Infof("Running WipeFS for disk: %s, partition %d", disk, partNum)
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionWipeFS, getPartitionPath(disk, partNum)), " "))
	if err != nil {
		klog.Errorf("WipeFS failed for disk: %s, partition: %d . Error: %s", disk, partNum, err)
	}
	return err
}

// GetVolumeDevPath returns the path to the volume.
// eg: /dev/sda1, /dev/nvme0n1p1
func GetVolumeDevPath(vol *apis.DeviceVolume) (string, error) {
	diskMetaName := vol.Spec.DevName
	partitionName := getPartitionName(vol.Name)
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

	return getPartitionPath(pList[0].DiskPath, pList[0].PartNum), nil
}

// RunCommand runs command and returns the output/error
func RunCommand(cList []string) (string, error) {
	cmd := exec.Command(cList[0], cList[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("Device LocalPV: could not Run command %+v\n", cList)
		return "", err
	}
	return string(out), nil
}

// GetPartitionList gets the list of free/used partitions on the given disk with
// the given meta partition name
func GetPartitionList(diskPath string, diskMetaName string, free bool) ([]partedOutput, error) {
	var command string
	if free {
		command = PartitionPrintFree
	} else {
		command = PartitionPrint
	}
	out, err := RunCommand(strings.Split(fmt.Sprintf(command, diskPath), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not get parts error: %s %v\n", out, err)
		return nil, err
	}
	// the output will be of the following format
	// BYT;
	// "path":"size":"transport-type":"logical-sector-size":"physical-sector-size":"partition-table-type":"model-name";
	// "number":"begin":"end":"size":"filesystem-type":"partition-name":"flags-set";
	//
	// CHS/CYL types are not considered as they are not that common.
	rows := strings.Split(out, "\n")

	deviceRow := strings.Split(rows[1], ":")
	if deviceRow[partedDiskInfoPartTypeIndex] != partitionTypeGPT {
		klog.Infof("Disk: %s Not a GPT Partitioned Disk", diskPath)
		return nil, errors.New("Wrong Partition type")
	}

	var result []partedOutput
	// skipping first 2 rows as they contain output units and disk information
	for _, value := range rows[2:] {

		if len(value) == 0 {
			continue
		}
		partitionRow, err := parsePartedPartitionRow(value)

		if err != nil {
			klog.Errorf("parsing parted output failed")
			return nil, fmt.Errorf("parsing parted output failed for disk %s, err: %v", diskPath, err)
		}

		devRegex, err := regexp.Compile(diskMetaName)
		if err != nil {
			klog.Infof("Disk: Regex compile failure %s, %+v", diskMetaName, err)
			return nil, err
		}
		if diskMetaName != "" &&
			partitionRow.fsType != freeSlotFSType &&
			partitionRow.partNum == metaPartitionNumber &&
			!devRegex.MatchString(partitionRow.partName) {
			klog.Errorf("Disk: %s DiskPath not correct, partition entry: %v", diskPath, partitionRow)
			return nil, errors.New("Wrong DiskMetaName")
		}

		result = append(result, partitionRow)
	}
	return result, nil
}

// getPartsFree lists the free slots on the disk and returns it as a slice
// of partFree type
func getPartsFree(diskPath string, diskMetaName string) ([]partFree, error) {
	var pList []partFree
	tmpList, err := GetPartitionList(diskPath, diskMetaName, true)
	if err != nil {
		klog.V(4).Infof("GetPartitionList error, path: %s metaName: %s, error: %v", diskPath, diskMetaName, err)
		return nil, errors.New("GetPartitionList Error")
	}
	for _, tmp := range tmpList {
		if tmp.fsType == freeSlotFSType {
			part := parsePartFree(tmp)
			part.DiskName = diskPath
			pList = append(pList, part)
		}
	}
	return pList, nil
}

// GetFreeCapacity returns the size of the maximum free slot available on the disk
func GetFreeCapacity(diskPath string) (uint64, error) {
	pList, err := getPartsFree(diskPath, "")
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
	return 0, nil
}

// getDiskList gets the list of disks on the node with path and size
func getDiskList() ([]diskDetail, error) {
	var result []diskDetail
	out, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDiskList), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not list blockDeviceEntry error: %s %v", out, err)
		return nil, err
	}
	blockDeviceList := strings.Split(out, "\n")
	for _, blockDeviceEntry := range blockDeviceList {
		// a sample blockDeviceEntry entry look like
		// sdb        8:16   0 17179869184  0 disk
		tmp := strings.Fields(blockDeviceEntry)
		if len(tmp) <= lsblkDevTypeIndex {
			continue
		}

		// loop is added here for testing purposes
		if tmp[lsblkDevTypeIndex] == deviceTypeDisk ||
			tmp[lsblkDevTypeIndex] == deviceTypeLoop {
			diskSize, err := strconv.ParseUint(tmp[lsblkSizeIndex], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing disk size from %v. error %s", tmp, err)
			}
			result = append(result, diskDetail{tmp[lsblkNameIndex], diskSize, tmp[5]})
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
		if tmp[0] == "Disklabel" && tmp[1] == "type:" && tmp[2] != partitionTypeGPT {
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
		klog.Errorf("GetPart Error, %s. err: %v", diskName, err)
		return "", err
	}
	for _, tmp := range tmpList {
		if partName, ok := getMetaPartition(tmp); ok {
			return partName, err
		}
	}
	return "", errors.New("Meta Partition not found")
}

// getMetaPartition checks if the given parted row is meta partition or not.
func getMetaPartition(partedRow partedOutput) (string, bool) {
	if partedRow.partNum == metaPartitionNumber &&
		// DiskMetaPartition will not contain Flags, Filesystem,
		// and the Name will not contain special characters
		regexp.MustCompile(`^[a-zA-Z0-9_.-]*$`).MatchString(partedRow.partName) &&
		partedRow.fsType == "" &&
		partedRow.flags == "" {
		return partedRow.partName, true
	}
	return "", false
}

// GetDiskDetails gets the list of all disks on the node along with disk metaname,
// unique identifier for the disk, free and total size of the disk.
func GetDiskDetails() ([]apis.Device, error) {
	var result []apis.Device
	diskList, err := getDiskList()
	if err != nil {
		klog.Errorf("Device LocalPV: could not list disk error: %+v", err)
		return nil, err
	}
	for _, diskIter := range diskList {

		metaName, err := getDiskMetaName(diskIter.DiskPath)
		if err != nil {
			klog.Errorf("Device LocalPV: getDiskMetaName Failed %s, error: %v", diskIter.DiskPath, err)
			continue
		}
		id, err := getDiskIdentifier(diskIter.DiskPath)
		if err != nil {
			klog.Errorf("Device LocalPV: getDiskIdentifier Failed %s, error: %v", diskIter.DiskPath, err)
			continue
		}
		free, err := GetFreeCapacity(diskIter.DiskPath)
		if err != nil {
			klog.Errorf("Device LocalPV: GetFreeCapacity Failed %s, error: %v", diskIter.DiskPath, err)
			continue
		}
		result = append(result, apis.Device{
			Name: metaName,
			UUID: id,
			Size: *resource.NewQuantity(int64(diskIter.Size), resource.BinarySI),
			Free: *resource.NewQuantity(int64(free*1024*1024), resource.BinarySI),
		})
	}

	return result, nil
}

// ListPartUsed lists all disk partitions created by plugin.
func ListPartUsed() ([]PartUsed, error) {
	diskList, err := getDiskList()
	if err != nil {
		return nil, fmt.Errorf("failed to list disk: %v", err)
	}
	plist := make([]PartUsed, 0)
	for _, disk := range diskList {
		tmpList, err := GetPartitionList(disk.DiskPath, "", false)
		if err != nil {
			klog.Errorf("failed to list partition for disk %q: %v", disk.DiskPath, err)
			continue
		}
		if len(tmpList) == 0 {
			continue
		}
		// see if the first partition is meta partition or not.
		if _, ok := getMetaPartition(tmpList[0]); !ok {
			continue
		}
		// ignoring first meta partition
		for i := 1; i < len(tmpList); i++ {
			part, err := parsePartUsed(disk.DiskPath, tmpList[i])
			if err != nil {
				return nil, fmt.Errorf("failed to parse parted output: %v", err)
			}
			plist = append(plist, part)
		}
	}
	return plist, nil
}

// getPartitionPath gets the partition path from disk name and partition number.
func getPartitionPath(diskName string, partNum uint32) string {
	r := regexp.MustCompile(".+[0-9]+$")
	// if the disk name ends in a number, then partition will be of the format /dev/nvme0n1p1
	if r.MatchString(diskName) {
		return fmt.Sprintf("/dev/%sp%d", diskName, partNum)
	}
	return fmt.Sprintf("/dev/%s%d", diskName, partNum)
}

// getPartitionName returns the partition name from volume name
func getPartitionName(volumeName string) string {
	return strings.TrimPrefix(volumeName, "pvc-")
}

// parsePartedPartitionRow parses a single partition row in the output of parted command
func parsePartedPartitionRow(partitionRow string) (partedOutput, error) {

	// trim the ; from machine parseable output
	partitionRow = strings.TrimSuffix(partitionRow, ";")

	tmp := strings.Split(partitionRow, ":")

	var partition partedOutput

	partitionNumber, err := strconv.ParseInt(tmp[partedPartInfoPartNumIndex], 10, 32)
	if err != nil {
		return partition, fmt.Errorf("invalid partition number. row: %s, err: %v", partitionRow, err)
	}

	beginBytes, err := strconv.ParseUint(tmp[partedPartInfoBeginBytesIndex][:len(tmp[partedPartInfoBeginBytesIndex])-1], 10, 64)
	if err != nil {
		return partition, fmt.Errorf("failed to parse begin bytes from %v. err: %s", partitionRow, err)
	}

	endBytes, err := strconv.ParseUint(tmp[partedPartInfoEndBytesIndex][:len(tmp[partedPartInfoEndBytesIndex])-1], 10, 64)
	if err != nil {
		return partition, fmt.Errorf("failed to parse end bytes from %v. err: %s", partitionRow, err)
	}

	size, err := strconv.ParseUint(tmp[partedPartInfoSizeIndex][:len(tmp[partedPartInfoSizeIndex])-1], 10, 64)
	if err != nil {
		return partition, fmt.Errorf("failed to parse size from %v. err: %s", partitionRow, err)
	}

	partition.partNum = uint32(partitionNumber)
	partition.beginBytes = beginBytes
	partition.endBytes = endBytes
	partition.size = size

	if len(tmp) == partedOutputDefaultNoOfColumns {
		partition.fsType = tmp[partedPartInfoFSTypeIndex]
		partition.partName = tmp[partedPartInfoPartNameIndex]
		partition.flags = tmp[partedPartInfoFlagSetIndex]
	} else if len(tmp) == partedOutputFreePartitionNoOfColumns {
		partition.fsType = tmp[partedPartInfoFSTypeIndex]
	} else {
		return partition, fmt.Errorf("unexpected result format while parsing: %s", partitionRow)
	}
	return partition, nil
}
