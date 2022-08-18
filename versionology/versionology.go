package versionology

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/collections"
)

func HasVersionToString(semverVersions []HasVersion) []string {
	return collections.TransformFunc(semverVersions, func(version HasVersion) string {
		return version.GetVersion().String()
	})
}

func SemverToString(semverVersions []*semver.Version) []string {
	return collections.TransformFunc(semverVersions, func(v *semver.Version) string {
		return v.String()
	})
}

func ConstraintsToString(semverVersions []Constraint) []string {
	return collections.TransformFunc(semverVersions, func(c Constraint) string {
		return c.Constraint.String()
	})
}

func LogAllVersions(id, description string, versions []HasVersion) {
	fmt.Printf("Found %d versions of %s %s\n", len(versions), id, description)

	if output, err := json.MarshalIndent(HasVersionToString(versions), "", " "); err != nil {
		fmt.Printf("error marshalling versions from upstream: %s", err)
	} else {
		fmt.Println(string(output))
	}
}

func FilterVersionsByConstraints(inputVersions []*semver.Version, constraints []*semver.Constraints) []*semver.Version {
	if len(constraints) == 0 {
		return inputVersions
	}

	var outputVersions []*semver.Version

	for _, version := range inputVersions {
		for _, constraint := range constraints {
			if constraint.Check(version) {
				outputVersions = append(outputVersions, version)
			}
		}
	}

	return outputVersions
}

func FilterHasVersionsByConstraints(id string, inputVersions []HasVersion, constraints []Constraint, dependencies []Dependency) []HasVersion {
	constraintsToDependencies := make(map[Constraint]DependencyArray)

	for _, dependency := range dependencies {
		for _, constraint := range constraints {
			if constraint.Check(dependency) {
				constraintsToDependencies[constraint] = append(constraintsToDependencies[constraint], dependency)
			}
		}
	}

	constraintsToInputVersion := make(map[Constraint][]HasVersion)

	for _, version := range inputVersions {
		for _, constraint := range constraints {
			if constraint.Check(version) {
				constraintsToInputVersion[constraint] = append(constraintsToInputVersion[constraint], version)
			}
		}
	}

	for constraint, versions := range constraintsToInputVersion {
		constraintDescription := fmt.Sprintf("for constraint %s", constraint.Constraint.String())
		LogAllVersions(id, constraintDescription, versions)
	}

	constraintsToOutputVersions := make(map[Constraint][]HasVersion)

	for constraint, inputVersionsForConstraint := range constraintsToInputVersion {
		existingDependencies := constraintsToDependencies[constraint]

	ConstraintsToInputVersionLoop:
		for _, inputVersionForConstraint := range inputVersionsForConstraint {
			for _, existingDependency := range existingDependencies {
				if inputVersionForConstraint.GetVersion().LessThan(existingDependency.GetVersion()) || inputVersionForConstraint.GetVersion().Equal(existingDependency.GetVersion()) {
					continue ConstraintsToInputVersionLoop
				}
			}
			constraintsToOutputVersions[constraint] = append(constraintsToOutputVersions[constraint], inputVersionForConstraint)
		}
	}

	var outputVersions []HasVersion

	for constraint, constraintsToOutputVersion := range constraintsToOutputVersions {
		sort.Slice(constraintsToOutputVersion, func(i, j int) bool {
			return constraintsToOutputVersion[i].GetVersion().LessThan(constraintsToOutputVersion[j].GetVersion())
		})

		if constraint.Patches < len(constraintsToOutputVersion) {
			constraintsToOutputVersion = constraintsToOutputVersion[len(constraintsToOutputVersion)-constraint.Patches:]
		}

		outputVersions = append(outputVersions, constraintsToOutputVersion...)
	}

	for constraint, versions := range constraintsToOutputVersions {
		constraintDescription := fmt.Sprintf("for constraint %s, after limiting for patches", constraint.Constraint.String())
		LogAllVersions(id, constraintDescription, versions)
	}

	if len(constraints) < 1 {
	ZeroConstraintsLoop:
		for _, inputVersion := range inputVersions {
			for _, dependency := range dependencies {
				if inputVersion.GetVersion().LessThan(dependency.GetVersion()) || inputVersion.GetVersion().Equal(dependency.GetVersion()) {
					continue ZeroConstraintsLoop
				}
			}
			outputVersions = append(outputVersions, inputVersion)
		}
	}

	LogAllVersions(id, "as new versions", outputVersions)
	return outputVersions
}
