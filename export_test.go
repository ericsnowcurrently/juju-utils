// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

var (
	GOMAXPROCS        = &gomaxprocs
	NumCPU            = &numCPU
	Dial              = dial
	NetDial           = &netDial
	ResolveSudoByFunc = resolveSudo
)

func RawEnvVars(env *OSEnv) map[string]string {
	return env.vars
}
