# Package Registry

Until I come up with a better solution, this folder contains the `registry.json` that is read from the tool in order to learn about the available packages.

## Entry Description

Each tool appears on the root `tool` object:

```json
    "tools": {
        "tool-name": {
            ...
        }
    }
```

### Tool Record

Each tool record has the following fields:

* `topics` : An array with topics related to this tool (keywords)
* `help` : Source of information to the user (see [Help Record](#help-record) for more details)
* `desc` : A short description of this tool
* `versions` : An array of [Version Rercord](#version-record) entries

```json
    ...
    "tool-name": {
        "topics": [ "marathon", "recovery" ],
        "desc": "Short Description",
        "help": {
            "text": "Help Message"
        },
        "versions": [        
            ...
        ]
    }
    ...
```

### Version Record

Each tool can have one or more released versions. This can be particular useful if you want to be backwards-compatible and you want to allow the user to install an older version.

Each version record has the following fields:

* `version` : An array with the major, minor and revision numbers
* `artifacts` : One or more download-able artifacts. The one compatible with the user's computer will be downloaded. Check the [Artifacts Record](#artifacts-record) for more details.

```json
    ...
    {
      "version": [1,5,5],
      "artifacts": [
        ...
      ]
    }
    ...
```

### Artifacts Record

Each artifact is an installable component in the user's computer. There are three kinds of artifacts that will end-up as command-line tools to the user's computer:

1. **Docker Images**

   The _Sonic Screwdriver_ can automatically wrap docker images as command-line tools. Docker artifacts have the following fields:

   * `type` : Must be `"docker"`
   * `image` : The name of the docker image
   * `tag` : The tag of the docker image
   * `dockerArgs` : You can optionally specify additional arguments to the docker command (ex. expose ports, mount volumes etc.)

    ```json
        ...
        {
          "type": "docker"
          "image": "mesosphere/marathon-storage-tool",
          "tag": "1.5.5",
          "dockerArgs": "-p 8080:8080"
        }
        ...
    ```
2. **Interpreted Script**

    An interpreted script is using a system interpreter for it's execution and therefore is by default platform-agnostic. 

    Such artifacts have the following fields:

    * `type` : Must be `"executable"`
    * `interpreter` : The name of the interpreter (ex. `python`, `python3`, `java` etc.)
    * `source` : Where to find the artifact, can be either:
        - A URL and a SHA256 checksum to validate the contents of the payload:
          ```json
          {
            "url": "http://path/to/artifact",
            "checksum": "010203040506..."
          }
          ```
        - A GIT repository, that will be checked-out during installation:
          ```json
          {
            "gitUrl": "http://github.com/mesosphere/awesome-tool.git"
          }
          ```
    * `extract` : A boolean flag, that if set to `true` assumes that the `soruce` is an archive and will be extracted. Otherwise, it's assumed to be a link to the script itself.
    * `entrypoint` : If the `source` is an archive or a git repository, this field should point to the path within the repository or the archive that should be interpreted.

    ```json
        ...
        {
          "source": {
            "url": "https://raw.githubusercontent.com/vishnu2kmohan/dcos-toolbox/master/aws/setup-aws-secrets.sh",
            "checksum": "5594c30450660c2f367701ac42f807f3521d3d1ebdaa7450878640224b03ac2d"
          },
          "extract": false,
          "interpreter": "bash"
        }
        ...
    ```
    ```json
        ...
        {
          "source": {
            "url": "https://some/path/to/archive.tar.gz",
            "checksum": "01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b"
          },
          "extract": true,
          "entrypoint": "/tools/tool.pyl",
          "interpreter": "python"
        }
        ...
    ```
    ```json
        ...
        {
          "source": {
            "gitUrl": "https://github.com/mesosphere/dcos-perf-test-driver.git"
          },
          "entrypoint": "/build/driver.jar",
          "interpreter": "java"
        }
        ...
    ```
3. **Executable Binary**

    An executable binary is a machine-dependent artifact and therefore you have to specify the `platform` and the cpu architecture that it targets. Apart from this, it's syntax is exactly the same as with the previous case.

    Such artifacts have the following fields:

    * `type` : Must be `"executable"`
    * `arch` : The CPU architecture (ex. `386`, `amd64`, `arm`, `s390x` etc.)
    * `platform` : The OS platform (ex. `darwin`, `freebsd`, `linux`, `windows` etc.)
    * `source` : _(The same as the Interpreted Script)_
    * `extract` :  _(The same as the Interpreted Script)_
    * `entrypoint` :  _(The same as the Interpreted Script)_

    ```json
        ...
        {
          "source": {
            "url": "https://some.repository.com/some-binary-x86_64.darwin",
            "checksum": "5594c30450660c2f367701ac42f807f3521d3d1ebdaa7450878640224b03ac2d"
          },
          "extract": false,
          "platform": "darwin",
          "arch": "amd64"
        }
        ...
    ```

### Help Record

The `help` record provides some useful information to the user, when (s)he queries it using the `ss help <tool>` command. You have two options:

1. **Embed the description**

    This should only be used on brief descriptions:

    ```json
    {
        "text": "Some\ndescription\nhere!"
    }
    ```
2. **Open the browser to the help page**

    If the description is quite extensive, you can instruct the tool to open a browser window and point the user to the designated URL:

    ```json
    {
        "url": "http://path/to/readme.html"
    }
    ```
3. **Download and display the message**

    You can also instruct the tool to download a description hosted in a remote location and display it in-line to the user.

    **Note:** If the document you are pointing to is in Markdown format, you can instruct the tool to format the console output, to match the markdown
    document as closely as possible (in the console :smile: )

    ```json
    {
        "url": "http://path/to/readme.md",
        "inline": true,
        "markdown": true
    }
    ```
