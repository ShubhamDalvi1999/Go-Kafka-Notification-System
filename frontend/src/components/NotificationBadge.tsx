import React, { useState, useEffect } from 'react';
import './NotificationBadge.css';

interface NotificationBadgeProps {
  userId: string;
  className?: string;
}

export const NotificationBadge: React.FC<NotificationBadgeProps> = ({ 
  userId, 
  className = '' 
}) => {
  // Demo data - replace with actual hook later
  const unreadCount = 3; // Mock unread count
  const [isAnimating, setIsAnimating] = useState(false);

  useEffect(() => {
    if (unreadCount > 0) {
      setIsAnimating(true);
      
      // Stop animation after 1 second
      const timer = setTimeout(() => setIsAnimating(false), 1000);
      return () => clearTimeout(timer);
    }
  }, [unreadCount]);

  return (
    <div 
      className={`notification-badge ${className} ${isAnimating ? 'animate' : ''}`}
      title={`${unreadCount} unread notification${unreadCount !== 1 ? 's' : ''}`}
    >
      <span className="badge-icon">🔔</span>
      {unreadCount > 0 && (
        <span className={`badge-count ${unreadCount > 10 ? 'urgent' : unreadCount > 5 ? 'high' : 'medium'}`}>
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      )}
    </div>
  );
};
