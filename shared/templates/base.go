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
        <aside class="sidebar">
            <nav>
                <div class="mb-md">
                    <div class="muted">Application</div>
                    <div>{{.AppName}}</div>
                </div>
                <div class="mb-md">
                    <a href="/">Home</a>
                </div>
                {{if .Username}}
                <div class="mb-sm muted">Signed in as {{.Username}}</div>
                <form action="/logout" method="POST">
                    <button type="submit">Logout</button>
                </form>
                {{end}}
            </nav>
        </aside>
        <main class="content">
            <div class="content-inner">
                <header class="mb-lg">
                    <h1>{{.Title}}</h1>
                </header>
                {{.Content}}
            </div>
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
        <aside class="sidebar">
            <nav>
                <div class="mb-md">
                    <div class="muted">Application</div>
                    <div>{{.AppName}}</div>
                </div>
            </nav>
        </aside>
        <main class="content">
            <div class="content-inner">
                <header class="mb-lg">
                    <h1>Sign in</h1>
                    <p class="muted">Use your account to continue.</p>
                </header>
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
            </div>
        </main>
    </div>
</body>
</html>`
