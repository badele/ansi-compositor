# love notebook

```bash
chafa -f symbols --symbols braille --colors none  ~/Downloads/logo4-BW.png --invert --size 80 | splitans -N 80 > love-notebook.neo
ansi-compositor original-calm.yaml > original-calm.neo
splitans -f neotex -F ansi -V -K original-calm.neo > output.ans
reset && \cat output.ans && magick import -window $(xdotool getactivewindow) screenshot.png && magick screenshot.png -crop +0-180 -trim +repage original-calm.png
```

![original-calm](./original-calm.png)

## Dynamic

```bash
while sleep 1 ; do x="$( ansi-compositor time-calm.yaml -F ansi 2>&1 )" ; clear ; echo "$x" ; done
```

![time-calm](./time-calm.gif)
