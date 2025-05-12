function toggleCommentForm(postId) {
    const commentForm = document.getElementById(`comment-form-${postId}`);
    const commentsSection = document.getElementById(`comments-${postId}`);
    
    // Always show "No comments" by default when opening
    if (commentForm.style.display === 'none' || !commentForm.style.display) {
        commentForm.style.display = 'block';
        commentsSection.innerHTML = '<p class="no-comments">Loading comments...</p>';
        commentsSection.style.display = 'block';
        loadComments(postId); // Load after showing the section
    } else {
        commentForm.style.display = 'none';
        commentsSection.style.display = 'none';
    }
}

async function loadComments(postId) {
    const commentsSection = document.getElementById(`comments-${postId}`);
    if (!commentsSection) return;

    try {
        const response = await fetch(`/api/comments?post_id=${postId}`);
        
        // First check if the response is successful
        if (!response.ok) {
            throw new Error(`Server returned ${response.status}`);
        }

        // Check if response has data
        const data = await response.json();
        
        // Handle empty comments array
        if (!data || !Array.isArray(data)) {
            commentsSection.innerHTML = '<p class="no-comments">No comments yet. Be the first to comment!</p>';
            return;
        }

        // If we get here, render the comments
        commentsSection.innerHTML = data.length > 0 
            ? renderComments(data) 
            : '<p class="no-comments">No comments yet. Be the first to comment!</p>';

    } catch (error) {
        console.error('Comment load error:', error);
        commentsSection.innerHTML = `
            <p class="no-comments">No comments yet</p>
            <button class="retry-btn" onclick="loadComments(${postId})">
                Try Again
            </button>
        `;
    }
}

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

function toggleReplyForm(commentId) {
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    if (replyForm.style.display === 'none' || !replyForm.style.display) {
        replyForm.style.display = 'block';
    } else {
        replyForm.style.display = 'none';
    }
}

// window.handleCommentLike = handleCommentLike;
window.toggleCommentForm = toggleCommentForm;
window.handleCommentSubmit = handleCommentSubmit;