async function checkLoginStatus() {
    try {
        const response = await fetch('/api/check-login');
        const data = await response.json();
        return data.isLoggedIn;
    } catch (error) {
        console.error('Error checking login status:', error);
        return false;
    }
}

async function fetchLoginContent() {
    const isLoggedIn = await checkLoginStatus();
    const homeLink = isLoggedIn ? '<p class="home-link"><a href="#/home">← Go to Homepage</a></p>' : '';

    return `
        <div class="auth-container">
        <h2>Real-Time Forum</h2>
            <h1>Login</h1>
            <!-- Traditional Login Form -->
            <form id="login-form" onsubmit="handleLogin(event)">
                <label for="identifier">Email or Nickname:</label>
                <input type="text" id="identifier" name="identifier" placeholder="Enter email or nickname" required>
                <br>
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
                <br>
                <button type="submit">Login</button>
            </form>
            <p>Don't have an account? <a href="#/register">Register here</a></p>
            ${homeLink}
        </div>
    `;
}

async function fetchRegisterContent() {
    return `
        <div class="auth-container">
            <h1>Register</h1>
            <!-- Traditional Registration Form -->
           <form id="registerForm" onsubmit="handleRegister(event)">
                <label for="nickname">Nickname:</label>
                <input type="text" id="nickname" name="nickname" required>
                <br>

                <label for="age">Age:</label>
                <input type="number" id="age" name="age" min="1" required>
                <br>

                <div class="gender-group">
                    <label class="gender-label">Gender:</label>
                    <div class="gender-options">
                        <label class="gender-option">
                            <input type="radio" name="gender" value="male" required>
                            <span class="radio-custom"></span>
                            <i class="fas fa-mars"></i>
                            Male
                        </label>
                        <label class="gender-option">
                            <input type="radio" name="gender" value="female" required>
                            <span class="radio-custom"></span>
                            <i class="fas fa-venus"></i>
                            Female
                        </label>
                        <label class="gender-option">
                            <input type="radio" name="gender" value="other" required>
                            <span class="radio-custom"></span>
                            <i class="fas fa-genderless"></i>
                            Other
                        </label>
                    </div>
                </div>

                <label for="first_name">First Name:</label>
                <input type="text" id="first_name" name="first_name" required>
                <br>

                <label for="last_name">Last Name:</label>
                <input type="text" id="last_name" name="last_name" required>
                <br>

                <label for="email">Email:</label>
                <input type="email" id="email" name="email" placeholder="example@gmail.com" required>
                <br>

                <label for="username">Username:</label>
                <input type="text" id="username" name="username" required>
                <br>

                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
                <br>

                <label for="confirm_password">Confirm Password:</label>
                <input type="password" id="confirm_password" name="confirm_password" required>
                <br>

                <button type="submit">Register</button>
            </form>
            <p>Already have an account? <a href="#/login">Login here</a></p>
            <p class="home-link"><a href="#/">← Back to Homepage</a></p>
        </div>
    `;
}

async function handleLogin(event) {
    event.preventDefault();
    const formData = new FormData(event.target);

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            body: JSON.stringify({
                identifier: formData.get('identifier'),
                password: formData.get('password'),
            }),
            headers: {
                'Content-Type': 'application/json',
            },
        });
        const data = await response.json();
        if (data.success) {
            window.location.hash = '/home';
        } else {
            alert(data.error);
        }
    } catch (error) {
        console.error('Login error:', error);
        alert('Login failed. Please try again.');
    }
}

async function handleRegister(event) {
    event.preventDefault();
    const formData = new FormData(event.target);

    // Validate passwords match
    const password = formData.get('password');
    const confirmPassword = formData.get('confirm_password');
    if (password !== confirmPassword) {
        alert("Passwords do not match.");
        return;
    }

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            body: JSON.stringify({
                email: formData.get('email'),
                username: formData.get('username'),
                password: password,
                nickname: formData.get('nickname'),
                age: Number(formData.get('age')),
                gender: formData.get('gender'),
                first_name: formData.get('first_name'),
                last_name: formData.get('last_name'),
            }),
            headers: {
                'Content-Type': 'application/json',
            },
        });
        const data = await response.json();
        if (data.success) {
            window.location.hash = '/login';
        } else {
            alert(data.error);
        }
    } catch (error) {
        console.error('Registration error:', error);
        alert('Registration failed. Please try again.');
    }
}

async function handleLogout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Logout failed');
        }

        localStorage.removeItem('authState');
        window.location.hash = '/login';
        window.location.reload();
        
    } catch (error) {
        console.error('Logout error:', error);
        alert('Error during logout. Please try again.');
    }
}

// Handle hash change
window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});

window.handleLogin = handleLogin;
window.handleRegister = handleRegister;
window.checkLoginStatus = checkLoginStatus;
window.handleLogout = handleLogout;