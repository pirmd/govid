package main

import (
	"strings"
	"testing"
)

func TestHtpasswdAuthenticate(t *testing.T) {
	testHtpwd, err := NewHtpasswd(strings.NewReader("rvid:$2b$10$ISdqfeODKyB4Qjd8KqA5BuP4whZY2bQlFmkrMoDhfLfyB1Xqx4c0Ov"))
	if err != nil {
		t.Fatalf("fail to parse htpassd file: %v", err)
	}

	if !testHtpwd.Authenticate("rvid", "TestingIsFun") {
		t.Errorf("Authentication of 'rvid' failed")
	}

	if testHtpwd.Authenticate("Unauthorized", "12345") {
		t.Errorf("Authentication of 'Unauthorized' succeed")
	}

	if testHtpwd.Authenticate("rvid", "12345") {
		t.Errorf("Authentication of 'Unauthorized' succeed")
	}
}
