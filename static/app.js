// Routes
const routes = {
    '/': 'login',
    '/home': 'home',
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
        case '/login':
            app.innerHTML = await fetchLoginContent();
            break;
        case '/register':
            app.innerHTML = await fetchRegisterContent();
            break;
        case '/home':
                app.innerHTML = await fetchHomeContent();
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
            <button class="auth-button create-post" onclick="toggleCreatePost()">Create Post</button>
            <a href="#/logout" class="logout-icon" title="Logout">
                <i class="fas fa-sign-out-alt" style="font-size: 24px; color: #4A7C8C; margin-top: 10px;"></i>
            </a>
        `
        : `
            <a href="#/login" class="auth-button login">Login</a>
            <a href="#/register" class="auth-button register">Register</a>
        `;
}

// To be replaced with actual fetch functions and backend api & query
const CATEGORIES = [
    "technology", "general", "lifestyle", "entertainment", "gaming",
    "food", "business", "religion", "health", "music",
    "sports", "beauty", "jobs"
];

async function fetchHomeContent() {
    const response = await fetch('/api/home');
    const data = await response.json();

    return `
        <div class="container">
            ${renderSidebar()}
            <main>
                 ${data.isLoggedIn ? renderCreatePostForm() : ''}
                <h1 id="postsHeading">All Posts</h1>
                <div id="posts">${renderPosts(data.posts)}</div>
            </main>
        </div>
    `;
}

// Helper Functions

function renderCreatePostForm() {
    return `
        <div id="createPostForm" style="display: none;">
            <h1>Create a New Post</h1>
            <form method="POST" action="/api/post" enctype="multipart/form-data" onsubmit="handlePostSubmission(event)">
                <label for="title">Title:</label>
                <input type="text" id="title" name="title" required>
                
                <label for="content">Content:</label>
                <textarea id="content" name="content" required></textarea>
                
                <label for="image">Image:</label>
                <input type="file" id="image" name="image" accept="image/jpeg, image/png, image/gif">
                
                <label for="category">Category:</label>
                <div id="category" class="checkbox-group">
                    ${CATEGORIES.map(cat => `
                        <label><input type="checkbox" name="category" value="${cat}"> ${capitalize(cat)}</label>
                    `).join('')}
                </div>

                <button type="submit">Post</button>
                <button type="button" onclick="toggleCreatePost()">Cancel</button>
            </form>
        </div>
    `;
}

function renderSidebar() {
    return `
        <aside class="sidebar" id="sidebar">
            <h3>Categories</h3>
            <ul>
                <li><a href="#/">All Posts</a></li>
                ${CATEGORIES.map(cat => `<li><a href="/filter?category=${cat}">${capitalize(cat)}</a></li>`).join('')}
            </ul>
        </aside>
    `;
}

function renderPosts(posts) {
    console.log(posts)
    return posts.map(post => `
        <div class="post" data-category="${post.Categories}">
            <p class="posted-on">${post.CreatedAtHuman}</p>
            <strong><p>${post.Username}</p></strong>
            <h3>${post.Title}</h3>
            <p>${post.Content}</p>
            ${post.ImagePath ? `<img src="${post.ImagePath}" alt="Post Image" class="post-image">` : ''}
            <p class="categories">Categories: <span>${post.Categories}</span></p>
            <div class="post-actions">
                <button class="like-button" data-post-id="${post.ID}" onclick="toggleLike('${post.ID}', true)">
                    <i class="fas fa-thumbs-up"></i> <span class="like-count">${post.LikeCount}</span>
                </button>
                <button class="dislike-button" data-post-id="${post.ID}" onclick="toggleLike('${post.ID}', false)">
                    <i class="fas fa-thumbs-down"></i> <span class="dislike-count">${post.DislikeCount}</span>
                </button>
                <button class="comment-button" onclick="toggleCommentForm('${post.ID}')">
                    <i class="fas fa-comment"></i> Comments
                </button>
            </div>
        </div>
    `).join('');
}

// Utility Function
function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

async function handlePostSubmission(event) {
    event.preventDefault(); // Prevent default form submission

    // Validate categories before proceeding
    if (!validateCategories()) {
        return; // Stop execution if no category is selected
    }

    const form = event.target;
    const formData = new FormData(form);

    try {
        const response = await fetch('/api/post', {
            method: 'POST',
            body: formData
        });

        const result = await response.json();

        if (response.ok) {
            alert('Post submitted successfully!');
            form.reset(); // Clear the form after submission
            fetchHomeContent().then(content => document.getElementById('app').innerHTML = content);
        } else {
            alert(`Error: ${result.message}`);
        }
    } catch (error) {
        console.error('Error submitting post:', error);
        alert('Failed to submit post. Please try again.');
    }
}



// Fetch profile content
async function fetchProfileContent() {
    
    const response = await fetch('/api/profile');
    const data = await response.json();
    return `
    <div class="profile-container">
        <div class="profile-header">
            <h1><i class="fas fa-user-circle"></i> ${data.Username}'s Profile</h1>
            <p><i class="fas fa-envelope"></i> ${data.Email}</p>
        </div>

        <div class="profile-sections">
            <!-- Created Posts Section -->
            <section class="profile-section">
                <h2><i class="fas fa-pencil-alt"></i> Your Posts</h2>
                ${data.CreatedPosts && data.CreatedPosts.length > 0 
                    ? data.CreatedPosts.map(post => `
                        <article class="post">
                            <h3>${post.Title}</h3>
                            <p class="post-content">${post.Content}</p>
        ${post.imagePath ? `<img src="${post.ImagePath}" alt="Post Image" class="post-image" />` : ""}
        <div class="post-meta">
            ${post.Categories ? `<span class="categories"><i class="fas fa-tags"></i> ${post.Categories}</span>` : ""}
            <span class="likes"><i class="fas fa-thumbs-up"></i> ${post.LikeCount}</span>
            <span class="dislikes"><i class="fas fa-thumbs-down"></i> ${post.DislikeCount}</span>
            <span class="date"><i class="far fa-clock"></i> ${post.CreatedAtHuman}</span>
        </div>
                        </article>
                    `).join('')
                    : '<p class="empty-message">You haven\'t created any posts yet.</p>'
                }
            </section>

            <!-- Liked Posts Section -->
            <section class="profile-section">
                <h2><i class="fas fa-heart"></i> Liked Posts</h2>
                ${data.LikedPosts && data.LikedPosts.length > 0 
                    ? data.LikedPosts.map(post => `
                        <article class="post">
                            <h3>${post.Title}</h3>
                            <p class="post-content">${post.Content}</p>
                            ${post.ImagePath 
                                ? `<img src="${post.ImagePath}" alt="Post Image" class="post-image" />`
                                : ''
                            }
                            <div class="post-meta">
                                ${post.Categories 
                                    ? `<span class="categories"><i class="fas fa-tags"></i> ${post.Categories}</span>`
                                    : ''
                                }
                                <span class="likes"><i class="fas fa-thumbs-up"></i> ${post.LikeCount}</span>
                                <span class="dislikes"><i class="fas fa-thumbs-down"></i> ${post.DislikeCount}</span>
                                <span class="date"><i class="far fa-clock"></i> ${post.CreatedAtHuman}</span>
                            </div>
                        </article>
                    `).join('')
                    : '<p class="empty-message">You haven\'t liked any posts yet.</p>'
                }
            </section>
        </div>
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
            <p class="home-link"><a href="#/home">← Back to Homepage</a></p>
        </div>
    `;
}

// Handle login form submission
async function handleLogin(event) {
    event.preventDefault();
    const formData = new FormData(event.target);

    // Send the login data to the backend
    const response = await fetch('/auth/login', {
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
        window.location.hash = '/home'; // Redirect to home page
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
            <p class="home-link"><a href="#/home">← Back to Homepage</a></p>
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
    const response = await fetch('/auth/register', {
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
    const response = await fetch('/auth/checklogin');
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


//  home utils
let isProcessing = false; // Debounce flag

function toggleLike(postId, isLike) {
    if (isProcessing) return; // Prevent multiple rapid clicks
    isProcessing = true;

    fetch('/like', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `post_id=${postId}&is_like=${isLike}`
    })
        .then(response => {
            if (response.status === 401) {
                // User is not logged in, redirect to login page
                window.location.href = '/login';
                return;
            }
            return response.json();
        })
        .then(data => {
            if (data && data.success) {
                // Update the like and dislike counts
                const likeCountElement = document.querySelector(`.like-button[data-post-id="${postId}"] .like-count`);
                const dislikeCountElement = document.querySelector(`.dislike-button[data-post-id="${postId}"] .dislike-count`);

                likeCountElement.textContent = data.like_count;
                dislikeCountElement.textContent = data.dislike_count;

                // Update button styles
                const likeButton = document.querySelector(`.like-button[data-post-id="${postId}"]`);
                const dislikeButton = document.querySelector(`.dislike-button[data-post-id="${postId}"]`);

                if (isLike) {
                    likeButton.classList.toggle('active');
                    dislikeButton.classList.remove('active');
                } else {
                    dislikeButton.classList.toggle('active');
                    likeButton.classList.remove('active');
                }
            } else if (data && data.success) {
                console.error('Error:', data.error);
                alert(data.error); // Optional: Show the error message
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('An error occurred. Please try again.');
        })
        .finally(() => {
            isProcessing = false; // Reset debounce flag
        });
}

function toggleCommentLike(commentId, isLike) {
    if (isProcessing) return; // Prevent multiple rapid clicks
    isProcessing = true;

    fetch('/comment/like', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `comment_id=${commentId}&is_like=${isLike}`
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    throw new Error(text);
                });
            }
            return response.json();
        })
        .then(data => {
            // Find the clicked button directly using the data-comment-id attribute
            const clickedButton = document.querySelector(`button[data-comment-id="${commentId}"][class*="-button"]`);
            if (!clickedButton) return;

            // Find the parent comment container
            const comment = clickedButton.closest('.comment');
            if (!comment) return;

            // Find the like/dislike buttons within this comment
            const likeButton = comment.querySelector('.like-button[data-comment-id="' + commentId + '"]');
            const dislikeButton = comment.querySelector('.dislike-button[data-comment-id="' + commentId + '"]');

            if (!likeButton || !dislikeButton) return;

            const likeCountElement = likeButton.querySelector('.like-count');
            const dislikeCountElement = dislikeButton.querySelector('.dislike-count');

            // Update the counts
            if (likeCountElement) likeCountElement.textContent = data.likeCount;
            if (dislikeCountElement) dislikeCountElement.textContent = data.dislikeCount;

            // Update button styles based on userLiked
            if (data.userLiked === true) {
                likeButton.classList.add('active');
                dislikeButton.classList.remove('active');
            } else if (data.userLiked === false) {
                dislikeButton.classList.add('active');
                likeButton.classList.remove('active');
            } else {
                // If userLiked is null, remove both active states
                likeButton.classList.remove('active');
                dislikeButton.classList.remove('active');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert(error.message || 'An error occurred. Please try again.');
        })
        .finally(() => {
            isProcessing = false; // Reset debounce flag
        });
}

function toggleCreatePost() {
    const createPostForm = document.getElementById('createPostForm');
    const postsList = document.getElementById('posts');
    const postsHeading = document.getElementById('postsHeading');

    if (createPostForm.style.display === 'none') {
        createPostForm.style.display = 'block';
        postsList.style.display = 'none';
        postsHeading.style.display = 'none'; // Hide the heading when the form is shown
    } else {
        createPostForm.style.display = 'none';
        postsList.style.display = 'block';
        postsHeading.style.display = 'block'; // Show the heading when the form is hidden
    }
}

function validateCategories() {
    const checkboxes = document.querySelectorAll('input[name="category"]');
    let isChecked = false;

    checkboxes.forEach((checkbox) => {
        if (checkbox.checked) {
            isChecked = true;
        }
    });

    if (!isChecked) {
        alert("Please select at least one category.");
        return false; // Prevent form submission
    }
    return true; // Allow form submission
}

function toggleCommentForm(postId) {
    const commentForm = document.getElementById(`comment-form-${postId}`);
    const commentsSection = document.getElementById(`comments-${postId}`);
    if (commentForm.style.display === 'none') {
        commentForm.style.display = 'block';
        commentsSection.style.display = 'block'; // Show comments section when the form is shown
    } else {
        commentForm.style.display = 'none';
        commentsSection.style.display = 'none'; // Hide comments section when the form is hidden
    }
}

function toggleReplyForm(commentId) {
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    if (replyForm.style.display === 'none') {
        replyForm.style.display = 'block';
    } else {
        replyForm.style.display = 'none';
    }
}

function toggleMenu() {
    const sidebar = document.getElementById('sidebar');
    sidebar.style.display = sidebar.style.display === 'block' ? 'none' : 'block';
}

// Function to validate the comment form
function validateCommentForm(event, form) {
    // Get the textarea element
    const textarea = form.querySelector('textarea[name="content"]');

    // Check if the textarea is empty or contains only whitespace
    if (!textarea.value.trim()) {
        // Prevent form submission
        event.preventDefault();

        // Alert the user
        alert("Comment cannot be empty. Please write something before submitting.");

        // Focus on the textarea so the user can continue typing
        textarea.focus();

        // Return false to prevent form submission
        return false;
    }

    // If the textarea is not empty, allow form submission
    return true;
}
