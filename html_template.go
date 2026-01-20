package main

const htmlTemplate = `<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>System Snapshot Report</title>
<style>
  :root { --bg:#0b0f17; --card:#121a26; --txt:#e6edf3; --muted:#9aa4b2; --line:#223044; }
  body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial; background:var(--bg); color:var(--txt); }
  .wrap { max-width: 1200px; margin: 0 auto; padding: 18px; }
  .row { display:flex; gap:12px; flex-wrap:wrap; }
  .card { background:var(--card); border:1px solid var(--line); border-radius:14px; padding:14px; }
  .grow { flex:1; min-width:280px; }
  h1 { margin:0 0 6px 0; font-size:20px; }
  h2 { margin:0 0 10px 0; font-size:16px; color:var(--muted); font-weight:600; }
  .meta { color:var(--muted); font-size:12px; line-height:1.5; }
  .pill { display:inline-block; padding:6px 10px; border-radius:999px; border:1px solid var(--line); background:#0e1522; margin-right:8px; font-size:12px; color:var(--muted); }
  .score { font-size:36px; font-weight:800; letter-spacing:-1px; }
  .grade { font-size:14px; color:var(--muted); }
  .reasons { margin:8px 0 0 0; padding-left:18px; color:var(--muted); font-size:13px; }
  .controls { display:flex; gap:10px; flex-wrap:wrap; align-items:center; }
  input, select { background:#0e1522; border:1px solid var(--line); color:var(--txt); padding:8px 10px; border-radius:10px; }
  .small { font-size:12px; color:var(--muted); }
  table { width:100%; border-collapse:collapse; }
  th, td { border-bottom:1px solid var(--line); padding:10px 8px; vertical-align:top; }
  th { text-align:left; color:var(--muted); font-size:12px; position:sticky; top:0; background:var(--card); }
  td { font-size:13px; }
  .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
  .muted { color:var(--muted); }
  .hide { display:none; }
  details summary { cursor:pointer; color:var(--muted); }
  pre { white-space:pre-wrap; word-break:break-word; background:#0e1522; border:1px solid var(--line); border-radius:12px; padding:10px; }
</style>
</head>
<body>
<div class="wrap">
  <div class="row">
    <div class="card grow">
      <h1>System Snapshot Report</h1>
      <div class="meta">
        Host: <span class="mono">{{.Meta.Hostname}}</span> · Kernel: <span class="mono">{{.Meta.Kernel}}</span><br>
        Generated: <span class="mono">{{.Meta.GeneratedAt}}</span> · Since: <span class="mono">{{.Meta.Since}}</span> · User: <span class="mono">{{.Meta.User}}</span>
      </div>
    </div>

    <div class="card" style="min-width:280px;">
      <div class="pill">Health score</div>
      <div class="score">{{.Health.Score}}</div>
      <div class="grade">Grade: <span class="mono">{{.Health.Grade}}</span></div>
      <ul class="reasons">
        {{range .Health.Reasons}}<li>{{.}}</li>{{end}}
      </ul>
      <div class="small" style="margin-top:8px;">
        failed units: <span class="mono">{{.Health.FailedUnits}}</span> ·
        OOM: <span class="mono">{{.Health.OOMCount}}</span> ·
        IO hits: <span class="mono">{{.Health.IOErrors}}</span> ·
        FS hits: <span class="mono">{{.Health.FSErrors}}</span> ·
        disk>=90%: <span class="mono">{{.Health.DiskCritical}}</span><br>
        unmounted fs: <span class="mono">{{.Health.UnmountedFSTypes}}</span> ·
        SMART bad: <span class="mono">{{.Health.SmartBadHealth}}</span>
      </div>
    </div>
  </div>

  <div class="row" style="margin-top:12px;">
    <div class="card grow">
      <h2>Quick Summary</h2>
      <details open><summary>Uptime / Load</summary><pre class="mono">{{.Summary.Uptime}}
{{.Summary.LoadAvg}}</pre></details>
      <details><summary>Memory</summary><pre class="mono">{{.Summary.MemInfo}}</pre></details>
      <details><summary>Disk (df -hT)</summary><pre class="mono">{{.Summary.DiskDf}}</pre></details>
      <details><summary>Block devices (lsblk)</summary><pre class="mono">{{.Summary.LsblkText}}</pre></details>
      <details><summary>Mounted / Unmounted (derived)</summary><pre class="mono">{{.Summary.LsblkMountView}}</pre></details>
      <details><summary>Mounts</summary><pre class="mono">{{.Summary.Mounts}}</pre></details>
      <details><summary>Network (ip/route/ss)</summary><pre class="mono">{{.Summary.NetOverview}}</pre></details>
      <details><summary>Failed systemd units</summary><pre class="mono">{{.Summary.FailedSystemd}}</pre></details>
      <details><summary>Recent reboots/logins</summary><pre class="mono">{{.Summary.RecentReboots}}</pre></details>
      <details><summary>SMART</summary><pre class="mono">{{.Summary.SmartctlInfo}}</pre></details>
      <details><summary>Top CPU</summary><pre class="mono">{{.Summary.TopCPU}}</pre></details>
      <details><summary>Top MEM</summary><pre class="mono">{{.Summary.TopMEM}}</pre></details>
      <div class="small muted" style="margin-top:8px;">
        {{.Summary.OOMKills}} · {{.Summary.DmesgErrorsBrief}}
      </div>
    </div>
  </div>

  <div class="row" style="margin-top:12px;">
    <div class="card grow">
      <h2>Findings (errors/failures) — <span class="mono" id="countAll">{{len .Findings}}</span></h2>

      <div class="controls">
        <input id="q" type="text" placeholder="Search in message/source..." style="min-width:260px;">
        <select id="sev">
          <option value="">Severity: any</option>
          <option value="ERR">ERR</option>
          <option value="WARN/ERR">WARN/ERR</option>
          <option value="WARN">WARN</option>
          <option value="MATCH">MATCH</option>
          <option value="INFO">INFO</option>
        </select>
        <select id="src">
          <option value="">Source: any</option>
        </select>
        <span class="small">Shown: <span class="mono" id="countShown">0</span></span>
      </div>

      <div style="margin-top:10px; overflow:auto; max-height: 560px; border:1px solid var(--line); border-radius:12px;">
        <table id="tbl">
          <thead>
            <tr>
              <th style="width:64px;">#</th>
              <th style="width:220px;">Source</th>
              <th style="width:110px;">Severity</th>
              <th>Message</th>
            </tr>
          </thead>
          <tbody>
            {{range $i, $f := .Findings}}
            <tr data-source="{{$f.Source}}" data-severity="{{$f.Severity}}">
              <td class="mono muted">{{$i}}</td>
              <td class="mono">{{$f.Source}}</td>
              <td class="mono muted">{{$f.Severity}}</td>
              <td>{{$f.Message}}</td>
            </tr>
            {{end}}
          </tbody>
        </table>
      </div>

      <div class="small muted" style="margin-top:8px;">
        Tip: full export in <span class="mono">errors.csv</span> and machine data in <span class="mono">report.json</span>.
      </div>
    </div>
  </div>
</div>

<script>
(function(){
  const q = document.getElementById('q');
  const sev = document.getElementById('sev');
  const src = document.getElementById('src');
  const tbody = document.querySelector('#tbl tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  const countShown = document.getElementById('countShown');

  const sources = Array.from(new Set(rows.map(r => r.dataset.source))).sort();
  for (const s of sources) {
    const opt = document.createElement('option');
    opt.value = s;
    opt.textContent = s;
    src.appendChild(opt);
  }

  function apply(){
    const qq = (q.value || '').toLowerCase().trim();
    const ss = sev.value;
    const so = src.value;

    let shown = 0;
    for (const r of rows) {
      const text = (r.innerText || '').toLowerCase();
      const okQ = !qq || text.includes(qq);
      const okS = !ss || r.dataset.severity === ss;
      const okO = !so || r.dataset.source === so;
      const ok = okQ && okS && okO;
      r.classList.toggle('hide', !ok);
      if (ok) shown++;
    }
    countShown.textContent = shown;
  }

  q.addEventListener('input', apply);
  sev.addEventListener('change', apply);
  src.addEventListener('change', apply);

  apply();
})();
</script>
</body>
</html>`
