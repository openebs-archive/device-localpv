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
	"testing"
)

func Test_getMetaPartition(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		partName string
		exists   bool
	}{
		{
			name:     "valid meta partition",
			args:     []string{"1", "1048576B", "2097151B", "1048576B", "HDD-JBOD-2216723-1"},
			partName: "HDD-JBOD-2216723-1",
			exists:   true,
		},
		{
			name:     "invalid meta partition",
			args:     []string{"1", "1048576B", "2097151B", "1048576B", "ext4", "HDD-JBOD-2216723-1"},
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
	tests := []struct {
		name     string
		diskName string
		row      []string
		partUsed PartUsed
		wantErr  bool
	}{
		{
			name:     "valid partition",
			diskName: "sdc",
			row:      []string{"2", "2097152B", "9500469755903B", "9500467658752B", "ext4", "5d8d56cb-e291-4dfd-81ac-fb664dd5ec75"},
			partUsed: PartUsed{
				DiskPath:   "sdc",
				PartNum:    2,
				Name:       "5d8d56cb-e291-4dfd-81ac-fb664dd5ec75",
				DevicePath: "/dev/sdc2",
				Size:       9500467658752,
			},
			wantErr: false,
		},
		{
			name:     "invalid partition",
			diskName: "sdc",
			row:      []string{"2", "2097152B", "9500469755903B", "ext4", "9500467658752B", "5d8d56cb-e291-4dfd-81ac-fb664dd5ec75"},
			partUsed: PartUsed{
				DiskPath: "sdc",
			},
			wantErr: true,
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
