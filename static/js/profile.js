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