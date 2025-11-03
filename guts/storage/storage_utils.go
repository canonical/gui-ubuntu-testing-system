package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ncw/swift/v2"
	"guts.ubuntu.com/v2/utils"
	"log"
	"os"
	"strings"
	"time"
)

func GetStorageBackend(strgCfg map[string]string) (StorageBackend, error) {
	var local LocalBackend
	provider := strgCfg["provider"]
	if provider == "swift" {
		var swift SwiftBackend
		swiftBknd, err := swift.Connect(&strgCfg)
		swift, ok := swiftBknd.(SwiftBackend)
		if !ok { // coverage-ignore
			return swift, fmt.Errorf("Something went wrong acquiring the %v backend", provider)
		}
		if err != nil {
			return swift, err
		}
		return swift, nil // coverage-ignore
	} else if provider == "local" {
		var local LocalBackend
		localBknd, err := local.Connect(&strgCfg)
		local, ok := localBknd.(LocalBackend)
		if !ok { // coverage-ignore
			return local, fmt.Errorf("Something went wrong acquiring the %v backend", provider)
		}
		if err != nil {
			return local, err
		}
		return local, nil
	} else {
		return local, fmt.Errorf("%v isn't a supported storage backend.", provider)
	}
}

type StorageBackend interface {
	Connect(strgCfg *map[string]string) (StorageBackend, error)
	Upload(namespace, remotePath string, data []byte) (string, error)
	RemoveObjectsOlderThan(duration time.Duration) ([]string, error)
}

type LocalBackend struct {
	Cfg LocalBackendCfg
}

type LocalBackendCfg struct {
	ObjectPath string `json:"object_path"`
	ObjectPort string `json:"object_port"`
	ObjectHost string `json:"object_host"`
}

func (l LocalBackendCfg) AssertConfigured() error {
	if l.ObjectPath == "" {
		return fmt.Errorf("object_path not set")
	}
	if l.ObjectPort == "" {
		return fmt.Errorf("object_port not set")
	}
	if l.ObjectHost == "" {
		return fmt.Errorf("object_host not set")
	}
	return nil
}

func (l LocalBackend) Connect(strgCfg *map[string]string) (StorageBackend, error) {
	if strgCfg != nil {
		jsonBody, err := json.Marshal(*strgCfg)
		// This cannot fail - even if the map is empty, it won't
		// return an error. But let's keep the check just in case
		if err != nil { // coverage-ignore
			return l, err
		}

		// unmarshal json into config struct
		cfg := LocalBackendCfg{}
		if err = json.Unmarshal(jsonBody, &cfg); err != nil {
			return l, err
		}
		err = cfg.AssertConfigured()
		if err != nil {
			return l, err
		}
		l.Cfg = cfg
	}

	create := false

	err := utils.FileOrDirExists(l.Cfg.ObjectPath)
	if err != nil {
		create = true
	}

	if create {
		err = os.MkdirAll(l.Cfg.ObjectPath, 0755)
		if err != nil { // coverage-ignore
			return l, err
		}
	}

	return l, nil
}

func (l LocalBackend) Upload(namespace, remotePath string, data []byte) (string, error) {
	localBknd, err := l.Connect(nil)
	if err != nil { // coverage-ignore
		return "", err
	}
	l, ok := localBknd.(LocalBackend)
	if !ok { // coverage-ignore
		return "", fmt.Errorf("couldn't connect to local backend")
	}

	fullPath := fmt.Sprintf("%v/%v/%v", l.Cfg.ObjectPath, namespace, remotePath)
	fileName := utils.GetFileNameFromUrl(fullPath)
	objectPathWithoutFn := strings.Replace(fullPath, fileName, "", -1)

	err = os.MkdirAll(objectPathWithoutFn, 0755)
	if err != nil { // coverage-ignore
		return "", err
	}
	err = os.WriteFile(fullPath, data, 0644)
	if err != nil { // coverage-ignore
		return "", err
	}

	fullObjectPath := fmt.Sprintf("%v:%v/%v", l.Cfg.ObjectHost, l.Cfg.ObjectPort, fmt.Sprintf("%v/%v", namespace, remotePath))

	return fullObjectPath, nil
}

func (l LocalBackend) RemoveObjectsOlderThan(duration time.Duration) ([]string, error) {
	var deletedObjects []string
	now := time.Now()
	log.Printf("removing objects older than %v\n", duration)
	// list directories at l.Cfg.ObjectPath - this is the equivalent of containers in swift
	entries, err := os.ReadDir(l.Cfg.ObjectPath)
	if err != nil { // coverage-ignore
		return deletedObjects, err
	}
	// for each directory, os.Stat -> FileInfo -> info.ModTime()
	for _, entry := range entries {
		fullPath := fmt.Sprintf("%v/%v", l.Cfg.ObjectPath, entry.Name())
		fi, err := os.Stat(fullPath)
		if err != nil { // coverage-ignore
			return deletedObjects, err
		}
		// ModTime returns time.Time, so can directly compare them with no parsing.
		modTime := fi.ModTime()
		timeSinceLastMod := now.Sub(modTime)
		if timeSinceLastMod > duration {
			log.Printf("%v is older than %v, removing...", entry.Name(), duration)
			// if older than duration, nuke the container/directory
			err = os.RemoveAll(fullPath)
			if err != nil { // coverage-ignore
				return deletedObjects, err
			}
			deletedObjects = append(deletedObjects, entry.Name())
		}
	}
	return deletedObjects, nil
}

type SwiftBackend struct {
	Con swift.Connection
	Cfg SwiftBackendConfig
}

type SwiftBackendConfig struct {
	User    string `json:"swift_user"`
	ApiKey  string `json:"swift_api_key"`
	AuthUrl string `json:"swift_auth_url"`
	Domain  string `json:"swift_domain"`
	Tenant  string `json:"swift_tenant"`
}

func (s SwiftBackendConfig) AssertConfigured() error {
	if s.User == "" {
		return fmt.Errorf("swift_user not set")
	}
	if s.ApiKey == "" {
		return fmt.Errorf("swift_api_key not set")
	}
	if s.AuthUrl == "" {
		return fmt.Errorf("swift_auth_url not set")
	}
	if s.Domain == "" {
		return fmt.Errorf("swift_domain not set")
	}
	if s.Tenant == "" {
		return fmt.Errorf("swift_tenant not set")
	}
	return nil
}

func (s SwiftBackend) Connect(strgCfg *map[string]string) (StorageBackend, error) {
	if strgCfg != nil {
		jsonBody, err := json.Marshal(*strgCfg)
		if err != nil { // coverage-ignore
			return s, err
		}

		// unmarshal json into config struct
		cfg := SwiftBackendConfig{}
		if err = json.Unmarshal(jsonBody, &cfg); err != nil {
			return s, err
		}
		err = cfg.AssertConfigured()
		if err != nil {
			return s, err
		}
		s.Cfg = cfg
	}

	s.Con = swift.Connection{
		UserName: s.Cfg.User,
		ApiKey:   s.Cfg.ApiKey,
		AuthUrl:  s.Cfg.AuthUrl,
		Domain:   s.Cfg.Domain,
		Tenant:   s.Cfg.Tenant,
	}

	ctxt := context.TODO()

	err := s.Con.Authenticate(ctxt)
	if err != nil {
		return s, err
	}

	return s, nil // coverage-ignore
}

// Since this function basically only uses swift internals, and swift is pretty
// intricate to deploy both locally and in CI, we'll avoid unit testing this func
// for the time being. In the future this should change.
func (s SwiftBackend) Upload(namespace, remotePath string, data []byte) (string, error) { // coverage-ignore
	ctxt := context.TODO()
	swiftBknd, err := s.Connect(nil)
	if err != nil {
		return "", err
	}
	s, ok := swiftBknd.(SwiftBackend)
	if !ok {
		return "", fmt.Errorf("couldn't connect to swift backend")
	}

	err = s.Con.ContainerCreate(ctxt, namespace, nil)
	if err != nil {
		return "", err
	}

	hashSum := utils.Md5SumOfBytes(data)
	ctxt = context.TODO()
	writeCloser, err := s.Con.ObjectCreate(ctxt, namespace, remotePath, true, hashSum, "", make(map[string]string))
	if err != nil {
		return "", err
	}

	_, err = writeCloser.Write(data)
	if err != nil {
		return "", err
	}

	err = writeCloser.Close()
	if err != nil {
		return "", err
	}

	ctxt = context.TODO()
	baseSwiftUrl, err := s.Con.GetStorageUrl(ctxt)
	if err != nil {
		return "", err
	}

	fullStorageUrl := fmt.Sprintf("%v/%v/%v", baseSwiftUrl, namespace, remotePath)

	return fullStorageUrl, nil
}

// TODO: Not testing for now since we have no swift for testing just yet.
func (s SwiftBackend) RemoveObjectsOlderThan(duration time.Duration) ([]string, error) { // coverage-ignore
	var deletedObjects []string
	// ensure swift connection is initialised
	ctxt := context.TODO()
	swiftBknd, err := s.Connect(nil)
	if err != nil {
		return deletedObjects, err
	}
	s, ok := swiftBknd.(SwiftBackend)
	if !ok {
		return deletedObjects, fmt.Errorf("couldn't connect to swift backend")
	}

	now := time.Now()

	// get complete list of containers
	ctxt = context.TODO()
	containers, err := s.Con.ContainersAll(ctxt, nil)
	if err != nil {
		return deletedObjects, err
	}
	// get container info
	for _, cont := range containers {
		ctxt = context.TODO()
		_, hdrs, err := s.Con.Container(ctxt, cont.Name)
		if err != nil {
			return deletedObjects, err
		}
		// Time format is:
		// Tue, 08 Apr 2025 11:09:15 GMT
		// Which is RFC1123
		lastMod := hdrs["Last-Modified"] // I've also seen this as last-modified, annoyingly
		timeObj, err := time.Parse(time.RFC1123, lastMod)
		if err != nil {
			return deletedObjects, err
		}
		if now.Sub(timeObj) > duration {
			ctxt = context.TODO()
			// remove the container
			err = s.Con.ContainerDelete(ctxt, cont.Name)
			if err != nil {
				return deletedObjects, err
			}
			deletedObjects = append(deletedObjects, cont.Name)
		}
	}
	return deletedObjects, nil
}
