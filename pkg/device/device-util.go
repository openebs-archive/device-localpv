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

package device

import (
	"fmt"
	"github.com/openebs/lib-csi/pkg/common/errors"
	"k8s.io/klog"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
)

// Partition Commands
const (
	PartitionDiskList  = "lsblk"
	PartitionPrintFree = "parted /dev/%s unit MiB print free --script"
	PartitionPrint     = "parted /dev/%s unit MiB print --script"
	PartitionCreate    = "parted /dev/%s mkpart %s %dMiB %dMiB --script"
	PartitionDelete    = "parted /dev/%s rm %d --script"
)

type partUsed struct {
	disk    string
	partNum int
}
type partFree struct {
	disk  string
	start int
	end   int
	size  int
}

// CreateVolume Todo
func CreateVolume(vol *apis.DeviceVolume) error {
	//func CreatePartition(diskName string, partitionName string, size int) error {
	diskName := vol.Spec.DevName
	partitionName := vol.Name
	// ToDo Capacity is assumed to be in MiB
	size, _ := strconv.Atoi(vol.Spec.Capacity)
	pList, err := getAllPartsUsed(diskName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return err
	}
	if len(pList) > 0 {
		klog.Errorf("Partition %s already exist", partitionName)
		return errors.New("Already exist")
	}
	disk, start, err := FindBestPart(diskName, size)
	if err != nil {
		klog.Errorf("FindBestPart Failed")
		return err
	}
	return CreatePart(disk, start, partitionName, size)
}

// CreatePart Todo
func CreatePart(disk string, start int, partitionName string, size int) error {
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionCreate, disk, partitionName, start, start+size), " "))
	return err
}

// GetAllPartsUsed Todo
func getAllPartsUsed(diskName string, partitionName string) ([]partUsed, error) {
	list, err := GetDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []partUsed
	for _, disk := range list {
		tmpList, err := GetPartitionList(disk, diskName, false)
		if err != nil {
			klog.Infof("GetPart Error, %s", disk)
			continue
		}
		for _, tmp := range tmpList {
			if tmp[len(tmp)-1] == partitionName {
				partNum, _ := strconv.ParseInt(tmp[0], 10, 32)
				pList = append(pList, partUsed{disk, int(partNum)})
			}
		}
	}
	return pList, nil

}

// GetPartitionList Todo
func GetPartitionList(disk string, diskName string, free bool) ([][]string, error) {
	var command string
	if free {
		command = PartitionPrintFree
	} else {
		command = PartitionPrint
	}
	out, err := RunCommand(strings.Split(fmt.Sprintf(command, disk), " "))
	if err != nil {
		klog.Errorf("Device LocalPV: could not get parts error: %s %v\n", string(out), err)
		fmt.Printf("Device LocalPV: could not get parts error: %s %v\n", string(out), err)
		return nil, err
	}
	sli := strings.Split(string(out), "\n")
	start := 0
	var result [][]string
	for _, value := range sli {
		tmp := strings.Fields(value)
		if len(tmp) == 0 {
			continue
		}
		if tmp[0] == "Partition" && tmp[1] == "Table" && tmp[2] != "gpt" {
			klog.Infof("Disk: %s Not an GPT Partitioned Disk", disk)
			return nil, errors.New("Wrong Partition type")
		}
		if start == 0 {
			if tmp[0] == "Number" {
				start = 1
			}
			continue
		}
		if diskName != "" && tmp[0] == "1" && tmp[len(tmp)-1] != diskName {
			klog.Errorf("Disk: DiskName not correct")
			return nil, errors.New("Wrong DiskName")
		}
		result = append(result, tmp)
	}
	fmt.Println(result)
	return result, nil
}

// DestroyVolume Todo
func DestroyVolume(vol *apis.DeviceVolume) error {
	diskName := vol.Spec.DevName
	partitionName := vol.Name
	pList, err := getAllPartsUsed(diskName, partitionName)
	if err != nil {
		klog.Errorf("GetAllPartsUsed failed %s", err)
		return err
	}
	fmt.Println(pList)
	if len(pList) > 1 {
		klog.Errorf("More than one partition of same name %s\n", partitionName)
		return errors.New("More than one partition of same name")
	}
	if len(pList) == 0 {
		klog.Errorf("%s Partition not found\n", partitionName)
		return errors.New("Partition not found")
	}
	return DeletePart(pList[0].disk, pList[0].partNum)

}

// DeletePart Todo
func DeletePart(disk string, partNum int) error {
	_, err := RunCommand(strings.Split(fmt.Sprintf(PartitionDelete, disk, partNum), " "))
	return err
}

// GetVolumeDevPath Todo
func GetVolumeDevPath(vol *apis.DeviceVolume) (string, error) {
	//func GetPath(diskName string, partitionName string) (string, error) {
	diskName := vol.Spec.DevName
	partitionName := vol.Name
	list, err := GetDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return "", err
	}
	for _, disk := range list {
		pList, err := GetPartitionList(disk, diskName, true)
		if err != nil {
			klog.Infof("GetPart Error, %s", disk)
			continue
		}
		for _, tmp := range pList {
			if tmp[len(tmp)-1] == partitionName {
				return disk + tmp[0], nil
			}
		}
	}
	return "", errors.New("Partition not found")
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

// GetDiskList Todo
func GetDiskList() ([]apis.Device, error) {
	var result []apis.Device
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
		if tmp[5] == "disk" {
			result = append(result, apis.Device{
				Name: tmp[0],
			})
		}
	}
	return result, nil
}

//func GetPartNum(disk string, name string) (int, error){
//	pList, err := GetPartitionList(disk, "",false)
//	if err != nil {
//		klog.Errorf("Device LocalPV: could not get parts error: %s", err)
//		return 0, err
//	}
//	for _, line := range pList {
//		if line[len(line)-1] == name {
//			num, _ := strconv.Atoi(line[0])
//			return num, nil
//		}
//	}
//	return 0, nil
//}

// getAllPartsFree Todo
func getAllPartsFree(diskName string) ([]partFree, error) {
	list, err := GetDiskList()
	if err != nil {
		klog.Errorf("GetDiskList failed %s", err)
		return nil, err
	}
	var pList []partFree
	for _, disk := range list {
		tmpList, err := GetPartitionList(disk, diskName, true)
		if err != nil {
			klog.Infof("GetPart Error, %s", disk)
			continue
		}
		for _, tmp := range tmpList {
			if tmp[3] == "Free" {
				begin, _ := strconv.ParseFloat(string(tmp[0][:len(tmp[0])-3]), 32)
				end, _ := strconv.ParseFloat(string(tmp[1][:len(tmp[1])-3]), 32)
				size := int(math.Floor(end) - math.Ceil(begin))
				fmt.Println(begin, end, size)
				pList = append(pList, partFree{disk, int(math.Ceil(begin)), int(math.Floor(end)), size})
			}
		}
	}
	return pList, nil

}

// GetCapacity Todo
func GetCapacity(diskName string) (int, error) {
	pList, err := getAllPartsFree(diskName)
	if err != nil {
		klog.Errorln("Device LocalPV: GetAllPartsFree error")
		return 0, err
	}
	fmt.Println(pList)
	if len(pList) > 0 {
		sort.Slice(pList, func(i, j int) bool {
			// ">" Descending order
			return pList[i].size > pList[j].size
		})
		return pList[0].size, nil
	}
	klog.Errorln("Device LocalPV: Free space for partition is not found")
	return 0, err

}

// FindBestPart Todo
func FindBestPart(diskName string, partSize int) (string, int, error) {
	pList, err := getAllPartsFree(diskName)
	if err != nil {
		klog.Errorln("Device LocalPV: GetAllPartsFree error")
		return "", 0, err
	}
	fmt.Println(pList)

	if len(pList) > 0 {
		sort.Slice(pList, func(i, j int) bool {
			// "<" Ascending order
			return pList[i].size < pList[j].size
		})
		for _, tmp := range pList {
			if tmp.size > partSize {
				return tmp.disk, tmp.start, nil
			}
		}
	}
	klog.Errorln("Device LocalPV: Free space for partition is not found")
	return "", 0, err

}
