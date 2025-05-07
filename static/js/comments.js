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

window.handleCommentLike = handleCommentLike;
window.toggleCommentForm = toggleCommentForm;
window.handleCommentSubmit = handleCommentSubmit;