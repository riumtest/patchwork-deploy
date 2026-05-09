package patch

import (
	"os"
	"testing"
)

func TestDefaultEnvPolicy_Empty(t *testing.T) {
	p := DefaultEnvPolicy()
	if len(p.Vars) != 0 {
		t.Fatalf("expected empty Vars, got %v", p.Vars)
	}
	if len(p.PassThrough) != 0 {
		t.Fatalf("expected empty PassThrough, got %v", p.PassThrough)
	}
}

func TestEnvResolver_ExplicitVars(t *testing.T) {
	p := DefaultEnvPolicy()
	p.Vars["DEPLOY_ENV"] = "staging"
	p.Vars["APP_VERSION"] = "1.2.3"

	r := NewEnvResolver(p)
	env := r.Environ()

	if env["DEPLOY_ENV"] != "staging" {
		t.Errorf("expected DEPLOY_ENV=staging, got %q", env["DEPLOY_ENV"])
	}
	if env["APP_VERSION"] != "1.2.3" {
		t.Errorf("expected APP_VERSION=1.2.3, got %q", env["APP_VERSION"])
	}
}

func TestEnvResolver_PassThrough(t *testing.T) {
	os.Setenv("PW_TEST_VAR", "hello")
	defer os.Unsetenv("PW_TEST_VAR")

	p := DefaultEnvPolicy()
	p.PassThrough = []string{"PW_TEST_VAR", "PW_MISSING_VAR"}

	r := NewEnvResolver(p)
	env := r.Environ()

	if env["PW_TEST_VAR"] != "hello" {
		t.Errorf("expected PW_TEST_VAR=hello, got %q", env["PW_TEST_VAR"])
	}
	if _, ok := env["PW_MISSING_VAR"]; ok {
		t.Error("expected PW_MISSING_VAR to be absent")
	}
}

func TestEnvResolver_PrefixFilter(t *testing.T) {
	os.Setenv("PWDEPLOY_HOST", "prod.example.com")
	os.Setenv("PWDEPLOY_PORT", "22")
	os.Setenv("OTHER_VAR", "ignored")
	defer func() {
		os.Unsetenv("PWDEPLOY_HOST")
		os.Unsetenv("PWDEPLOY_PORT")
		os.Unsetenv("OTHER_VAR")
	}()

	p := DefaultEnvPolicy()
	p.Prefix = "PWDEPLOY_"

	r := NewEnvResolver(p)
	env := r.Environ()

	if env["PWDEPLOY_HOST"] != "prod.example.com" {
		t.Errorf("expected PWDEPLOY_HOST, got %q", env["PWDEPLOY_HOST"])
	}
	if env["PWDEPLOY_PORT"] != "22" {
		t.Errorf("expected PWDEPLOY_PORT, got %q", env["PWDEPLOY_PORT"])
	}
	if _, ok := env["OTHER_VAR"]; ok {
		t.Error("OTHER_VAR should not be included")
	}
}

func TestEnvResolver_ExplicitOverridesPassThrough(t *testing.T) {
	os.Setenv("PW_OVERRIDE", "original")
	defer os.Unsetenv("PW_OVERRIDE")

	p := DefaultEnvPolicy()
	p.PassThrough = []string{"PW_OVERRIDE"}
	p.Vars["PW_OVERRIDE"] = "overridden"

	r := NewEnvResolver(p)
	env := r.Environ()

	if env["PW_OVERRIDE"] != "overridden" {
		t.Errorf("expected overridden, got %q", env["PW_OVERRIDE"])
	}
}

func TestEnvResolver_ResolveFormat(t *testing.T) {
	p := DefaultEnvPolicy()
	p.Vars["KEY"] = "VALUE"

	r := NewEnvResolver(p)
	slice := r.Resolve()

	if len(slice) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(slice))
	}
	if slice[0] != "KEY=VALUE" {
		t.Errorf("expected KEY=VALUE, got %q", slice[0])
	}
}
