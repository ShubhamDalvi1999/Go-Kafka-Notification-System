import { useState, useEffect } from 'react';

export interface NotificationPreference {
  id: string;
  type: string;
  channel: string;
  enabled: boolean;
  quietHoursStart?: string;
  quietHoursEnd?: string;
  maxPerDay?: number;
}

export interface UserPreferences {
  userId: string;
  preferences: NotificationPreference[];
  globalQuietHours: {
    enabled: boolean;
    start: string;
    end: string;
  };
  maxNotificationsPerDay: number;
}

export function useNotificationPreferences(userId: string) {
  const [preferences, setPreferences] = useState<UserPreferences>({
    userId,
    preferences: [],
    globalQuietHours: {
      enabled: false,
      start: '22:00',
      end: '08:00'
    },
    maxNotificationsPerDay: 10
  });
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPreferences = async () => {
      try {
        setLoading(true);
        
        // Mock data - replace with actual API call
        const mockPreferences: UserPreferences = {
          userId,
          preferences: [
            {
              id: '1',
              type: 'daily_reminder',
              channel: 'in_app',
              enabled: true,
              maxPerDay: 1
            },
            {
              id: '2',
              type: 'daily_reminder',
              channel: 'push',
              enabled: true,
              maxPerDay: 1
            },
            {
              id: '3',
              type: 'streak_reminder',
              channel: 'in_app',
              enabled: true,
              maxPerDay: 2
            },
            {
              id: '4',
              type: 'streak_reminder',
              channel: 'push',
              enabled: false,
              maxPerDay: 0
            },
            {
              id: '5',
              type: 'achievement',
              channel: 'in_app',
              enabled: true,
              maxPerDay: 5
            },
            {
              id: '6',
              type: 'achievement',
              channel: 'push',
              enabled: true,
              maxPerDay: 3
            }
          ],
          globalQuietHours: {
            enabled: true,
            start: '22:00',
            end: '08:00'
          },
          maxNotificationsPerDay: 10
        };

        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 300));
        
        setPreferences(mockPreferences);
        setError(null);
      } catch (err) {
        setError('Failed to fetch notification preferences');
        console.error('Error fetching preferences:', err);
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchPreferences();
    }
  }, [userId]);

  const updatePreference = (preferenceId: string, updates: Partial<NotificationPreference>) => {
    setPreferences(prev => ({
      ...prev,
      preferences: prev.preferences.map(pref => 
        pref.id === preferenceId ? { ...pref, ...updates } : pref
      )
    }));
  };

  const updateGlobalQuietHours = (updates: Partial<UserPreferences['globalQuietHours']>) => {
    setPreferences(prev => ({
      ...prev,
      globalQuietHours: { ...prev.globalQuietHours, ...updates }
    }));
  };

  const updateMaxNotificationsPerDay = (max: number) => {
    setPreferences(prev => ({
      ...prev,
      maxNotificationsPerDay: max
    }));
  };

  const savePreferences = async () => {
    try {
      // Simulate API call - replace with actual save logic
      await new Promise(resolve => setTimeout(resolve, 500));
      console.log('Preferences saved successfully');
      return true;
    } catch (err) {
      console.error('Failed to save preferences:', err);
      return false;
    }
  };

  return {
    preferences,
    loading,
    error,
    updatePreference,
    updateGlobalQuietHours,
    updateMaxNotificationsPerDay,
    savePreferences
  };
}
