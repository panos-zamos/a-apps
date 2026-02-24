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
        <header class="top-bar">
            <h1>{{.AppName}}</h1>
            {{if .Username}}
            <form action="/logout" method="POST">
                <button type="submit">Logout</button>
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
