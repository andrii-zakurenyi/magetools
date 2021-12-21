// Docker package provides common docker commands.
package docker

import (
	"fmt"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

const (
	//nolint:godox // godox: this is initial version, will replace with better path input/config options later. 2021-12-21.
	// TODO: make this a configuration value, or auto discover.
	dockerComposeFile = "docker-compose.yml"

	// DockerDir contains all the dockerfiles and docker-compose files.
	dockerDir = "docker"
)

// Docker namespace contains all the tasks for Docker commands.
type Docker mg.Namespace

// Compose namespace contains all the tasks for Docker compose commands.
type Compose mg.Namespace

// 🧹 DockerPrune cleans up images.
//
// parameter: dockerMaintainer is the label used to filter to images maintained by this maintainer (ok for small teams).
//
// This is useful to narrow the pruning to just the images being build by the designated maintainer over `n` hours old.
func (Docker) Prune(maintainer string) error {
	pterm.Info.Println("Pruning images over 24 hours old and maintained by ", maintainer)
	return sh.RunV(
		"docker",
		"image",
		"prune",
		"--filter",
		"until=24h",
		"--filter",
		"label=maintainer="+maintainer,
	)
}

// invokeDockerCompose runs the docker command with the given args.
// It checks for docker compose being a valid command, and if not it runs docker-compose.
//
// This command uses RunV because streaming the output of Docker Compose is typically a user driven action and feedback is useful.
func invokeDockerCompose(args ...string) error {
	magetoolsutils.CheckPtermDebug()
	if len(args) == 0 {
		return nil
	}
	// Check for docker compose version to return without error, and if so proceed with newer docker compose command
	if err := sh.Run("docker", "compose", "version"); err != nil {
		// if docker compose returned and error, then default to the older python based docker-compose command
		pterm.Debug.Println("docker compose version command failed. Defaulting to docker-compose")
		return sh.RunV("docker-compose", args...)
	}
	// Since it didn't fail, we are safe to use docker compose.
	pterm.Debug.Println("docker compose version is a valid command, defaulting to docker compose based invoke")
	args = append([]string{"compose"}, args...)
	return sh.RunV("docker", args...)
}

// Build runs docker build command against the provided dockerfile.
func (Docker) Build(dockerfile, imagename, tag string) error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultSection.Println("Building docker image")
	dockerfileDirectory := filepath.Dir(dockerfile)
	imagetag := fmt.Sprintf("%s:%s", imagename, tag)
	if err := sh.Run("docker", "build", "--pull", "--rm", "-f", dockerfile, "-t", imagetag, dockerfileDirectory); err != nil {
		return err
	}

	return nil
}

// ⬆ Build runs docker compose stack build. //noinspection.
func (Compose) Build() error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultSection.Println("docker compose build")
	df := filepath.Join(dockerDir, dockerComposeFile)
	return invokeDockerCompose("-f", df, "build")
}
