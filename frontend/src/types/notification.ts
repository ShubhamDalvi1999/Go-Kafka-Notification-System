export interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'reminder' | 'achievement' | 'streak' | 'system';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  isRead: boolean;
  isDelivered: boolean;
  createdAt: string;
  updatedAt: string;
  userId: string;
  metadata?: Record<string, any>;
}

export interface NotificationPreferences {
  id: string;
  userId: string;
  type: string;
  channel: string;
  enabled: boolean;
  quietHoursStart?: string;
  quietHoursEnd?: string;
  maxPerDay?: number;
}

export interface UserPreferences {
  userId: string;
  preferences: NotificationPreferences[];
  globalQuietHours: {
    enabled: boolean;
    start: string;
    end: string;
  };
  maxNotificationsPerDay: number;
}

export interface NotificationFilters {
  type?: string;
  priority?: string;
  isRead?: boolean;
  isDelivered?: boolean;
  search?: string;
  dateRange?: {
    start: string;
    end: string;
  };
}

export interface NotificationStats {
  total: number;
  unread: number;
  unreadByPriority: {
    urgent: number;
    high: number;
    medium: number;
    low: number;
  };
  unreadByType: {
    reminder: number;
    achievement: number;
    streak: number;
    system: number;
  };
}
