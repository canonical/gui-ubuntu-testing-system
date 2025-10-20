package api

import (
	"reflect"
	"testing"
)

func TestUuidNotFoundError(t *testing.T) {
	var UuidError UuidNotFoundError
	UuidError.uuid = "4ce9189f-561a-4886-aeef-1836f28b073b"
	ExpectedString := "No jobs with uuid 4ce9189f-561a-4886-aeef-1836f28b073b found!"
	if !reflect.DeepEqual(UuidError.Error(), ExpectedString) {
		t.Errorf("Uuid failure string not as expected!\nExpected: %v\nActual: %v", ExpectedString, UuidError.Error())
	}
}

func TestBadUrlError(t *testing.T) {
	urlError := BadUrlError{url: "https://planet-express.nny", code: 404}
	desiredErrString := "Url https://planet-express.nny returned 404"
	if urlError.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, urlError.Error())
	}
}

func TestNonWhitelistedDomainError(t *testing.T) {
	domainErr := NonWhitelistedDomainError{url: "https://inspector-5.com"}
	desiredErrString := "Url https://inspector-5.com not from accepted list of domains"
	if domainErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, domainErr.Error())
	}
}

func TestApiKeyNotAcceptedError(t *testing.T) {
	keyErr := ApiKeyNotAcceptedError{}
	desiredErrString := "Api key not accepted!"
	if keyErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, keyErr.Error())
	}
}

func TestEmptyApiKeyError(t *testing.T) {
	keyErr := EmptyApiKeyError{}
	desiredErrString := "Api key passed is empty"
	if keyErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, keyErr.Error())
	}
}

func TestPlanFileNonexistentError(t *testing.T) {
	planFileErr := PlanFileNonexistentError{planFile: "dummy/file"}
	desiredErrString := "Plan file dummy/file doesn't exist!"
	if planFileErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, planFileErr.Error())
	}
}

func TestInvalidArtifactTypeError(t *testing.T) {
	artifactErr := InvalidArtifactTypeError{url: "https://central-bureaucracy.gov/hello.rpm"}
	desiredErrString := "url https://central-bureaucracy.gov/hello.rpm contains an invalid artifact type"
	if artifactErr.Error() != desiredErrString {
		t.Errorf("Unexpected error string!\nExpected: %v\nActual: %v", desiredErrString, artifactErr.Error())
	}
}
