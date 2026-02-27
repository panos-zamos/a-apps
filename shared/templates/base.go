package templates

// BaseHTML provides a base HTML template with shared layout and HTMX
const BaseHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/custom.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <div class="app">
                <header class="top-bar row space-between" style="align-items: center;">
                        <h1>{{.AppName}}</h1>
                        {{if .Username}}
                        <form action="/logout" method="POST" style="margin:0;">
                                <button type="submit" title="Log out" style="background:none;border:none;padding:0;cursor:pointer;">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-power">
                                        <path d="M18.36 6.64a9 9 0 1 1-12.73 0"/>
                                        <line x1="12" y1="2" x2="12" y2="12"/>
                                    </svg>
                                </button>
                        </form>
                        {{end}}
                </header>
        <main class="content">
            {{.Content}}
        </main>
    </div>
</body>
</html>`

// LoginHTML provides a login form template
const LoginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - {{.AppName}}</title>
    <link rel="stylesheet" href="/custom.css">
</head>
<body>
    <div class="app">
        <header class="top-bar">
            <h1>{{.AppName}}</h1>
        </header>
        <main class="content">
            <h2 class="mb-sm">Sign in</h2>
            <p class="muted mb-md">Use your account to continue.</p>
            {{if .Error}}
            <div class="panel mb-md">
                <p>{{.Error}}</p>
            </div>
            {{end}}
            <div class="panel">
                <form action="/login" method="POST">
                    <label for="username">Username</label>
                    <input id="username" name="username" type="text" required>
                    <div class="mt-md">
                        <label for="password">Password</label>
                        <input id="password" name="password" type="password" required>
                    </div>
                    <div class="mt-md">
                        <button type="submit" class="primary">Sign in</button>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`
