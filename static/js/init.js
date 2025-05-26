window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});

(async () => {
    const isLoggedIn = await checkLoginStatus();
    if (isLoggedIn) {
        try {
            await initChat();
        } catch (error) {
            console.error('Chat initialization failed:', error);
            const chatContainer = document.getElementById('chat-sidebar');
            if (chatContainer) {
                chatContainer.innerHTML = '<div class="error-message">Chat unavailable</div>';
            }
        }
    } else {
        // hide chat sidebar
        const chatContainer = document.getElementById('chat-sidebar');
        if (chatContainer) {
            chatContainer.style.display = 'none';
        }
    }
})();


// Attach global functions
window.toggleCreatePost = toggleCreatePost;
window.handleLikeAction = handleLikeAction;
window.toggleCommentForm = toggleCommentForm;
window.toggleReplyForm = toggleReplyForm;
window.handleCommentSubmit = handleCommentSubmit;
// window.handleCommentLike = handleCommentLike;

const initialPath = window.location.hash.replace('#', '') || '/';
render(initialPath);