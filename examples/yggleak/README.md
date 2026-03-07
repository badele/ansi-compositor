# yggleak the cron is not dead

```bash
ansi-compositor yggleak.yaml -K  -F ansi | splitans  -S  -W 120 -N 76 | sed 's/F0DBC79/Fg/g' | sed 's/FFFFFFF/FW/g' | sed 's/F11A8CD/Fc/g' > yggleak.neo
splitans -f neotex -E cp437 -F ansi -S yggleak.neo > yggleak.ans
ansilove -c 120 yggleak.ans
```

![cron](./yggleak.ans.png)
