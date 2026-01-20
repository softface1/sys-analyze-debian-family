#######BUILD
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o sys-snapshot .

output -  "sys-snapshot" binary

#######RUN
sudo ./sys-snapshot --outdir /tmp/snap --since 24h --pack


--since 2h|24h|7d — окно для journalctl и анализов

--outdir — куда писать

--timeout — общий таймаут команд

--raw=false — не сохранять raw

--html=false — не генерить html

--pack — +tar
