package objects

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
)

func TestBlockDevices(t *testing.T) {
	tmpBlockDevices, err := os.CreateTemp("/tmp", "block-devices")
	if err != nil {
		t.Fatalf("error generate temp file for block devices yaml: %v", err)
	}

	defer os.Remove(tmpBlockDevices.Name())
	_, err = tmpBlockDevices.Write([]byte(sampleBlockDevices))
	if err != nil {
		t.Fatalf("error writing to temp file %s: %v", tmpBlockDevices.Name(), err)
	}

	objs, err := GenerateObjects(tmpBlockDevices.Name())
	if err != nil {
		t.Fatalf("error reading temp block device file %s %v", tmpBlockDevices.Name(), err)
	}
	for _, obj := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(obj)
		if err != nil {
			t.Fatal(err)
		}

		err = cleanupObjects(unstructObj.Object)
		if err != nil {
			t.Fatalf("error during cleanup of block device object %v", err)
		}

		err = objectHousekeeping(unstructObj)
		if err != nil {
			t.Fatalf("error during block device housekeeping: %v", err)
		}
		// check mountPoint values to ensure they have not been removed when they are null
		_, ok, err := unstructured.NestedFieldNoCopy(unstructObj.Object, "spec", "fileSystem", "mountPoint")
		if err != nil {
			t.Fatalf("error fetching spec.fileSystem.mountPoint on block device: %v ", err)
		}

		if !ok {
			t.Fatalf("could not find spec.fileSystem.mountPoint for %s \n%v", unstructObj.GetName(), unstructObj)
		}

		// check mountPoint in status
		_, ok, err = unstructured.NestedFieldNoCopy(unstructObj.Object, "status", "deviceStatus", "fileSystem", "mountPoint")
		if err != nil {
			t.Fatalf("error fetchign spec.fileSystem.mountPoint on block device: %v", err)
		}

		if !ok {
			t.Fatalf("could not find spec.fileSystem.mountPoint for %s", unstructObj.GetName())
		}

		// check for type in status
		_, ok, err = unstructured.NestedFieldNoCopy(unstructObj.Object, "status", "deviceStatus", "fileSystem", "type")
		if err != nil {
			t.Fatalf("error fetchign spec.fileystem.mountPath on block device: %v", err)
		}

		if !ok {
			t.Fatalf("could not find spec.filesystem.mountPath for %s", unstructObj.GetName())
		}
	}
}

const sampleBlockDevices = `
apiVersion: v1
items:
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-15T20:06:43Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 2833
    labels:
      kubernetes.io/hostname: r620-3
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: b2bf49dab59d30c911d1b51b6f464c18
    name: 002de72aa8fa90f69eb811a78ac07a5a
    namespace: longhorn-system
    resourceVersion: "15317509"
    uid: d60b4e68-54df-4264-9346-8ece853e5424
  spec:
    devPath: /dev/sda1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/002de72aa8fa90f69eb811a78ac07a5a
      provisioned: true
    nodeName: r620-3
  status:
    conditions:
    - lastUpdateTime: "2022-01-25T14:23:16Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-15T20:07:10Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-15T20:07:10Z"
      message: Added disk 002de72aa8fa90f69eb811a78ac07a5a to longhorn node r620-3
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498674085376e+12
      details:
        deviceType: part
        driveType: HDD
        partUUID: 3b99d689-e9c2-b940-a857-a80d006996e7
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-15T20:07:10Z"
        mountPoint: /var/lib/harvester/extra-disks/002de72aa8fa90f69eb811a78ac07a5a
        type: ext4
      parentDevice: /dev/sda
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-12T00:08:43Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 53
    labels:
      kubernetes.io/hostname: r620-6
      ndm.harvesterhci.io/device-type: disk
    name: 0951d18c526fb066f5fdc0a682fab681
    namespace: longhorn-system
    resourceVersion: "15317934"
    uid: 20d62b29-5c7b-4600-8c4e-8d3e0ffe6924
  spec:
    devPath: /dev/sdb
    fileSystem:
      mountPoint: "null"
    nodeName: r620-6
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T15:01:57Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.99116767232e+11
      details:
        busPath: pci-0000:02:00.0-scsi-0:2:1:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710
        ptUUID: 9df1eb36-f70c-b949-af9e-3553b217e1c7
        serialNumber: 6848f690e5001700297081a01f8976d9
        storageController: SCSI
        vendor: DELL
        wwn: 0x6848f690e5001700297081a01f8976d9
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-11T08:56:07Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 59
    labels:
      kubernetes.io/hostname: r620-2
      ndm.harvesterhci.io/device-type: disk
      manager: node-disk-manager
      operation: Update
      time: "2022-01-11T08:56:07Z"
    name: 0e47b784aca115df5d18cd1e12e953bc
    namespace: longhorn-system
    resourceVersion: "15314546"
    uid: 21ffdfdc-6ece-43f2-ae9e-896f5a0fa1af
  spec:
    devPath: /dev/sda
    fileSystem:
      mountPoint: "null"
    nodeName: r620-2
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T14:59:47Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498675150848e+12
      details:
        busPath: pci-0000:03:00.0-scsi-0:2:0:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710
        ptUUID: 4a7e9e30-a6fa-d440-b76d-8b8e481cdc92
        serialNumber: 6b083fe0ceb7640028f1a27b0785c958
        storageController: SCSI
        vendor: DELL
        wwn: 0x6b083fe0ceb7640028f1a27b0785c958
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-17T22:51:26Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 39
    labels:
      kubernetes.io/hostname: r620-4
      ndm.harvesterhci.io/device-type: disk
      manager: node-disk-manager
      operation: Update
      time: "2022-01-17T22:52:56Z"
    name: 1d527487729e5fac4d1370de26fc6b07
    namespace: longhorn-system
    resourceVersion: "15317316"
    uid: 86deebaf-dae3-41df-8f2c-f6f66ab2b690
  spec:
    devPath: /dev/sda
    fileSystem:
      mountPoint: "null"
    nodeName: r620-4
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T15:01:41Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.99116767232e+11
      details:
        busPath: pci-0000:02:00.0-scsi-0:2:0:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710
        ptUUID: 8ab6ea25-31ad-f344-8501-581ca69606e6
        serialNumber: 6848f690ea60da00297869624937afdf
        storageController: SCSI
        vendor: DELL
        wwn: 0x6848f690ea60da00297869624937afdf
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-11T09:01:56Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 48
    labels:
      kubernetes.io/hostname: r620-2
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: 0e47b784aca115df5d18cd1e12e953bc
      manager: node-disk-manager
      operation: Update
      time: "2022-03-31T14:59:47Z"
    name: 34775714a36d777f43a216dcd892dcf4
    namespace: longhorn-system
    resourceVersion: "15314547"
    uid: 7fd5ed0b-5670-4335-b616-17f7c13e3eab
  spec:
    devPath: /dev/sda1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/34775714a36d777f43a216dcd892dcf4
      provisioned: true
    nodeName: r620-2
  status:
    conditions:
    - lastUpdateTime: "2022-01-11T09:02:51Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-11T09:02:51Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-11T09:02:51Z"
      message: Added disk 34775714a36d777f43a216dcd892dcf4 to longhorn node r620-2
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498674085376e+12
      details:
        deviceType: part
        driveType: HDD
        partUUID: 6f95f5ee-ef00-004e-ae24-1745c27f5fe9
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-11T09:02:51Z"
        mountPoint: /var/lib/harvester/extra-disks/34775714a36d777f43a216dcd892dcf4
        type: ext4
      parentDevice: /dev/sda
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-12T00:11:43Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 49
    labels:
      kubernetes.io/hostname: r620-6
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: 0951d18c526fb066f5fdc0a682fab681
      manager: node-disk-manager
      operation: Update
      time: "2022-03-31T15:01:57Z"
    name: 5d2336bc53d3cf290d61a7b5a2a25aa0
    namespace: longhorn-system
    resourceVersion: "15317944"
    uid: 4a8cf4af-8172-4d81-aac3-6aaf5fd22280
  spec:
    devPath: /dev/sdb1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/0951d18c526fb066f5fdc0a682fab681
      provisioned: true
    nodeName: r620-6
  status:
    conditions:
    - lastUpdateTime: "2022-01-12T00:11:43Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-12T00:12:03Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-12T00:12:03Z"
      message: Added disk 5d2336bc53d3cf290d61a7b5a2a25aa0 to longhorn node r620-6
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.9911570176e+11
      details:
        deviceType: part
        driveType: HDD
        partUUID: 1f9407ac-e240-a547-a1fd-0adbcb6e1a98
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-12T00:12:03Z"
        mountPoint: /var/lib/harvester/extra-disks/0951d18c526fb066f5fdc0a682fab681
        type: ext4
      parentDevice: /dev/sdb
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-17T22:52:40Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 38
    labels:
      kubernetes.io/hostname: r620-4
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: 1d527487729e5fac4d1370de26fc6b07
      manager: node-disk-manager
      operation: Update
      time: "2022-03-31T15:01:41Z"
    name: 5eb48662ba0cff1da4f7bd982327f71c
    namespace: longhorn-system
    resourceVersion: "15317322"
    uid: 12dca94b-1976-480f-a2e4-1ddbe11c8130
  spec:
    devPath: /dev/sda1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/5eb48662ba0cff1da4f7bd982327f71c
      provisioned: true
    nodeName: r620-4
  status:
    conditions:
    - lastUpdateTime: "2022-01-17T22:53:19Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-17T22:53:19Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-17T22:53:19Z"
      message: Added disk 5eb48662ba0cff1da4f7bd982327f71c to longhorn node r620-4
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.9911570176e+11
      details:
        deviceType: part
        driveType: HDD
        partUUID: 34ab2dbd-b707-584e-a69a-3faf6a2f1e76
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-17T22:53:19Z"
        mountPoint: /var/lib/harvester/extra-disks/5eb48662ba0cff1da4f7bd982327f71c
        type: ext4
      parentDevice: /dev/sda
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-15T14:36:59Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 51
    labels:
      kubernetes.io/hostname: r620-3
      ndm.harvesterhci.io/device-type: disk
      manager: node-disk-manager
      operation: Update
      time: "2022-01-15T20:06:59Z"
    name: b2bf49dab59d30c911d1b51b6f464c18
    namespace: longhorn-system
    resourceVersion: "15317504"
    uid: f16fd995-4d5c-4fc0-af43-59a52bcdbab4
  spec:
    devPath: /dev/sda
    fileSystem:
      mountPoint: "null"
    nodeName: r620-3
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T15:01:45Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498675150848e+12
      details:
        busPath: pci-0000:03:00.0-scsi-0:2:0:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710
        ptUUID: fb0f0709-dc9b-7244-b4ee-c2a88481a954
        serialNumber: 6b8ca3a0f04ed80028f1a23a08256368
        storageController: SCSI
        vendor: DELL
        wwn: 0x6b8ca3a0f04ed80028f1a23a08256368
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-11T08:56:09Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 64
    labels:
      kubernetes.io/hostname: r620-1
      ndm.harvesterhci.io/device-type: disk
      manager: node-disk-manager
      operation: Update
      time: "2022-01-11T08:56:09Z"
    name: be5d3d1a9d2193e156a64244ef194120
    namespace: longhorn-system
    resourceVersion: "15363204"
    uid: e4276794-3337-4c94-89e4-d314292958da
  spec:
    devPath: /dev/sda
    fileSystem:
      mountPoint: "null"
    nodeName: r620-1
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T15:56:06Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498675150848e+12
      details:
        busPath: pci-0000:03:00.0-scsi-0:2:0:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710
        ptUUID: 9a146982-4feb-4843-9165-67e8363f7a84
        serialNumber: 6d4ae520b6e0f30028f1a1504461c437
        storageController: SCSI
        vendor: DELL
        wwn: 0x6d4ae520b6e0f30028f1a1504461c437
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-11T08:56:09Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 52
    labels:
      kubernetes.io/hostname: r620-1
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: be5d3d1a9d2193e156a64244ef194120
      manager: node-disk-manager
      operation: Update
      time: "2022-03-31T15:00:11Z"
    name: c74489c4b3ec30913195371a19085a25
    namespace: longhorn-system
    resourceVersion: "15314956"
    uid: 2987b9ef-ad4a-4f51-a699-90be5244906e
  spec:
    devPath: /dev/sda1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/c74489c4b3ec30913195371a19085a25
      provisioned: true
    nodeName: r620-1
  status:
    conditions:
    - lastUpdateTime: "2022-01-18T16:53:58Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-11T08:57:13Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-11T08:57:13Z"
      message: Added disk c74489c4b3ec30913195371a19085a25 to longhorn node r620-1
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 1.498673053696e+12
      details:
        deviceType: part
        driveType: HDD
        partUUID: 7a8cf793-4429-46a3-9e2a-8708170ca69d
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-11T08:57:13Z"
        mountPoint: /var/lib/harvester/extra-disks/c74489c4b3ec30913195371a19085a25
        type: ext4
      parentDevice: /dev/sda
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-01-17T19:35:48Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 37
    labels:
      kubernetes.io/hostname: r620-5
      ndm.harvesterhci.io/device-type: part
      ndm.harvesterhci.io/parent-device: f751e3b167a83f7d624988a26c699457
      manager: node-disk-manager
      operation: Update
      time: "2022-03-31T15:02:54Z"
    name: c8bbf9787bf856a0f765728217d83d34
    namespace: longhorn-system
    resourceVersion: "15319726"
    uid: 3c51e74b-56ef-4cac-939c-105af3fbbd03
  spec:
    devPath: /dev/sda1
    fileSystem:
      forceFormatted: true
      mountPoint: /var/lib/harvester/extra-disks/c8bbf9787bf856a0f765728217d83d34
      provisioned: true
    nodeName: r620-5
  status:
    conditions:
    - lastUpdateTime: "2022-01-17T19:36:15Z"
      status: "True"
      type: Mounted
    - lastUpdateTime: "2022-01-17T19:36:15Z"
      message: Done device ext4 filesystem formatting
      status: "False"
      type: Formatting
    - lastUpdateTime: "2022-01-17T19:36:15Z"
      message: Added disk c8bbf9787bf856a0f765728217d83d34 to longhorn node r620-5
        as an additional disk
      status: "True"
      type: AddedToNode
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.9911570176e+11
      details:
        deviceType: part
        driveType: HDD
        partUUID: 1ce50ec2-c238-5f4f-8707-82be8fea31d8
        storageController: SCSI
      fileSystem:
        LastFormattedAt: "2022-01-17T19:36:15Z"
        mountPoint: /var/lib/harvester/extra-disks/c8bbf9787bf856a0f765728217d83d34
        type: ext4
      parentDevice: /dev/sda
      partitioned: false
    provisionPhase: Provisioned
    state: Active
- apiVersion: harvesterhci.io/v1beta1
  kind: BlockDevice
  metadata:
    creationTimestamp: "2022-03-30T14:02:53Z"
    finalizers:
    - wrangler.cattle.io/harvester-block-device-handler
    generation: 3
    labels:
      kubernetes.io/hostname: r620-5
      ndm.harvesterhci.io/device-type: disk
      manager: node-disk-manager
      operation: Update
      time: "2022-03-30T14:02:53Z"
    name: ddd1b39c7ccaff2802f0b3f86bc76c90
    namespace: longhorn-system
    resourceVersion: "15319724"
    uid: 9b0af07b-9396-4447-b9f3-5fc4ffe3c532
  spec:
    devPath: /dev/sda
    fileSystem:
      mountPoint: "null"
    nodeName: r620-5
  status:
    conditions:
    - lastUpdateTime: "2022-03-31T15:02:54Z"
      status: "False"
      type: Mounted
    deviceStatus:
      capacity:
        physicalBlockSizeBytes: 512
        sizeBytes: 9.99116767232e+11
      details:
        busPath: pci-0000:02:00.0-scsi-0:2:0:0
        deviceType: disk
        driveType: HDD
        model: PERC_H710P
        ptUUID: a75c0d98-6ee8-c14a-b522-e741df2bd3c9
        serialNumber: 6b82a720d12ab50029786564093dd35e
        storageController: SCSI
        vendor: DELL
        wwn: 0x6b82a720d12ab50029786564093dd35e
      fileSystem:
        isReadOnly: true
        mountPoint: "null"
        type: "null"
      partitioned: true
    provisionPhase: Unprovisioned
    state: Active
kind: List
metadata:
  continue: "null"
  resourceVersion: "15380285"
`
