// Routes
const routes = {
    '/': 'login',
    '/login': 'login',
    '/register': 'register',
    '/profile': 'profile',
    '/home': 'home',
    '/logout': 'logout',
    '/filter': 'filter'
};

const validCategories = [
    "technology",
    "general",
    "lifestyle",
    "entertainment",
    "gaming",
    "food",
    "business",
    "religion",
    "health",
    "music",
    "sports",
    "beauty",
    "jobs"
];

// Toggle create post form
function toggleCreatePost() {
    // console.log("Toggle create post function called");
    
    const createPostForm = document.getElementById('createPostForm');
    const postsList = document.getElementById('posts');
    const postsHeading = document.getElementById('postsHeading');

    if (!createPostForm) {
        console.error("Create Post form element not found!");
        return;
    }

    const shouldShowForm = createPostForm.style.display === 'none' || 
                          !createPostForm.style.display;
    
    createPostForm.style.display = shouldShowForm ? 'block' : 'none';
    
    if (postsList) postsList.style.display = shouldShowForm ? 'none' : 'block';
    if (postsHeading) postsHeading.style.display = shouldShowForm ? 'none' : 'block';
}

// Render content based on the route
async function render(path) {
    // console.log(`Rendering path: ${path}`);
    const app = document.getElementById('app');
    const authButtons = document.getElementById('auth-buttons');

    try {
        if (path.startsWith('/filter')) {
            const queryString = path.split('?')[1];
            const params = new URLSearchParams(queryString);
            const category = params.get('category');
            if (!validCategories.includes(category) && category !== 'all') {
                app.innerHTML = '<p class="error-message">Invalid category selected</p>';
                return;
            }
            app.innerHTML = await fetchFilteredContent(category);
        } else {
            switch (path) {
                case '/':
                case '/login':
                    app.innerHTML = await fetchLoginContent();
                    break;
                case '/register':
                    app.innerHTML = await fetchRegisterContent();
                    break;
                case '/profile':
                    app.innerHTML = await fetchProfileContent();
                    break;
                case '/logout':
                    await handleLogout();
                    window.location.hash = '/login';
                    break;
                case '/home':
                    app.innerHTML = await fetchHomeContent();
                    document.getElementById('post-form')?.addEventListener('submit', handlePostSubmit);
                    break;
                default:
                    app.innerHTML = '<h1>404 Not Found</h1>';
            }
        }
    } catch (error) {
        console.error('Error rendering content:', error);
        app.innerHTML = '<p class="error-message">Error loading content. Please try again.</p>';
    }

    // Update auth buttons based on login status
    const isLoggedIn = await checkLoginStatus();
    authButtons.innerHTML = isLoggedIn
        ? `
            <div class="profile-icon" style="position: relative;">
                <a href="#/profile" class="material-icons" style="font-size:30px; color: #4A7C8C; margin-top: 10px; vertical-align: middle;">person</a>
            </div>
            <button class="auth-button create-post" id="create-post-btn">Create Post</button>
            <a href="#/home" class="auth-button home">Home</a>
            <a href="#/logout" class="logout-icon" title="Logout">
                <i class="fas fa-sign-out-alt" style="font-size: 24px; color: #4A7C8C; margin-top: 10px;"></i>
            </a>
        `
        : `
            <a href="#/" class="auth-button login">Login</a>
            <a href="#/register" class="auth-button register">Register</a>
        `;

    // Attach create post button event listener
    if (isLoggedIn) {
        document.getElementById('create-post-btn')?.addEventListener('click', function(e) {
            e.preventDefault();
            toggleCreatePost();
        });
    }
}

// Helper function to render individual posts
function renderPost(post) {
    const p = normalizePost(post);
    return `
        <div class="post" data-category="${p.categories}">
            <p class="posted-on">${p.createdAtHuman}</p>
            <strong><p>${p.username}</p></strong>
            <h3>${p.title}</h3>
            <p>${p.content}</p>
            ${p.imagePath ? `<img src="${p.imagePath}" alt="Post Image" class="post-image">` : ''}
            <p class="categories">Categories: <span>${p.categories}</span></p>
            <div class="post-actions">
                <button class="like-button" data-post-id="${p.id}" onclick="toggleLike('${p.id}', true)">
                    <i class="fas fa-thumbs-up"></i> <span class="like-count">${p.likeCount}</span>
                </button>
                <button class="dislike-button" data-post-id="${p.id}" onclick="toggleLike('${p.id}', false)">
                    <i class="fas fa-thumbs-down"></i> <span class="dislike-count">${p.dislikeCount}</span>
                </button>
                <button class="comment-button" onclick="toggleCommentForm('${p.id}')">
                    <i class="fas fa-comment"></i> Comments
                </button>
            </div>
        </div>
    `;
}

// Fetch filtered content by category
async function fetchFilteredContent(category) {
    try {
        const response = await fetch(`/api/filter?category=${encodeURIComponent(category)}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const data = await response.json();
        const isLoggedIn = await checkLoginStatus();
        const posts = data.Posts || data.posts || [];

        return `
            <div class="container">
                <aside class="sidebar" id="sidebar">
                    <h3>Categories</h3>
                    <ul>
                        <li><a href="#/home">All Posts</a></li>
                        ${validCategories.map(cat => `
                            <li><a href="#/filter?category=${cat}" class="${cat === category ? 'active' : ''}">
                                ${cat.charAt(0).toUpperCase() + cat.slice(1)}
                            </a></li>
                        `).join('')}
                    </ul>
                </aside>
                <main>
                    ${isLoggedIn ? `
                    <div id="createPostForm" style="display:none; background:white; padding:20px; margin-bottom:20px; border-radius:8px;">
                        <h1>Create a New Post</h1>
                        <form id="post-form" enctype="multipart/form-data">
                            <label for="post-title">Title:</label>
                            <input type="text" id="post-title" name="title" required>
                            <br>
                            <label for="post-content">Content:</label>
                            <textarea id="post-content" name="content" required></textarea>
                            <br>
                            <label for="post-image">Image (optional, max 20MB):</label>
                            <input type="file" id="post-image" name="image" accept="image/jpeg,image/png,image/gif">
                            <br>
                            <label>Categories (select at least one):</label>
                            <div class="checkbox-group">
                                ${validCategories.map(cat => `
                                    <label><input type="checkbox" name="category" value="${cat}"> ${cat.charAt(0).toUpperCase() + cat.slice(1)}</label>
                                `).join('')}
                            </div>
                            <br>
                            <button type="submit">Post</button>
                            <button type="button" onclick="toggleCreatePost()">Cancel</button>
                        </form>
                    </div>
                    ` : ''}
                    
                    <h1 id="postsHeading">${category === 'all' ? 'All Posts' : `Posts in ${category.charAt(0).toUpperCase() + category.slice(1)}`}</h1>
                    <div id="posts">
                        ${posts.length > 0 ? 
                            posts.map(post => renderPost(post)).join('') :
                            '<p class="empty-message">No posts found in this category</p>'
                        }
                    </div>
                </main>
            </div>
        `;
    } catch (error) {
        console.error('Error fetching filtered content:', error);
        return '<p class="error-message">Error loading filtered posts. Please try again.</p>';
    }
}

// Fetch profile content
async function fetchProfileContent() {
    try {
        const response = await fetch('/api/profile');
        if (!response.ok) {
            if (response.status === 401) {
                window.location.hash = '/login';
                return '<p class="error-message">Please login to view your profile.</p>';
            }
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const profileData = await response.json();

        // Generate HTML for created posts
        const createdPostsHTML = profileData.CreatedPosts && profileData.CreatedPosts.length > 0 
            ? profileData.CreatedPosts.map(post => `
                <article class="post">
                    <h3>${post.Title}</h3>
                    <p class="post-content">${post.Content}</p>
                    ${post.ImagePath ? `<img src="${post.ImagePath}" alt="Post Image" class="post-image">` : ''}
                    <div class="post-meta">
                        ${post.Categories ? `<span class="categories"><i class="fas fa-tags"></i> ${post.Categories}</span>` : ''}
                        <span class="likes"><i class="fas fa-thumbs-up"></i> ${post.LikeCount}</span>
                        <span class="dislikes"><i class="fas fa-thumbs-down"></i> ${post.DislikeCount}</span>
                        <span class="date"><i class="far fa-clock"></i> ${post.CreatedAtHuman}</span>
                    </div>
                </article>
            `).join('')
            : `<p class="empty-message">You haven't created any posts yet.</p>`;

        // Generate HTML for liked posts
        const likedPostsHTML = profileData.LikedPosts && profileData.LikedPosts.length > 0 
            ? profileData.LikedPosts.map(post => `
                <article class="post">
                    <h3>${post.Title}</h3>
                    <p class="post-content">${post.Content}</p>
                    ${post.ImagePath ? `<img src="${post.ImagePath}" alt="Post Image" class="post-image">` : ''}
                    <div class="post-meta">
                        <span class="author"><i class="fas fa-user"></i> ${post.Username}</span>
                        ${post.Categories ? `<span class="categories"><i class="fas fa-tags"></i> ${post.Categories}</span>` : ''}
                        <span class="likes"><i class="fas fa-thumbs-up"></i> ${post.LikeCount}</span>
                        <span class="dislikes"><i class="fas fa-thumbs-down"></i> ${post.DislikeCount}</span>
                        <span class="date"><i class="far fa-clock"></i> ${post.CreatedAtHuman}</span>
                    </div>
                </article>
            `).join('')
            : `<p class="empty-message">You haven't liked any posts yet.</p>`;

        return `
            <div class="profile-container">
                <div class="profile-header">
                    <h1><i class="fas fa-user-circle"></i> ${profileData.Username}'s Profile</h1>
                    <p><i class="fas fa-envelope"></i> ${profileData.Email}</p>
                </div>

                <div class="profile-sections">
                    <section class="profile-section">
                        <h2><i class="fas fa-pencil-alt"></i> Your Posts</h2>
                        ${createdPostsHTML}
                    </section>

                    <section class="profile-section">
                        <h2><i class="fas fa-heart"></i> Posts You've Liked</h2>
                        ${likedPostsHTML}
                    </section>
                </div>
            </div>
        `;
    } catch (error) {
        console.error('Error fetching profile:', error);
        return '<p class="error-message">Failed to load profile. Please try again.</p>';
    }
}

// Fetch home content
async function fetchHomeContent() {
    try {
        const response = await fetch('/api/home');
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        
        const data = await response.json();
        const isLoggedIn = await checkLoginStatus();
        const posts = data.Posts || data.posts || [];

        return `
            <div class="container">
                <aside class="sidebar" id="sidebar">
                    <h3>Categories</h3>
                    <ul>
                        <li><a href="#/home">All Posts</a></li>
                        ${validCategories.map(cat => `
                            <li><a href="#/filter?category=${cat}">
                                ${cat.charAt(0).toUpperCase() + cat.slice(1)}
                            </a></li>
                        `).join('')}
                    </ul>
                </aside>
                <main>
                    ${isLoggedIn ? `
                    <div id="createPostForm" style="display:none; background:white; padding:20px; margin-bottom:20px; border-radius:8px;">
                        <h1>Create a New Post</h1>
                        <form id="post-form" enctype="multipart/form-data">
                            <label for="post-title">Title:</label>
                            <input type="text" id="post-title" name="title" required>
                            <br>
                            <label for="post-content">Content:</label>
                            <textarea id="post-content" name="content" required></textarea>
                            <br>
                            <label for="post-image">Image (optional, max 20MB):</label>
                            <input type="file" id="post-image" name="image" accept="image/jpeg,image/png,image/gif">
                            <br>
                            <label>Categories (select at least one):</label>
                            <div class="checkbox-group">
                                ${validCategories.map(cat => `
                                    <label><input type="checkbox" name="category" value="${cat}"> ${cat.charAt(0).toUpperCase() + cat.slice(1)}</label>
                                `).join('')}
                            </div>
                            <br>
                            <button type="submit">Post</button>
                            <button type="button" onclick="toggleCreatePost()">Cancel</button>
                        </form>
                    </div>
                    ` : ''}
                    
                    <h1 id="postsHeading">All Posts</h1>
                    <div id="posts">
                        ${posts.length > 0 ? 
                            posts.map(post => renderPost(post)).join('') :
                            '<p class="empty-message">No posts found</p>'
                        }
                    </div>
                </main>
            </div>
        `;
    } catch (error) {
        console.error('Error fetching home content:', error);
        return '<p class="error-message">Error loading posts. Please try again.</p>';
    }
}

function normalizePost(post) {
    return {
        id: post.id || post.ID,
        title: post.title || post.Title,
        content: post.content || post.Content,
        username: post.username || post.Username,
        categories: post.categories || post.Categories,
        imagePath: post.imagePath || post.ImagePath,
        likeCount: post.likeCount || post.LikeCount || 0,
        dislikeCount: post.dislikeCount || post.DislikeCount || 0,
        createdAtHuman: post.createdAtHuman || post.CreatedAtHuman || formatDate(post.createdAt || post.CreatedAt)
    };
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Validate categories before submission
function validateCategories() {
    const checkboxes = document.querySelectorAll('input[name="category"]:checked');
    if (checkboxes.length === 0) {
        alert("Please select at least one category.");
        return false;
    }
    return true;
}

// Handle post form submission
async function handlePostSubmit(event) {
    event.preventDefault();
    
    if (!validateCategories()) {
        return;
    }

    const form = event.target;
    const formData = new FormData(form);
    
    // Add categories to form data
    document.querySelectorAll('input[name="category"]:checked').forEach(checkbox => {
        formData.append('category', checkbox.value);
    });

    try {
        const response = await fetch('/api/posts', {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            toggleCreatePost();
            // Refresh the posts
            if (window.location.hash === '#/home' || window.location.hash === '') {
                const app = document.getElementById('app');
                app.innerHTML = await fetchHomeContent();
                // Re-attach event listeners
                document.getElementById('post-form')?.addEventListener('submit', handlePostSubmit);
            }
        } else {
            const error = await response.json();
            alert(error.error || 'Failed to create post');
        }
    } catch (error) {
        console.error('Error submitting post:', error);
        alert('Error submitting post. Please try again.');
    }
}

// Fetch login form content
async function fetchLoginContent() {
    const isLoggedIn = await checkLoginStatus();
    const homeLink = isLoggedIn ? '<p class="home-link"><a href="#/home">← Go to Homepage</a></p>' : '';

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
            ${homeLink}
        </div>
    `;
}

// Handle login form submission
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

// Fetch register form content
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

// Check login status
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

// Handle hash change
window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});

// Initial render
const initialPath = window.location.hash.replace('#', '') || '/';
render(initialPath);

// Make functions available globally
window.toggleCreatePost = toggleCreatePost;
window.toggleLike = toggleLike; // Make sure this function exists
window.toggleCommentForm = toggleCommentForm; // Make sure this function exists