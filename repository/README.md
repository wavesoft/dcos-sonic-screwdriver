# Local Repository

This package contains the operations to access and manage the local package repository in the user's machine.

## Repository Structure

The repository is located on the user's data directory (usually `$HOME/.mesosphere/toolbox`) and has the following structure:

```
~/.mesosphere/toolbox
├── pkg
│   ├── <artifact ID>
│   ...
├── tools
│   ├── <name>
│   │   ├── <version>-<artifact ID>
│   │   │   ├── run
│   │   │   ...
│   │   ...
│   ...
└── registry.json
```

Where:

* `artifact ID` : A SHA256 checksum that uniquely identifies a particular artifact that was downloaded and cached.
* `name` : Is the name of the tool
* `version` : Is the version of the tool

The directory structure has the following purpose:

* `pkg/<artifact ID>` : Contains the files downloaded and used to satisfy the installation needs of a tool. This can vary from code sources all the way down to pre-compiled binaries.
* `tool/<name>` : Contains one or more installed versions of the tool. Some versions might re-use the same artifact and just update the run-time details.
* `tool/<name>/<version>-<artifact ID>` : Contains the run-time environment for the tool. For example this could be the directory where the python virtualenv is located, or could be the build directory where the sources were compiled.
* `tool/<name>/<version>-<artifact ID>/run` : Is the entry point to the tool, that is going to be symlinked to the user's `bin` directory.

## API Interface



