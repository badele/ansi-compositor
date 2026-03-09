# commands

```bash
ansi-compositor commands.yaml -F ansi
```

![commands](./commands.png)

## logo generation

```bash
nix shell github:badele/splitans nixpkgs#ansilove
curl "https://codef-ansi-logo-maker-api.santo.fr/api.php?text=ANSI&font=161&spacing=1&spacesize=5&vary=0" > logo.ans
icy_draw logo.ans
splitans -W 120 logo.ans > logo.neo
splitans -f neotex -E cp437 -F ansi -S logo.neo > logo.ans
ansilove -c 120 -o logo.png logo.ans
```

![logo](./logo.png)

## Source

- Tools
  - [ansilove](https://github.com/ansilove/ansilove)
  - [CODEF](https://n0namen0.github.io/CODEF_Ansi_Logo_Maker/)
  - [icy_draw](https://github.com/mkrueger/icy_tools)
