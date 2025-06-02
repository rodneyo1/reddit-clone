// Routes
const routes = {
    '/': 'login',
    '/login': 'login',
    '/register': 'register',
    '/profile': 'profile',
    '/home': 'home',
    '/logout': 'logout',
    '/filter': 'filter',
    '/api/comments': 'comments',
    '/api/comment': 'comment',
    '/api/comment/like': 'commentLike'
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
    const isLoggedIn = await checkLoginStatus();

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
                    if (isLoggedIn) {
                        window.location.hash = '/home';
                        return;
                    }
                    app.innerHTML = await fetchLoginContent();
                    break;
                case '/register':
                    if (isLoggedIn) {
                        window.location.hash = '/home';
                        return;
                    }
                    app.innerHTML = await fetchRegisterContent();
                    break;
                case '/profile':
                    app.innerHTML = await fetchProfileContent();
                    break;
                case '/logout':
                    await handleLogout();
                    window.location.hash = '/login';
                    window.location.reload();
                    return
                    break;
                case '/messages':
                        if (!isLoggedIn) {
                            window.location.hash = '/login';
                            return;
                        }
                        app.innerHTML = await fetchMessagesContent();
                        initChat(); // Initialize chat functionality
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


    
    // const isLoggedIn = await checkLoginStatus();
    
    const logo = document.getElementById('logo');
    // Update logo link
if (logo) {
    logo.innerHTML = `<a href="${isLoggedIn ? '#/home' : '#/'}" class="logo-link">Forum</a>`;
}

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

// Add to your utility functions
function updateLikeUI(postId, likeCount, dislikeCount, isLike) {
    const likeBtn = document.querySelector(`.like-button[data-post-id="${postId}"]`);
    const dislikeBtn = document.querySelector(`.dislike-button[data-post-id="${postId}"]`);
    
    // Update counts
    if (likeBtn) likeBtn.querySelector('.like-count').textContent = likeCount;
    if (dislikeBtn) dislikeBtn.querySelector('.dislike-count').textContent = dislikeCount;
    
    // Update active states
    if (isLike !== undefined) {
        if (isLike) {
            likeBtn?.classList.add('active');
            dislikeBtn?.classList.remove('active');
        } else {
            dislikeBtn?.classList.add('active');
            likeBtn?.classList.remove('active');
        }
    }
}

function normalizeLikePost(post) {
    return {
        ...post,
        likeCount: post.likeCount || post.LikeCount || 0,
        dislikeCount: post.dislikeCount || post.DislikeCount || 0,
        userLiked: post.userLiked || false,
        userDisliked: post.userDisliked || false
    };
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
                <button class="like-button ${p.userLiked ? 'active' : ''}" data-post-id="${p.id}" onclick="handleLikeAction('${p.id}', true)">
                    <i class="fas fa-thumbs-up"></i> <span class="like-count">${p.likeCount}</span>
                </button>
                <button class="dislike-button ${p.userDisliked ? 'active' : ''}" data-post-id="${p.id}" onclick="handleLikeAction('${p.id}', false)">
                    <i class="fas fa-thumbs-down"></i> <span class="dislike-count">${p.dislikeCount}</span>
                </button>
                <button class="comment-button" onclick="toggleCommentForm('${p.id}')">
                    <i class="fas fa-comment"></i> Comments
                </button>
            </div>
            <div id="comment-form-${p.id}" style="display:none;" class="comment-form">
                <form onsubmit="return handleCommentSubmit('${p.id}', event)">
                    <textarea name="content" required placeholder="Write your comment..."></textarea>
                    <button type="submit">Post Comment</button>
                    <button type="button" onclick="toggleCommentForm('${p.id}')">Cancel</button>
                </form>
            </div>
            <div id="comments-${p.id}" style="display:none;" class="comments-section"></div>
        </div>
    `;
}

let likeProcessing = false;

async function handleLikeAction(postId, isLike) {
    if (likeProcessing) return;
    likeProcessing = true;
    
    try {
        const response = await fetch('/api/like', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: `post_id=${postId}&is_like=${isLike}`
        });

        if (response.status === 401) {
            window.location.hash = '#/login';
            return;
        }

        const data = await response.json();
        
        if (data.success) {
            updateLikeUI(postId, data.like_count, data.dislike_count, isLike);
        } else {
            alert(data.error || 'Failed to process like/dislike');
        }
    } catch (error) {
        console.error('Like action failed:', error);
        alert('An error occurred. Please try again.');
    } finally {
        likeProcessing = false;
    }
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
        userLiked: post.userLiked || false,
        userDisliked: post.userDisliked || false,
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

// Toggle comment form visibility
function toggleCommentForm(postId) {
    const commentForm = document.getElementById(`comment-form-${postId}`);
    const commentsSection = document.getElementById(`comments-${postId}`);
    
    if (commentForm.style.display === 'none' || !commentForm.style.display) {
        commentForm.style.display = 'block';
        if (commentsSection) commentsSection.style.display = 'block';
        loadComments(postId);
    } else {
        commentForm.style.display = 'none';
        if (commentsSection) commentsSection.style.display = 'none';
    }
}

// Toggle reply form visibility
function toggleReplyForm(commentId) {
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    if (replyForm.style.display === 'none' || !replyForm.style.display) {
        replyForm.style.display = 'block';
    } else {
        replyForm.style.display = 'none';
    }
}

// Load comments for a post
async function loadComments(postId) {
    const commentsSection = document.getElementById(`comments-${postId}`);
    if (!commentsSection) return;

    try {
        const response = await fetch(`/api/comments?post_id=${postId}`, {
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to load comments');
        }

        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
            throw new Error('Invalid response format');
        }

        const comments = await response.json();
        
        // Debug log to check the received data
        console.log('Received comments:', comments);
        
        if (!Array.isArray(comments)) {
            throw new Error('Invalid comments data');
        }

        commentsSection.innerHTML = comments.length > 0 
            ? renderComments(comments) 
            : '<p class="no-comments">No comments yet. Be the first to comment!</p>';
            
    } catch (error) {
        console.error('Error loading comments:', error);
        commentsSection.innerHTML = `
            <div class="error-message">
                Error loading comments. Please try again.
                <button onclick="loadComments(${postId})">Retry</button>
            </div>
        `;
    }
}

// Render comments HTML
function renderComments(comments) {
    if (!comments || comments.length === 0) {
        return '<p class="no-comments">No comments yet. Be the first to comment!</p>';
    }

    return comments.map(comment => `
        <div class="comment" data-comment-id="${comment.ID}">
            <div class="comment-header">
                <span class="comment-author">${comment.Username || 'Anonymous'}</span>
                <span class="comment-time">${comment.CreatedAtHuman || formatDate(comment.CreatedAt) || 'Just now'}</span>
            </div>
            <div class="comment-content">${comment.Content || ''}</div>
            ${comment.Replies && comment.Replies.length > 0 ? `
                <div class="replies">
                    ${renderComments(comment.Replies)}
                </div>
            ` : ''}
        </div>
    `).join('');
}

// Add this helper function if you don't have it
function formatDate(dateString) {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('en-US', options);
}

// Render a single comment with its replies
function renderSingleComment(comment) {
    return `
        <div class="comment" data-comment-id="${comment.id}">
            <div class="comment-header">
                <span class="comment-author">${comment.username}</span>
                <span class="comment-time">${comment.createdAtHuman}</span>
            </div>
            <div class="comment-content">${comment.content}</div>
            <div class="comment-actions">
                <button class="like-button ${comment.userLiked ? 'active' : ''}" 
                        data-comment-id="${comment.id}" 
                        onclick="handleCommentLike('${comment.id}', true)">
                    <i class="fas fa-thumbs-up"></i> 
                    <span class="like-count">${comment.likeCount || 0}</span>
                </button>
                <button class="dislike-button ${comment.userDisliked ? 'active' : ''}" 
                        data-comment-id="${comment.id}" 
                        onclick="handleCommentLike('${comment.id}', false)">
                    <i class="fas fa-thumbs-down"></i> 
                    <span class="dislike-count">${comment.dislikeCount || 0}</span>
                </button>
                <button class="reply-button" data-comment-id="${comment.id}">
                    <i class="fas fa-reply"></i> Reply
                </button>
            </div>
            
            <!-- Reply form (hidden by default) -->
            <div id="reply-form-${comment.id}" style="display:none;" class="reply-form">
                <form onsubmit="handleCommentSubmit(event)">
    <textarea name="content" required placeholder="Write your comment..."></textarea>
    <button type="submit">Post Comment</button>
    <button type="button" onclick="toggleCommentForm('${postId}')">Cancel</button>
</form>
            </div>
            
            <!-- Replies section -->
            ${comment.replies && comment.replies.length > 0 ? `
                <div class="replies">
                    ${comment.replies.map(reply => renderSingleComment(reply)).join('')}
                </div>
            ` : ''}
        </div>
    `;
}

// Handle comment form submission
// Replace your handleCommentSubmit function with:
// Handle comment submission
async function handleCommentSubmit(postId, event) {
    event.preventDefault();
    
    const form = event.target;
    const content = form.content.value.trim();
    
    if (!content) {
        alert('Comment cannot be empty');
        return false;
    }

    try {
        // Convert postId to number
        const numericPostId = Number(postId);
        if (isNaN(numericPostId)) {
            throw new Error('Invalid post ID');
        }

        const response = await fetch('/api/comment', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                post_id: numericPostId,
                content: content
            })
        });

        const responseData = await response.json();
        
        if (!response.ok) {
            throw new Error(responseData.error || 'Failed to post comment');
        }

        form.reset();
        await loadComments(postId);
    } catch (error) {
        console.error('Error submitting comment:', error);
        alert(`Error: ${error.message}`);
    }
    return false;
}

// Handle comment like/dislike
async function handleCommentLike(commentId, isLike) {
    try {
        const response = await fetch('/api/comment/like', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                comment_id: commentId,
                is_like: isLike
            })
        });

        if (response.ok) {
            const data = await response.json();
            if (data.success) {
                // Find the comment element
                const commentElement = document.querySelector(`.comment[data-comment-id="${commentId}"]`);
                if (commentElement) {
                    // Update like/dislike counts
                    const likeCountElement = commentElement.querySelector('.like-count');
                    const dislikeCountElement = commentElement.querySelector('.dislike-count');
                    
                    if (likeCountElement) likeCountElement.textContent = data.likeCount;
                    if (dislikeCountElement) dislikeCountElement.textContent = data.dislikeCount;
                    
                    // Update button states
                    const likeButton = commentElement.querySelector('.like-button');
                    const dislikeButton = commentElement.querySelector('.dislike-button');
                    
                    if (data.userLiked) {
                        likeButton.classList.add('active');
                        dislikeButton.classList.remove('active');
                    } else if (data.userDisliked) {
                        dislikeButton.classList.add('active');
                        likeButton.classList.remove('active');
                    } else {
                        likeButton.classList.remove('active');
                        dislikeButton.classList.remove('active');
                    }
                }
            }
        } else {
            const error = await response.json();
            alert(error.error || 'Failed to process like/dislike');
        }
    } catch (error) {
        console.error('Error liking comment:', error);
        alert('Error processing your request. Please try again.');
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
window.handleLikeAction = handleLikeAction;
window.toggleCommentForm = toggleCommentForm;
window.toggleReplyForm = toggleReplyForm;
window.handleCommentSubmit = handleCommentSubmit;
window.handleCommentLike = handleCommentLike;