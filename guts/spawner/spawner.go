package main

import (
	"database/sql"
	"fmt"
  "time"
	"github.com/google/uuid"
  "strings"
  "io"
  "os"
  "os/exec"
  "net/http"
)

type TestRequirements struct {
  tpmRequired bool
  liveImage bool
  defaultDiskSizeGb int
}

func CreateCacheIfNotExists() error {
  if !FileOrDirExists(SpawnerCfg.General.ImageCachePath) {
    err = os.MkdirAll(SpawnerCfg.General.ImageCachePath, 0755)
    if err != nil { // coverage-ignore
      return err
    }
  }
  return nil
}

func FindHighestPrioUuid() (string, error) {
	var uuid string
	jobQuery := `SELECT tests.uuid FROM tests JOIN jobs ON jobs.uuid=tests.uuid WHERE state='requested' ORDER BY priority DESC LIMIT 1`
	fmt.Println("Running:")
	fmt.Println(jobQuery)
	row, err := Driver.RunQueryRow(jobQuery)
	if err != nil { // coverage-ignore
		return uuid, err
	}
	err = row.Scan(
		&uuid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		} else {
			return "", err
		}
	}
	return uuid, nil
}

func FindRowIdForUuidInStateRequested(uuid string) (int, error) {
  var id int
  idQuery := fmt.Sprintf(`SELECT id FROM tests WHERE uuid='%v' AND state='requested' LIMIT 1`, uuid)
  row, err := Driver.RunQueryRow(idQuery)
	if err != nil { // coverage-ignore
		return uuid, err
	}
  err = row.Scan(
    &id,
  )
	if err != nil {
    return id, err
	}
	return id, nil
}

func SetTestStateTo(id int, state string) error {
	stateUpdateQuery := fmt.Sprintf(`UPDATE tests SET state='%v' WHERE id='%v'`, state, id)
	err := Driver.UpdateRow(stateUpdateQuery)
	return err
}

func UpdateUpdatedAt(id int) error {
  err := Driver.TestsUpdateUpdatedAt(id)
}

func SetVncAddressForId(id int) error {
  addressString := fmt.Sprintf("%v:%v", VncHost, VncPort)
  updateQuery := fmt.Sprintf(`UPDATE tests SET vnc_address='%v' WHERE id='%v'`, addressString, id)
  err := Driver.UpdateRow(updateQuery)
  return err
}

func GetImageUrl(id int) (string, error) {
  var imageUrl string
  imageUrlQuery := fmt.Sprintf(`SELECT image_url FROM jobs JOIN tests ON jobs.uuid=tests.uuid WHERE id=%v`, id)
  err := Driver.RunQueryRow(tpmQuery)
	if err != nil { // coverage-ignore
		return "", err
	}
  err = row.Scan(
    &imageUrl,
  )
  return imageUrl, err
}

func GetTestRequirements(id int, imageUrl string) (TestRequirements, error) {
  var requirements TestRequirements
  tpmQuery := fmt.Sprintf(`SELECT tpm FROM tests WHERE id=%v`, id)
  err := Driver.RunQueryRow(tpmQuery)
	if err != nil { // coverage-ignore
		return requirements, err
	}
  err = row.Scan(
    &requirements.tpmRequired,
  )
	if err != nil {
    return requirements, err
	}

  requirements.liveImage = strings.HasSuffix(imageUrl, ".iso")
  requirements.defaultDiskSizeGb = 40

	return requirements, nil
}

func DownloadImage(imageUrl string) (string, error) {
  // takes url, downloads image to cache, returns image path
  // - parse file/image name
  // - if image already exists in cache:
  //    - check image domain - have a list of domains we know about shasum files for
  //      - check to see if imageUrl is accompanied with a SHA256sums file
  //      - check shasum of local image and remote image
  //      - if same, return image path, and continue without downloading
  // - download image to .new, then move so it's atomic
  // - return image path
  imageName := strings.Split(imageUrl, "/")[-1]
  imagePath := fmt.Sprintf("%v/%v", SpawnerCfg.General.ImageCachePath, imageName)

  if FileOrDirExists(imagePath) {
    remoteShasum, err := GetRemoteShaSum(imageUrl)
    if err != nil {
      return "", err
    }
    localShasum, err := GetLocalShaSum(imagePath)
    if err != nil {
      return "", err
    }
    if localShasum == remoteShasum {
      return imagePath, nil
    }
  }

  err = AtomicDownloadImageToPath(imageUrl, imagePath)
  if err != nil {
    return "", err
  }

  return imagePath, nil
}

func AtomicDownloadImageToPath(imageUrl, imagePath string) error {
  newFile := fmt.Sprintf("%v.new", imagePath)
  resp, err := http.Get(imageUrl)
  if err != nil {
    return err
  }
  defer DeferredErrCheck(resp.Body.Close)
  b, err := io.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    return err
  }
  err = os.WriteFile(newFile, b, 0644)
  if err != nil {
    return err
  }
  err = os.Rename(newFile, imagePath)
  return err
}

func GetQemuCmdLine(imagePath, diskPath string, req TestRequirements) []string {
  executable := "qemu-system-x86_64"
  defaultFlags := fmt.Sprintf("-m %v -smp %v -enable-kvm -machine pc,accel=kvm -usbdevice tablet -vga virtio -vnc :%v,share=ignore", SpawnerCfg.Virtualisation.Memory, SpawnerCfg.Virtualisation.Cores, VncPort)
  var imageArgs string
  if req.liveImage {
    imageArgs = fmt.Sprintf("-boot once=d -cdrom %v -hda %v", imagePath, diskPath)
  } else {
    imageArgs = fmt.Sprintf("-drive format=raw,file=%v", imagePath)
  }
  cmdLineStr := fmt.Sprintf("%v %v %v", executable, defaultFlags, imageArgs)
  return strings.Split(cmdLineStr, " ")
}

func CreateQcowDisk(requirements TestRequirements, uuid string) (string, string, error) {
  secondaryUuid := uuid.New().String()
  diskName := fmt.Sprintf("%v-%v.qcow2", uuid, secondaryUuid)
  diskPath := fmt.Sprintf("%v/%v", SpawnerCfg.General.ImageCachePath, diskName)
  qcowCreateCmd := exec.Command(
    "qemu-img",
    "create",
    "-f",
    "qcow2",
    diskPath,
    fmt.Sprintf("%vG", requirements.defaultDiskSizeGb),
  )
	if err := qcowCreateCmd.Run(); err != nil {
    return "", "", err
  return diskPath, diskName, nil
}

func SpawnVm(cmdLine []string) (Cmd, error) {
  // use pkg.go.dev/os#ProcessState
  // Cmd.ProcessState
  // Return the Cmd
  // and then outside this function, use the Cmd to handle the process
  qemuVmCreateCmd := exec.Command(
    cmdLine,
  )
  err := qemuVmCreateCmd.Start()
  return qemuVmCreateCmd, err
}

func GetTestState(id int) (string, error) {
  var state string
  stateQuery := fmt.Sprintf(`SELECT state FROM tests WHERE id=%v`, id)
  err := Driver.RunQueryRow(stateQuery)
  if err != nil {
    return "", err
  }
  err = row.Scan(
    &state,
  )
  if err != nil {
    return "", err
  }
  return state, nil
}

func SpawnerLoop() error {
  // Find the requested job with the highest priority
  uuid, err := FindHighestPrioUuid()
  // Perform a standard error check
  if err != nil {
    return err
  }
  // if the uuid is empty, there are no tests waiting
  if uuid == "" {
    return nil
  }
  // Get the id of the individual test
  id, err := FindRowIdForUuidInStateRequested(uuid)
  if err != nil {
    return err
  }
  // Set the test state to spawning to indicate we are spawning the VM
  err = SetTestStateTo(id, "spawning")
  if err != nil {
    return err
  }
  // Update the heartbeat timestamp
  err = UpdateUpdatedAt(id)
  if err != nil {
    return err
  }
  // Set the vncaddress field to state where the test is running
  err = SetVncAddressForId(id)
  if err != nil {
    return err
  }
  // Update the heartbeat timestamp
  err = UpdateUpdatedAt(id)
  if err != nil {
    return err
  }
  // Get the url for the image for the test
  imageUrl, err := GetImageUrl(id)
  if err != nil {
    return err
  }
  // Parse test requirements from the db
  requirements, err := GetTestRequirements(id, imageUrl)
  if err != nil {
    return err
  }
  // Download the image to a local path
  imagePath, err := DownloadImage(imageUrl)
  if err != nil {
    return err
  }
  // the diskpath and image path are the same if an image is pre-installed
  // otherwise they differ
  diskPath := imagePath
  if requirements.liveImage {
    // Create the qcow2 disk for qemu to use as storage for the test VM
    diskPath, diskName, err := CreateQcowDisk(requirements, uuid)
    if err != nil {
      return err
    }
  }

  // Get the appropriate qemu command line given the test requirements
  qemuCmdLine := GetQemuCmdLine(imagePath, diskPath, requirements)

  // spawn the qemu VM
  vmProcess, err := SpawnVm(qemuCmdLine)
  if err != nil {
    return err
  }
  // set state to spawned
  err = SetTestStateTo(id, "spawned")
  if err != nil {
    return err
  }
  // update the heartbeat ts
  err = UpdateUpdatedAt(id)
  if err != nil {
    return err
  }
  // declare the states the spawner considers finished
  finishStates := []string{"pass", "fail", "requested"}
  finished := false

  // define how often we check the test state
  heartbeatDuration := time.Second * 5
  // wait for either the qemu process to die or the test to finish
  for !vmProcess.ProcessState.Exited() || finished {
    // get the test state
    state := GetTestState(id)
    // see if it's in a "finished" state
    if slices.Contains(finishStates, state) {
      finished = true
    }
    // update the heartbeat ts
    err = UpdateUpdatedAt(id)
    if err != nil {
      return err
    }
    // wait
    time.Sleep(heartbeatDuration)
  }
  if !finished {
    // we reach this if the VM dies unexpectedly, set the state back to requested
    err = SetTestStateTo(id, "requested")
    return err
  }
  // kill the VM
  err = vmProcess.Process.Kill()
  return err
}

