# Power Control Service(pcs)

This is a fork of the original PCS code from [Cray-HPE/hms-power-control](https://github.com/Cray-HPE/hms-power-control), suitable only for experimentation and demo purposes at this point.

## Build/Install with goreleaser

This project uses [GoReleaser](https://goreleaser.com/) to automate releases and include additional build metadata such as commit info, build time, and versioning. Below is a guide on how to set up and build the project locally using GoReleaser.

### Environment Variables

To include detailed build metadata, ensure the following environment variables are set:

* __GIT_STATE__: Indicates whether there are uncommitted changes in the working directory. Set to clean if the repository is clean, or dirty if there are uncommitted changes.
* __BUILD_HOST__: The hostname of the machine where the build is being performed.
* __GO_VERSION__: The version of Go used for the build. GoReleaser uses this to ensure consistent Go versioning information.
* __BUILD_USER__: The username of the person or system performing the build.

Set all the environment variables with:
```bash
export GIT_STATE=$(if git diff-index --quiet HEAD --; then echo 'clean'; else echo 'dirty'; fi)
export BUILD_HOST=$(hostname)
export GO_VERSION=$(go version | awk '{print $3}')
export BUILD_USER=$(whoami)
```

### Building Locally with GoReleaser

Once the environment variables are set, you can build the project locally using GoReleaser in snapshot mode (to avoid publishing).


Follow the installation instructions from [GoReleaser’s documentation](https://goreleaser.com/install/).

1. Run GoReleaser in snapshot mode with the --snapshot flag to create a local build without attempting to release it:
  ```bash
  goreleaser release --snapshot --clean
  ```
2.	Check the dist/ directory for the built binaries, which will include the metadata from the environment variables. You can inspect the binary output to confirm that the metadata was correctly embedded.

__NOTE__ If you see errors, ensure that you are using the same version of goreleaser that is being used in the [Release Action](.github/workflows/Release.yml)