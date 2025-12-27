# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the DocuSeal CLI.

## Authentication Problems

### Invalid API Key

**Symptoms:**
- `authentication failed: invalid API key or insufficient permissions`
- HTTP 401 or 403 errors

**Solutions:**

1. **Verify your API key is correct:**
   ```bash
   docuseal auth status
   ```

2. **Check if you're using the right instance URL:**
   - For DocuSeal Cloud: `https://api.docuseal.com`
   - For self-hosted: Your instance URL (e.g., `https://docuseal.example.com`)

3. **Re-configure authentication:**
   ```bash
   docuseal auth logout
   docuseal auth setup --url YOUR_URL --api-key YOUR_KEY
   ```

4. **If using environment variables, check they're set correctly:**
   ```bash
   echo $DOCUSEAL_URL
   echo $DOCUSEAL_API_KEY
   ```
   Note: Environment variables override keychain credentials.

### Expired or Rotated Credentials

**Symptoms:**
- `WARNING: Credentials are X days old`
- Previously working commands now fail with auth errors

**Solution:**
Rotate your API key:
1. Generate a new API key in your DocuSeal dashboard
2. Update the CLI:
   ```bash
   docuseal auth setup --url YOUR_URL --api-key NEW_KEY
   ```

### Keychain Access Issues

**Symptoms:**
- `failed to open keyring`
- `failed to save credentials`

**Platform-specific solutions:**

**macOS:**
- Grant Terminal/iTerm2 access to Keychain when prompted
- If denied, go to System Preferences → Security & Privacy → Privacy → Full Disk Access

**Linux:**
- Install `libsecret` or `gnome-keyring`:
  ```bash
  # Ubuntu/Debian
  sudo apt-get install libsecret-1-0

  # Fedora
  sudo dnf install libsecret
  ```

**Windows:**
- Ensure Windows Credential Manager is accessible
- Run PowerShell/Terminal as your user (not Administrator)

## Connection Issues

### Network Timeouts

**Symptoms:**
- `request failed: context deadline exceeded`
- Operations hang for 30 seconds then fail

**Solutions:**

1. **Check your internet connection**

2. **Verify the instance URL is reachable:**
   ```bash
   curl https://your-instance.com/api
   ```

3. **Check for proxy/firewall issues:**
   - If behind a corporate proxy, set proxy environment variables:
     ```bash
     export HTTP_PROXY=http://proxy.example.com:8080
     export HTTPS_PROXY=http://proxy.example.com:8080
     ```

4. **For self-hosted instances, verify SSL certificates are valid:**
   - If using self-signed certificates (development only):
     ```bash
     # NOT RECOMMENDED FOR PRODUCTION
     curl -k https://your-instance.com/api
     ```

### Circuit Breaker Activated

**Symptoms:**
- `circuit breaker open: too many consecutive failures, requests temporarily blocked`

**What it means:**
The CLI detected 5+ consecutive failures and paused requests to prevent overwhelming the server.

**Solutions:**

1. **Wait 30 seconds** - the circuit breaker resets automatically

2. **Check server status:**
   ```bash
   curl https://your-instance.com/api
   ```

3. **If server is healthy, retry your command**

4. **If problem persists, check server logs for issues**

### SSL/TLS Certificate Errors

**Symptoms:**
- `x509: certificate signed by unknown authority`
- `tls: failed to verify certificate`

**Solutions:**

1. **For production: Ensure your instance has a valid SSL certificate**
   - Use Let's Encrypt or a trusted CA

2. **For development only (NOT production):**
   - Fix the certificate issue rather than disabling verification
   - Self-signed certificates should be added to your system's trust store

3. **Verify the URL uses HTTPS:**
   ```bash
   docuseal auth status
   ```

## Rate Limiting

**Symptoms:**
- `rate limit exceeded, retry after X seconds`
- HTTP 429 errors

**What it means:**
You've exceeded the API rate limit for your instance.

**Solutions:**

1. **Wait for the retry period** - the CLI automatically retries with exponential backoff

2. **Reduce request frequency:**
   - Use pagination limits (`--limit`) to fetch fewer results
   - Add delays between batch operations

3. **Check your plan's rate limits** in the DocuSeal dashboard

4. **For self-hosted instances, adjust rate limits** in your server configuration

## File Upload Problems

### File Size Exceeded

**Symptoms:**
- `file size X bytes exceeds maximum allowed size of Y bytes (50MB)`

**Solution:**
- **Maximum file size: 50MB** for PDF/DOCX files
- **Maximum HTML size: 10MB**
- Compress large PDFs before uploading
- Split large documents into smaller templates

### Invalid File Format

**Symptoms:**
- `failed to read file`
- `unexpected API response format`

**Solutions:**

1. **Verify file format:**
   - Use `create-pdf` for PDF files only
   - Use `create-docx` for Word documents only
   - Use `create-html` for HTML content

2. **Check file integrity:**
   ```bash
   file your-document.pdf  # Should show "PDF document"
   ```

3. **Ensure file exists and is readable:**
   ```bash
   ls -lh your-document.pdf
   ```

### HTML Content Validation Errors

**Symptoms:**
- `HTML content does not appear to contain valid HTML tags`
- `HTML content cannot be empty`

**Solutions:**

1. **Verify HTML contains proper tags:**
   - Must include `<` and `>` characters
   - Cannot be empty

2. **Escape special characters if passing HTML inline:**
   ```bash
   docuseal templates create-html --name "Form" --html '<html><body>Content</body></html>'
   ```

## Common Error Messages

### "Not authenticated"

**Full error:** `not authenticated: run 'docuseal auth setup' or set DOCUSEAL_API_KEY and DOCUSEAL_URL environment variables`

**Solution:**
```bash
docuseal auth setup
```

### "Validation error on field 'email'"

**Cause:** Invalid email address format

**Solution:** Use valid email addresses for submitters:
```bash
docuseal submissions create --submitters "user@example.com:Signer"
```

### "Failed to verify credentials"

**Cause:** API key works but cannot access resources

**Solutions:**
- Check API key permissions in DocuSeal dashboard
- Ensure API key hasn't been revoked
- Verify you're using the correct DocuSeal instance

### "Unexpected API response format"

**Causes:**
- DocuSeal version incompatibility
- Network proxy modifying responses
- Server returning non-JSON content

**Solutions:**
1. **Check your DocuSeal version** (CLI supports DocuSeal v1.0+)
2. **Verify no proxy is modifying responses:**
   ```bash
   unset HTTP_PROXY HTTPS_PROXY
   ```
3. **Test API directly:**
   ```bash
   curl -H "X-Auth-Token: YOUR_KEY" https://your-instance.com/api/templates
   ```

## Debug Mode

### Viewing Request Details

**Check authentication status:**
```bash
docuseal auth status
```

**Use JSON output for detailed error information:**
```bash
docuseal templates list --output json
```

**Test connectivity:**
```bash
docuseal auth whoami
```

### Environment Variables for Debugging

```bash
# Override authentication
export DOCUSEAL_URL=https://your-instance.com
export DOCUSEAL_API_KEY=your-api-key

# Force JSON output
export DOCUSEAL_OUTPUT=json

# Disable color output
export DOCUSEAL_COLOR=never
export NO_COLOR=1
```

## Reset and Clean Installation

### Complete Credential Reset

```bash
# 1. Remove stored credentials
docuseal auth logout

# 2. Clear environment variables
unset DOCUSEAL_URL DOCUSEAL_API_KEY DOCUSEAL_OUTPUT DOCUSEAL_COLOR

# 3. Re-setup authentication
docuseal auth setup
```

### Verify Installation

```bash
# Check version
docuseal version

# Test basic functionality
docuseal auth status
docuseal templates list --limit 5
```

## Platform-Specific Issues

### macOS

**Issue:** "docuseal cannot be opened because the developer cannot be verified"

**Solution:**
```bash
xattr -d com.apple.quarantine /path/to/docuseal
```

### Linux

**Issue:** "command not found"

**Solution:**
```bash
# Ensure binary is in PATH
echo $PATH
# Add to PATH if needed
export PATH="$PATH:/path/to/docuseal/bin"
```

### Windows

**Issue:** PowerShell execution policy

**Solution:**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Still Having Issues?

If your problem isn't covered here:

1. **Check the version:**
   ```bash
   docuseal version
   ```

2. **Verify your instance is running:**
   - DocuSeal Cloud: Check [status page](https://status.docuseal.com)
   - Self-hosted: Check server logs

3. **Review recent changes:**
   - Did you update DocuSeal instance?
   - Did you rotate API keys?
   - Did network/firewall rules change?

4. **Collect diagnostic information:**
   ```bash
   docuseal version
   docuseal auth status
   echo "OS: $(uname -s)"
   echo "URL: $DOCUSEAL_URL"
   ```

5. **Report the issue:**
   - GitHub: https://github.com/salmonumbrella/docuseal-cli/issues
   - Include: CLI version, OS, error message, command that failed
   - DO NOT share your API key
