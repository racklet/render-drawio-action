# render-drawio-action

A GitHub Action for rendering `*.drawio` files generated with [diagrams.net](https://www.diagrams.net/) (formerly [draw.io](https://draw.io/)).

## How to use

1. Go to [diagrams.net](https://www.diagrams.net/)
2. Choose to `Save diagrams to:` GitHub
   - <img src="docs/images/save_diagrams_dialog.png" alt="Save diagrams dialog" style="vertical-align:middle;width:300px"/>
3. Choose `Create New Diagram` or `Open Existing Diagram`, depending on what you want to do
   - <img src="docs/images/open_diagram_dialog.png" alt="Open diagram dialog" style="vertical-align:middle;width:300px"/>
4. Authorize the app through OAuth2 if asked
   - <img src="docs/images/authorize_github_dialog.png" alt="Authorize GitHub dialog" style="vertical-align:middle;width:300px"/>
5. Choose a repository you want to browse files in
6. Choose the branch the file you want to edit is in (the branch needs to already exist)
7. Choose the file you want to edit (or create) in that repository folder structure (`special-computing-machine` here is just an example, you can choose any repo you have given diagrams.net access to)
   - <img src="docs/images/select_file_dialog.png" alt="Select file dialog" style="vertical-align:middle;width:300px"/>
8. You will now see the online editor, feel free to edit your diagram as you like
   - <img src="docs/images/example_diagram.png" alt="Example diagram" style="vertical-align:middle;width:800px"/>
9. As you make changes you will see a button stating `Unsaved changes. Click here to save.` appear
    - <img src="docs/images/unsaved_changes_button.png" alt="Unsaved changes button" style="vertical-align:middle;width:800px"/>
10. When you are ready to save your changes into a commit, click that button and write your commit message
11. A commit has now been created in the given branch on your behalf
    - <img src="docs/images/diagrams_net_commit.png" alt="diagrams.net commit" style="vertical-align:middle;width:800px"/>
12. This GitHub Action detects the change, and automatically renders the "raw" `.drawio` file into the format(s) of your liking

## Inputs

### `formats`

**Optional:** A comma-separated list of the formats to render. All supported formats: `svg,pdf,png,jpg`

**Default:** `svg`

Examples:

- `png,svg`
- `pdf`
- `png,pdf,svg,jpg`

### `sub-dirs`

**Optional:** A comma-separated list of directories to consider when searching for `.drawio` files

**Default:** `.`

Examples:

- `.`
- `docs,content/sketches`

### `skip-dirs`

**Optional:** A comma-separated list of directories to skip when searching for `.drawio` files

**Default:** `.git`

Examples:

- `.git`
- `foo/dont_include,bar/dont_include`

### `files`

**Optional:** A comma-separated list of specific files to convert, in the form `dest-file=src-file`

**Default:** Empty

Examples:

- `docs/images/sketch.png:docs/drawings/sketch.drawio`
- `docs/backup-sketch.pdf:docs/drawings/sketch.drawio.bak`

### `log-level`

**Optional:** Specify the log level, recognized levels are `info` and `debug`

**Default:** `info`

Examples:

- `info`
- `debug`

## Output

### `rendered-files`

A space-separated list of files that were rendered, can be passed to e.g. `git add`

Example:

- `test/sketch.svg test/sketch.png diagrams/intro.png`

## Usage Example

The following example GitHub Actions workflow pushes a new commit with the generated files.

```yaml
name: Render .drawio files

on: [push]

jobs:
  render_drawio:
    runs-on: ubuntu-latest
    name: Render .drawio files
    steps:
    - name: Checkout
      uses: actions/checkout@v2
    - name: Render .drawio files
      uses: docker://ghcr.io/racklet/render-drawio-action:v1
      with: # Showcasing the default values here
        formats: 'svg'
        sub-dirs: '.'
        skip-dirs: '.git'
        # files: '' # unset, specify `dest-file=src-file` mappings here
        log-level: 'info'
      id: render
    - name: List the rendered files
      run: 'ls -l ${{ steps.render.outputs.rendered-files }}'
    - name: Commit the rendered files
      uses: EndBug/add-and-commit@v7
      with:
        # This makes the GH Actions user/bot the author of the commit
        default_author: github_actor
        message: 'Automatically render .drawio files'
        add: "${{ steps.render.outputs.rendered-files }}"
      if: "${{ steps.render.outputs.rendered-files != ''}}"
```

## Docker

You can use it standalone as well, by running the Docker container directly:

```console
$ docker run -it -v $(pwd):/files ghcr.io/racklet/render-drawio-action:v1 --help
Usage of /render-drawio:
  -f, --files stringToString   Comma-separated list of files to render, of form 'dest-file=src-file'. The extension for src-file can be any of [drawio *], and for dest-file any of [pdf png jpg svg] (default [])
      --formats strings        Comma-separated list of formats to render the *.drawio files as, for use with --subdirs (default [svg])
      --log-level Level        What log level to use (default info)
  -r, --root-dir string        Where the root directory for the files that should be rendered are. (default "/files")
  -s, --skip-dirs strings      Comma-separated list of sub-directories of --root-dir to skip when recursively checking for files to convert (default [.git])
  -d, --sub-dirs strings       Comma-separated list of sub-directories of --root-dir to recursively search for files to render (default [.])
pflag: help requested
```

Sample usage with Docker:

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

## Getting Help

If you have any questions about, feedback for or problems with Racklet:

- Invite yourself to the [Open Source Firmware Slack](https://slack.osfw.dev/).
- Ask a question on the [#racklet](https://osfw.slack.com/messages/racklet/) Slack channel.
- Ask a question on the [discussions forum](https://github.com/racklet/racklet/discussions).
- [File an issue](https://github.com/racklet/racklet/issues/new).
- Join our [community meetings](https://github.com/racklet/meeting-notes).

Your feedback is always welcome!

## Maintainers

In alphabetical order:

- Dennis Marttinen, [@twelho](https://github.com/twelho)
- Jaakko Sirén, [@Jaakkonen](https://github.com/Jaakkonen)
- Lucas Käldström, [@luxas](https://github.com/luxas)
- Verneri Hirvonen, [@chiplet](https://github.com/chiplet)

## License

[Apache 2.0](LICENSE)
