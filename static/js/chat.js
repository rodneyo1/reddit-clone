let chatSocket = null;
let currentRecipient = null;
let messageOffset = 0;
let isLoadingMessages = false;
let hasMoreMessages = true;
const renderedMessages = new Set();
let typingTimeout = null;
let typingUsers = new Map();
let lastTypingTime = 0;
const TYPING_TIMER_LENGTH = 3000; // How long to wait after last keystroke
let loadedMessageIds = new Set();
const INITIAL_MESSAGE_COUNT = 30; // Number of messages to load initially
const MESSAGES_PER_SCROLL = 20;  // Number of messages to load per scroll

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
        console.log('WebSocket message received:', data); // Debug log
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

    userList.innerHTML = users
        .sort((a, b) => {
            // Sort by last message time, then alphabetically
            if (a.last_message_time && b.last_message_time) {
                return new Date(b.last_message_time) - new Date(a.last_message_time);
            }
            return a.username.localeCompare(b.username);
        })
        .map(user => `
        <div class="chat-user" data-user-id="${user.id}">
            <div class="user-avatar">
                ${user.avatar_url ? 
                    `<img src="${user.avatar_url}" alt="${user.username}">` : 
                    `<div class="default-avatar">${user.username.charAt(0).toUpperCase()}</div>`}
                <span class="status-indicator ${user.is_online ? 'online' : 'offline'}"></span>
            </div>
            <div class="user-info">
                <span class="username">${user.username}</span>
                <div class="user-status">
                    <span class="last-message">${user.last_message ? 
                        user.last_message.substring(0,30) : 'No messages yet'}</span>
                    <div class="typing-indicator" data-user-id="${user.id}" style="display: none;">
                        is typing<span class="typing-dots">...</span>
                    </div>
                </div>
            </div>
            ${user.unread_count > 0 ? `<span class="unread-count">${user.unread_count}</span>` : ''}
        </div>
    `).join('');

    // Add event listeners
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
    loadedMessageIds.clear();
    
    // Reset typing indicators
    const chatTypingIndicator = document.getElementById('chat-typing-indicator');
    if (chatTypingIndicator) {
        chatTypingIndicator.style.display = 'none';
    }
    
    // Update UI
    document.getElementById('chat-recipient').textContent = username;
    document.getElementById('chat-container').style.display = 'block';
    document.getElementById('messages-list').innerHTML = '<div class="loading">Loading messages...</div>';

    const userEl = document.querySelector(`.chat-user[data-user-id="${userId}"]`);
    if (userEl) {
        const badge = userEl.querySelector('.unread-count');
        if (badge) {
            badge.remove();
        }
    }
    
    // Load initial messages
    await loadMessages(true);
    
    // Mark messages as read via WebSocket
    if (chatSocket && chatSocket.readyState === WebSocket.OPEN) {
        chatSocket.send(JSON.stringify({
            type: 'mark_read',
            sender_id: userId
        }));
    }
}

// Load messages with pagination
async function loadMessages(isInitialLoad = false) {
    if (isLoadingMessages || !hasMoreMessages) return;
    
    isLoadingMessages = true;
    try {
        // Add loading indicator at top when scrolling up
        const messagesList = document.getElementById('messages-list');
        if (!isInitialLoad && messagesList) {
            const loadingDiv = document.createElement('div');
            loadingDiv.className = 'loading-more';
            loadingDiv.textContent = 'Loading more messages...';
            messagesList.insertAdjacentElement('afterbegin', loadingDiv);
        }

        const response = await fetch(`/api/chat/messages?recipient_id=${currentRecipient}&offset=${messageOffset}`);
        if (!response.ok) throw new Error('Failed to load messages');
        
        const messages = await response.json();
        
        // Remove loading indicator if it exists
        const loadingIndicator = document.querySelector('.loading-more');
        if (loadingIndicator) {
            loadingIndicator.remove();
        }

        if (messages.length === 0) {
            hasMoreMessages = false;
            if (messageOffset === 0) {
                messagesList.innerHTML = '<div class="no-messages">No messages yet. Start the conversation!</div>';
            }
            return;
        }

        // Filter out any duplicate messages
        const newMessages = messages.filter(msg => !loadedMessageIds.has(msg.id));
        
        if (newMessages.length === 0) {
            hasMoreMessages = messages.length === 10; // If we got 10 messages but all were duplicates, there might be more
            messageOffset += messages.length; // Still increment offset even if all were duplicates
            return;
        }

        // Add new message IDs to our tracking set
        newMessages.forEach(msg => loadedMessageIds.add(msg.id));
        
        // Reverse to show oldest first (since we load newest first)
        newMessages.reverse();
        
        const messagesHTML = newMessages.map(msg => createMessageElement(msg)).join('');
        
        if (messageOffset === 0) {
            messagesList.innerHTML = messagesHTML;
            messagesList.scrollTop = messagesList.scrollHeight;
        } else {
            const scrollHeightBefore = messagesList.scrollHeight;
            const scrollTopBefore = messagesList.scrollTop;
            
            messagesList.insertAdjacentHTML('afterbegin', messagesHTML);
            
            // Maintain scroll position when loading older messages
            const scrollHeightAfter = messagesList.scrollHeight;
            messagesList.scrollTop = scrollTopBefore + (scrollHeightAfter - scrollHeightBefore);
        }
        
        // Increment offset by the number of messages we received
        messageOffset += messages.length;
        
        // Update hasMoreMessages based on whether we received a full page
        hasMoreMessages = messages.length === 10;
        
        console.log('Loaded messages:', {
            offset: messageOffset,
            newMessages: newMessages.length,
            hasMore: hasMoreMessages
        });
    } catch (error) {
        console.error('Error loading messages:', error);
        // Remove loading indicator on error
        document.querySelector('.loading-more')?.remove();
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
    
    // Unique identifier for deduplication
    const messageId = msg.id ? msg.id : `temp-${msg.temp_id || Date.now()}`;

    // Check if this message should show the sender (if it's different from the previous message)
    const previousMessage = document.querySelector(`[data-message-id="${messageId}"]`)?.previousElementSibling;
    const showSender = !previousMessage || previousMessage.dataset.sender !== msg.sender_id;
    
    // Handle avatar with fallbacks
    const avatarContent = msg.sender_avatar ? 
        `<img src="${msg.sender_avatar}" class="message-avatar" alt="${username}">` :
        `<div class="message-avatar-default">${username.charAt(0).toUpperCase()}</div>`;

    return `
        <div class="message ${isCurrentUser ? 'sent' : 'received'} ${showSender ? 'new-sender' : 'same-sender'}" 
             data-message-id="${messageId}"
             data-sender="${msg.sender_id}"
             data-timestamp="${msg.created_at}"
             data-recipient="${msg.recipient_id}">
            ${showSender ? `
                <div class="message-header">
                    ${avatarContent}
                    <div class="message-metadata">
                        <span class="sender">${username}</span>
                    </div>
                </div>
            ` : ''}
            <div class="message-content">
                ${msg.content}
                <span class="time">${messageTime}</span>
            </div>
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
        temp_id: tempId,
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
        if (!renderedMessages.has(tempId)) {
            messagesList.insertAdjacentHTML('beforeend', createMessageElement(tempMessage));
            renderedMessages.add(tempId); //Avoid future duplication
            messagesList.scrollTop = messagesList.scrollHeight;
        }
    }
    // messagesList.insertAdjacentHTML('beforeend', createMessageElement(tempMessage));
    // messagesList.scrollTop = messagesList.scrollHeight;
    
    // Clear typing status when sending message
    if (typingTimeout) {
        clearTimeout(typingTimeout);
    }
    chatSocket.send(JSON.stringify({
        type: 'stop_typing',
        recipient_id: currentRecipient
    }));
    typingUsers.clear();
    updateTypingIndicator();
    
    // Send via WebSocket
    chatSocket.send(JSON.stringify({
        recipient_id: currentRecipient,
        content: content,
        temp_id: tempId
    }));
    
    input.value = '';
}

// Handle WebSocket messages
function handleWebSocketMessage(data) {
    console.log('Received WebSocket message:', data);

    const messageId = data.temp_id || data.id;

    // Early return if message already exists
    if (renderedMessages.has(messageId) || (data.id && loadedMessageIds.has(data.id))) {
        console.log('Duplicate message detected, skipping:', messageId);
        return;
    }
    
    if (data.type === 'messages_read') {
        if (data.recipient_id === currentRecipient) {
            document.querySelectorAll('.message.sent').forEach(msg => {
                msg.classList.add('read');
            });
        }
        return;
    }

    if (data.type === 'status_update') {
        updateUserStatus(data.user_id, data.is_online);
        handleTypingStatus(data);
        // loadChatUsers();
        return;
    }

        // Handle notification messages (not full chat message)
    if (data.type === 'new_message_notification') {
        if (data.sender_id !== getCurrentUserId() && data.sender_id !== currentRecipient) {
            playNotificationSound();
            highlightUserInList(data.sender_id);
        }
        return;
    }

    if (data.type === 'typing_status') {
        console.log('Received typing status:', data); // Add debug logging
        handleTypingStatus(data);
        return;
    }
    
    // Handle chat message
    const message = data;
    const isCurrentUser = message.sender_id === getCurrentUserId();
    
    // Remove temporary message if it exists
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

        renderedMessages.add(messageId);
        
        // Update user list and notifications
        updateUserLastMessage(message.sender_id, message.content);
        if (!isCurrentUser && message.sender_id !== currentRecipient) {
            playNotificationSound();
            incrementUnreadCount(message.sender_id);
        }
    }

    // Add message ID to tracking sets when adding new message
    if (data.id) {
        loadedMessageIds.add(data.id);
    }
}

function incrementUnreadCount(userId) {
    const userEl = document.querySelector(`.chat-user[data-user-id="${userId}"]`);
    if (!userEl) return;

    let badge = userEl.querySelector('.unread-count');
    if (badge) {
        let count = parseInt(badge.textContent || '0', 10);
        count += 1;
        badge.textContent = count;
    } else {
        badge = document.createElement('span');
        badge.className = 'unread-count';
        badge.textContent = '1';
        userEl.appendChild(badge);
    }
}


function playNotificationSound() {
    new Audio('/static/sounds/notification.mp3').play().catch(() => {});
}

function highlightUserInList(userId) {
    const userElem = document.querySelector(`#user-${userId}`);
    if (userElem) {
        userElem.classList.add('has-new-message');
        // Optionally add a badge
        const badge = userElem.querySelector('.unread-count');
        if (badge) {
            let count = parseInt(badge.textContent) || 0;
            badge.textContent = count + 1;
            badge.style.display = 'inline';
        }
    }
}

// Update user status in the UI
function updateUserStatus(userId, isOnline) {
    const userElement = document.querySelector(`.chat-user[data-user-id="${userId}"]`);
    if (userElement) {
        const indicator = userElement.querySelector('.status-indicator');
        if (indicator) {
           indicator.classList.remove('online', 'offline');
            indicator.classList.add(isOnline ? 'online' : 'offline');

            // Update last seen tooltip
            const lastSeen = isOnline ? 'Online now' : `Last seen ${formatLastSeen(data.last_seen)}`;
            indicator.title = lastSeen;
        }
    }
}

function formatLastSeen(timestamp) {
    const now = new Date();
    const lastSeen = new Date(timestamp);
    const diffMinutes = Math.round((now - lastSeen) / 60000);
    
    if (diffMinutes < 1) return 'just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffMinutes < 1440) return `${Math.floor(diffMinutes/60)}h ago`;
    return `${Math.floor(diffMinutes/1440)}d ago`;
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
        let scrollTimeout;
        messagesList.addEventListener('scroll', () => {
            if (scrollTimeout) {
                clearTimeout(scrollTimeout);
            }

            scrollTimeout = setTimeout(() => {
                // Check if we're near the top (within 100px) and should load more messages
                if (messagesList.scrollTop < 100 && hasMoreMessages && !isLoadingMessages) {
                    console.log('Loading more messages...', {
                        scrollTop: messagesList.scrollTop,
                        hasMore: hasMoreMessages,
                        offset: messageOffset
                    });
                    loadMessages(false);
                }
            }, 150); // Debounce time of 150ms
        });
    }
    
    const messageInput = document.getElementById('message-input');
    if (messageInput) {
        messageInput.addEventListener('input', handleTyping);
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });
    }
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

// Add this function to handle typing events
function handleTyping() {
    if (!currentRecipient || !chatSocket) return;

    const now = Date.now();

    // Only send typing event if it's been more than 1 second since last keystroke
    if (now - lastTypingTime > 1000) {
        lastTypingTime = now;
        
        console.log('Sending typing status...'); // Debug log
        chatSocket.send(JSON.stringify({
            type: 'typing',
            recipient_id: currentRecipient
        }));
    }

    // Clear existing timeout
    if (typingTimeout) {
        clearTimeout(typingTimeout);
    }

    // Set new timeout
    typingTimeout = setTimeout(() => {
        console.log('Sending stop typing status...'); // Debug log
        chatSocket.send(JSON.stringify({
            type: 'stop_typing',
            recipient_id: currentRecipient
        }));
    }, TYPING_TIMER_LENGTH);
}

// Add this new function to handle typing status updates
function handleTypingStatus(data) {
    console.log('Handling typing status:', data);

    // Update typing indicator in user list
    const userTypingIndicator = document.querySelector(`.typing-indicator[data-user-id="${data.user_id}"]`);
    if (userTypingIndicator) {
        const userInfo = userTypingIndicator.closest('.user-info');
        if (userInfo) {
            const lastMessage = userInfo.querySelector('.last-message');
            if (lastMessage) {
                lastMessage.style.display = data.is_typing ? 'none' : 'block';
            }
            userTypingIndicator.style.display = data.is_typing ? 'block' : 'none';
        }
    }

    // Update typing indicator in active chat
    const chatTypingIndicator = document.getElementById('chat-typing-indicator');
    if (chatTypingIndicator && data.user_id === currentRecipient) {
        if (data.is_typing) {
            chatTypingIndicator.style.display = 'block';
            chatTypingIndicator.textContent = `${data.username || 'Someone'} is typing`;
            
            // Scroll to show typing indicator if near bottom
            const messagesList = document.getElementById('messages-list');
            if (messagesList) {
                const isNearBottom = messagesList.scrollHeight - messagesList.scrollTop - messagesList.clientHeight < 100;
                if (isNearBottom) {
                    messagesList.scrollTop = messagesList.scrollHeight;
                }
            }
        } else {
            chatTypingIndicator.style.display = 'none';
        }
    }
}

// Add this function to update the typing indicator display
function updateTypingIndicator() {
    const typingIndicator = document.getElementById('typing-indicator');
    if (!typingIndicator) return;

    // Clean up old typing statuses (older than 3 seconds)
    const now = Date.now();
    for (const [userId, data] of typingUsers.entries()) {
        if (now - data.timestamp > 3000) {
            typingUsers.delete(userId);
        }
    }

    if (typingUsers.size === 0) {
        typingIndicator.style.display = 'none';
        return;
    }

    const typingUsernames = Array.from(typingUsers.values())
        .map(data => data.username)
        .join(', ');

    console.log('Updating typing indicator:', typingUsernames); // Add debug logging

    typingIndicator.style.display = 'block';
    typingIndicator.textContent = typingUsers.size === 1 
        ? `${typingUsernames} is typing...`
        : `${typingUsernames} are typing...`;
}

// Add functions to store and retrieve chat state
function storeChatState() {
    const chatState = {
        messageOffset,
        loadedMessageIds: Array.from(loadedMessageIds),
        lastAccessed: new Date().getTime()
    };
    localStorage.setItem(`chat_state_${currentRecipient}`, JSON.stringify(chatState));
}

function retrieveChatState() {
    const storedState = localStorage.getItem(`chat_state_${currentRecipient}`);
    if (storedState) {
        const state = JSON.parse(storedState);
        // Only restore state if it's from the last 24 hours
        if (new Date().getTime() - state.lastAccessed < 24 * 60 * 60 * 1000) {
            messageOffset = state.messageOffset;
            loadedMessageIds = new Set(state.loadedMessageIds);
            return true;
        }
    }
    return false;
}

// Add CSS for loading indicator
const style = document.createElement('style');
style.textContent = `
.loading-more {
    text-align: center;
    padding: 10px;
    color: #666;
    font-style: italic;
}
`;
document.head.appendChild(style);