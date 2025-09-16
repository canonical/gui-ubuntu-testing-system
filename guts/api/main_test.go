package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Setup() error {
	ParseArgs()
	err := ParseConfig(configFilePath)
	CheckError(err)
	Driver, err = NewDbDriver(GutsCfg)
	CheckError(err)
	return err
}

func ServeDirectory() {
	pwd, err := os.Getwd()
	CheckError(err)
	port := "9999"
	testFilesDir := pwd + "/../../postgres/test-data/test-files/"
	serveCmd := exec.Command("php", "-S", "localhost:"+port)
	serveCmd.Dir = testFilesDir
	go serveCmd.Run() //nolint:all
	for i := 0; i < 60; i++ {
		timeout := time.Second * 5
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
		if err != nil {
			time.Sleep(timeout)
		} else {
			if conn != nil {
				err := conn.Close()
				CheckError(err)
				return
			}
		}
	}
	CheckError(fmt.Errorf("Port never came up when trying to serve directory with command:\n%v", serveCmd))
}

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestJobEndpoint(t *testing.T) {
	r := SetUpRouter()
	r.GET("/job/:uuid", JobEndpoint)
	ExpectedResponse := `"{\"Job\":{\"uuid\":\"4ce9189f-561a-4886-aeef-1836f28b073b\",\"artifact_url\":null,\"tests_repo\":\"https://github.com/canonical/ubuntu-gui-testing.git\",\"tests_repo_branch\":\"main\",\"tests_plans\":[\"tests/firefox-example/plans/extended.yaml\",\"tests/firefox-example/plans/regular.yaml\"],\"image_url\":\"https://cdimage.ubuntu.com/daily-live/current/questing-desktop-amd64.iso\",\"reporter\":\"test_observer\",\"status\":\"running\",\"submitted_at\":\"2025-07-23T14:17:14.632177Z\",\"requester\":\"andersson123\",\"debug\":false,\"priority\":8},\"results\":{\"Firefox-Example-Basic\":\"running\",\"Firefox-Example-New-Tab\":\"spawning\"}}"`
	Uuid := "4ce9189f-561a-4886-aeef-1836f28b073b"
	reqFound, _ := http.NewRequest("GET", "/job/"+Uuid, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)
	ActualResponse := w.Body.String()
	if !reflect.DeepEqual(ExpectedResponse, ActualResponse) {
		t.Errorf("Json response not as expected!\nExpected:%v\nActual:%v", ExpectedResponse, ActualResponse)
	}
}

func TestJobEndpointUnknownUuid(t *testing.T) {
	r := SetUpRouter()
	r.GET("/job/:uuid", JobEndpoint)

	Uuid := "3676ead0-6d93-422d-91cc-0da81d6f594a"
	reqFound, _ := http.NewRequest("GET", "/job/"+Uuid, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 404

	if !reflect.DeepEqual(w.Code, expectedCode) {
		t.Errorf("Unexpected exit code!\nExpected: %v\nActual: %v", expectedCode, w.Code)
	}
}

func TestJobEndpointInvalidUuid(t *testing.T) {
	r := SetUpRouter()
	r.GET("/job/:uuid", JobEndpoint)

	Uuid := "asdf"
	reqFound, _ := http.NewRequest("GET", "/job/"+Uuid, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400

	if !reflect.DeepEqual(w.Code, expectedCode) {
		t.Errorf("Unexpected exit code!\nExpected: %v\nActual: %v", expectedCode, w.Code)
	}
}

func TestArtifactsEndpoint(t *testing.T) {
	r := SetUpRouter()
	r.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
	Uuid := "27549483-e8f5-497f-a05d-e6d8e67a8e8a"
	reqFound, _ := http.NewRequest("GET", "/artifacts/"+Uuid+"/results.tar.gz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)
	expectedCode := 200
	if !reflect.DeepEqual(w.Code, expectedCode) {
		t.Errorf("Unexpected exit code!\nExpected: %v\nActual: %v", expectedCode, w.Code)
	}
}

func TestArtifactsEndpointUnknownUuid(t *testing.T) {
	r := SetUpRouter()
	r.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
	Uuid := "3676ead0-6d93-422d-91cc-0da81d6f594a"
	reqFound, _ := http.NewRequest("GET", "/artifacts/"+Uuid+"/results.tar.gz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)
	expectedCode := 404

	if !reflect.DeepEqual(w.Code, expectedCode) {
		t.Errorf("Unexpected exit code!\nExpected: %v\nActual: %v", expectedCode, w.Code)
	}
}

func TestArtifactsEndpointInvalidUuid(t *testing.T) {
	r := SetUpRouter()
	r.GET("/artifacts/:uuid/results.tar.gz", ArtifactsEndpoint)
	Uuid := "asdf"
	reqFound, _ := http.NewRequest("GET", "/artifacts/"+Uuid+"/results.tar.gz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)
	expectedCode := 400

	if !reflect.DeepEqual(w.Code, expectedCode) {
		t.Errorf("Unexpected exit code!\nExpected: %v\nActual: %v", expectedCode, w.Code)
	}
}

func CreateAcceptableJobRequest() JobRequest {
	var req JobRequest
	myString := "https://launchpad.net/ubuntu/+archive/primary/+files/hello_2.10-5_amd64.deb"
	req.ArtifactUrl = &myString
	req.TestsRepo = "https://github.com/canonical/ubuntu-gui-testing.git"
	req.TestsRepoBranch = "main"
	req.TestsPlans = []string{"tests/firefox-example/plans/regular.yaml", "tests/firefox-example/plans/extended.yaml"}
	req.TestBed = "https://releases.ubuntu.com/noble/ubuntu-24.04.3-desktop-amd64.iso"
	req.Debug = false
	req.Priority = 9
	req.Reporter = ""
	return req
}

func TestRequestEndpointSuccess(t *testing.T) {
	request := CreateAcceptableJobRequest()

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 200
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointBadJson(t *testing.T) {
	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader("asdf"))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointEmptyApiKey(t *testing.T) {
	request := CreateAcceptableJobRequest()

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 401
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointUnauthorizedApiKey(t *testing.T) {
	request := CreateAcceptableJobRequest()

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "asdf")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 401
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointBadUrlError(t *testing.T) {
	request := CreateAcceptableJobRequest()
	myString := "https://launchpad.net/ubuntu/+archive/primary/+files/dingus_2.10-5_amd64.deb"
	request.ArtifactUrl = &myString

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointInvalidArtifactType(t *testing.T) {
	request := CreateAcceptableJobRequest()
	myString := "https://launchpad.net/ubuntu/+archive/primary/+files/hello_2.10-5_amd64.rpm"
	request.ArtifactUrl = &myString

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointBadArtifactDomain(t *testing.T) {
	request := CreateAcceptableJobRequest()
	myString := "https://momcorp.com/ubuntu/+archive/primary/+files/hello_2.10-5_amd64.deb"
	request.ArtifactUrl = &myString

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 403
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointBadTestbedUrl(t *testing.T) {
	request := CreateAcceptableJobRequest()
	request.TestBed = "https://releases.ubuntu.com/24.04.3/ubuntu-24.04.3-besktop-amd64.iso"

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointGitError(t *testing.T) {
	request := CreateAcceptableJobRequest()
	request.TestsRepo = "https://github.com/momcorp-bending-unit-ocr.git"

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}

func TestRequestEndpointBadPlanFile(t *testing.T) {
	request := CreateAcceptableJobRequest()
	request.TestsPlans = []string{"non/existant/plan.yaml"}

	r := SetUpRouter()
	r.POST("/request/", RequestEndpoint)

	reqFound, _ := http.NewRequest("POST", "/request/", strings.NewReader(request.toJson()))
	reqFound.Header.Add("X-Api-Key", "4c126f75-c7d8-4a89-9370-f065e7ff4208")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	expectedCode := 400
	if w.Code != expectedCode {
		t.Errorf("wtf! code is expected to be %v but is actually %v, and response string is:\n%v", expectedCode, w.Code, w.Body.String())
	}
}
