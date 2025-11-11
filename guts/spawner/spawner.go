package spawner

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"guts.ubuntu.com/v2/database"
	"guts.ubuntu.com/v2/utils"
	"io"
	"net/http"
	"os"
	"os/exec"
  "log"
	"slices"
	"strings"
	"time"
)

type TestRequirements struct {
	tpmRequired       bool
	liveImage         bool
	defaultDiskSizeGb int
}

func CreateCacheIfNotExists(SpawnerCfg GutsSpawnerConfig) error {
	err := utils.FileOrDirExists(SpawnerCfg.General.ImageCachePath)
	if err != nil {
		err = os.MkdirAll(SpawnerCfg.General.ImageCachePath, 0755)
	}
	return err
}

func FindHighestPrioUuid(Driver database.DbDriver) (string, error) {
	var uuid string
	jobQuery := `SELECT tests.uuid FROM tests JOIN jobs ON jobs.uuid=tests.uuid WHERE state='requested' ORDER BY priority DESC LIMIT 1`
  log.Printf("running query: %v", jobQuery)
	row, err := Driver.RunQueryRow(jobQuery)
	if err != nil { // coverage-ignore
		return uuid, err
	}
	err = row.Scan(
		&uuid,
	)
  log.Printf("query complete")
	if err != nil { // coverage-ignore
    log.Printf(err.Error())
		if err == sql.ErrNoRows {
      log.Printf("found no uuids")
			return "", nil
		} else {
			return "", err
		}
	}
	return uuid, nil
}

func FindRowIdForUuidInStateRequested(uuid string, Driver database.DbDriver) (int, error) {
	var id int
	idQuery := fmt.Sprintf(`SELECT id FROM tests WHERE uuid='%v' AND state='requested' LIMIT 1`, uuid)
	row, err := Driver.RunQueryRow(idQuery)
	if err != nil { // coverage-ignore
		return id, err
	}
	err = row.Scan(
		&id,
	)
	if err != nil {
		return id, err
	}
	return id, nil
}

func SetVncAddressForId(id int, Driver database.DbDriver) error {
	addressString := fmt.Sprintf("%v:%v", VncHost, VncPort)
	updateQuery := fmt.Sprintf(`UPDATE tests SET vnc_address='%v' WHERE id='%v'`, addressString, id)
	err := Driver.UpdateRow(updateQuery)
	return err
}

func GetImageUrl(id int, Driver database.DbDriver) (string, error) {
	var imageUrl string
	imageUrlQuery := fmt.Sprintf(`SELECT image_url FROM jobs JOIN tests ON jobs.uuid=tests.uuid WHERE id=%v`, id)
	row, err := Driver.RunQueryRow(imageUrlQuery)
	if err != nil { // coverage-ignore
		return "", err
	}
	err = row.Scan(
		&imageUrl,
	)
	return imageUrl, err
}

func GetTestRequirements(id int, imageUrl string, Driver database.DbDriver) (TestRequirements, error) {
	var requirements TestRequirements
	tpmQuery := fmt.Sprintf(`SELECT tpm FROM tests WHERE id=%v`, id)
	row, err := Driver.RunQueryRow(tpmQuery)
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

func DownloadImage(imageUrl string, SpawnerCfg GutsSpawnerConfig) (string, error) {
	// takes url, downloads image to cache, returns image path
	// - parse file/image name
	// - if image already exists in cache:
	//    - check image domain - have a list of domains we know about shasum files for
	//      - check to see if imageUrl is accompanied with a SHA256sums file
	//      - check shasum of local image and remote image
	//      - if same, return image path, and continue without downloading
	// - download image to .new, then move so it's atomic
	// - return image path
  log.Printf("downloading image at %v", imageUrl)
	splitUrl := strings.Split(imageUrl, "/")
	imageName := splitUrl[len(splitUrl)-1]
  log.Printf("image name: %v", imageName)
	imagePath := fmt.Sprintf("%v%v", SpawnerCfg.General.ImageCachePath, imageName)
  log.Printf("image path: %v", imagePath)

  log.Printf("checking if image already exists...")
	err := utils.FileOrDirExists(imagePath)
	if err == nil {
		if IdenticalLocalAndRemoteShasum(imageUrl, imagePath) {
      log.Printf("image already cached!")
			return imagePath, nil
		}
	}

  log.Printf("image doesn't exist, downloading...")
	err = AtomicDownloadImageToPath(imageUrl, imagePath)
	if err != nil { // coverage-ignore
    log.Printf(err.Error())
		return "", err
	}

  log.Printf("%v downloaded to %v", imageName, imagePath)

	return imagePath, nil
}

func IdenticalLocalAndRemoteShasum(imageUrl, imagePath string) bool {
  log.Printf("checking local vs remote shasum...")
	remoteShasum, err := GetRemoteShaSum(imageUrl)
	if err != nil {
    log.Printf(err.Error())
		return false
	}
  log.Printf("remote shasum: %v", remoteShasum)
	localShasum, err := GetLocalShaSum(imagePath)
	if err != nil {
		return false
	}
  log.Printf("local sha sum: %v", localShasum)
	if localShasum == remoteShasum {
		return true
	}
	return false
}

func AtomicDownloadImageToPath(imageUrl, imagePath string) error {
  log.Printf("preparing atomic image download from url %v...", imageUrl)
	newFile := fmt.Sprintf("%v.new", imagePath)
	resp, err := http.Get(imageUrl)
	if err != nil { // coverage-ignore
		return err
	}
	defer utils.DeferredErrCheck(resp.Body.Close)
  // log.Printf("image downloaded!")
	// b, err := io.ReadAll(resp.Body)
	// if err != nil { // coverage-ignore
  //   log.Printf(err.Error())
	// 	return err
	// }
	// err = resp.Body.Close()
	// if err != nil { // coverage-ignore
  //   log.Printf(err.Error())
	// 	return err
	// }

  log.Printf("creating parent directories if necessary...")
	splitPath := strings.Split(imagePath, "/")
	fileName := splitPath[len(splitPath)-1]
	directoryNoFn := strings.Replace(imagePath, fileName, "", -1)
	err = os.MkdirAll(directoryNoFn, 0755)
	if err != nil { // coverage-ignore
		return err
	}

  log.Printf("writing to disk...")
  out, err := os.Create(newFile)

  _, err = io.Copy(out, resp.Body)
	if err != nil { // coverage-ignore
		return err
	}

  err = out.Close()
	if err != nil { // coverage-ignore
		return err
	}

	// err = os.WriteFile(newFile, b, 0644)
	// if err != nil { // coverage-ignore
	// 	return err
	// }

	err = os.Rename(newFile, imagePath)
  if err != nil { // coverage-ignore
    return err
  }
  log.Printf("image downloaded to %v", imagePath)
	return nil
}

func GetQemuCmdLine(imagePath, DiskPath string, req TestRequirements, SpawnerCfg GutsSpawnerConfig) []string {
	executable := "qemu-system-x86_64"
	defaultFlags := fmt.Sprintf("-m %v -smp %v -enable-kvm -machine pc,accel=kvm -usbdevice tablet -vga virtio -vnc :%v,share=ignore", SpawnerCfg.Virtualisation.Memory, SpawnerCfg.Virtualisation.Cores, VncPort - 5900)
	var imageArgs string
	if req.liveImage {
		imageArgs = fmt.Sprintf("-boot once=d -cdrom %v -hda %v", imagePath, DiskPath)
	} else {
		imageArgs = fmt.Sprintf("-drive format=raw,file=%v", imagePath)
	}
	cmdLineStr := fmt.Sprintf("%v %v %v", executable, defaultFlags, imageArgs)
	return strings.Split(cmdLineStr, " ")
}

func CreateQcowDisk(requirements TestRequirements, Uuid string, SpawnerCfg GutsSpawnerConfig) (string, string, error) {
	secondaryUuid := uuid.New().String()
	diskName := fmt.Sprintf("%v-%v.qcow2", Uuid, secondaryUuid)
	DiskPath := fmt.Sprintf("%v/%v", SpawnerCfg.General.ImageCachePath, diskName)
	qcowCreateCmd := exec.Command(
		"qemu-img",
		"create",
		"-f",
		"qcow2",
		DiskPath,
		fmt.Sprintf("%vG", requirements.defaultDiskSizeGb),
	)
	if err := qcowCreateCmd.Run(); err != nil { // coverage-ignore
		return "", "", err
	}
	return DiskPath, diskName, nil
}

func SpawnVm(cmdLine []string) (*exec.Cmd, error) { // coverage-ignore
  log.Printf("command line:\n%v", cmdLine)
	qemuVmCreateCmd := exec.Command(cmdLine[0])
	qemuVmCreateCmd.Args = cmdLine
  log.Println(qemuVmCreateCmd)
	err := qemuVmCreateCmd.Start()
	return qemuVmCreateCmd, err
}

func GetTestState(id int, Driver database.DbDriver) (string, error) {
	var state string
	stateQuery := fmt.Sprintf(`SELECT state FROM tests WHERE id=%v`, id)
	row, err := Driver.RunQueryRow(stateQuery)
	if err != nil { // coverage-ignore
		return "", err
	}
	err = row.Scan(
		&state,
	)
	if err != nil { // coverage-ignore
		return "", err
	}
	return state, nil
}

func SpawnerLoop(Driver database.DbDriver, SpawnerCfg GutsSpawnerConfig) error { // coverage-ignore
  log.Printf("starting spawner loop...")
	// Find the requested job with the highest priority
  log.Printf("finding highest prio uuid...")
	uuid, err := FindHighestPrioUuid(Driver)
	// Perform a standard error check
	if err != nil {
		return err
	}
	// if the uuid is empty, there are no tests waiting
	if uuid == "" {
		return nil
	}
  log.Printf("found uuid: %v", uuid)
	// Get the id of the individual test
  log.Printf("finding row ids for %v in state 'requested'", uuid)
	id, err := FindRowIdForUuidInStateRequested(uuid, Driver)
	if err != nil {
		return err
	}
  log.Printf("found id: %v", id)
	// Set the test state to spawning to indicate we are spawning the VM
  log.Printf("setting state to 'spawning'")
	err = Driver.SetTestStateTo(id, "spawning")
	if err != nil {
		return err
	}
  log.Printf("updating heartbeat timestamp")
	// Update the heartbeat timestamp
	err = database.UpdateUpdatedAt(id, Driver)
	if err != nil {
		return err
	}
	// Set the vncaddress field to state where the test is running
  log.Printf("setting vnc address")
	err = SetVncAddressForId(id, Driver)
	if err != nil {
		return err
	}
	// Update the heartbeat timestamp
  log.Printf("updating heartbeat timestamp")
	err = database.UpdateUpdatedAt(id, Driver)
	if err != nil {
		return err
	}
	// Get the url for the image for the test
  log.Printf("getting image url")
	imageUrl, err := GetImageUrl(id, Driver)
	if err != nil {
		return err
	}
  log.Printf("got url: %v", imageUrl)
	// Parse test requirements from the db
  log.Printf("getting test requirements...")
	requirements, err := GetTestRequirements(id, imageUrl, Driver)
	if err != nil {
		return err
	}
  log.Printf("got requirements: %v", requirements)
	// Download the image to a local path
  log.Printf("downloading image...")
	imagePath, err := DownloadImage(imageUrl, SpawnerCfg)
	if err != nil {
		return err
	}
  log.Printf("image downloaded to: %v", imagePath)
	// the diskpath and image path are the same if an image is pre-installed
	// otherwise they differ
	DiskPath := imagePath
	if requirements.liveImage {
		// Create the qcow2 disk for qemu to use as storage for the test VM
		DiskPath, _, err = CreateQcowDisk(requirements, uuid, SpawnerCfg)
		if err != nil {
			return err
		}
	}

	// Get the appropriate qemu command line given the test requirements
	qemuCmdLine := GetQemuCmdLine(imagePath, DiskPath, requirements, SpawnerCfg)

  log.Printf("got qemu cmd line: %v", qemuCmdLine)

	// spawn the qemu VM
  log.Printf("spawning VM...")
	vmProcess, err := SpawnVm(qemuCmdLine)
	if err != nil {
		return err
	}
	// set state to spawned
  log.Printf("setting state to 'spawned'...")
	err = Driver.SetTestStateTo(id, "spawned")
	if err != nil {
		return err
	}
	// update the heartbeat ts
  log.Printf("updating heartbeat timestamp")
	err = database.UpdateUpdatedAt(id, Driver)
	if err != nil {
		return err
	}
  log.Printf("declaring finish states")
	// declare the states the spawner considers finished
	finishStates := []string{"pass", "fail", "requested"}
	finished := false

	// define how often we check the test state
  log.Printf("defining heartbeat duration")
	heartbeatDuration := time.Second * 5
	// wait for either the qemu process to die or the test to finish
	for !vmProcess.ProcessState.Exited() || finished {
		// get the test state
    log.Printf("getting test state...")
		state, err := GetTestState(id, Driver)
		if err != nil {
			return err
		}
    log.Printf("got test state: %v", state)
		// see if it's in a "finished" state
		if slices.Contains(finishStates, state) {
      log.Printf("job finished!")
			finished = true
		}
		// Only update the heartbeat timestamp
		// when the runner is not already running the test
		if state != "running" {
			// update the heartbeat ts
      log.Printf("updating heartbeat timestamp")
			err = database.UpdateUpdatedAt(id, Driver)
			if err != nil {
				return err
			}
		}
		// wait
		time.Sleep(heartbeatDuration)
	}
	if !finished {
		// we reach this if the VM dies unexpectedly, set the state back to requested
    log.Printf("setting state back to 'requested'")
		err = Driver.SetTestStateTo(id, "requested")
		return err
	}
	// kill the VM
  log.Printf("killing vm")
	err = vmProcess.Process.Kill()
	if err != nil {
		return err
	}
  log.Printf("spawner loop over")
	// remove the disk
	err = os.Remove(DiskPath)
	return err
}
