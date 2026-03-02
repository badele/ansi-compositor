<div align="center">
    <h1>
        <img src="./docs/logo.png"/>
    </h1>
</div>

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

## Example

A sample `ansi-compositor` result

![example.png](./docs/example.png)

For more information about this example, see the files in [./docs](./docs).

### Complete example

```yaml
term:
  width: 180
  height: 180
  encoding: utf8

defaults:
  inputFormat: ansi
  inputEncoding: utf8

output:
  format: neotex

layers:
  - name: logo
    x: 1
    y: 1
    alignH: center
    # cmd: curl 'https://codef-ansi-logo-maker-api.santo.fr/api.php?text=ansi%20compositor&font=78&spacing=2&spacesize=5&vary=2'
    file: logo.neo
    inputFormat: neotex
  - name: slogan
    x: 1
    y: 22
    width: 180
    height: 1
    alignH: center
    content: "—————————---- Compose, color, and craft ANSI art cleanly ----—————————"
  - name: weather-title
    x: 3
    y: 26
    alignH: center
    cmd: bit -fit-scales 0.5,1,2,4  -fit-height 3 -fit-priority height -fit-limit 1 "WEATHER EXAMPLE"
  - name: weather
    x: 1
    y: 30
    alignH: center
    cmd: >-
      sh -c 'set -euo pipefail; curl --fail --max-time 1 -o /tmp/weather.ansi wttr.in; splitans -W 125 /tmp/weather.ansi > docs/weather.neo; splitans -f neotex -F ansi docs/weather.neo'
      || { [ -f docs/weather.neo ] && splitans docs/weather.neo -f neotex -F ansi || printf "[weather unavailable]\n"; }
  - name: ratesx-title
    x: 1
    y: 68
    alignH: center
    cmd: bit -fit-scales 0.5,1,2,4  -fit-height 5 -fit-priority height -fit-limit 1  "RATE SX EXAMPLE"
  - name: ratesx
    x: 1
    y: 74
    alignH: center
    cmd: >-
      sh -c 'set -euo pipefail; curl --fail --max-time 1 -o /tmp/ratesx.ansi https://eur.rate.sx/; splitans -W 104 /tmp/ratesx.ansi > docs/ratesx.neo; splitans -f neotex -F ansi docs/ratesx.neo'
      || { [ -f docs/ratesx.neo ] && splitans docs/ratesx.neo -f neotex -F ansi || printf "[rates unavailable]\n"; }
```
