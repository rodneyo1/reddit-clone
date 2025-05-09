// Global state management
let currentChatUserId = null;
let messagePage = 1;
let isLoadingMessages = false;
let chatSocket = null;
let currentUserId = null;
let currentUsername = null;

// Main initialization (call this after user logs in)
function initMessagingSystem(userId, username) {
    currentUserId = userId;
    currentUsername = username;
    
    // Initialize WebSocket connection
    initWebSocket();
    
    // Load and display user list
    loadUserList()
        .then(users => {
            if (users.length > 0) {
                updateOnlineStatus(users);
            }
        })
        .catch(error => {
            console.error('Failed to load user list:', error);
        });
    
    // Setup all event listeners
    setupChatEventListeners();
    
    // Show the chat sidebar
    document.getElementById('chat-sidebar').style.display = 'flex';
}

// WebSocket Management
function initWebSocket() {
    // Prevent duplicate connections
    if (chatSocket && [WebSocket.OPEN, WebSocket.CONNECTING].includes(chatSocket.readyState)) {
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    chatSocket = new WebSocket(wsUrl);
    
    chatSocket.onopen = () => {
        console.log('WebSocket connected');
        authenticateWebSocket();
    };
    
    chatSocket.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            handleSocketMessage(message);
        } catch (error) {
            console.error('Error parsing message:', error);
        }
    };
    
    chatSocket.onclose = () => {
        console.log('WebSocket disconnected');
        setTimeout(initWebSocket, 5000); // Reconnect after 5 seconds
    };
    
    chatSocket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

function authenticateWebSocket() {
    const sessionToken = getCookie('session-name');
    if (sessionToken && chatSocket.readyState === WebSocket.OPEN) {
        chatSocket.send(JSON.stringify({
            type: 'auth',
            token: sessionToken,
            userId: currentUserId
        }));
    }
}

function handleSocketMessage(message) {
    if (!message.type) {
        console.warn('Received message without type:', message);
        return;
    }

    switch(message.type) {
        case 'message':
            handleIncomingMessage(message);
            break;
        case 'status_update':
            updateUserStatus(message.userId, message.isOnline);
            break;
        case 'message_read':
            handleReadReceipt(message);
            break;
        case 'typing_indicator':
            handleTypingIndicator(message);
            break;
        case 'error':
            console.error('Server error:', message.error);
            break;
        default:
            console.warn('Unhandled message type:', message.type);
    }
}

// User List Management
async function loadUserList() {
    try {
        const response = await fetch('/api/users/status');
        if (!response.ok) throw new Error('Failed to fetch users');
        return await response.json();
    } catch (error) {
        console.error('Error loading user list:', error);
        throw error;
    }
}

function renderUserList(users) {
    const userListContainer = document.getElementById('user-list');
    if (!userListContainer) return;

    userListContainer.innerHTML = '';

    // Sort by online status then alphabetically
    users.sort((a, b) => {
        if (a.is_online !== b.is_online) return b.is_online - a.is_online;
        return a.username.localeCompare(b.username);
    });

    users.forEach(user => {
        const userElement = document.createElement('div');
        userElement.className = `user-item ${user.is_online ? 'online' : 'offline'}`;
        userElement.dataset.userId = user.id;
        userElement.innerHTML = `
            <div class="user-avatar">${user.username.charAt(0).toUpperCase()}</div>
            <div class="user-info">
                <span class="username">${user.username}</span>
                <span class="status">${user.is_online ? 'Online' : 'Offline'}</span>
                ${user.unread_count > 0 ? `<span class="unread-badge">${user.unread_count}</span>` : ''}
            </div>
        `;
        userElement.addEventListener('click', () => openChat(user.id));
        userListContainer.appendChild(userElement);
    });
}

function updateUserStatus(userId, isOnline) {
    const userElement = document.querySelector(`.user-item[data-user-id="${userId}"]`);
    if (userElement) {
        userElement.classList.toggle('online', isOnline);
        userElement.classList.toggle('offline', !isOnline);
        const statusElement = userElement.querySelector('.status');
        if (statusElement) {
            statusElement.textContent = isOnline ? 'Online' : 'Offline';
        }
    }
}

// Chat Management
function openChat(userId) {
    if (!userId || userId === currentChatUserId) return;

    currentChatUserId = userId;
    messagePage = 1;

    // Update UI
    document.querySelectorAll('.user-item').forEach(el => {
        el.classList.toggle('active', el.dataset.userId === userId);
    });

    // Show chat container
    const chatContainer = document.getElementById('chat-container');
    if (chatContainer) chatContainer.style.display = 'block';

    // Clear and load messages
    const messagesContainer = document.getElementById('messages-container');
    if (messagesContainer) messagesContainer.innerHTML = 'Loading...';

    loadMessages()
        .then(() => {
            scrollToBottom();
            markMessagesAsRead();
            document.getElementById('message-input')?.focus();
        })
        .catch(error => {
            console.error('Error opening chat:', error);
            if (messagesContainer) messagesContainer.innerHTML = 'Failed to load messages';
        });
}

async function loadMessages() {
    if (isLoadingMessages || !currentChatUserId) return;
    isLoadingMessages = true;

    try {
        const response = await fetch(`/api/messages/${currentChatUserId}?page=${messagePage}`);
        if (!response.ok) throw new Error('Failed to fetch messages');
        
        const messages = await response.json();
        const messagesContainer = document.getElementById('messages-container');
        if (!messagesContainer) return;

        if (messagePage === 1) {
            messagesContainer.innerHTML = '';
        }

        if (messages.length > 0) {
            messages.reverse().forEach(message => {
                prependMessage(message);
            });

            if (messagePage === 1 || messagesContainer.scrollTop === 0) {
                setTimeout(() => {
                    messagesContainer.children[0]?.scrollIntoView();
                }, 50);
            }
            
            messagePage++;
        } else if (messagePage === 1) {
            messagesContainer.innerHTML = '<div class="no-messages">No messages yet</div>';
        }
    } catch (error) {
        console.error('Error loading messages:', error);
        throw error;
    } finally {
        isLoadingMessages = false;
    }
}

function sendMessage() {
    const input = document.getElementById('message-input');
    if (!input) return;

    const content = input.value.trim();
    if (!content || !currentChatUserId || !chatSocket) return;

    const message = {
        type: 'message',
        sender_id: currentUserId,
        receiver_id: currentChatUserId,
        content: content,
        timestamp: new Date().toISOString()
    };

    // Optimistic UI update
    appendMessage({
        ...message,
        sender: { id: currentUserId, username: currentUsername },
        status: 'sending'
    });

    input.value = '';
    scrollToBottom();

    // Send to server
    try {
        if (chatSocket.readyState === WebSocket.OPEN) {
            chatSocket.send(JSON.stringify(message));
        } else {
            throw new Error('Connection not ready');
        }
    } catch (error) {
        console.error('Failed to send message:', error);
        updateMessageStatus(message.timestamp, 'failed');
    }
}

function handleIncomingMessage(message) {
    if (![message.sender_id, message.receiver_id].includes(currentUserId)) return;

    // If this is for the current chat, display it
    if (message.sender_id === currentChatUserId || message.receiver_id === currentChatUserId) {
        appendMessage(message);
        scrollToBottom();
    }

    // Update unread count in user list
    if (message.sender_id !== currentUserId) {
        updateUnreadCount(message.sender_id);
    }
}

// UI Helpers
function appendMessage(message) {
    const messagesContainer = document.getElementById('messages-container');
    if (!messagesContainer) return;

    const messageElement = createMessageElement(message);
    messagesContainer.appendChild(messageElement);
}

function prependMessage(message) {
    const messagesContainer = document.getElementById('messages-container');
    if (!messagesContainer) return;

    const messageElement = createMessageElement(message);
    messagesContainer.insertBefore(messageElement, messagesContainer.firstChild);
}

function createMessageElement(message) {
    const isCurrentUser = message.sender_id === currentUserId;
    const messageDate = formatMessageDate(message.timestamp || message.created_at);
    
    const messageElement = document.createElement('div');
    messageElement.className = `message ${isCurrentUser ? 'sent' : 'received'}`;
    messageElement.dataset.timestamp = message.timestamp || message.created_at;
    
    messageElement.innerHTML = `
        <div class="message-content">${message.content}</div>
        <div class="message-meta">
            <span class="message-sender">${isCurrentUser ? 'You' : message.sender?.username || 'Unknown'}</span>
            <span class="message-time">${messageDate}</span>
            ${message.status === 'sending' ? '<span class="message-status sending"></span>' : ''}
            ${message.status === 'failed' ? '<span class="message-status failed" title="Failed to send"></span>' : ''}
        </div>
    `;
    
    return messageElement;
}

function updateMessageStatus(timestamp, status) {
    const messageElement = document.querySelector(`.message[data-timestamp="${timestamp}"]`);
    if (messageElement) {
        const statusElement = messageElement.querySelector('.message-status');
        if (statusElement) {
            statusElement.className = `message-status ${status}`;
            statusElement.title = status === 'failed' ? 'Failed to send' : '';
        }
    }
}

function scrollToBottom() {
    const container = document.getElementById('messages-container');
    if (container) {
        setTimeout(() => {
            container.scrollTop = container.scrollHeight;
        }, 50);
    }
}

// Event Handling
function setupChatEventListeners() {
    // Send message on button click or Enter key
    document.addEventListener('click', (e) => {
        if (e.target?.id === 'send-button') {
            sendMessage();
        }
    });

    const messageInput = document.getElementById('message-input');
    if (messageInput) {
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Typing indicators
        messageInput.addEventListener('input', throttle(() => {
            if (chatSocket?.readyState === WebSocket.OPEN && currentChatUserId) {
                chatSocket.send(JSON.stringify({
                    type: 'typing_indicator',
                    receiver_id: currentChatUserId,
                    is_typing: true
                }));
            }
        }, 1000));
    }

    // Infinite scroll
    const messagesContainer = document.getElementById('messages-container');
    if (messagesContainer) {
        messagesContainer.addEventListener('scroll', throttle(() => {
            if (messagesContainer.scrollTop === 0) {
                loadMessages();
            }
        }, 200));
    }
}

// Utility Functions
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

function formatMessageDate(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// Export the initialization function
export { initMessagingSystem };