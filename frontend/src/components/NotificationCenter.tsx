import React, { useState, useEffect, useRef } from 'react';
import { useNotifications } from '../hooks/useNotifications';
import { NotificationItem } from './NotificationItem';
import { NotificationPreferences } from './NotificationPreferences';
import './NotificationCenter.css';

interface NotificationCenterProps {
  userId: string;
  isOpen: boolean;
  onClose: () => void;
}

export const NotificationCenter: React.FC<NotificationCenterProps> = ({
  userId,
  isOpen,
  onClose
}) => {
  const { 
    notifications, 
    unreadCount, 
    markAsRead, 
    markAllAsRead,
    loading,
    error 
  } = useNotifications(userId);
  
  const [activeTab, setActiveTab] = useState<'notifications' | 'preferences'>('notifications');
  const [filter, setFilter] = useState<'all' | 'unread' | 'read'>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const centerRef = useRef<HTMLDivElement>(null);

  // Close on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  // Close on click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (centerRef.current && !centerRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, onClose]);

  // Filter notifications based on current filter and search
  const filteredNotifications = notifications.filter(notification => {
    // Apply filter
    if (filter === 'unread' && notification.isRead) return false;
    if (filter === 'read' && !notification.isRead) return false;
    
    // Apply search
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase();
      return (
        notification.title?.toLowerCase().includes(searchLower) ||
        notification.message.toLowerCase().includes(searchLower) ||
        notification.type.toLowerCase().includes(searchLower)
      );
    }
    
    return true;
  });

  const handleMarkAllAsRead = () => {
    markAllAsRead();
  };

  if (!isOpen) return null;

  return (
    <div className="notification-center-overlay">
      <div className="notification-center" ref={centerRef}>
        {/* Header */}
        <div className="notification-center-header">
          <div className="header-left">
            <h2>Notifications</h2>
            {unreadCount > 0 && (
              <span className="unread-badge">{unreadCount}</span>
            )}
          </div>
          
          <div className="header-right">
            <button onClick={onClose} className="close-button">
              ✕
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="notification-tabs">
          <button
            className={`tab-button ${activeTab === 'notifications' ? 'active' : ''}`}
            onClick={() => setActiveTab('notifications')}
          >
            Notifications
          </button>
          <button
            className={`tab-button ${activeTab === 'preferences' ? 'active' : ''}`}
            onClick={() => setActiveTab('preferences')}
          >
            Preferences
          </button>
        </div>

        {/* Content */}
        <div className="notification-content">
          {activeTab === 'notifications' ? (
            <div className="notifications-tab">
              {/* Filters and Search */}
              <div className="notification-filters">
                <div className="filter-controls">
                  <select
                    value={filter}
                    onChange={(e) => setFilter(e.target.value as any)}
                    className="filter-select"
                  >
                    <option value="all">All</option>
                    <option value="unread">Unread</option>
                    <option value="read">Read</option>
                  </select>
                  
                  <input
                    type="text"
                    placeholder="Search notifications..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="search-input"
                  />
                </div>
                
                {unreadCount > 0 && (
                  <button onClick={handleMarkAllAsRead} className="mark-all-read-btn">
                    Mark all as read
                  </button>
                )}
              </div>

              {/* Notifications List */}
              <div className="notifications-list">
                {loading ? (
                  <div className="loading-state">
                    <p>Loading notifications...</p>
                  </div>
                ) : error ? (
                  <div className="error-state">
                    <p>Error: {error}</p>
                  </div>
                ) : filteredNotifications.length === 0 ? (
                  <div className="empty-state">
                    <p>No notifications found.</p>
                  </div>
                ) : (
                  filteredNotifications.map((notification) => (
                    <NotificationItem
                      key={notification.id}
                      notification={notification}
                      onMarkAsRead={markAsRead}
                    />
                  ))
                )}
              </div>
            </div>
          ) : (
            <NotificationPreferences userId={userId} />
          )}
        </div>
      </div>
    </div>
  );
};
