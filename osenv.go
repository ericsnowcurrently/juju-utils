// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/juju/errors"
)

// OSEnv is a snapshot of an OS environment.
type OSEnv struct {
	vars  map[string]string
	names []string // ...because order matters.
}

// NewOSEnv creates a new OSEnv, prepopulated with the initial vars.
func NewOSEnv(initial ...string) *OSEnv {
	env := &OSEnv{
		vars: make(map[string]string),
	}
	env.Update(initial...)
	return env
}

// ReadOSEnv creates a new OSEnv and prepoulates it with the values
// from the current OS environment.
func ReadOSEnv() *OSEnv {
	initial := os.Environ()
	env := NewOSEnv(initial...)
	return env
}

// Names returns the list of env var names, in order.
func (env OSEnv) Names() []string {
	names := make([]string, len(env.names))
	copy(names, env.names)
	return names
}

// EmptyNames returns the list of env var names for which the value
// is the empty string.
func (env OSEnv) EmptyNames() []string {
	var empty []string
	for _, name := range env.Names() {
		if env.vars[name] == "" {
			empty = append(empty, name)
		}
	}
	return empty
}

// Get returns the value of the named environment variable. If it is
// not set then the empty string is returned.
func (env OSEnv) Get(name string) string {
	envVar, _ := env.vars[name]
	return envVar
}

// Set updates the value of the named environment variable. The old
// value, if any, is returned.
func (env *OSEnv) Set(name, value string) string {
	existing, ok := env.vars[name]
	if !ok {
		env.names = append(env.names, name)
		existing = ""
	} // We otherwise preserve the original order.
	env.vars[name] = value
	return existing
}

// Unset ensures the named env var is removed. The old values, if any,
// are returned.
func (env *OSEnv) Unset(names ...string) []string {
	var values []string
	for _, name := range names {
		value := env.unset(name)
		values = append(values, value)
	}
	return values
}

func (env *OSEnv) unset(name string) string {
	value, ok := env.vars[name]
	if !ok {
		return ""
	}
	for i, existing := range env.names {
		if existing == name {
			env.names = append(env.names[:i], env.names[i+1:]...)
			break
		}
	}
	delete(env.vars, name)
	return value
}

// Update sets all the provided env vars, in order. If any is already
// set then it is overwritten, though its original order is preserved.
// To reset an env var's order, unset it before calling Update.
func (env *OSEnv) Update(vars ...string) {
	for _, envVar := range vars {
		name, value := SplitEnvVar(envVar)
		if _, ok := env.vars[name]; !ok {
			env.names = append(env.names, name)
		} // We otherwise preserve the original order.
		env.vars[name] = value
	}
}

// Reduce filters out all env vars that don't match the provided filters
// and returns the remainder in a new OSEnv.
func (env OSEnv) Reduce(filters ...func(name string) bool) *OSEnv {
	newEnv := NewOSEnv()
	for _, name := range env.Names() {
		if filtersAnd(name, filters) {
			value := env.Get(name)
			newEnv.Set(name, value)
		}
	}
	return newEnv
}

func filtersAnd(name string, filters []func(string) bool) bool {
	matched := false
	for _, filter := range filters {
		if !filter(name) {
			return false
		}
		matched = true
	}
	return matched
}

// TODO(ericsnow) Add an equivalent method to os.ExpandEnv.

// Copy returns a copy of the env.
func (env OSEnv) Copy() *OSEnv {
	return NewOSEnv(env.AsList()...)
}

// AsMap copies the environment variables into a new map.
func (env OSEnv) AsMap() map[string]string {
	envVars := make(map[string]string)
	for name, value := range env.vars {
		envVars[name] = value
	}
	return envVars
}

// AsList copies the environment variables into a new list of raw
// env var strings.
func (env OSEnv) AsList() []string {
	var envVars []string
	for _, name := range env.names {
		value := env.vars[name]
		envVar := JoinEnvVar(name, value)
		envVars = append(envVars, envVar)
	}
	return envVars
}

// PushOSEnv updates the current OS environment.
func PushOSEnv(env OSEnv) error {
	for _, name := range env.names {
		value := env.vars[name]
		if err := os.Setenv(name, value); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// PushOSEnvFresh updates the current OS environment after clearing it.
func PushOSEnvFresh(env OSEnv) error {
	os.Clearenv()
	if err := PushOSEnv(env); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// SplitEnvVar converts a raw env var string into a (name, value) pair.
func SplitEnvVar(envVar string) (string, string) {
	parts := strings.SplitN(envVar, "=", 2)
	if len(parts) == 1 {
		return envVar, ""
	}
	return parts[0], parts[1]
}

// JoinEnvVar converts a (name, value) pair into a raw env var string.
func JoinEnvVar(name, value string) string {
	return fmt.Sprintf("%s=%s", name, value)
}