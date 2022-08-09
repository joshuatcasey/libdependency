package versionology

import "github.com/Masterminds/semver/v3"

type HasVersion interface {
	GetVersion() *semver.Version
}

type SimpleHasVersion struct {
	version *semver.Version
}

func NewSimpleHasVersion(version *semver.Version) SimpleHasVersion {
	return SimpleHasVersion{
		version: version,
	}
}

func (s SimpleHasVersion) GetVersion() *semver.Version {
	return s.version
}
