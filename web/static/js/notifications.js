class NotificationManager {
    constructor() {
        this.bell = document.getElementById('notificationBell');
        this.dropdown = document.getElementById('notificationDropdown');
        this.badge = document.querySelector('.notification-badge');
        this.list = document.querySelector('.notification-list');
        this.markAllReadBtn = document.getElementById('markAllRead');
        
        // Initialize with empty array
        this.notifications = [];
        this.unreadCount = 0;
        
        // Only initialize if all required elements are present
        if (this.bell && this.dropdown && this.badge && this.list && this.markAllReadBtn) {
            this.initialize();
        } else {
            console.warn('Some notification elements are missing from the DOM');
        }
    }
    
    initialize() {
        // Toggle dropdown on bell click
        this.bell.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleDropdown();
        });
        
        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.dropdown.contains(e.target) && !this.bell.contains(e.target)) {
                this.hideDropdown();
            }
        });
        
        // Mark all as read
        this.markAllReadBtn.addEventListener('click', () => {
            this.markAllAsRead();
        });
        
        // Load initial notifications
        this.loadNotifications();
        
        // Start WebSocket connection for real-time notifications
        this.connectWebSocket();
    }
    
    toggleDropdown() {
        if (this.dropdown.style.display === 'none') {
            this.showDropdown();
        } else {
            this.hideDropdown();
        }
    }
    
    showDropdown() {
        this.dropdown.style.display = 'block';
    }
    
    hideDropdown() {
        this.dropdown.style.display = 'none';
    }
    
    async loadNotifications() {
        try {
            const response = await fetch('/api/notifications', {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error('Failed to load notifications');
            }
            
            const notifications = await response.json();
            this.notifications = notifications || [];
            this.updateUnreadCount();
            this.renderNotifications();
            
        } catch (error) {
            console.error('Failed to load notifications:', error);
            this.notifications = [];
            this.updateUnreadCount();
            this.renderNotifications();
        }
    }
    
    updateUnreadCount() {
        if (!this.notifications) {
            this.unreadCount = 0;
            this.badge.style.display = 'none';
            return;
        }
        
        this.unreadCount = this.notifications.filter(n => !n.read).length;
        if (this.unreadCount > 0) {
            this.badge.textContent = this.unreadCount;
            this.badge.style.display = 'block';
        } else {
            this.badge.style.display = 'none';
        }
    }
    
    renderNotifications() {
        this.list.innerHTML = '';
        
        if (!this.notifications || this.notifications.length === 0) {
            this.list.innerHTML = '<div class="notification-item">No notifications</div>';
            return;
        }
        
        this.notifications.forEach(notification => {
            const item = document.createElement('div');
            item.className = `notification-item ${notification.read ? '' : 'unread'}`;
            
            const title = notification.title || 'New Notification';
            const content = notification.content || '';
            
            item.innerHTML = `
                <div class="notification-title">${title}</div>
                <div class="notification-content">${content}</div>
                <div class="notification-time">${this.formatTime(notification.timestamp)}</div>
            `;
            
            item.addEventListener('click', () => {
                this.markAsRead(notification.id);
                if (notification.link) {
                    window.location.href = notification.link;
                }
            });
            
            this.list.appendChild(item);
        });
    }
    
    async markAsRead(notificationId) {
        try {
            const response = await fetch(`/api/notifications/${notificationId}/read`, {
                method: 'POST',
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error('Failed to mark notification as read');
            }
            
            const notification = this.notifications.find(n => n.id === notificationId);
            if (notification) {
                notification.read = true;
                this.updateUnreadCount();
                this.renderNotifications();
            }
            
        } catch (error) {
            console.error('Failed to mark notification as read:', error);
        }
    }
    
    async markAllAsRead() {
        try {
            const response = await fetch('/api/notifications/read-all', {
                method: 'POST',
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error('Failed to mark all notifications as read');
            }
            
            this.notifications.forEach(n => n.read = true);
            this.updateUnreadCount();
            this.renderNotifications();
            
        } catch (error) {
            console.error('Failed to mark all notifications as read:', error);
        }
    }
    
    formatTime(timestamp) {
        if (!timestamp) return 'Unknown date';
        
        try {
            const date = new Date(timestamp);
            if (isNaN(date.getTime())) return 'Invalid date';
            
            const now = new Date();
            const diff = now - date;
            
            if (diff < 60000) { // Less than 1 minute
                return 'Just now';
            } else if (diff < 3600000) { // Less than 1 hour
                const minutes = Math.floor(diff / 60000);
                return `${minutes}m ago`;
            } else if (diff < 86400000) { // Less than 1 day
                const hours = Math.floor(diff / 3600000);
                return `${hours}h ago`;
            } else {
                return date.toLocaleDateString();
            }
        } catch (error) {
            console.error('Error formatting date:', error);
            return 'Invalid date';
        }
    }
    
    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(`${protocol}//${window.location.host}/ws/notifications`);
        
        ws.onmessage = (event) => {
            const notification = JSON.parse(event.data);
            this.notifications.unshift(notification);
            this.updateUnreadCount();
            this.renderNotifications();
        };
        
        ws.onclose = () => {
            // Reconnect after 5 seconds
            setTimeout(() => this.connectWebSocket(), 5000);
        };
    }
}

// Initialize notifications when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new NotificationManager();
}); 