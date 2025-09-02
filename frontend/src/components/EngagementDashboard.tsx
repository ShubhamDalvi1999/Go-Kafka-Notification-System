import React, { useState } from 'react';
import { useEngagementData } from '../hooks/useEngagementData';
import { StreakCard } from './StreakCard';
import { ProgressChart } from './ProgressChart';
import { AchievementsList } from './AchievementsList';
import './EngagementDashboard.css';

interface EngagementDashboardProps {
  userId: string;
}

export const EngagementDashboard: React.FC<EngagementDashboardProps> = ({
  userId
}) => {
  const {
    data,
    loading,
    error,
    updateStreak,
    addSession
  } = useEngagementData(userId);

  const [selectedTimeframe, setSelectedTimeframe] = useState<'7d' | '30d' | '90d'>('7d');

  if (loading) {
    return (
      <div className="engagement-dashboard">
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading engagement data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="engagement-dashboard">
        <div className="error-state">
          <p>Error loading engagement data: {error}</p>
          <button onClick={() => window.location.reload()}>Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="engagement-dashboard">
      {/* Header */}
      <div className="dashboard-header">
        <h2 className="dashboard-title">🎯 Engagement Dashboard</h2>
        <p className="dashboard-subtitle">Track your progress and stay motivated</p>
        
        <div className="header-controls">
          <select
            value={selectedTimeframe}
            onChange={(e) => setSelectedTimeframe(e.target.value as any)}
            className="input"
            style={{ width: 'auto', marginRight: '12px' }}
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>
          
          <button onClick={addSession} className="btn btn-primary">
            ➕ Add Session
          </button>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="dashboard-grid">
        {/* Streak Card */}
        <div className="dashboard-section">
          <h3 className="section-title">🔥 Streak Tracking</h3>
          <StreakCard 
            currentStreak={data.currentStreak}
            longestStreak={data.longestStreak}
            totalSessions={data.totalSessions}
          />
        </div>

        {/* Progress Chart */}
        <div className="dashboard-section">
          <h3 className="section-title">📊 Weekly Progress</h3>
          <ProgressChart weeklyProgress={data.weeklyProgress} />
        </div>

        {/* Achievements */}
        <div className="dashboard-section">
          <h3 className="section-title">🏆 Achievements</h3>
          <AchievementsList achievements={data.achievements} />
        </div>
      </div>

      {/* Quick Actions */}
      <div className="dashboard-section">
        <h3 className="section-title">⚡ Quick Actions</h3>
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <button 
            onClick={addSession} 
            className="btn btn-primary"
          >
            Start Practice Session
          </button>
          <button 
            onClick={() => updateStreak(data.currentStreak + 1)} 
            className="btn btn-secondary"
          >
            Extend Streak
          </button>
          <button className="btn btn-secondary">
            View All Achievements
          </button>
        </div>
      </div>

      {/* Motivation Section */}
      <div className="dashboard-section">
        <h3 className="section-title">💡 Stay Motivated</h3>
        <div style={{ display: 'grid', gap: '16px', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))' }}>
          <div className="card">
            <h4 style={{ marginBottom: '8px' }}>💡 Set Daily Goals</h4>
            <p style={{ margin: 0, color: '#6b7280' }}>
              Even 15 minutes of practice can maintain your streak!
            </p>
          </div>
          
          <div className="card">
            <h4 style={{ marginBottom: '8px' }}>🎯 Track Progress</h4>
            <p style={{ margin: 0, color: '#6b7280' }}>
              Monitor your streaks and celebrate small wins
            </p>
          </div>
          
          <div className="card">
            <h4 style={{ marginBottom: '8px' }}>🚀 Level Up</h4>
            <p style={{ margin: 0, color: '#6b7280' }}>
              Complete sessions to earn achievements and unlock rewards
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
