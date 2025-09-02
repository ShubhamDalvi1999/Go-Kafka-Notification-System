import { useState, useEffect } from 'react';

export interface EngagementData {
  currentStreak: number;
  longestStreak: number;
  totalSessions: number;
  weeklyProgress: number[];
  achievements: Achievement[];
}

export interface Achievement {
  id: string;
  name: string;
  description: string;
  unlockedAt: string;
  icon: string;
}

export function useEngagementData(userId: string) {
  const [data, setData] = useState<EngagementData>({
    currentStreak: 0,
    longestStreak: 0,
    totalSessions: 0,
    weeklyProgress: [0, 0, 0, 0, 0, 0, 0],
    achievements: []
  });
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Simulate API call - replace with actual API call
    const fetchEngagementData = async () => {
      try {
        setLoading(true);
        
        // Mock data for now
        const mockData: EngagementData = {
          currentStreak: 7,
          longestStreak: 21,
          totalSessions: 45,
          weeklyProgress: [5, 7, 6, 8, 7, 9, 7],
          achievements: [
            {
              id: '1',
              name: 'First Week',
              description: 'Complete 7 days in a row',
              unlockedAt: '2024-01-15',
              icon: '🎯'
            },
            {
              id: '2',
              name: 'Streak Master',
              description: 'Maintain a 21-day streak',
              unlockedAt: '2024-02-05',
              icon: '🔥'
            },
            {
              id: '3',
              name: 'Consistent Learner',
              description: 'Complete 30 total sessions',
              unlockedAt: '2024-01-28',
              icon: '📚'
            }
          ]
        };

        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 500));
        
        setData(mockData);
        setError(null);
      } catch (err) {
        setError('Failed to fetch engagement data');
        console.error('Error fetching engagement data:', err);
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchEngagementData();
    }
  }, [userId]);

  const updateStreak = (newStreak: number) => {
    setData(prev => ({
      ...prev,
      currentStreak: newStreak,
      longestStreak: Math.max(prev.longestStreak, newStreak)
    }));
  };

  const addSession = () => {
    setData(prev => ({
      ...prev,
      totalSessions: prev.totalSessions + 1
    }));
  };

  return {
    data,
    loading,
    error,
    updateStreak,
    addSession
  };
}
