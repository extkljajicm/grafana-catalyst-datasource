# Troubleshooting Guide - Grafana Catalyst Datasource

This guide helps diagnose and resolve common issues with the Grafana Catalyst Datasource plugin.

## Table of Contents

1. [Connection Issues](#connection-issues)
2. [Authentication Errors](#authentication-errors)
3. [Query Problems](#query-problems)
4. [Configuration Issues](#configuration-issues)
5. [Performance Issues](#performance-issues)
6. [Data Display Issues](#data-display-issues)
7. [Getting Help](#getting-help)

---

## Connection Issues

### "Unable to connect to Catalyst Center"

**Symptoms:**
- Error message: "Unable to connect to Catalyst Center"
- Health check fails with connection timeout
- No data returned from queries

**Possible Causes & Solutions:**

1. **Network connectivity**
   - ✅ Verify Grafana can reach your Catalyst Center:
     ```bash
     # From Grafana server
     curl -k https://catalyst.example.com
     ```
   - Check firewall rules
   - Verify VPN connection if required

2. **Incorrect Base URL**
   - ✅ Ensure Base URL is in format: `https://hostname` or `https://hostname:port`
   - ❌ Don't include API paths like `/dna/system/api/v1/`
   - Examples:
     - ✅ `https://catalyst.example.com`
     - ✅ `https://192.168.1.100:443`
     - ❌ `https://catalyst.example.com/dna`

3. **TLS/SSL certificate issues**
   - For self-signed certificates, enable "Skip TLS verification" in datasource config
   - ⚠️ Only use this in development/lab environments
   - For production, import the CA certificate into Grafana's trust store

4. **Proxy configuration**
   - If using a reverse proxy, include the proxy prefix in Base URL
   - Example: `https://proxy.company.com/catalyst`

---

## Authentication Errors

### "Authentication failed. Please check your credentials."

**Symptoms:**
- Error code: 401 or 403
- Message about authentication failure
- Queries return unauthorized error

**Solutions:**

1. **Verify credentials**
   - Check username and password are correct
   - Ensure user has API access in Catalyst Center
   - Required minimum role: Observer (read-only)

2. **Token expiry**
   - The plugin automatically refreshes tokens
   - If seeing frequent auth errors, there may be a token caching issue
   - Try removing and re-entering credentials

3. **API Token override**
   - If using manual API token (override field):
     - Ensure token is still valid
     - Token format should be a long string (no `X-Auth-Token:` prefix needed)
     - Clear the API Token field to use username/password instead

4. **User permissions**
   - Verify user account is not locked
   - Check user has appropriate RBAC permissions in Catalyst Center
   - Minimum required permissions:
     - Read access to Issues
     - Read access to Site Health (if using that endpoint)

---

## Query Problems

### "No data" or empty results

**Possible Causes:**

1. **Time range too narrow**
   - ✅ Widen the time range in Grafana
   - Most useful range: Last 24 hours or longer
   - Note: Issues are created at specific points in time

2. **Filters too restrictive**
   - Remove all filters and try again
   - Add filters back one at a time to identify the issue
   - Common mistakes:
     - Site ID instead of Site Name
     - Incorrect priority format (should be P1, P2, etc.)

3. **No data in Catalyst Center**
   - Verify data exists in Catalyst Center UI
   - Check the same time range and filters

4. **Wrong endpoint selected**
   - Verify endpoint in datasource config matches your query type
   - Alerts endpoint: `/dna/data/api/v1/assuranceIssues`
   - Site Health endpoint: `/dna/intent/api/v1/site-health`

### "Query timeout" or very slow queries

**Solutions:**

1. **Reduce query limit**
   - Default limit: 100
   - Try smaller limits (25-50) first
   - Increase gradually if needed

2. **Narrow time range**
   - Large time ranges return more data
   - Use appropriate intervals for your dashboard refresh rate

3. **Add filters**
   - Filter by specific sites or priorities
   - Use Status filter to exclude RESOLVED issues

4. **Network performance**
   - Check latency between Grafana and Catalyst Center
   - Consider deploying Grafana closer to Catalyst Center

---

## Configuration Issues

### "Invalid configuration" errors

**Common Issues:**

1. **Base URL validation fails**
   - Error: "Base URL is not a valid URL"
   - ✅ Must start with `http://` or `https://`
   - ✅ Must be a valid URL format
   - ❌ Don't use file:// or other protocols

2. **Missing required fields**
   - Base URL is required
   - Either username/password OR API token is required

3. **Configuration not saving**
   - Check Grafana server logs for permission errors
   - Ensure Grafana has write access to its database
   - Verify no browser extensions blocking requests

### Template variables not working

**Symptoms:**
- Variable shows "No data"
- Variable options don't update
- Queries using variables fail

**Solutions:**

1. **Check datasource selection**
   - Ensure variable uses correct datasource
   - Variable query syntax: `priorities()`, `sites(search:"text")`

2. **Verify data exists**
   - Variables fetch data from last 24 hours
   - If no recent issues, variables may be empty
   - Use `priorities()` or `issueStatuses()` for static values

3. **Search syntax**
   - Search parameter is case-insensitive
   - Examples:
     - `sites(search:"building")`
     - `devices(search:"192.168")`

---

## Performance Issues

### Plugin uses too much memory

**Solutions:**

1. **Reduce query frequency**
   - Increase dashboard refresh interval
   - Typical: 30s - 5m for production dashboards

2. **Limit data retention**
   - Use shorter time ranges
   - Archive old panels/dashboards

3. **Optimize queries**
   - Use specific filters instead of fetching all data
   - Reduce number of queries per dashboard

### Slow dashboard loading

**Solutions:**

1. **Enable caching**
   - The plugin caches API responses (5-minute TTL)
   - Multiple panels with same query share cache

2. **Reduce panel count**
   - Each panel = one query to backend
   - Consider combining data in fewer panels

3. **Use table instead of time series**
   - For alerts, table view is more efficient
   - Time series requires more processing

---

## Data Display Issues

### Data shows wrong time zone

**Solution:**
- Grafana displays times in your browser's time zone
- Check Grafana profile settings
- Check browser time zone settings

### Fields missing or incorrect names

**Possible Causes:**

1. **API response changed**
   - Catalyst Center API may vary by version
   - Check DEVELOPER_GUIDE.md for supported versions

2. **Enrichment not enabled**
   - Enable "Enrich with Site Names" in query editor
   - This resolves site IDs to names

### Values showing as "undefined" or null

**Solutions:**

1. **Field doesn't exist in response**
   - Not all issues have all fields
   - Use table view to see available fields

2. **API version mismatch**
   - Verify Catalyst Center version compatibility
   - Check plugin documentation for supported versions

---

## Debugging Steps

### Enable Debug Logging

**Frontend (Browser):**
```javascript
// Open browser console and run:
localStorage.setItem('catalyst_log_level', 'DEBUG');
// Reload Grafana
```

**Backend (Grafana Server):**
```bash
# Check Grafana logs
docker compose logs -f grafana

# Or if running directly:
tail -f /var/log/grafana/grafana.log
```

### Test API Connectivity

```bash
# 1. Test base URL
curl -k https://catalyst.example.com

# 2. Test authentication
curl -k -X POST https://catalyst.example.com/dna/system/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -u username:password

# 3. Test issues endpoint (replace TOKEN with actual token)
curl -k https://catalyst.example.com/dna/data/api/v1/assuranceIssues?limit=1 \
  -H "X-Auth-Token: TOKEN"
```

### Capture Network Traffic

Use browser DevTools:
1. Open DevTools (F12)
2. Go to Network tab
3. Filter by "Fetch/XHR"
4. Execute query
5. Check request/response details

---

## Getting Help

### Before Opening an Issue

1. Check this troubleshooting guide
2. Review DEVELOPER_GUIDE.md
3. Enable debug logging and capture logs
4. Test API connectivity directly

### When Opening an Issue

Include:
- Plugin version
- Grafana version
- Catalyst Center version
- Error messages (full text)
- Debug logs (if available)
- Steps to reproduce
- Expected vs actual behavior

### Community Resources

- GitHub Issues: https://github.com/extkljajicm/grafana-catalyst-datasource/issues
- Plugin Documentation: [README.md](src/README.md)
- Grafana Community: https://community.grafana.com

---

## Known Issues

### Issue: Site Health metrics return zero

**Status:** Fixed in v1.2.0

**Workaround:** Update to latest version

### Issue: Priority filter accepts multiple values but API doesn't

**Status:** Implemented in plugin (filters applied after fetch)

**Note:** This is expected behavior due to API limitation

---

## License

Apache-2.0 © extkljajicm
