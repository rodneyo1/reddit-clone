// Shared constants and utilities
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

window.validateCategories = function() {
    const checkboxes = document.querySelectorAll('input[name="category"]:checked');
    if (checkboxes.length === 0) {
        alert("Please select at least one category.");
        return false;
    }
    return true;
};

window.validateCategories = validateCategories;
window.formatDate = formatDate;
window.normalizePost = normalizePost;