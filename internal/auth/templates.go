package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DocuSeal CLI Setup</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --docuseal-blue: #236cff;
            --docuseal-blue-dark: #1a56cc;
            --docuseal-blue-light: #e8f0ff;
            --seal-tan: #AA968C;
            --seal-light: #C8AF9B;
            --seal-dark: #8C7873;
            --bg-white: #ffffff;
            --bg-gray: #f8fafc;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --text-muted: #94a3b8;
            --border: #e2e8f0;
            --border-focus: #236cff;
            --error: #dc2626;
            --success: #16a34a;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: linear-gradient(135deg, var(--bg-gray) 0%, #f1f5f9 100%);
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem;
        }

        .container {
            width: 100%;
            max-width: 480px;
        }

        .header {
            text-align: center;
            margin-bottom: 2rem;
            animation: fadeInDown 0.6s ease-out;
        }

        .logo {
            width: 64px;
            height: 64px;
            margin: 0 auto 1.25rem;
            animation: bounceIn 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
        }

        .logo svg {
            width: 100%;
            height: 100%;
            filter: drop-shadow(0 4px 12px rgba(170, 150, 140, 0.3));
        }

        @keyframes bounceIn {
            0% { transform: scale(0); opacity: 0; }
            50% { transform: scale(1.1); }
            100% { transform: scale(1); opacity: 1; }
        }

        .brand {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            margin-bottom: 0.5rem;
        }

        .brand-text {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--text-primary);
            letter-spacing: -0.02em;
        }

        .cli-badge {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.6875rem;
            font-weight: 500;
            background: var(--docuseal-blue);
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        h1 {
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--text-secondary);
            margin-top: 0.75rem;
        }

        .card {
            background: var(--bg-white);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 2rem;
            box-shadow:
                0 1px 3px rgba(0, 0, 0, 0.04),
                0 6px 16px rgba(0, 0, 0, 0.04);
            animation: fadeInUp 0.6s ease-out 0.1s both;
        }

        .form-group {
            margin-bottom: 1.5rem;
        }

        label {
            display: block;
            font-size: 0.875rem;
            font-weight: 500;
            color: var(--text-primary);
            margin-bottom: 0.5rem;
        }

        input {
            width: 100%;
            padding: 0.75rem 1rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            background: var(--bg-gray);
            border: 1.5px solid var(--border);
            border-radius: 10px;
            color: var(--text-primary);
            transition: all 0.2s ease;
        }

        input::placeholder {
            color: var(--text-muted);
        }

        input:focus {
            outline: none;
            border-color: var(--border-focus);
            background: var(--bg-white);
            box-shadow: 0 0 0 3px rgba(35, 108, 255, 0.1);
        }

        .input-hint {
            font-size: 0.8125rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
            line-height: 1.5;
        }

        .input-hint code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--bg-gray);
            padding: 0.125rem 0.375rem;
            border-radius: 4px;
            font-size: 0.75rem;
            color: var(--docuseal-blue);
        }

        .btn-group {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 0.75rem;
            margin-top: 1.75rem;
        }

        button {
            padding: 0.875rem 1.5rem;
            font-family: 'Inter', sans-serif;
            font-size: 0.9375rem;
            font-weight: 600;
            border-radius: 10px;
            cursor: pointer;
            transition: all 0.2s ease;
            border: none;
        }

        .btn-secondary {
            background: var(--bg-white);
            border: 1.5px solid var(--border);
            color: var(--text-secondary);
        }

        .btn-secondary:hover {
            background: var(--bg-gray);
            border-color: var(--text-muted);
            color: var(--text-primary);
        }

        .btn-primary {
            background: var(--docuseal-blue);
            color: white;
            box-shadow: 0 2px 8px rgba(35, 108, 255, 0.25);
        }

        .btn-primary:hover {
            background: var(--docuseal-blue-dark);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(35, 108, 255, 0.35);
        }

        .btn-primary:active {
            transform: translateY(0);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none !important;
        }

        .status {
            margin-top: 1.25rem;
            padding: 0.875rem 1rem;
            border-radius: 10px;
            font-size: 0.875rem;
            display: none;
            align-items: center;
            gap: 0.625rem;
            animation: slideDown 0.3s ease;
        }

        @keyframes slideDown {
            from { opacity: 0; transform: translateY(-8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .status.show { display: flex; }

        .status.loading {
            background: var(--docuseal-blue-light);
            color: var(--docuseal-blue-dark);
        }

        .status.success {
            background: #dcfce7;
            color: #166534;
        }

        .status.error {
            background: #fef2f2;
            color: #dc2626;
        }

        .spinner {
            width: 16px;
            height: 16px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .help-section {
            margin-top: 1.75rem;
            padding-top: 1.75rem;
            border-top: 1px solid var(--border);
        }

        .help-title {
            font-size: 0.75rem;
            font-weight: 600;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.08em;
            margin-bottom: 1rem;
        }

        .help-steps {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
        }

        .help-step {
            display: flex;
            align-items: flex-start;
            gap: 0.75rem;
            font-size: 0.875rem;
            color: var(--text-secondary);
            line-height: 1.5;
        }

        .step-num {
            flex-shrink: 0;
            width: 22px;
            height: 22px;
            background: var(--docuseal-blue-light);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 0.75rem;
            font-weight: 600;
            color: var(--docuseal-blue);
        }

        .footer {
            text-align: center;
            margin-top: 2rem;
            font-size: 0.8125rem;
            color: var(--text-muted);
            animation: fadeIn 0.6s ease-out 0.3s both;
        }

        .footer a {
            color: var(--text-muted);
            text-decoration: none;
            transition: color 0.2s;
        }

        .footer a:hover {
            color: #2D52F6;
        }

        .github-link {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }

        .github-link svg {
            opacity: 0.7;
            transition: opacity 0.2s;
        }

        .github-link:hover svg {
            opacity: 1;
        }

        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        @keyframes fadeInDown {
            from { opacity: 0; transform: translateY(-16px); }
            to { opacity: 1; transform: translateY(0); }
        }
        @keyframes fadeInUp {
            from { opacity: 0; transform: translateY(16px); }
            to { opacity: 1; transform: translateY(0); }
        }

        @media (max-width: 480px) {
            .card { padding: 1.5rem; }
            .btn-group { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">
                <svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
                    <path fill="#AA968C" d="M504,248c0,128.13-111.03,240-248,240S8,376.13,8,248s111.03-224,248-224S504,119.87,504,248z"/>
                    <path fill="#AA968C" d="M256,24C119.03,24,8,119.87,8,248c0,24.63,4.14,48.65,11.74,71.4c25.52,34.29,66.23,56.6,112.26,56.6c53.89,0,100.6-30.5,124-75.13c23.4,44.64,70.11,75.13,124,75.13c46.03,0,86.74-22.31,112.26-56.6c7.6-22.75,11.74-46.77,11.74-71.4C504,119.87,392.97,24,256,24z"/>
                    <circle fill="#C8AF9B" cx="256" cy="352" r="136"/>
                    <circle fill="#464655" cx="132" cy="204" r="28"/>
                    <circle fill="#464655" cx="380" cy="204" r="28"/>
                    <path fill="#464655" d="M270,284.5c-7.67,10.74-20.23,10.74-27.9,0l-12.1-16.94c-7.67-10.74-3.15-19.53,10.05-19.53h32c13.2,0,17.72,8.79,10.05,19.53L270,284.5z"/>
                    <path fill="#AA968C" d="M351,400c-34.34,0-52-48-95-48c-43.14,0-60.75,48-95,48c-15.4,0-29.07-5.85-40.17-25.2C131.6,439.03,187.9,488,255.88,488c67.99,0,124.29-48.97,135.26-113.2C380.04,394.15,366.37,400,351,400z"/>
                    <g fill="#8C7873">
                        <path d="M32,424c-3.17,0-6.18-1.91-7.43-5.03c-1.64-4.11,0.36-8.76,4.46-10.4l160-64c4.06-1.62,8.76,0.35,10.4,4.46c1.64,4.1-0.36,8.76-4.46,10.4l-160,64C34,423.81,33,424,32,424z"/>
                        <path d="M16,376c-3.55,0-6.78-2.38-7.73-5.97c-1.13-4.27,1.42-8.65,5.7-9.77l152-40c4.29-1.12,8.65,1.43,9.77,5.7c1.13,4.27-1.42,8.65-5.7,9.77l-152,40C17.35,375.91,16.67,376,16,376z"/>
                        <path d="M8,336c-3.81,0-7.19-2.73-7.87-6.61c-0.77-4.35,2.13-8.5,6.48-9.27l136-24c4.33-0.77,8.51,2.14,9.27,6.49c0.77,4.35-2.13,8.5-6.48,9.27l-136,24C8.92,335.96,8.45,336,8,336z"/>
                        <path d="M480,424c3.17,0,6.18-1.91,7.43-5.03c1.64-4.11-0.36-8.76-4.46-10.4l-160-64c-4.06-1.62-8.76,0.35-10.4,4.46c-1.64,4.1,0.36,8.76,4.46,10.4l160,64C478,423.81,479,424,480,424z"/>
                        <path d="M496,376c3.55,0,6.78-2.38,7.73-5.97c1.13-4.27-1.42-8.65-5.7-9.77l-152-40c-4.29-1.12-8.65,1.43-9.77,5.7c-1.13,4.27,1.42,8.65,5.7,9.77l152,40C494.65,375.91,495.33,376,496,376z"/>
                        <path d="M504,336c3.81,0,7.19-2.73,7.87-6.61c0.77-4.35-2.13-8.5-6.48-9.27l-136-24c-4.33-0.77-8.51,2.14-9.27,6.49c-0.77,4.35,2.13,8.5,6.48,9.27l136,24C503.08,335.96,503.55,336,504,336z"/>
                    </g>
                </svg>
            </div>
            <div class="brand">
                <span class="brand-text">DocuSeal</span>
                <span class="cli-badge">CLI</span>
            </div>
            <h1>Connect your instance</h1>
        </div>

        <div class="card">
            <form id="setupForm" autocomplete="off">
                <div class="form-group">
                    <label for="baseUrl">Instance URL</label>
                    <input
                        type="url"
                        id="baseUrl"
                        name="baseUrl"
                        placeholder="https://docuseal.example.com"
                        required
                    >
                    <div class="input-hint">Your self-hosted DocuSeal URL or <code>https://api.docuseal.com</code> for cloud</div>
                </div>

                <div class="form-group">
                    <label for="apiKey">API Key</label>
                    <input
                        type="password"
                        id="apiKey"
                        name="apiKey"
                        placeholder="Enter your API key"
                        required
                    >
                    <div class="input-hint">Found in Settings → API in your DocuSeal dashboard</div>
                </div>

                <div class="btn-group">
                    <button type="button" id="testBtn" class="btn-secondary">Test</button>
                    <button type="submit" id="submitBtn" class="btn-primary">Connect</button>
                </div>

                <div id="status" class="status"></div>
            </form>

            <div class="help-section">
                <div class="help-title">Quick Setup</div>
                <div class="help-steps">
                    <div class="help-step">
                        <span class="step-num">1</span>
                        <span>Open your DocuSeal instance in a browser</span>
                    </div>
                    <div class="help-step">
                        <span class="step-num">2</span>
                        <span>Go to <strong>Settings → API</strong> and copy your key</span>
                    </div>
                    <div class="help-step">
                        <span class="step-num">3</span>
                        <span>Paste your URL and API key above</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/docuseal-cli" target="_blank" class="github-link">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>

    <script>
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');
        const csrfToken = '{{.CSRFToken}}';

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                const icon = type === 'success' ? '✓' : '✕';
                status.innerHTML = '<span style="font-weight:600">' + icon + '</span><span>' + message + '</span>';
            }
        }

        function getFormData() {
            return {
                base_url: document.getElementById('baseUrl').value.trim().replace(/\/$/, ''),
                api_key: document.getElementById('apiKey').value.trim()
            };
        }

        testBtn.addEventListener('click', async () => {
            const data = getFormData();
            if (!data.base_url || !data.api_key) {
                showStatus('error', 'Please fill in all fields');
                return;
            }

            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection...');

            try {
                const response = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                showStatus(result.success ? 'success' : 'error', result.success ? result.message : result.error);
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
            }
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const data = getFormData();
            if (!data.base_url || !data.api_key) {
                showStatus('error', 'Please fill in all fields');
                return;
            }

            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Connecting...');

            try {
                const response = await fetch('/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                if (result.success) {
                    showStatus('success', 'Connected! Redirecting...');
                    setTimeout(() => window.location.href = '/success', 800);
                } else {
                    showStatus('error', result.error);
                    testBtn.disabled = false;
                    submitBtn.disabled = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
                testBtn.disabled = false;
                submitBtn.disabled = false;
            }
        });
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connected - DocuSeal CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --docuseal-blue: #236cff;
            --docuseal-blue-dark: #1a56cc;
            --seal-tan: #AA968C;
            --bg-white: #ffffff;
            --bg-gray: #f8fafc;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --text-muted: #94a3b8;
            --border: #e2e8f0;
            --success: #16a34a;
            --success-light: #dcfce7;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: 'Inter', -apple-system, sans-serif;
            background: linear-gradient(135deg, var(--bg-gray) 0%, #f1f5f9 100%);
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem;
        }

        .container {
            width: 100%;
            max-width: 520px;
            text-align: center;
        }

        .success-icon {
            width: 88px;
            height: 88px;
            margin: 0 auto 1.5rem;
            background: var(--success);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 8px 24px rgba(22, 163, 74, 0.3);
            animation: scaleIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
        }

        @keyframes scaleIn {
            0% { transform: scale(0); }
            100% { transform: scale(1); }
        }

        .success-icon svg {
            width: 44px;
            height: 44px;
            stroke: white;
            stroke-width: 3;
            fill: none;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .success-icon svg path {
            stroke-dasharray: 60;
            stroke-dashoffset: 60;
            animation: checkDraw 0.4s ease-out 0.3s forwards;
        }

        @keyframes checkDraw {
            to { stroke-dashoffset: 0; }
        }

        h1 {
            font-size: 1.75rem;
            font-weight: 700;
            color: var(--text-primary);
            margin-bottom: 0.5rem;
            animation: fadeIn 0.5s ease-out 0.2s both;
        }

        .subtitle {
            font-size: 1.0625rem;
            color: var(--text-secondary);
            margin-bottom: 2rem;
            animation: fadeIn 0.5s ease-out 0.3s both;
        }

        .terminal {
            background: var(--bg-white);
            border: 1px solid var(--border);
            border-radius: 12px;
            overflow: hidden;
            text-align: left;
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
            animation: fadeInUp 0.5s ease-out 0.4s both;
            margin-bottom: 1.5rem;
        }

        .terminal-header {
            background: var(--bg-gray);
            padding: 0.75rem 1rem;
            display: flex;
            gap: 0.5rem;
            border-bottom: 1px solid var(--border);
        }

        .terminal-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            background: var(--border);
        }

        .terminal-dot:nth-child(1) { background: #ef4444; }
        .terminal-dot:nth-child(2) { background: #fbbf24; }
        .terminal-dot:nth-child(3) { background: #22c55e; }

        .terminal-body {
            padding: 1.25rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            line-height: 1.8;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .terminal-prompt {
            color: var(--docuseal-blue);
            font-weight: 500;
        }

        .terminal-output {
            color: var(--text-secondary);
            padding-left: 1.25rem;
            margin: 0.25rem 0 0.75rem;
        }

        .terminal-output.success {
            color: var(--success);
        }

        .terminal-cursor {
            display: inline-block;
            width: 8px;
            height: 16px;
            background: var(--docuseal-blue);
            animation: blink 1.2s step-end infinite;
        }

        @keyframes blink {
            0%, 50% { opacity: 1; }
            50.01%, 100% { opacity: 0; }
        }

        .message {
            background: var(--success-light);
            border-radius: 10px;
            padding: 1rem 1.25rem;
            font-size: 0.9375rem;
            color: #166534;
            animation: fadeIn 0.5s ease-out 0.5s both;
        }

        .footer {
            text-align: center;
            margin-top: 2rem;
            font-size: 0.8125rem;
            color: var(--text-muted);
            animation: fadeIn 0.5s ease-out 0.6s both;
        }

        .footer a {
            color: var(--text-muted);
            text-decoration: none;
            transition: color 0.2s;
        }

        .footer a:hover {
            color: #2D52F6;
        }

        .github-link {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }

        .github-link svg {
            opacity: 0.7;
            transition: opacity 0.2s;
        }

        .github-link:hover svg {
            opacity: 1;
        }

        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
        @keyframes fadeInUp {
            from { opacity: 0; transform: translateY(16px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 24 24">
                <path d="M5 13l4 4L19 7"/>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">DocuSeal CLI is now connected</p>

        <div class="terminal">
            <div class="terminal-header">
                <span class="terminal-dot"></span>
                <span class="terminal-dot"></span>
                <span class="terminal-dot"></span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span>docuseal auth status</span>
                </div>
                <div class="terminal-output success">✓ Connected to {{.InstanceURL}}</div>

                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span>docuseal templates list</span>
                </div>
                <div class="terminal-output">ID    NAME              FOLDER</div>

                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            You can now <strong>close this window</strong> and return to your terminal.
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/docuseal-cli" target="_blank" class="github-link">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>

    <script>
        fetch('/complete', { method: 'POST' }).catch(() => {});
        setTimeout(() => window.close(), 4000);
    </script>
</body>
</html>`
