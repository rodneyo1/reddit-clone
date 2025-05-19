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
        console.log(profileData)
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
        <div class="avatar">
            ${profileData.AvatarURL 
                ? `<img src="${profileData.AvatarURL}" alt="Avatar" class="avatar-img">` 
                : `<i class="fas fa-user-circle fa-4x"></i>`}
        </div>
        <h1><i class="fas fa-user-circle"></i> ${profileData.Username}'s Profile</h1>
        <p><i class="fas fa-envelope"></i> ${profileData.Email}</p>
        ${profileData.Nickname ? `<p><i class="fas fa-smile"></i> Nickname: ${profileData.Nickname}</p>` : ''}
        ${profileData.FirstName || profileData.LastName 
            ? `<p><i class="fas fa-id-card"></i> Name: ${profileData.FirstName || ''} ${profileData.LastName || ''}</p>` 
            : ''}
        ${profileData.Age ? `<p><i class="fas fa-birthday-cake"></i> Age: ${profileData.Age}</p>` : ''}
        ${profileData.Gender ? `<p><i class="fas fa-venus-mars"></i> Gender: ${profileData.Gender}</p>` : ''}
        
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
<div class="update-profile">
       <form id="profile-update-form">
    <h2>Update Your Profile</h2>

    <!-- Hidden field for user ID (optional if using sessions/JWTs) -->
    <input type="hidden" name="id" value="${profileData.ID}">

    <label>
        Nickname:
        <input type="text" name="nickname" value="${profileData.Nickname || ''}">
    </label>

    <label>
        Avatar URL:
        <input type="url" name="avatar_url" value="${profileData.AvatarURL || ''}">
    </label>

    <label>
        Age:
        <input type="number" name="age" min="0" value="${profileData.Age || ''}">
    </label>

    <label>
        Gender:
        <select name="gender">
            <option value="">-- Select Gender --</option>
            <option value="Male" ${profileData.Gender === "Male" ? "selected" : ""}>Male</option>
            <option value="Female" ${profileData.Gender === "Female" ? "selected" : ""}>Female</option>
            <option value="Other" ${profileData.Gender === "Other" ? "selected" : ""}>Other</option>
        </select>
    </label>

    <label>
        First Name:
        <input type="text" name="first_name" value="${profileData.FirstName || ''}">
    </label>

    <label>
        Last Name:
        <input type="text" name="last_name" value="${profileData.LastName || ''}">
    </label>

    <button type="submit">Update Profile</button>
</form>
     
</div>

        `;
    } catch (error) {
        console.error('Error fetching profile:', error);
        return '<p class="error-message">Failed to load profile. Please try again.</p>';
    }
}

// profile.js
window.attachProfileFormHandler = function () {
    const form = document.getElementById("profile-update-form");
    if (!form) return;

    form.addEventListener("submit", async function (e) {
        e.preventDefault();
        const formData = new FormData(this);
        const payload = {};
        for (const [key, value] of formData.entries()) {
            if (value !== "") {
                payload[key] = key === "age" ? parseInt(value, 10) : value;
            }
        }

        try {
            const res = await fetch("/api/profile/update", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            const result = await res.json();
            if (result.success) {
                alert("Profile updated successfully!");
                location.reload();
            } else {
                alert("Error updating profile: " + (result.message || "Unknown error"));
            }
        } catch (err) {
            console.error("Failed to update profile:", err);
            alert("Something went wrong. Please try again.");
        }
    });
};
