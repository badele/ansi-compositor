# yggleak the cron is not dead

```bash
ansi-compositor yggleak.yaml > yggleak.neo
splitans -f neotex -F ansi -V -K yggleak.neo > output.ans
reset && \cat output.ans && magick import -window $(xdotool getactivewindow) screenshot.png && magick screenshot.png -crop +0-130 -trim +repage yggleak.png
```

![cron](./yggleak.ans.png)
