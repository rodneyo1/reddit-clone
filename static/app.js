// Function to render the Home page
function renderHomePage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <header>
            <div class="hamburger" onclick="toggleMenu()">
                <i class="fas fa-bars"></i>
            </div>
            <div class="logo">
                <a href="/" class="logo-link" data-link>Forum</a>
            </div>
            <nav>
                <div class="profile-icon" style="position: relative;">
                    <a href="/profile" class="material-icons" data-link
                        style="font-size:30px; color: #4A7C8C; margin-top: 10px; vertical-align: middle;">person</a>
                </div>
                <a href="#" class="auth-button create-post" data-link>Create Post</a>
                <a href="/logout" class="logout-icon" title="Logout" data-link>
                    <i class="fas fa-sign-out-alt" style="font-size: 24px; color: #4A7C8C; margin-top: 10px;"></i>
                </a>
                <a href="/login" class="auth-button login" data-link>Login</a>
                <a href="/register" class="auth-button register" data-link>Register</a>
            </nav>
        </header>
        <div class="container">
            <aside class="sidebar" id="sidebar">
                <h3>Categories</h3>
                <ul>
                    <li><a href="/" data-link>All Posts</a></li>
                    <li><a href="/filter?category=technology" data-link>Technology</a></li>
                    <li><a href="/filter?category=general" data-link>General</a></li>
                    <li><a href="/filter?category=lifestyle" data-link>Lifestyle</a></li>
                    <li><a href="/filter?category=entertainment" data-link>Entertainment</a></li>
                    <li><a href="/filter?category=gaming" data-link>Gaming</a></li>
                    <li><a href="/filter?category=food" data-link>Food</a></li>
                    <li><a href="/filter?category=business" data-link>Business</a></li>
                    <li><a href="/filter?category=religion" data-link>Religion</a></li>
                    <li><a href="/filter?category=health" data-link>Health</a></li>
                    <li><a href="/filter?category=music" data-link>Music</a></li>
                    <li><a href="/filter?category=sports" data-link>Sports</a></li>
                    <li><a href="/filter?category=beauty" data-link>Beauty</a></li>
                    <li><a href="/filter?category=jobs" data-link>Jobs</a></li>
                </ul>
            </aside>
            <main>
                <h1 id="postsHeading">All Posts</h1>
                <div id="posts">
                    <p>No posts available.</p>
                </div>
            </main>
        </div>
    `;

    // Attach event listeners after rendering
    attachEventListeners();
}

// Function to render the Login page
function renderLoginPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <h1>Login</h1>
        <form id="login-form">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" required>
            <br>
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
            <br>
            <button type="submit">Login</button>
        </form>
    `;

    // Attach event listeners for the login form
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', (e) => {
            e.preventDefault();
            alert('Login functionality to be implemented.');
        });
    }
}

// Function to render the Register page
function renderRegisterPage() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <h1>Register</h1>
        <form id="register-form">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" required>
            <br>
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
            <br>
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
            <br>
            <button type="submit">Register</button>
        </form>
    `;

    // Attach event listeners for the register form
    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', (e) => {
            e.preventDefault();
            alert('Register functionality to be implemented.');
        });
    }
}

// Function to attach event listeners
function attachEventListeners() {
    const createPostButton = document.querySelector('.create-post');
    if (createPostButton) {
        createPostButton.addEventListener('click', toggleCreatePost);
    }
}

// Function to toggle the "Create Post" form
function toggleCreatePost() {
    const createPostForm = document.getElementById('createPostForm');
    if (createPostForm) {
        createPostForm.style.display = createPostForm.style.display === 'none' ? 'block' : 'none';
    }
}

// Client-side routing
const routes = {
    '/': renderHomePage,
    '/login': renderLoginPage,
    '/register': renderRegisterPage,
};

function navigateTo(path) {
    window.history.pushState({}, '', path);
    loadPage(path);
}

function loadPage(path) {
    const routeHandler = routes[path];
    if (routeHandler) {
        routeHandler();
    } else {
        // Handle 404 - Page Not Found
        document.getElementById('app').innerHTML = '<h1>404 - Page Not Found</h1>';
    }
}

// Handle browser back/forward buttons
window.onpopstate = () => {
    loadPage(window.location.pathname);
};

// Attach event listeners for navigation
document.addEventListener('DOMContentLoaded', () => {
    document.body.addEventListener('click', (e) => {
        if (e.target.matches('[data-link]')) {
            e.preventDefault();
            navigateTo(e.target.getAttribute('href'));
        }
    });

    loadPage(window.location.pathname);
});
// // Router object: Maps routes to functions
// const router = {
//     home: async function() {
//         let data = await fetchData("home");
//         renderPage(data);
//     },
// };

// // Fetch JSON data from Go backend
// async function fetchData(route) {
//     let response = await fetch(`/api/${route}`);
//     return response.json();
// }

// // Update the page content inside the #app div
// function renderPage(data) {
//     document.getElementById("app").innerHTML = `
//         <h1>${data.title}</h1>
//         <p>${data.content}</p>
//     `;
// }

// // Navigate function: Calls the correct handler
// function navigate(route) {
//     if (router[route]) {
//         router[route]();
//         window.history.pushState({}, "", `/${route}`); // Update URL
//     } else {
//         console.error("Route not found:", route);
//     }
// }

// // Handle browser back/forward buttons
// window.onpopstate = () => {
//     let path = window.location.pathname.substring(1);
//     navigate(path || "home");
// };

// // Load the default page when the app starts
// window.onload = () => {
//     let path = window.location.pathname.substring(1); // allow refreshing on any page
//     navigate(path || "home");
// };