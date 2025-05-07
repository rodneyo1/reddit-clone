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

async function handleLogin(event) {
    event.preventDefault();
    const formData = new FormData(event.target);

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            body: JSON.stringify({
                email: formData.get('email'),
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
        window.location.reload();
        
    } catch (error) {
        console.error('Logout error:', error);
        alert('Error during logout. Please try again.');
    }
}