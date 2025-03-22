// Routes
const routes = {
    '/': 'home',
    '/login': 'login',
    '/register': 'register',
    '/profile': 'profile'
};

// Render content based on the route
async function render(path) {
    const app = document.getElementById('app');
    const authButtons = document.getElementById('auth-buttons');

    switch (path) {
        case '/':
            app.innerHTML = await fetchHomeContent();
            break;
        case '/login':
            app.innerHTML = await fetchLoginContent();
            break;
        case '/register':
            app.innerHTML = await fetchRegisterContent();
            break;
        case '/profile':
            app.innerHTML = await fetchProfileContent();
            break;
        default:
            app.innerHTML = '<h1>404 Not Found</h1>';
    }

    // Update auth buttons based on login status
    const isLoggedIn = await checkLoginStatus();
    authButtons.innerHTML = isLoggedIn
        ? `
            <div class="profile-icon" style="position: relative;">
                <a href="#/profile" class="material-icons" style="font-size:30px; color: #4A7C8C; margin-top: 10px; vertical-align: middle;">person</a>
            </div>
            <a href="#" class="auth-button create-post" onclick="toggleCreatePost()">Create Post</a>
            <a href="#/logout" class="logout-icon" title="Logout">
                <i class="fas fa-sign-out-alt" style="font-size: 24px; color: #4A7C8C; margin-top: 10px;"></i>
            </a>
        `
        : `
            <a href="#/login" class="auth-button login">Login</a>
            <a href="#/register" class="auth-button register">Register</a>
        `;
}

// Fetch home content
async function fetchHomeContent() {
    const response = await fetch('/api/home');
    const data = await response.json();
    return `
        <div class="container">
            <aside class="sidebar" id="sidebar">
                <h3>Categories</h3>
                <ul>
                    <li><a href="#/">All Posts</a></li>
                    <li><a href="#/filter?category=technology">Technology</a></li>
                     <li><a href="/filter?category=general">General</a></li>
                    <li><a href="/filter?category=lifestyle">Lifestyle</a></li>
                    <li><a href="/filter?category=entertainment">Entertainment</a></li>
                    <li><a href="/filter?category=gaming">Gaming</a></li>
                    <li><a href="/filter?category=food">Food</a></li>
                    <li><a href="/filter?category=business">Business</a></li>
                    <li><a href="/filter?category=religion">Religion</a></li>
                    <li><a href="/filter?category=health">Health</a></li>
                    <li><a href="/filter?category=music">Music</a></li>
                    <li><a href="/filter?category=sports">Sports</a></li>
                    <li><a href="/filter?category=beauty">Beauty</a></li>
                    <li><a href="/filter?category=jobs">Jobs</a></li>
                </ul>
            </aside>
            <main>
                <h1 id="postsHeading">All Posts</h1>
                <div id="posts">
                    ${data.posts.map(post => `
                        <div class="post" data-category="${post.categories}">
                            <p class="posted-on">${post.createdAtHuman}</p>
                            <strong><p>${post.username}</p></strong>
                            <h3>${post.title}</h3>
                            <p>${post.content}</p>
                            ${post.imagePath ? `<img src="${post.imagePath}" alt="Post Image" class="post-image">` : ''}
                            <p class="categories">Categories: <span>${post.categories}</span></p>
                            <div class="post-actions">
                                <button class="like-button" data-post-id="${post.id}" onclick="toggleLike('${post.id}', true)">
                                    <i class="fas fa-thumbs-up"></i> <span class="like-count">${post.likeCount}</span>
                                </button>
                                <button class="dislike-button" data-post-id="${post.id}" onclick="toggleLike('${post.id}', false)">
                                    <i class="fas fa-thumbs-down"></i> <span class="dislike-count">${post.dislikeCount}</span>
                                </button>
                                <button class="comment-button" onclick="toggleCommentForm('${post.id}')">
                                    <i class="fas fa-comment"></i> Comments
                                </button>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </main>
        </div>
    `;
}

// Fetch login form content
async function fetchLoginContent() {
    return `
        <div class="auth-container">
            <h1>Login</h1>

            <!-- Google Sign-In Button -->
            <a href="/auth/google/login" class="google-btn">
                <img src="/src/google.jpeg" alt="Google Logo">
                <span>Sign in with Google</span>
            </a>

            <!-- GitHub Sign-In Button -->
            <a href="/auth/github/login" class="github-btn">
                <img src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" alt="GitHub Logo">
                <span>Sign in with GitHub</span>
            </a>

            <div class="oauth-divider">
                <span>or</span>
            </div>

            <!-- Traditional Login Form -->
            <form id="login-form" onsubmit="handleLogin(event)">
                <label for="email">Email:</label>
                <input type="email" id="email" name="email" placeholder="example@gmail.com" required>
                <br>
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
                <br>
                <button type="submit">Login</button>
            </form>
            <p>Don't have an account? <a href="#/register">Register here</a></p>
            <p class="home-link"><a href="#/">← Back to Homepage</a></p>
        </div>
    `;
}

// Handle login form submission
async function handleLogin(event) {
    event.preventDefault();
    const formData = new FormData(event.target);

    // Send the login data to the backend
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
        alert('Login successful!');
        window.location.hash = '/'; // Redirect to home page
    } else {
        alert(data.error); // Show error message
    }
}

async function fetchRegisterContent() {
    return `
        <div class="auth-container">
            <h1>Register</h1>

            <!-- Google Sign-In Button -->
            <a href="/auth/google/login" class="google-btn">
                <img src="/src/google.jpeg" alt="Google Logo">
                <span>Sign up with Google</span>
            </a>

            <!-- GitHub Sign-In Button -->
            <a href="/auth/github/login" class="github-btn">
                <img src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" alt="GitHub Logo">
                <span>Sign up with GitHub</span>
            </a>

            <div class="oauth-divider">
                <span>or</span>
            </div>

            <!-- Traditional Registration Form -->
            <form id="register-form" onsubmit="handleRegister(event)">
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

// Handle registration form submission
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

    // Send the registration data to the backend
    try{
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
        alert('Registration successful! Please login.');
        window.location.hash = '/login'; // Redirect to login page
    } else {
        alert(data.error); // Show error message
    }
} catch (error) {
    console.error('Error during registration:', error);
    alert('An error occurred during registration. Please try again.');
}
}

// Check login status
async function checkLoginStatus() {
    const response = await fetch('/api/check-login');
    const data = await response.json();
    return data.isLoggedIn;
}

// Handle hash change
window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});

// Initial render
const initialPath = window.location.hash.replace('#', '') || '/';
render(initialPath);