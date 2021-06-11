# render-drawio-action

A GitHub Action for rendering `*.drawio` files generated with [diagrams.net](https://www.diagrams.net/) (formerly [draw.io](https://draw.io/)).

## How to use

1. Go to [diagrams.net](https://www.diagrams.net/)
2. Choose to "Save diagrams to" GitHub
   - <img src="docs/images/2.png" width="300"/>
3. Choose "Create New Diagram" or "Open Existing Diagram", depending on what you want to do
   - <img src="docs/images/3.png" width="300"/>
4. Authorize the app through OAuth2 if asked
   - <img src="docs/images/4.png" width="300"/>
5. Choose what repository you want to browse for files in
6. Choose what branch the file you want to edit is on (the branch already needs to exist)
7. Choose what file you want to edit (or create) in that repository folder structure (`special-computing-machine` here is just an example; you can choose any repo you have given draw.io access to)
   - <img src="docs/images/7.png" width="300"/>
8. You will now see the online editor; you can now edit your diagram as you like
   - ![diagrams.net editor](docs/images/8.png)
9. When you make any changes; you will see a "Unsaved changes. Click here to save"-button.
   - ![diagrams.net editor](docs/images/9.png)
10. When you are ready to save your changes into a commit, click that button and write your commit message.
11. A commit has now been created to the given branch on your behalf.
    - ![diagrams.net editor](docs/images/11.png)
12. This GitHub Action detects the change, and automatically renders the "raw" `.drawio` file into the format of your liking

## Inputs

### `formats`

**Optional:** A comma-separated list of the formats to render. Supported formats are: `svg,pdf,png,jpg`.

**Default:** `svg`

Examples:

- `png,svg`
- `pdf`
- `png,pdf,svg,jpg`

### `sub-dirs`

**Optional:** A comma-separated list of what directories to search for .drawio files

**Default:** `.`

Examples:

- `.`
- `docs,content/sketches`

### `skip-dirs`

**Optional:** A comma-separated list of what directories to skip when searching for .drawio files

**Default:** `.git`

Examples:

- `.git`
- `foo/dont_include,bar/dont_include`

### `files`

**Optional:** A comma-separated list of specific files to convert, in the form: "dest-file:src-file"

**Default:** Empty

Examples:

- `docs/images/sketch.png:docs/drawings/sketch.drawio`
- `docs/backup-sketch.pdf:docs/drawings/sketch.drawio.bak`

### `log-level`

**Optional:** What log level to use. Recognized levels are "info" and "debug".

**Default:** `info`

Examples:

- `info`
- `debug`

## Output

### `rendered-files`

A space-separated list of files that were rendered, can be passed to e.g. "git add"

Example:

- `test/sketch.svg test/sketch.png diagrams/intro.png`

## Usage Example

The following example GitHub Action pushes a new commit with the generated files.

```yaml
name: Render draw.io files

on: [push]

jobs:
  render_drawio:
    runs-on: ubuntu-latest
    name: Render draw.io files
    steps:
    - name: Checkout
      uses: actions/checkout@v2
    - name: Render draw.io files
      uses: docker://ghcr.io/racklet/render-drawio-action:v1.0.3
      id: render
      with: # Showcasing the default values here
        formats: 'svg'
        sub-dirs: '.'
        skip-dirs: '.git'
        # files: '' # unset, specify "dest-file:src-file" mappings here
        log-level: 'info'
    - name: List the rendered files
      run: 'ls -l ${{ steps.render.outputs.rendered-files }}'
    - uses: EndBug/add-and-commit@v7
      with:
        # This "special" author name and email will show up as the GH Actions user/bot in the UI
        author_name: github-actions
        author_email: 41898282+github-actions[bot]@users.noreply.github.com
        message: 'Automatically render .drawio files'
        add: "${{ steps.render.outputs.rendered-files }}"
      if: "${{ steps.render.outputs.rendered-files != ''}}"
```

## Docker

You can use it standalone as well, through the Docker container:

```console
$ docker run -it -v $(pwd):/files ghcr.io/racklet/render-drawio-action:v1 --help
Usage of /render-drawio:
  -f, --files stringToString   Comma-separated list of files to render, of form 'dest-file=src-file'. The extension for src-file can be any of [drawio *], and for dest-file any of [pdf png jpg svg] (default [])
      --formats strings        Comma-separated list of formats to render the *.drawio files as, for use with --subdirs (default [svg])
      --log-level Level        What log level to use (default info)
  -r, --root-dir string        Where the root directory for the files that should be rendered are. (default "/files")
  -s, --skip-dirs strings      Comma-separated list of sub-directories of --root-dir to skip when recursively checking for files to convert (default [.git])
  -d, --sub-dirs strings       Comma-separated list of sub-directories of --root-dir to recursively search for files to render (default [.])
pflag: help requested
```

Sample Docker usage:

```console
$ docker run -it -v $(pwd):/files ghcr.io/racklet/render-drawio-action:v1
{"level":"info","msg":"Got config","cfg":{"RootDir":"/files","SubDirs":["."],"SkipDirs":[".git"],"Files":{},"SrcFormats":["drawio"],"ValidSrcFormats":["drawio","*"],"DestFormats":["svg"],"ValidDestFormats":["pdf","png","jpg","svg"]}}
{"level":"info","msg":"Created os.DirFS at /files"}
{"level":"info","msg":"Walking subDir ."}
{"level":"info","msg":"Rendering test/foo.drawio -> test/foo.svg"}
{"level":"info","msg":"Setting Github Action output","rendered-files":"/files/test/foo.svg"}
::set-output name=rendered-files::/files/test/foo.svg
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) and our [Code Of Conduct](CODE_OF_CONDUCT.md).

Other interesting resources include:

- [The issue tracker](https://github.com/racklet/racklet/issues)
- [The discussions forum](https://github.com/racklet/racklet/discussions)
- [The list of milestones](https://github.com/racklet/racklet/milestones)
- [The roadmap](https://github.com/orgs/racklet/projects/1)
- [The changelog](CHANGELOG.md)

## Getting Help

If you have any questions about, feedback for or problems with Racklet:

- Invite yourself to the [Open Source Firmware Slack](https://slack.osfw.dev/).
- Ask a question on the [#racklet](https://osfw.slack.com/messages/racklet/) slack channel.
- Ask a question on the [discussions forum](https://github.com/racklet/racklet/discussions).
- [File an issue](https://github.com/racklet/racklet/issues/new).
- Join our [community meetings](https://hackmd.io/@racklet/Sk8jHHc7_) (see also the [meeting-notes](https://github.com/racklet/meeting-notes) repo).

Your feedback is always welcome!

## Maintainers

In alphabetical order:

- Dennis Marttinen, [@twelho](https://github.com/twelho)
- Jaakko Sirén, [@Jaakkonen](https://github.com/Jaakkonen)
- Lucas Käldström, [@luxas](https://github.com/luxas)
- Verneri Hirvonen, [@chiplet](https://github.com/chiplet)

## License

[Apache 2.0](LICENSE)
