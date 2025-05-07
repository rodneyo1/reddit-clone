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

async function handlePostSubmit(event) {
    event.preventDefault();

    if (!window.validateCategories || typeof window.validateCategories !== 'function') {
        console.error("validateCategories is not available!");
        return;
    }
    
    if (!window.validateCategories()) {
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

window.renderPost = renderPost;
window.fetchHomeContent = fetchHomeContent;
window.handleLikeAction = handleLikeAction;
window.handlePostSubmit = handlePostSubmit;
window.toggleCreatePost = toggleCreatePost;
// window.validateCategories = validateCategories;