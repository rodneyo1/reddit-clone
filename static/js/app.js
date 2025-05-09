window.addEventListener('DOMContentLoaded', () => {
    const initialPath = window.location.hash.replace('#', '') || '/login'; // Force login as default
    render(initialPath);
});

// Then keep your existing hashchange listener
window.addEventListener('hashchange', () => {
    const path = window.location.hash.replace('#', '');
    render(path);
});
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

async function render(path) {
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
                case '/messages':
                    if (!isLoggedIn) {
                        window.location.hash = '/login';
                        return;
                    }
                    app.innerHTML = await fetchMessagesContent();
                    initChat();
                    break;
                case '/logout':
                    await handleLogout();
                    window.location.hash = '/login';
                    window.location.reload();
                    return;
                case '/home':
                case '/filter':
                default:
                    // Main content view (homepage or filtered posts)
                    app.innerHTML = await fetchHomeContent();
                    document.getElementById('post-form')?.addEventListener('submit', window.handlePostSubmit);
                    
                    // Initialize chat sidebar if logged in (persistent chat)
                    if (isLoggedIn) {
                        initChat();
                    }
                    break;
            }
        }
    } catch (error) {
        console.error('Error rendering content:', error);
        app.innerHTML = '<p class="error-message">Error loading content. Please try again.</p>';
    }

    // Update navigation elements
    const logo = document.getElementById('logo');
    if (logo) {
        logo.innerHTML = `<a href="${isLoggedIn ? '#/home' : '#/'}" class="logo-link">Forum</a>`;
    }

    // Update auth buttons - now includes messages link
    authButtons.innerHTML = isLoggedIn
        ? `
            <div class="profile-icon" style="position: relative;">
                <a href="#/profile" class="material-icons" style="font-size:30px; color: #4A7C8C; margin-top: 10px; vertical-align: middle;">person</a>
            </div>
            <a href="#/messages" class="auth-button messages" title="Messages">
                <i class="fas fa-envelope" style="font-size: 24px; color: #4A7C8C; margin-top: 10px;"></i>
            </a>
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
