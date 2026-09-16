IASI Builder prototype

Usage:
  iasi-builder
  iasi-builder <path>

Contract:
  - Working directory defaults to current directory.
  - Exactly one *.toml configuration is expected.
  - The TOML must define template.
  - Template name must end in .tpl.
  - Output defaults to _outputs/<template name without .tpl>.
  - All TOML values are exposed to Go text/template.
  - iasi-builder does not dispatch by type; iasi-dev does that.
