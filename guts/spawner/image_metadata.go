package main

import (
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
  "regexp"
)

func GetRemoteShaSum(imageUrl string) (string, error) {
  // This is now looking okay to me.
  domainsFunctions := DomainsAndInterfaces()
  supportedDomains := make([]string, len(domainsFunctions))
  for idx, entry := range domainsFunctions {
    supportedDomains[idx] = entry
  }
  for _, domain := range supportedDomains {
		thisRegex := fmt.Sprintf(`(http|https):\/\/%v\/(.*)`, domain)
		match := regexp.MustCompile(thisRegex).MatchString(imageUrl)
    if match {
      shasum, err := domainsFunctions[domain]["shasum"](imageUrl)
      return shasum, err
    }
  }
  return "", fmt.Errorf("Couldn't acquire shasum of image at %v!", imageUrl)
}

func DomainsAndInterfaces() map[string]map[string]func(string)(string, error) {
  m := make(map[string]map[string]func(string)(string, error))
  cdimageMap := map[string]func(string)(string, error){"shasum": CdImageGetShasumOfImage}
  m["cdimage.ubuntu.com"] = cdimageMap
  m["localhost"] = cdimageMap
  return m
}

func CdImageGetShasumOfImage(imageUrl string) (string, error) {
  allShaSums, err := CdImageDownloadCheckSumFileForImage(imageUrl)
  if err != nil {
    return "", err
  }
  thisShaSum, err := CdImageParseShasumForImage(allShaSums, strings.Split(imageUrl, "/")[-1])
  return thisShaSum, err
}

func CdImageDownloadCheckSumFileForImage(imageUrl string) (string, error) {
  imageName := strings.Split(imageUrl, "/")[-1]
  baseDirUrl := strings.Replace(imageName, "")
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
  if err != nil {
    return "", err
  }
  return string(body), nil
}

func CdImageParseShasumForImage(allShaSums, imageName string) (string, error) {
  r, err := regexp.Compile(fmt.Sprintf(`([a-zA-Z0-9]{64}) \*%v`, imageName))
  if err != nil {
    return "", err
  }
  imageChkSum := r.FindString(allShaSums)
  if imageChkSum == "" {
    return "", fmt.Errorf("Couldn't find checksum at %v for image %v", shasumUrl, imageName)
  }
  return imageChkSum, nil
}

