package spawner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"guts.ubuntu.com/v2/utils"
	"io/ioutil"
  "log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func GetLocalShaSum(pathToFile string) (string, error) {
  log.Printf("getting shasum of %v", pathToFile)
	err := utils.FileOrDirExists(pathToFile)
	if err != nil {
    log.Printf("%v doesn't exist!", pathToFile)
		return "", err
	}
	dat, err := os.ReadFile(pathToFile)
	if err != nil { // coverage-ignore
		return "", err
	}
	h := sha256.New()
	h.Write(dat)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func GetRemoteShaSum(imageUrl string) (string, error) {
	// This is now looking okay to me.
	domainsFunctions := DomainsAndInterfaces()
	supportedDomains := make([]string, len(domainsFunctions))
	ctr := 0
	for idx, _ := range domainsFunctions {
		supportedDomains[ctr] = idx
		ctr = ctr + 1
	}
  log.Printf("supported domains: %v", supportedDomains)
	for _, domain := range supportedDomains {
    log.Printf("checking against domain: %v", domain)
		thisRegex := fmt.Sprintf(`(http|https):\/\/%v(.*)`, domain)
		match := regexp.MustCompile(thisRegex).MatchString(imageUrl)
		if match {
      log.Printf("domain supported!")
			shasum, err := domainsFunctions[domain]["shasum"](imageUrl)
			return shasum, err
		}
	}
  log.Printf("%v is not from a supported domain", imageUrl)
	return "", fmt.Errorf("Couldn't acquire shasum of image at %v", imageUrl)
}

func DomainsAndInterfaces() map[string]map[string]func(string) (string, error) {
	m := make(map[string]map[string]func(string) (string, error))
	cdimageMap := map[string]func(string) (string, error){"shasum": CdImageGetShasumOfImage}
	m["cdimage.ubuntu.com"] = cdimageMap
	m["releases.ubuntu.com"] = cdimageMap
	m["localhost"] = cdimageMap
	return m
}

func CdImageGetShasumOfImage(imageUrl string) (string, error) {
	allShaSums, err := CdImageDownloadCheckSumFileForImage(imageUrl)
	if err != nil {
		return "", err
	}
	imageName := utils.GetFileNameFromUrl(imageUrl)
	thisShaSum, err := CdImageParseShasumForImage(allShaSums, imageName)
	return thisShaSum, err
}

func CdImageDownloadCheckSumFileForImage(imageUrl string) (string, error) {
	imageName := utils.GetFileNameFromUrl(imageUrl)
	baseDirUrl := strings.Replace(imageUrl, imageName, "", -1)
	shasumUrl := fmt.Sprintf("%v%v", baseDirUrl, "SHA256SUMS")
	response, err := http.Get(shasumUrl)
	if err != nil { // coverage-ignore
		return "", err
	}
	if response.StatusCode != 200 {
		response.Body.Close()
		return "", fmt.Errorf("%v returned %v instead of 200", shasumUrl, response.StatusCode)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil { // coverage-ignore
		return "", err
	}
	return string(body), nil
}

func CdImageParseShasumForImage(allShaSums, imageName string) (string, error) {
	shasumRegex := fmt.Sprintf(`([a-zA-Z0-9]{64}) \*%v`, imageName)
	r, err := regexp.Compile(shasumRegex)
	if err != nil { // coverage-ignore
		return "", err
	}
	imageChkSum := r.FindStringSubmatch(allShaSums)
	if len(imageChkSum) < 2 || imageChkSum[1] == "" {
		return "", fmt.Errorf("No matches for regex %v in:\n%v", shasumRegex, allShaSums)
	}
	return imageChkSum[1], nil
}
