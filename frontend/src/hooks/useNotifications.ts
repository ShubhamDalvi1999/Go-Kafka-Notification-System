import { useState, useEffect } from 'react';

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'reminder' | 'achievement' | 'streak' | 'system';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  isRead: boolean;
  createdAt: string;
  userId: string;
  metadata?: Record<string, any>;
}

export interface NotificationFilters {
  type?: string;
  priority?: string;
  isRead?: boolean;
  search?: string;
}

export function useNotifications(userId: string) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<NotificationFilters>({});

  useEffect(() => {
    const fetchNotifications = async () => {
      try {
        setLoading(true);
        
        // Mock data - replace with actual API call
        const mockNotifications: Notification[] = [
          {
            id: '1',
            title: 'Daily Practice Reminder',
            message: 'Time for your daily practice session! Keep your streak alive.',
            type: 'reminder',
            priority: 'medium',
            isRead: false,
            createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
            userId
          },
          {
            id: '2',
            title: 'Streak Milestone! 🔥',
            message: 'Congratulations! You\'ve maintained a 7-day streak!',
            type: 'streak',
            priority: 'high',
            isRead: false,
            createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
            userId
          },
          {
            id: '3',
            title: 'Achievement Unlocked! 🏆',
            message: 'You\'ve earned the "First Week" badge for completing 7 days in a row.',
            type: 'achievement',
            priority: 'high',
            isRead: true,
            createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(), // 2 days ago
            userId
          },
          {
            id: '4',
            title: 'System Update',
            message: 'New features have been added to your dashboard.',
            type: 'system',
            priority: 'low',
            isRead: true,
            createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(), // 3 days ago
            userId
          }
        ];

        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 300));
        
        setNotifications(mockNotifications);
        setError(null);
      } catch (err) {
        setError('Failed to fetch notifications');
        console.error('Error fetching notifications:', err);
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchNotifications();
    }
  }, [userId]);

  const markAsRead = (notificationId: string) => {
    setNotifications(prev => 
      prev.map(notification => 
        notification.id === notificationId 
          ? { ...notification, isRead: true }
          : notification
      )
    );
  };

  const markAllAsRead = () => {
    setNotifications(prev => 
      prev.map(notification => ({ ...notification, isRead: true }))
    );
  };

  const deleteNotification = (notificationId: string) => {
    setNotifications(prev => 
      prev.filter(notification => notification.id !== notificationId)
    );
  };

  const addNotification = (notification: Omit<Notification, 'id' | 'createdAt'>) => {
    const newNotification: Notification = {
      ...notification,
      id: Date.now().toString(),
      createdAt: new Date().toISOString()
    };
    setNotifications(prev => [newNotification, ...prev]);
  };

  const filteredNotifications = notifications.filter(notification => {
    if (filters.type && notification.type !== filters.type) return false;
    if (filters.priority && notification.priority !== filters.priority) return false;
    if (filters.isRead !== undefined && notification.isRead !== filters.isRead) return false;
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      return (
        notification.title.toLowerCase().includes(searchLower) ||
        notification.message.toLowerCase().includes(searchLower)
      );
    }
    return true;
  });

  const unreadCount = notifications.filter(n => !n.isRead).length;

  return {
    notifications: filteredNotifications,
    allNotifications: notifications,
    loading,
    error,
    filters,
    unreadCount,
    setFilters,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    addNotification
  };
}
