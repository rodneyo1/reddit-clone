// Routes
const routes = {
    '/': 'login',
    '/login': 'login',
    '/register': 'register',
    '/profile': 'profile',
    '/home': 'home'
};

// Toggle create post form
function toggleCreatePost() {
    console.log("Toggle create post function called");
    
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
    console.log(`Rendering path: ${path}`);
    const app = document.getElementById('app');
    const authButtons = document.getElementById('auth-buttons');

    try {
        switch (path) {
            case '/':
                app.innerHTML = await fetchLoginContent();
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
            case '/home':
                app.innerHTML = await fetchHomeContent();
                // Attach form submission handler
                document.getElementById('post-form')?.addEventListener('submit', handlePostSubmit);
                break;
            default:
                app.innerHTML = '<h1>404 Not Found</h1>';
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
    const response = await fetch('/api/home');
    const data = await response.json();
    const isLoggedIn = await checkLoginStatus();
    
    return `
        <div class="container">
            <aside class="sidebar" id="sidebar">
                <h3>Categories</h3>
                <ul>
                    <li><a href="#/home">All Posts</a></li>
                    <li><a href="#/filter?category=technology">Technology</a></li>
                    <li><a href="#/filter?category=general">General</a></li>
                    <li><a href="#/filter?category=lifestyle">Lifestyle</a></li>
                    <li><a href="#/filter?category=entertainment">Entertainment</a></li>
                    <li><a href="#/filter?category=gaming">Gaming</a></li>
                    <li><a href="#/filter?category=food">Food</a></li>
                    <li><a href="#/filter?category=business">Business</a></li>
                    <li><a href="#/filter?category=religion">Religion</a></li>
                    <li><a href="#/filter?category=health">Health</a></li>
                    <li><a href="#/filter?category=music">Music</a></li>
                    <li><a href="#/filter?category=sports">Sports</a></li>
                    <li><a href="#/filter?category=beauty">Beauty</a></li>
                    <li><a href="#/filter?category=jobs">Jobs</a></li>
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
                            <label><input type="checkbox" name="category" value="technology"> Technology</label>
                            <label><input type="checkbox" name="category" value="general"> General</label>
                            <label><input type="checkbox" name="category" value="lifestyle"> Lifestyle</label>
                            <label><input type="checkbox" name="category" value="entertainment"> Entertainment</label>
                            <label><input type="checkbox" name="category" value="gaming"> Gaming</label>
                            <label><input type="checkbox" name="category" value="food"> Food</label>
                            <label><input type="checkbox" name="category" value="business"> Business</label>
                            <label><input type="checkbox" name="category" value="religion"> Religion</label>
                            <label><input type="checkbox" name="category" value="health"> Health</label>
                            <label><input type="checkbox" name="category" value="music"> Music</label>
                            <label><input type="checkbox" name="category" value="sports"> Sports</label>
                            <label><input type="checkbox" name="category" value="beauty"> Beauty</label>
                            <label><input type="checkbox" name="category" value="jobs"> Jobs</label>
                        </div>
                        <br>
                        <button type="submit">Post</button>
                        <button type="button" onclick="toggleCreatePost()">Cancel</button>
                    </form>
                </div>
                ` : ''}
                
                <h1 id="postsHeading">All Posts</h1>
                <div id="posts">
                    ${data.posts.map(post => `
                        <div class="post" data-category="${post.categories || post.Categories}">
                            <p class="posted-on">${post.createdAtHuman || post.CreatedAtHuman}</p>
                            <strong><p>${post.username || post.Username}</p></strong>
                            <h3>${post.title || post.Title}</h3>
                            <p>${post.content  || post.Content}</p>
                            ${post.imagePath || post.ImagePath  ? `<img src="${post.imagePath || post.ImagePath}" alt="Post Image" class="post-image">` : ''}
                            <p class="categories">Categories: <span>${post.categories || post.Categories}</span></p>
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
            const error = await response.json();z
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

// Make toggleCreatePost available globally
window.toggleCreatePost = toggleCreatePost;