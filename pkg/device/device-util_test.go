/*
 Copyright © 2021 The OpenEBS Authors

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
	"reflect"
	"strconv"
	"testing"
)

func Test_getMetaPartition(t *testing.T) {

	beginBytes, _ := strconv.ParseUint("1048576", 10, 64)
	endBytes, _ := strconv.ParseUint("2097151", 10, 64)
	sizeBytes, _ := strconv.ParseUint("1048576", 10, 64)

	tests := []struct {
		name     string
		args     partedOutput
		partName string
		exists   bool
	}{
		{
			name: "valid meta partition",
			args: partedOutput{
				partNum:    1,
				beginBytes: beginBytes,
				endBytes:   endBytes,
				size:       sizeBytes,
				partName:   "HDD-JBOD-2216723-1",
			},
			partName: "HDD-JBOD-2216723-1",
			exists:   true,
		},
		{
			name: "invalid meta partition",
			args: partedOutput{
				partNum:    1,
				beginBytes: beginBytes,
				endBytes:   endBytes,
				size:       sizeBytes,
				partName:   "HDD-JBOD-2216723-1",
			},
			partName: "",
			exists:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partName, exists := getMetaPartition(tt.args)
			if partName != tt.partName {
				t.Errorf("getMetaPartition() got = %v, want %v", partName, tt.partName)
			}
			if exists != tt.exists {
				t.Errorf("getMetaPartition() got = %v, want %v", exists, tt.exists)
			}
		})
	}
}

func Test_parsePartUsed(t *testing.T) {

	beginBytes, _ := strconv.ParseUint("2097152", 10, 64)
	endBytes, _ := strconv.ParseUint("9500469755903", 10, 64)
	sizeBytes, _ := strconv.ParseUint("9500467658752", 10, 64)

	tests := []struct {
		name     string
		diskName string
		row      partedOutput
		partUsed PartUsed
		wantErr  bool
	}{
		{
			name:     "valid partition",
			diskName: "sdc",
			row: partedOutput{
				partNum:    2,
				beginBytes: beginBytes,
				endBytes:   endBytes,
				size:       sizeBytes,
				fsType:     "ext4",
				partName:   "5d8d56cb-e291-4dfd-81ac-fb664dd5ec75",
			},
			partUsed: PartUsed{
				DiskPath:   "sdc",
				PartNum:    2,
				Name:       "5d8d56cb-e291-4dfd-81ac-fb664dd5ec75",
				DevicePath: "/dev/sdc2",
				Size:       9500467658752,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partUsed, err := parsePartUsed(tt.diskName, tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePartUsed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(partUsed, tt.partUsed) {
				t.Errorf("parsePartUsed() got = %v, want %v", partUsed, tt.partUsed)
			}
		})
	}
}

func Test_parsePartedPartitionRow(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    partedOutput
		wantErr bool
	}{
		{
			name: "valid partition row",
			args: "1:1048576B:10485759B:9437184B::test-device:;",
			want: partedOutput{
				1,
				1048576,
				10485759,
				9437184,
				"",
				"test-device",
				"",
			},
			wantErr: false,
		},
		{
			name: "valid free slot",
			args: "1:10485760B:17179852287B:17169366528B:free;",
			want: partedOutput{
				partNum:    1,
				beginBytes: 10485760,
				endBytes:   17179852287,
				size:       17169366528,
				fsType:     freeSlotFSType,
				partName:   "",
				flags:      "",
			},
			wantErr: false,
		},
		{
			name: "partition with fs and flags",
			args: "1:1048576B:511705087B:510656512B:fat32::boot, esp;",
			want: partedOutput{
				partNum:    1,
				beginBytes: 1048576,
				endBytes:   511705087,
				size:       510656512,
				fsType:     "fat32",
				partName:   "",
				flags:      "boot, esp",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePartedPartitionRow(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePartedPartitionRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePartedPartitionRow() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePartFree(t *testing.T) {
	tests := []struct {
		name string
		args partedOutput
		want partFree
	}{
		{
			name: "valid free slot",
			args: partedOutput{
				partNum:    1,
				beginBytes: 10485760,
				endBytes:   17179852287,
				size:       17169366528,
				fsType:     freeSlotFSType,
				partName:   "",
				flags:      "",
			},
			want: partFree{
				StartMiB: 10,
				EndMiB:   16383,
				SizeMiB:  16373,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePartFree(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePartFree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPartitionPath(t *testing.T) {
	type args struct {
		diskName string
		partNum  uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "partition of loop device",
			args: args{
				diskName: "loop0",
				partNum:  1,
			},
			want: "/dev/loop0p1",
		},
		{
			name: "partition of scsi device",
			args: args{
				diskName: "sda",
				partNum:  2,
			},
			want: "/dev/sda2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPartitionPath(tt.args.diskName, tt.args.partNum); got != tt.want {
				t.Errorf("getPartitionPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
