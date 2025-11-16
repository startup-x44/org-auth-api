#!/usr/bin/env node

/**
 * Simple OAuth Callback Server for Testing
 * Listens on http://localhost:3000/callback
 * Displays authorization code and state parameters
 */

const http = require('http');
const url = require('url');

const PORT = 3001;

const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);
  
  if (parsedUrl.pathname === '/callback') {
    const { code, state, error, error_description } = parsedUrl.query;
    
    res.writeHead(200, { 'Content-Type': 'text/html' });
    
    if (error) {
      res.end(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>OAuth Error</title>
          <style>
            body {
              font-family: system-ui, -apple-system, sans-serif;
              max-width: 800px;
              margin: 50px auto;
              padding: 20px;
              background: #f5f5f5;
            }
            .container {
              background: white;
              padding: 40px;
              border-radius: 8px;
              box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            }
            h1 { color: #dc2626; }
            .error { 
              background: #fee2e2; 
              border-left: 4px solid #dc2626;
              padding: 15px;
              margin: 20px 0;
            }
            code {
              background: #f3f4f6;
              padding: 2px 6px;
              border-radius: 4px;
              font-family: monospace;
            }
            .back-btn {
              display: inline-block;
              margin-top: 20px;
              padding: 10px 20px;
              background: #3b82f6;
              color: white;
              text-decoration: none;
              border-radius: 6px;
            }
            .back-btn:hover {
              background: #2563eb;
            }
          </style>
        </head>
        <body>
          <div class="container">
            <h1>❌ Authorization Failed</h1>
            <div class="error">
              <strong>Error:</strong> <code>${error}</code><br>
              <strong>Description:</strong> ${error_description || 'No description provided'}
            </div>
            <p>The authorization request was denied or failed.</p>
            <a href="http://localhost:5173/oauth/client-apps" class="back-btn">← Back to OAuth Apps</a>
          </div>
        </body>
        </html>
      `);
    } else if (code) {
      res.end(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>OAuth Callback</title>
          <style>
            body {
              font-family: system-ui, -apple-system, sans-serif;
              max-width: 800px;
              margin: 50px auto;
              padding: 20px;
              background: #f5f5f5;
            }
            .container {
              background: white;
              padding: 40px;
              border-radius: 8px;
              box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            }
            h1 { color: #16a34a; }
            .success { 
              background: #dcfce7; 
              border-left: 4px solid #16a34a;
              padding: 15px;
              margin: 20px 0;
            }
            .code-block {
              background: #1e293b;
              color: #e2e8f0;
              padding: 20px;
              border-radius: 8px;
              overflow-x: auto;
              margin: 20px 0;
              font-family: 'Monaco', 'Courier New', monospace;
              font-size: 14px;
            }
            .copy-btn {
              background: #3b82f6;
              color: white;
              border: none;
              padding: 8px 16px;
              border-radius: 4px;
              cursor: pointer;
              font-size: 14px;
              margin-top: 10px;
            }
            .copy-btn:hover {
              background: #2563eb;
            }
            .next-steps {
              background: #eff6ff;
              border-left: 4px solid #3b82f6;
              padding: 15px;
              margin: 20px 0;
            }
            .next-steps h3 {
              margin-top: 0;
              color: #1e40af;
            }
            code {
              background: #f3f4f6;
              padding: 2px 6px;
              border-radius: 4px;
              font-family: monospace;
            }
            .back-btn {
              display: inline-block;
              margin-top: 20px;
              padding: 10px 20px;
              background: #3b82f6;
              color: white;
              text-decoration: none;
              border-radius: 6px;
            }
            .back-btn:hover {
              background: #2563eb;
            }
          </style>
        </head>
        <body>
          <div class="container">
            <h1>✅ Authorization Successful!</h1>
            
            <div class="success">
              Your application has been authorized successfully.
            </div>
            
            <h3>Authorization Code:</h3>
            <div class="code-block" id="codeBlock">${code}</div>
            <button class="copy-btn" onclick="copyCode()">📋 Copy Code</button>
            
            <h3>State Parameter:</h3>
            <div class="code-block">${state || 'Not provided'}</div>
            
            <div class="next-steps">
              <h3>Next Steps:</h3>
              <p>Exchange this authorization code for an access token:</p>
              <div class="code-block">
curl -X POST http://localhost:8080/oauth/token \\
  -H "Content-Type: application/x-www-form-urlencoded" \\
  -d "grant_type=authorization_code" \\
  -d "code=${code}" \\
  -d "redirect_uri=http://localhost:3000/callback" \\
  -d "client_id=YOUR_CLIENT_ID" \\
  -d "client_secret=YOUR_CLIENT_SECRET" \\
  -d "code_verifier=YOUR_CODE_VERIFIER"
              </div>
            </div>
            
            <a href="http://localhost:5173/oauth/client-apps" class="back-btn">← Back to OAuth Apps</a>
          </div>
          
          <script>
            function copyCode() {
              const code = document.getElementById('codeBlock').textContent;
              navigator.clipboard.writeText(code).then(() => {
                const btn = event.target;
                const originalText = btn.textContent;
                btn.textContent = '✓ Copied!';
                setTimeout(() => {
                  btn.textContent = originalText;
                }, 2000);
              });
            }
          </script>
        </body>
        </html>
      `);
    } else {
      res.end(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>Invalid Callback</title>
          <style>
            body {
              font-family: system-ui, -apple-system, sans-serif;
              max-width: 800px;
              margin: 50px auto;
              padding: 20px;
              background: #f5f5f5;
            }
            .container {
              background: white;
              padding: 40px;
              border-radius: 8px;
              box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            }
            h1 { color: #ea580c; }
          </style>
        </head>
        <body>
          <div class="container">
            <h1>⚠️ Invalid Callback</h1>
            <p>No authorization code or error was received.</p>
          </div>
        </body>
        </html>
      `);
    }
  } else {
    res.writeHead(404, { 'Content-Type': 'text/html' });
    res.end(`
      <!DOCTYPE html>
      <html>
      <head>
        <title>404 Not Found</title>
        <style>
          body {
            font-family: system-ui, -apple-system, sans-serif;
            text-align: center;
            padding: 50px;
          }
          h1 { color: #6b7280; }
        </style>
      </head>
      <body>
        <h1>404 - Not Found</h1>
        <p>The OAuth callback endpoint is at <code>/callback</code></p>
      </body>
      </html>
    `);
  }
});

server.listen(PORT, () => {
  console.log('');
  console.log('🚀 OAuth Callback Server Running');
  console.log('================================');
  console.log(`📍 Callback URL: http://localhost:${PORT}/callback`);
  console.log('');
  console.log('Add this URL to your OAuth client app redirect URIs.');
  console.log('Press Ctrl+C to stop the server.');
  console.log('');
});
