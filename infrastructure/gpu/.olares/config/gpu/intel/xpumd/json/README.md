# Converting old-XPUM dashboards to new-XPUM metrics

First convert metric + attribute names that have direct enough mapping
in the given dashboard JSON file from old to new XPUM names:

```bash
./convert-metrics.sh old-xpum.json dashboard.json
```

Then check which panels are using the ambiguous old-XPUM "type"
attribute, so one can update their legends to suitable OTel ones:

```bash
grep -e expr -e legendFormat dashboard.json
```

Unless new XPUM dashboard is intended to replace the old one, its
`title` and `uid` fields need also to be updated.


# Generating Grafana dashboard Helm chart configMap templates

User would need to manually load Grafana dashboard JSON file, but if
Helm chart installs it to Grafana namespace inside a suitably labeled
configMap, Grafana will load it automatically.

Dashboard JSON spec file(s) in here are used as sources for those
configMaps. After dashboard is updated in Grafana, it can be saved
again to a JSON file here, and the corresponding configMap updated
with the provided conversion script.

Switch to template dir:

```bash
cd ../templates/
```

Update dashboard configMap:

```bash
../json/convert-dashboard.sh ../json/dashboard.json
```
