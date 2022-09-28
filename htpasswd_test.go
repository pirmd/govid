package main

import (
	"strings"
	"testing"
)

func TestHtpasswdAuthenticate(t *testing.T) {
	testHtpwd, err := NewHtpasswd(strings.NewReader("govid:$2b$10$ISdqfeODKyB4Qjd8KqA5BuP4whZY2bQlFmkrMoDhfLfyB1Xqx4c0Ov"))
	if err != nil {
		t.Fatalf("fail to parse htpassd file: %v", err)
	}

	if !testHtpwd.Authenticate("govid", "TestingIsFun") {
		t.Errorf("Authentication of 'govid' failed")
	}

	if testHtpwd.Authenticate("Unauthorized", "12345") {
		t.Errorf("Authentication of 'Unauthorized' succeed")
	}

	if testHtpwd.Authenticate("govid", "12345") {
		t.Errorf("Authentication of 'govid' succeed")
	}
}
