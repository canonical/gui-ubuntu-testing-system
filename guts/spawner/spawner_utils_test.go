package spawner

import (
	"os"
	"reflect"
	"strings"
	"testing"
	// "fmt"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
)

func TestFindHighestPrioUuid(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	actualUuid, err := FindHighestPrioUuid(Driver)
	utils.CheckError(err)
	expectedUuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
	if actualUuid != expectedUuid {
		t.Errorf("Unexpected uuid! Expected: %v\nActual: %v", expectedUuid, actualUuid)
	}
}

func TestFindRowIdForUuidInStateRequested(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	searchUuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
	rowId, err := FindRowIdForUuidInStateRequested(searchUuid, Driver)
	utils.CheckError(err)
	expectedRowId := 1
	if rowId != expectedRowId {
		t.Errorf("Unexpected row id! Expected: %v\nActual: %v", expectedRowId, rowId)
	}
}

func TestFindRowIdForUuidInStateRequestedFailure(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	searchUuid := "3017c20c-8d0b-42fc-986b-c558683e72fa"
	rowId, err := FindRowIdForUuidInStateRequested(searchUuid, Driver)
	if err == nil {
		t.Errorf("Finding row id for uuid %v should have failed but didn't", searchUuid)
	}
	expectedRowId := 0
	if rowId != expectedRowId {
		t.Errorf("Unexpected row id! Expected: %v\nActual: %v", expectedRowId, rowId)
	}
}

func TestSetTestStateTo(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 1
	err = SetTestStateTo(rowId, "running", Driver)
	utils.CheckError(err)
	err = SetTestStateTo(rowId, "requested", Driver)
	utils.CheckError(err)
}

func TestUpdateUpdatedAt(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 1
	err = UpdateUpdatedAt(rowId, Driver)
	utils.CheckError(err)
}

func TestSetVncAddressForId(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 22
	err = SetVncAddressForId(rowId, Driver)
	utils.CheckError(err)
}

func TestGetImageUrl(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 1
	imageUrl, err := GetImageUrl(rowId, Driver)
	utils.CheckError(err)
	expectedImageUrl := "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	if expectedImageUrl != imageUrl {
		t.Errorf("unexpected image url!\nExpected: %v\nActual: %v", expectedImageUrl, imageUrl)
	}
}

func TestGetTestRequirements(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	var expectedRequirements TestRequirements
	rowId := 1
	expectedRequirements.tpmRequired = false
	expectedRequirements.liveImage = true
	expectedRequirements.defaultDiskSizeGb = 40

	imageUrl := "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	actualRequirements, err := GetTestRequirements(rowId, imageUrl, Driver)
	utils.CheckError(err)
	if !reflect.DeepEqual(expectedRequirements, actualRequirements) {
		t.Errorf("Expected test requirements not the same as actual!\nExpected: %v\nActual: %v", expectedRequirements, actualRequirements)
	}
}

func TestGetTestRequirementsBadRowId(t *testing.T) {
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}
	rowId := 1007

	imageUrl := "https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso"
	_, err = GetTestRequirements(rowId, imageUrl, Driver)
	if err == nil {
		t.Errorf("Getting test requirements succeeded when it should have failed")
	}
}

func TestDownloadImageSuccess(t *testing.T) {
	spawnerCfg, err := ParseConfig("./guts-spawner.yaml")
	utils.CheckError(err)
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	imagePath, err := DownloadImage(imageUrl, spawnerCfg)
	utils.CheckError(err)
	expectedImagePath := "/srv/guts/images/questing-mini-iso-amd64.iso"
	err = os.Remove(expectedImagePath)
	utils.CheckError(err)
	if imagePath != expectedImagePath {
		t.Errorf("expected image path not the same as actual!\nExpected: %v\nActual: %v", expectedImagePath, imagePath)
	}
}

func TestDownloadImageAlreadyExists(t *testing.T) {
	spawnerCfg, err := ParseConfig("./guts-spawner.yaml")
	utils.CheckError(err)
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	imagePath, err := DownloadImage(imageUrl, spawnerCfg)
	utils.CheckError(err)
	expectedImagePath := "/srv/guts/images/questing-mini-iso-amd64.iso"
	if imagePath != expectedImagePath {
		t.Errorf("expected image path not the same as actual!\nExpected: %v\nActual: %v", expectedImagePath, imagePath)
	}
	imagePath, err = DownloadImage(imageUrl, spawnerCfg)
	utils.CheckError(err)
	err = os.Remove(expectedImagePath)
	utils.CheckError(err)
}

func TestIdenticalLocalAndRemoteShasum(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/resting-mini-iso-amd64.iso"
	imagePath := "/srv/guts/images/resting-mini-iso-amd64.iso"
	if IdenticalLocalAndRemoteShasum(imageUrl, imagePath) {
		t.Errorf("Retrieving shasum for %v should have failed but didn't!", imageUrl)
	}
}

func TestIdenticalLocalAndRemoteShasumNoLocal(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	imagePath := "/srv/guts/images/questing-mini-iso-amd64.iso"
	if IdenticalLocalAndRemoteShasum(imageUrl, imagePath) {
		t.Errorf("comparing shasum for %v should have failed but didn't!", imageUrl)
	}
}

func TestIdenticalLocalAndRemoteShasumWrongLocal(t *testing.T) {
	utils.ServeDirectory("/../../postgres/test-data/test-files/")
	imageUrl := "http://localhost:9999/questing-mini-iso-amd64.iso"
	imagePath := "/srv/guts/images/questing-mini-iso-amd64.iso"
	dummyBytes := []byte("Welcome to guts")
	err := os.WriteFile(imagePath, dummyBytes, 0644)
	utils.CheckError(err)
	identical := IdenticalLocalAndRemoteShasum(imageUrl, imagePath)
	os.Remove(imagePath)
	if identical {
		t.Errorf("Retrieving shasum for %v should have failed but didn't!", imageUrl)
	}
}

func TestGetQemuCmdLine(t *testing.T) {
	spawnerCfg, err := ParseConfig("./guts-spawner.yaml")
	utils.CheckError(err)

	imagePath := "/srv/guts/images/questing-mini-iso-amd64.iso"
	diskPath := "/srv/guts/disks/mydisk.qcow2"

	var dummyRequirements TestRequirements
	dummyRequirements.tpmRequired = false
	dummyRequirements.liveImage = true
	dummyRequirements.defaultDiskSizeGb = 40

	cmdLine := GetQemuCmdLine(imagePath, diskPath, dummyRequirements, spawnerCfg)
	expectedCmdLine := "qemu-system-x86_64 -m 4096 -smp 8 -enable-kvm -machine pc,accel=kvm -usbdevice tablet -vga virtio -vnc :0,share=ignore -boot once=d -cdrom /srv/guts/images/questing-mini-iso-amd64.iso -hda /srv/guts/disks/mydisk.qcow2"
	if strings.Join(cmdLine, " ") != expectedCmdLine {
		t.Errorf("unexpected qemu command line!\nExpected: %v\nActual: %v", expectedCmdLine, strings.Join(cmdLine, " "))
	}

	dummyRequirements.liveImage = false
	imagePath = "/srv/guts/images/questing-mini-iso-amd64-preinstalled.img"
	diskPath = "/srv/guts/images/questing-mini-iso-amd64-preinstalled.img"

	cmdLine = GetQemuCmdLine(imagePath, diskPath, dummyRequirements, spawnerCfg)
	expectedCmdLine = "qemu-system-x86_64 -m 4096 -smp 8 -enable-kvm -machine pc,accel=kvm -usbdevice tablet -vga virtio -vnc :0,share=ignore -drive format=raw,file=/srv/guts/images/questing-mini-iso-amd64-preinstalled.img"

	if strings.Join(cmdLine, " ") != expectedCmdLine {
		t.Errorf("unexpected qemu command line!\nExpected: %v\nActual: %v", expectedCmdLine, strings.Join(cmdLine, " "))
	}
}

func TestCreateQcowDisk(t *testing.T) {
	spawnerCfg, err := ParseConfig("./guts-spawner.yaml")
	utils.CheckError(err)

	var dummyRequirements TestRequirements
	dummyRequirements.tpmRequired = false
	dummyRequirements.liveImage = true
	dummyRequirements.defaultDiskSizeGb = 10

	thisUuid := "92e234a6-6172-44b5-a690-086fe3037c46"

	diskPath, _, err := CreateQcowDisk(dummyRequirements, thisUuid, spawnerCfg)
	utils.CheckError(err)

	err = utils.FileOrDirExists(diskPath)
	utils.CheckError(err)
	err = os.Remove(diskPath)
	utils.CheckError(err)
}

func TestGetTestState(t *testing.T) {
	id := 4
	Driver, err := database.TestDbDriver("guts_spawner", "guts_spawner")
	if database.SkipTestIfPostgresInactive(err) {
		t.Skip("Skipping test as postgresql service is not up")
	} else {
		utils.CheckError(err)
	}

	state, err := GetTestState(id, Driver)
	utils.CheckError(err)
	expectedState := "requested"
	if state != expectedState {
		t.Errorf("State for row %v should be %v but is %v", id, expectedState, state)
	}

	err = SetTestStateTo(id, "spawning", Driver)
	utils.CheckError(err)

	state, err = GetTestState(id, Driver)
	utils.CheckError(err)

	expectedState = "spawning"
	if state != expectedState {
		t.Errorf("State for row %v should be %v but is %v", id, expectedState, state)
	}

	err = SetTestStateTo(id, "requested", Driver)
	utils.CheckError(err)
}

func TestCreateCacheIfNotExists(t *testing.T) {
	spawnerCfg, err := ParseConfig("./guts-spawner.yaml")
	utils.CheckError(err)
	err = os.RemoveAll(spawnerCfg.General.ImageCachePath)
	utils.CheckError(err)
	err = CreateCacheIfNotExists(spawnerCfg)
	utils.CheckError(err)
	err = utils.FileOrDirExists(spawnerCfg.General.ImageCachePath)
	utils.CheckError(err)
	err = CreateCacheIfNotExists(spawnerCfg)
	utils.CheckError(err)
}
