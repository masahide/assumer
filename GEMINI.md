# Project Overview

This project, `assumer`, is a command-line utility written in Go. Its primary purpose is to simplify the process of assuming AWS IAM roles, including those configured for AWS Single Sign-On (SSO). It acts as a wrapper, allowing tools that do not natively support AWS profiles to seamlessly use assumed role credentials.

The tool reads AWS configuration from `~/.aws/credentials` and `~/.aws/config` to manage role assumption.

# Building and Running

The project uses a `Makefile` to streamline common development tasks.

-   **Build:** To build the `assumer` binary, run:
    ```bash
    make build
    ```
    This will create an `assumer` executable in the project's root directory.

-   **Run Tests:** To run the test suite:
    ```bash
    make test
    ```

-   **Linting:** To check the code for style and errors:
    ```bash
    make lint
    ```

-   **CI Pipeline:** To run the same checks as the CI/CD pipeline:
    ```bash
    make ci
    ```

# Development Conventions

-   **Dependency Management:** The project uses Go modules for managing dependencies.
-   **Code Style:** Code is formatted using `gofmt` and `goimports`. The `make fmt` command can be used to format the code automatically.
-   **CI/CD:** The project has a CI/CD pipeline configured in `.github/workflows/buildpkg.yml`. This pipeline automates building, testing, and packaging the application for different operating systems (Linux and macOS) and architectures (amd64 and arm64).
-   **Packaging:** The application is packaged into Debian (`.deb`), RPM (`.rpm`), and Homebrew formulas for easy distribution. The packaging process is handled by Go scripts in the `pack/` directory and automated in the GitHub Actions workflow.
