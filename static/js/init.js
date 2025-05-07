window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});

// Attach global functions
window.toggleCreatePost = toggleCreatePost;
window.handleLikeAction = handleLikeAction;
window.toggleCommentForm = toggleCommentForm;
window.toggleReplyForm = toggleReplyForm;
window.handleCommentSubmit = handleCommentSubmit;
window.handleCommentLike = handleCommentLike;

const initialPath = window.location.hash.replace('#', '') || '/';
render(initialPath);