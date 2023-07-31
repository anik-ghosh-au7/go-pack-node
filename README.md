# Go-Pack-Node

Go-Pack-Node is a Node.js package manager written in Go. It provides a lightweight, efficient, and fast alternative to npm, yarn, or pnpm, specifically designed for managing Node.js packages.

## Features

- **Project Initialization**: Set up a new Node.js project with default or user-provided configurations.
- **Dependency Management**: Install, update, and manage Node.js packages for your project.
- **Script Execution**: Run scripts defined in your project's `dependencies.json` file.

## How it works

### Project Initialization

The `Initialize` function in `controller/initializer.go` sets up a new Node.js project. It creates necessary directories and files such as `.cache`, `dependencies.json`, `dependencies-lock.json`, and `node_modules`. It prompts the user to enter project details like name, version, description, and entry point unless the `-y` flag is provided, in which case it uses default values.

### Dependency Management

The `Install` function in `controller/installer.go` manages the Node.js packages for your project. It reads the `dependencies.json` file to get the list of dependencies and their versions. If no arguments are provided, it installs all the dependencies listed in the `dependencies.json` file. Otherwise, it installs the packages provided as arguments. It fetches the package information from the npm registry, downloads the package, and copies it to the `node_modules` directory.

### Script Execution

The `Run` function in `controller/runner.go` executes scripts defined in your project's `dependencies.json` file. It reads the `dependencies.json` file, retrieves the script command, and executes it.

## Usage

To use this package manager, you need to have Go installed on your machine. Clone the repository and build the application using `go build`. The application accepts the following commands:

- `init`: Initializes a new Node.js project.
- `install`: Installs Node.js packages as dependencies.
- `start`: Runs the `start` script.
- `run <script>`: Runs the specified script.

## Contributing

Contributions are welcome. Please submit a pull request or create an issue to discuss the changes.

## License

This project is licensed under the MIT License.
