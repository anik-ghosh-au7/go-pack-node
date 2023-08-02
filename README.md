# Go-Pack-Node

Go-Pack-Node is a demonstration of a Node.js package manager, built in Go. It provides a lightweight and efficient way to manage Node.js dependencies. This project is intended to illustrate the inner workings of existing package managers like `npm`, `yarn`, etc.

## Features

- **Project Initialization**: Sets up a new Node.js project with either default or user-provided configurations.
- **Dependency Management**: Installs, updates, and manages Node.js packages for your project.
- **Script Execution**: Runs scripts defined in your project's `package.json` file.

## How it works

### Project Initialization

The `Initialize` function in `controller/initializer.go` sets up a new Node.js project. It creates the necessary directories and files such as `.cache`, `package.json`, `package-lock.json`, and `node_modules`. The function prompts the user to enter project details like name, version, description, and entry point unless the `-y` flag is provided, in which case it uses default values.

### Dependency Management

The `Install` function in `controller/installer.go` manages the Node.js packages for your project. It reads the `package.json` file to get the list of dependencies and their versions. If no arguments are provided, it installs all the dependencies listed in the `package.json` file. Otherwise, it installs the packages provided as arguments. The function fetches the package information from the npm registry, downloads the package, and copies it to the `node_modules` directory.

### Script Execution

The `Run` function in `controller/runner.go` executes scripts defined in your project's `package.json` file. It reads the file, retrieves the script command, and executes it.

## Usage

To use this package manager, you need to have Go installed on your machine. Clone the repository and build the application using `go build`. The application accepts the following commands:

- `init`: Initializes a new Node.js project.
- `install`: Installs Node.js packages as dependencies.
- `start`: Runs the `start` script.
- `run <script>`: Runs the specified script.

## Known Issues

- Some packages with internal dependencies may cause bugs. This is a known issue and it will be addressed in the future.

## Roadmap

- While copying from cache check .package-lock.json file if version already exists in node_modules, if yes skip
- Add functionality to uninstall, update package.json, package-lock.json & .package-lock.json accordingly
- Add sha type and hash for the packages to verify integrity in package-lock.json & .package-lock.json

## Contributing

Although this project is primarily for demonstration purposes, contributions are still welcome. If you have found a bug, have a question, or want to propose a feature, feel free to submit a pull request or create an issue to discuss the changes.
