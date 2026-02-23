![ansi compositor](./docs/logo.png)

Compose ANSI/Neotex files from a YAML description. ansi-compositor relies on
splitans for parsing/tokenizing ANSI or Neotex sources and exporting the final
buffer.

## Usage

```bash
ansi-compositor path/to/config.yaml
```

CLI options: `-o` (output file), `-F` (format: ansi|neotex|plaintext), `-I`
(inline), `-v` (VGA colors), `-V` (version). These override `output.file`,
`output.format`, `output.inline`, and `term.vgaColors` from the YAML. SAUCE is
controlled only via YAML.

## Configuration

Key YAML fields:

- `term.width`, `term.height`: workspace dimensions (required).
- `term.fill`: optional background (char + neotex style).
- `term.vgaColors`: enable VGA palette for splitans rendering.
- `defaults.inputFormat`: auto|ansi|neotex|plaintext; `defaults.inputEncoding`:
  utf8|cp437|cp850|iso-8859-1.
- `layers[]`: each layer has `x`, `y` (1-indexed) and exactly one source among
  `file` | `cmd` | `content`; `cmd` can be a shell string or a list of args.
  Alignment options: `alignH`, `alignV`, `crop`.
- `output.format`: ansi|neotex|plaintext; `output.inline`: bool; `output.file`:
  path.
- `sauce`: optional block (see below).

## SAUCE via YAML

- If the `sauce` block is absent: no SAUCE is exported.
- If present: SAUCE is exported (unless `enabled: false`).
- Strict length limits (error if exceeded): `title` ≤ 35, `author` ≤ 20, `group`
  ≤ 20, `font` (TInfoS) ≤ 22.
- Date: `YYYYMMDD` or `YYYY-MM-DD`.
- Supported for `output.format` ansi and neotex; ignored for plaintext.

Available fields:

```yaml
sauce:
  enabled: true # optional, defaults to true when block exists
  title: "My Art"
  author: "Bruno Adele"
  group: "Demo"
  date: "20250208"
  font: "80x25"
  iceColors: true
```

## Complete example

```yaml
defaults:
  inputFormat: neotex
  inputEncoding: utf8

term:
  width: 80
  height: 25
layers:
  - name: base
    x: 1
    y: 1
    file: art.neo
output:
  format: ansi
sauce:
  title: "My Art"
  author: "Bruno Adele"
  group: "Demo"
  date: "20250208"
  font: "80x25"
  iceColors: true
```
