# Copyright 2020-2021 The OpenEBS Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM ubuntu:20.04

RUN apt-get update

RUN apt-get install parted -y

CMD [ "bash" ]

###################################################################################
# This Dockerfile is used to create the image `quay.io/w3aman/device-meta-partition:ci`#
# which is being used in the daemonset in the file `device_partition_ds.yml`      #
# Using this privileged daemonset we will run `parted` command on nodes.
###################################################################################
