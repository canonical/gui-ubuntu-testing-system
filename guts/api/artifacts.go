package main

import (
  "fmt"
  "database/sql"
  "errors"
  "io"
  "compress/gzip"
  "archive/tar"
  "net/http"
  "path/filepath"
  "bytes"
  "strings"
  "reflect"
  "net/url"
  "time"
  "os"
  "sort"
)


func CollateArtifacts(uuidToFind string) ([]byte, error) { // coverage-ignore
  var gzippedTarBytes []byte
  uuidCacheDir := fmt.Sprintf("%v/%v", GutsCfg.Tarball.TarBallCachePath, uuid)
  cachedTarFile := fmt.Sprintf("%v/results.tar.gz", uuidCacheDir)
  cachedLastDownloadedFile := fmt.Sprintf("%v/%v.last_downloaded", uuidCacheDir, uuid)

  if AllFilesExist(uuidCacheDir, cachedTarFile, cachedLastDownloadedFile) {
    dat, err := os.ReadFile(cachedTarFile)
    err := RefreshLastDownloadedFile(cachedLastDownloadedFile)
    if err != nil {
      return err
    }
    return dat, err
  }

  urls, err := FindArtifactUrlsByUuid(uuidToFind)
  if err != nil {
    return gzippedTarBytes, err
  }

  directoryNames, err := CreateOutputDirectoriesFromUrls(urls)
  if err != nil {
    return gzippedTarBytes, err
  }

  allTarFilesBytes, err := DownloadTarFiles(urls)
  if err != nil {
    return gzippedTarBytes, err
  }

  allFilesInTarBytes, err := TarUpFilesInGivenDirectories(directoryNames, allTarFilesBytes)
  if err != nil {
    return gzippedTarBytes, err
  }

  gzippedTarBytes, err = GzipTarArchiveBytes(allFilesInTarBytes)
  if err != nil {
    return gzippedTarBytes, err
  }

  err = WriteTarballToCache(gzippedTarBytes, uuidToFind, uuidCacheDir, cachedTarFile, cachedLastDownloadedFile)
  if err != nil {
    return gzippedTarBytes, err
  }

  err = CacheRetentionPolicy(uuidCacheDir)
  if err != nil {
    return gzippedTarBytes, err
  }

  return gzippedTarBytes, nil
}

func CacheRetentionPolicy(uuidCacheDir string) error {

  entries, err := os.ReadDir(uuidCacheDir)
  now := time.Now().Unix()
  if err != nil {
    return err
  }

  uuidEpochMap := make(map[int64]string)
  sliceOfStamps := make([]int64, len(entries))

  for idx, e := range(entries) {
    entryLastUpdatedFile := fmt.Sprintf("%v/%v.last_downloaded", e.Name(), e.Name())
    dat, err := os.ReadFile(entryLastUpdatedFile)
    lastUpdated := int64(string(dat))
    uuidEpochMap[lastUpdated] = e.Name()
    sliceOfStamps[idx] = lastUpdated
  }

  sort.Ints(sliceOfStamps)

  for _, lastUpdated := range(sliceOfStamps) {
    thisSpecificDir := fmt.Sprintf("%v/%v", uuidCacheDir, uuidEpochMap[lastUpdated])
    dirSize, err := GetDirSize(uuidCacheDir)
    if err != nil {
      return err
    }
    if dirSize < GutsCfg.Tarball.TarBallCacheReductionThreshold {
      return nil
    }
    err = os.RemoveAll(thisSpecificDir)
    if err != nil {
      return err
    }
  }
  return nil
}

func GetDirSize(directory string) (int64, error) {
  var directorySize int64
  dir, err := os.Open(directory)
  if err != nil {
    return directorySize, err
  }
  defer DeferredErrCheck(dir.Close)
  fileInfo, err := dir.Stat()
  if err != nil {
    return directorySize, err
  }
  directorySize = fileInfo.Size()
  return directorySize, err
}

func RefreshLastDownloadedFile(cachedLastDownloadedFile string) error {
  now := time.Now().Unix()
  lastDownloaded := []byte(string(now))
  err := AtomicWrite(lastDownloaded, cachedLastDownloadedFile)
  if err != nil {
    return err
  }
  return nil
}

func WriteTarballToCache(tarBall []byte, uuid string, uuidCacheDir string, cachedTarFile string, cachedLastDownloadedFile string) error {
  if AllFilesExist(uuidCacheDir, cachedTarFile, cachedLastDownloadedFile) {
    err := RefreshLastDownloadedFile(cachedLastDownloadedFile)
    if err != nil {
      return err
    }
    return nil
  }

  err := os.RemoveAll(uuidCacheDir)
  if err != nil {
    return err
  }

  err = os.Mkdir(uuidCacheDir, 0755)
  if err != nil {
    return err
  }

  err = AtomicWrite(tarBall, cachedTarFile)
  if err != nil {
    return err
  }

  err = RefreshLastDownloadedFile(cachedLastDownloadedFile)
  if err != nil {
    return err
  }

  return nil
}

func GzipTarArchiveBytes(tarArchive []byte) ([]byte, error) {
  var buf bytes.Buffer
  var returnBytes []byte
  gz := gzip.NewWriter(&buf)
  _, err := gz.Write(tarArchive)
  if err != nil { // coverage-ignore
    return returnBytes, err
  }
  err = gz.Close()
  if err != nil { // coverage-ignore
    return returnBytes, err
  }
  returnBytes = buf.Bytes()
  return returnBytes, nil
}

func IsValidUrl(urlToTest string) bool {
  u, err := url.Parse(urlToTest)
  if err != nil {
    return false
  }
  return u.Scheme != "" && u.Host != ""
}

func CreateOutputDirectoriesFromUrls(urls []string) ([]string, error) {
  var listOfFilenames []string
  if len(urls) == 0 {
    return listOfFilenames, errors.New("list of urls is empty! can't create output directories")
  }
  for _, thisUrl := range(urls) {
    if !IsValidUrl(thisUrl) {
      return listOfFilenames, fmt.Errorf("%v is not a valid url", thisUrl)
    }
    splitUrl := strings.Split(thisUrl, `/`)
    filename := splitUrl[len(splitUrl)-1]
    directory := strings.Split(filename, `.`)[0]
    listOfFilenames = append(listOfFilenames, directory)
  }
  return listOfFilenames, nil
}

func DownloadTarFiles(tarfileUrls []string) ([]map[string][]byte, error) {
  var allFilesMaps []map[string][]byte

  for _, url := range(tarfileUrls) {
    var b []byte
    b, err := DownloadFile(url)
    if err != nil {
      return allFilesMaps, err
    }
    tarFiles, err := ExtractTarfiles(b)
    if err != nil {
      return allFilesMaps, err
    }
    allFilesMaps = append(allFilesMaps, tarFiles)
  }
  return allFilesMaps, nil
}

func TarUpFilesInGivenDirectories(dirsForFiles []string, inputFiles []map[string][]byte) ([]byte, error) {
  var returnBytes []byte
  if !reflect.DeepEqual(len(dirsForFiles), len(inputFiles)) {
    return returnBytes, errors.New("length of variables doesn't add up")
  }

  // initialise tar writer
  tarBuffer := &bytes.Buffer{}
  tarWriter := tar.NewWriter(tarBuffer)

  for idx, entry := range(dirsForFiles) {
    tarFiles := inputFiles[idx]
    for fileName, fileBytes := range(tarFiles) {
      hdr := &tar.Header{
        Name: entry + "/" + fileName,
        Mode: 0644,
        Size: int64(len(fileBytes)),
      }
      // write tar header
      err := tarWriter.WriteHeader(hdr)
      if err != nil { // coverage-ignore
        return returnBytes, err
      }
      // write tar data
      _, err = tarWriter.Write(fileBytes)
      if err != nil { // coverage-ignore
        return returnBytes, err
      }
    }
  }

  // close tar writer
  err := tarWriter.Close()
  if err != nil { // coverage-ignore
    return returnBytes, err
  }

  returnBytes = tarBuffer.Bytes()

  return returnBytes, nil
}

func FindArtifactUrlsByUuid(uuidToFind string) ([]string, error) {
  var result_urls []string
  var params = [...]string{"results_url"}
  rows, err := Driver.Query("tests", uuidToFind, params)
  if err != nil { // coverage-ignore
    return result_urls, err 
  }
  defer DeferredErrCheck(rows.Close)
  for rows.Next() {
    var result_url string
    err := rows.Scan(&result_url)
    if err != nil { // coverage-ignore
      return result_urls, err 
    }
    result_urls = append(result_urls, result_url)
  }

  if len(result_urls) == 0 {
    return result_urls, UuidNotFoundError{uuid: uuidToFind}
  }

  return result_urls, nil
}

func DownloadFile(url string) ([]byte, error) {
  var b []byte
  resp, err := http.Get(url)
  if err != nil {
    return b, err
  }
  defer DeferredErrCheck(resp.Body.Close)
  b, err = io.ReadAll(resp.Body)
  if err != nil { // coverage-ignore
    return b, err
  }
  if len(b) == 0 {
    return b, fmt.Errorf("file at %v is empty", url)
  }
  return b, nil
}

func ExtractTarfiles(tarFileBytes []byte) (map[string][]byte, error) {
  returnData := make(map[string][]byte)

  r := bytes.NewReader(tarFileBytes)
  gzipReader, err := gzip.NewReader(r)
  if err != nil {
    return returnData, err
  }
  defer DeferredErrCheck(gzipReader.Close)
  tarReader := tar.NewReader(gzipReader)
  var maxBytes int64 = 4000000
  for {
    header, err := tarReader.Next()
    if err == io.EOF {
      break
    }
    if err != nil { // coverage-ignore
      return returnData, err
    }
    limFileReader := io.LimitReader(tarReader, maxBytes)
    data, err := io.ReadAll(limFileReader)
    if err != nil {
      return returnData, err
    }
    if len(data) != 0 {
      returnData[filepath.Base(header.Name)] = data
    }
  }
  return returnData, nil
}

