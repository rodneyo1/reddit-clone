let chatSocket = null;
let currentRecipient = null;
let messageOffset = 0;
let isLoadingMessages = false;
let hasMoreMessages = true;

// Initialize chat functionality
function initChat() {
    console.log('initChat is running');
    connectWebSocket();
    loadChatUsers();
    setupEventListeners();
}

// Connect to WebSocket
function connectWebSocket() {
    if (chatSocket && chatSocket.readyState !== WebSocket.CLOSED) {
        chatSocket.close();
    }
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    chatSocket = new WebSocket(`${protocol}//${window.location.host}/ws/chat`);
    chatSocket.onopen = () => {
        console.log('WS connected');
        loadChatUsers(); // Refresh on reconnect
    };

    chatSocket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
    };

    chatSocket.onclose = (event) => {
         clearInterval(pingInterval);
         if (event.code !== 1000 && event.code !== 1001) { // Only log abnormal closes
            console.log('WebSocket closed abnormally:', event);
        }
        setTimeout(connectWebSocket, 3000);
    };

    chatSocket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

// Load chat users list
async function loadChatUsers() {
    try {
        const response = await fetch('/api/chat/users');
        if (!response.ok) {
            console.error('Server response:', response.status);
            return;
        }
        
        const data = await response.json();
        console.log('Fetched users:', data);
        
        // Ensure data is an array before mapping
        if (!Array.isArray(data)) {
            throw new Error('Invalid user data format');
        }
        
        renderChatUsers(data);
    } catch (error) {
        console.error('Error loading chat users:', error);
        renderChatUsers([]); // Render empty list on error
    }
}

// Render chat users list
function renderChatUsers(users = []) {
    const userList = document.getElementById('user-list');
    if (!userList) return;

    userList.innerHTML = users.map(user => `
        <div class="chat-user" data-user-id="${user.id}">
            <div class="user-avatar">
                ${user.avatar_url ? 
                    `<img src="${user.avatar_url}" alt="${user.username}">` : 
                    `<div class="default-avatar">${user.username.charAt(0).toUpperCase()}</div>`}
                <span class="status-indicator ${user.is_online ? 'online' : 'offline'}"></span>
            </div>
            <div class="user-info">
                <span class="username">${user.username}</span>
                ${user.last_message ? 
                    `<span class="last-message">${user.last_message.substring(0,30)}${user.last_message.length > 30 ? '...' : ''}</span>` : 
                    ''}
            </div>
            ${user.unread_count > 0 ? `<span class="unread-count">${user.unread_count}</span>` : ''}
        </div>
    `).join('');

    // Add event listeners properly
    document.querySelectorAll('.chat-user').forEach(userEl => {
        userEl.addEventListener('click', () => {
            const userId = userEl.dataset.userId;
            const username = userEl.querySelector('.username').textContent;
            openChat(userId, username);
        });
    });
}
// Open chat with a specific user
async function openChat(userId, username) {
    currentRecipient = userId;
    messageOffset = 0;
    hasMoreMessages = true;
    
    // Update UI
    document.getElementById('chat-recipient').textContent = username;
    document.getElementById('chat-container').style.display = 'block';
    document.getElementById('messages-list').innerHTML = '<div class="loading">Loading messages...</div>';
    
    // Load initial messages
    await loadMessages();
    
    // Scroll to bottom
    setTimeout(() => {
        const messagesList = document.getElementById('messages-list');
        messagesList.scrollTop = messagesList.scrollHeight;
    }, 100);
}

// Load messages with pagination
async function loadMessages() {
    if (isLoadingMessages || !hasMoreMessages) return;
    
    isLoadingMessages = true;
    try {
        const response = await fetch(`/api/chat/messages?recipient_id=${currentRecipient}&offset=${messageOffset}`);
        if (!response.ok) throw new Error('Failed to load messages');
        
        const messages = await response.json();
        
        if (messages.length === 0) {
            hasMoreMessages = false;
            if (messageOffset === 0) {
                document.getElementById('messages-list').innerHTML = '<div class="no-messages">No messages yet. Start the conversation!</div>';
            }
            return;
        }
        
        // Reverse to show oldest first (since we load newest first)
        messages.reverse();
        
        const messagesHTML = messages.map(msg => createMessageElement(msg)).join('');
        
        if (messageOffset === 0) {
            document.getElementById('messages-list').innerHTML = messagesHTML;
        } else {
            const messagesList = document.getElementById('messages-list');
            const scrollHeightBefore = messagesList.scrollHeight;
            const scrollTopBefore = messagesList.scrollTop;
            
            messagesList.insertAdjacentHTML('afterbegin', messagesHTML);
            
            const scrollHeightAfter = messagesList.scrollHeight;
            messagesList.scrollTop = scrollTopBefore + (scrollHeightAfter - scrollHeightBefore);
        }
        
        messageOffset += messages.length;
    } catch (error) {
        console.error('Error loading messages:', error);
    } finally {
        isLoadingMessages = false;
    }
}

// Create message element
function createMessageElement(msg) {
    // Determine message ownership using both server flag and client check
    const isCurrentUser = msg.is_owner || msg.sender_id === getCurrentUserId();
    const messageTime = formatMessageTime(new Date(msg.created_at));
    const username = isCurrentUser ? 'You' : (msg.sender_username || 'Unknown');
    
    // Handle avatar with fallbacks
    const avatarContent = msg.sender_avatar ? 
        `<img src="${msg.sender_avatar}" class="message-avatar" alt="${username}">` :
        `<div class="message-avatar-default">${username.charAt(0).toUpperCase()}</div>`;

    // Unique identifier for deduplication
    const messageId = msg.id ? msg.id : `temp-${msg.temp_id || Date.now()}`;

    return `
        <div class="message ${isCurrentUser ? 'sent' : 'received'}" 
             data-message-id="${messageId}"
             data-sender="${msg.sender_id}"
             data-recipient="${msg.recipient_id}">
            <div class="message-header">
                ${avatarContent}
                <div class="message-metadata">
                    <span class="sender">${username}</span>
                    <span class="time">${messageTime}</span>
                </div>
            </div>
            <div class="message-content">${msg.content}</div>
        </div>
    `;
}

// Format message time
function formatMessageTime(date) {
    const now = new Date();
    const diff = now - date;
    
    if (diff < 86400000) { // Less than 24 hours
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diff < 604800000) { // Less than 7 days
        return date.toLocaleDateString([], { weekday: 'short' });
    } else {
        return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
}

// Send message
function sendMessage() {
    const input = document.getElementById('message-input');
    const content = input.value.trim();
    
    if (!content || !currentRecipient || !chatSocket) return;
    
    const tempId = Date.now().toString();
    const tempMessage = {
        id: tempId,
        recipient_id: currentRecipient,
        content: content,
        created_at: new Date().toISOString(),
        sender_id: getCurrentUserId(),
        sender_username: 'You',
        sender_avatar: '',
        is_owner: true
    };
    
    // Add temp message
    const messagesList = document.getElementById('messages-list');

    if (currentRecipient === tempMessage.recipient_id) {
        messagesList.insertAdjacentHTML('beforeend', createMessageElement({
            ...tempMessage,
            is_owner: true
        }));
        messagesList.scrollTop = messagesList.scrollHeight;
    }
    // messagesList.insertAdjacentHTML('beforeend', createMessageElement(tempMessage));
    messagesList.scrollTop = messagesList.scrollHeight;
    
    // Send via WebSocket
    chatSocket.send(JSON.stringify({
        recipient_id: currentRecipient,
        content: content,
        temp_id: tempId
    }));
    
    input.value = '';
}

// Handle WebSocket messages
// Update handleWebSocketMessage
function handleWebSocketMessage(data) {
    if (data.type === 'status_update') {
        updateUserStatus(data.user_id, data.is_online);
        loadChatUsers();
        return;
    }
    
    // Handle message
    const message = data;
    const isCurrentUser = message.sender_id === getCurrentUserId();
    
    // Only add to UI if relevant
     if (message.temp_id) {
        const tempElement = document.querySelector(`[data-message-id="temp-${message.temp_id}"]`);
        if (tempElement) {
            tempElement.remove();
        }
    }
    
   // Check for existing message
    const existing = document.querySelector(`[data-message-id="${message.id}"]`);
    if (!existing) {
        const messagesList = document.getElementById('messages-list');
        messagesList.insertAdjacentHTML('beforeend', createMessageElement(message));
        messagesList.scrollTop = messagesList.scrollHeight;
    }

    // Update user list and notifications
    updateUserLastMessage(message.sender_id, message.content);
    if (!isCurrentUser && message.sender_id !== currentRecipient) {
        new Audio('/static/sounds/notification.mp3').play().catch(() => {});
    }
}

// Update user status in the UI
function updateUserStatus(userId, isOnline) {
    const userElement = document.querySelector(`.chat-user[data-user-id="${userId}"]`);
    if (userElement) {
        const indicator = userElement.querySelector('.status-indicator');
        if (indicator) {
            indicator.classList.toggle('online', isOnline);
            indicator.classList.toggle('offline', !isOnline);
        }
    }
}

// Update last message in user list
function updateUserLastMessage(userId, content) {
    const userElement = document.querySelector(`.chat-user[data-user-id="${userId}"]`);
    if (userElement) {
        const lastMessageElement = userElement.querySelector('.last-message');
        if (lastMessageElement) {
            lastMessageElement.textContent = content.substring(0, 30) + (content.length > 30 ? '...' : '');
        }
        
        // Move user to top of list
        userElement.parentNode.insertBefore(userElement, userElement.parentNode.firstChild);
    }
}

// Setup scroll event listener for infinite scroll
function setupEventListeners() {
    const messagesList = document.getElementById('messages-list');
    if (messagesList) {
        messagesList.addEventListener('scroll', throttle(() => {
            if (messagesList.scrollTop < 100 && hasMoreMessages && !isLoadingMessages) {
                loadMessages();
            }
        }, 200));
    }
    
    const messageInput = document.getElementById('message-input');
    if (messageInput) {
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });
    }
}

// Throttle function to prevent spamming scroll events
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

// Get current user ID from session
function getCurrentUserId() {
    return new Promise((resolve) => {
        const sessionCookie = document.cookie.split('; ')
            .find(row => row.startsWith('session_id='))
            ?.split('=')[1];

        if (!sessionCookie) return resolve(null);

        fetch('/api/current-user')
            .then(response => response.json())
            .then(data => resolve(data.userId))
            .catch(() => resolve(null));
    });
}

// Close chat
function closeChat() {
    document.getElementById('chat-container').style.display = 'none';
    currentRecipient = null;
}

// Expose functions to global scope
window.initChat = initChat;
window.openChat = openChat;
window.sendMessage = sendMessage;
window.closeChat = closeChat;