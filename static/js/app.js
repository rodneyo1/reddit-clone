window.addEventListener('DOMContentLoaded', () => {
    const initialPath = window.location.hash.replace('#', '') || '/login'; // Force login as default
    render(initialPath);
});

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
    // console.log(`Rendering path: ${path}`);
    const app = document.getElementById('app');
    const authButtons = document.getElementById('auth-buttons');
    const isLoggedIn = await checkLoginStatus();

    // if (path !== '/login' && path !== '/register' && !isLoggedIn) {
    //     window.location.hash = '/login';
    //     return;
    // }

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
                    case '/profile':
                    app.innerHTML = await fetchProfileContent();
                    window.attachProfileFormHandler();
                    break;
                case '/logout':
                    await handleLogout();
                    window.location.hash = '/login';
                    window.location.reload();
                    return
                    break;
                case '/home':
                    case '/filter':
                    app.innerHTML = await fetchHomeContent();
                    document.getElementById('post-form')?.addEventListener('submit', window.handlePostSubmit);
                    break;
                default:
                    app.innerHTML = '<h1>404 Not Found</h1>';
            }
        }
    } catch (error) {
        console.error('Error rendering content:', error);
        app.innerHTML = '<p class="error-message">Error loading content. Please try again.</p>';
    }


    
    // const isLoggedIn = await checkLoginStatus();
    
    const logo = document.getElementById('logo');
    // Update logo link
if (logo) {
    logo.innerHTML = `<a href="${isLoggedIn ? '#/home' : '#/'}" class="logo-link">Forum</a>`;
}

    authButtons.innerHTML = isLoggedIn
        ? `
            <div class="profile-icon" style="position: relative;">
                <a href="#/profile" class="material-icons" style="font-size:30px; color: #4A7C8C; margin-top: 10px; vertical-align: middle;">person</a>
            </div>
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
        const chatSidebar = document.getElementById('chat-sidebar');
        if (chatSidebar){
            chatSidebar.style.display = 'flex'; //show it after login
        }
         setTimeout(initChat, 100);
        document.getElementById('create-post-btn')?.addEventListener('click', function(e) {
            e.preventDefault();
            toggleCreatePost();
        });
    }
}

// Mobile menu functionality
function toggleMenu() {
    const nav = document.querySelector('nav');
    const sidebar = document.querySelector('.sidebar');
    nav.classList.toggle('active');
    
    // Close sidebar if it's open
    if (sidebar && sidebar.classList.contains('active')) {
        sidebar.classList.remove('active');
    }
}

// Close mobile menu when clicking outside
document.addEventListener('click', (e) => {
    const nav = document.querySelector('nav');
    const hamburger = document.querySelector('.hamburger');
    
    if (nav.classList.contains('active') && 
        !nav.contains(e.target) && 
        !hamburger.contains(e.target)) {
        nav.classList.remove('active');
    }
});

// Toggle chat sidebar on mobile
function toggleChatSidebar() {
    const chatSidebar = document.querySelector('.chat-sidebar');
    if (chatSidebar) {
        chatSidebar.classList.toggle('active');
    }
}

// Handle touch events for better mobile interaction
document.addEventListener('DOMContentLoaded', () => {
    // Add touch feedback to buttons and links
    const interactiveElements = document.querySelectorAll('button, .auth-button, .post-actions button, nav a');
    
    interactiveElements.forEach(element => {
        element.addEventListener('touchstart', () => {
            element.style.opacity = '0.7';
        });
        
        element.addEventListener('touchend', () => {
            element.style.opacity = '1';
        });
    });

    // Handle swipe gestures for mobile navigation
    let touchStartX = 0;
    let touchEndX = 0;
    
    document.addEventListener('touchstart', (e) => {
        touchStartX = e.changedTouches[0].screenX;
    });
    
    document.addEventListener('touchend', (e) => {
        touchEndX = e.changedTouches[0].screenX;
        handleSwipe();
    });
    
    function handleSwipe() {
        const swipeThreshold = 50;
        const nav = document.querySelector('nav');
        const chatSidebar = document.querySelector('.chat-sidebar');
        
        // Swipe right to open nav menu
        if (touchEndX - touchStartX > swipeThreshold) {
            nav.classList.add('active');
        }
        // Swipe left to close nav menu
        else if (touchStartX - touchEndX > swipeThreshold) {
            nav.classList.remove('active');
            if (chatSidebar) {
                chatSidebar.classList.remove('active');
            }
        }
    }
});
