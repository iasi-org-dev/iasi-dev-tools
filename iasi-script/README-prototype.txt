IASI Script prototype

Usage:
  iasi-script
  iasi-script [working-dir]
  iasi-script --file <file> [working-dir]
  iasi-script --root <root> [working-dir]
  iasi-script -h

Contract:
  - Working directory defaults to the current directory.
  - The working directory, when supplied, is the last positional argument.
  - Without --file or --root, config.toml is used from the working directory.
  - --file selects one TOML configuration file.
  - --root recursively discovers and processes all *.toml files below a root.
  - Relative --file and --root paths are resolved from the working directory.
  - --file and --root are mutually exclusive.
  - The TOML may define template or templates.
  - All non-template TOML values are exposed to Go text/template.
  - iasi-script does not dispatch by type; iasi-dev does that.
