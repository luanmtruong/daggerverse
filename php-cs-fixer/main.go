// A tool to automatically fix PHP Coding Standards issues.

package main

import (
	"dagger/php-cs-fixer/internal/dagger"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "ghcr.io/php-cs-fixer/php-cs-fixer"

type PhpCsFixer struct {
	Container *dagger.Container
}

func New(
	// Version (image tag) to use from the official image repository as a base container.
	//
	// +optional
	// +default="3"
	version string,

	// Customize PHP version (supported versions: https://github.com/PHP-CS-Fixer/PHP-CS-Fixer?tab=readme-ov-file#supported-php-versions).
	//
	// +optional
	phpVersion string,

	// Custom container to use as a base container. Takes precedence over version and phpVersion.
	//
	// +optional
	container *dagger.Container,
) *PhpCsFixer {
	if container == nil {
		if version == "" {
			version = "3"
		}

		if phpVersion != "" {
			version = version + "-php" + phpVersion
		}

		container = dag.Container().From(defaultImageRepository + ":" + version)
	}

	return &PhpCsFixer{
		Container: container,
	}
}

// Check if configured files/directories comply with configured rules.
func (m *PhpCsFixer) Check(
	source *dagger.Directory,

	// Paths with source code to run analysis on.
	//
	// +optional
	paths []string,
) *dagger.Container {
	args := []string{"php-cs-fixer", "check", "--show-progress", "none", "--no-interaction"}

	if len(paths) > 0 {
		args = append(args, paths...)
	}

	return m.Container.
		WithWorkdir("/work/src").
		WithMountedDirectory("/work/src", source).
		WithExec(args)
}
