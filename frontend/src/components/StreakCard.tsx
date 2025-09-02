import React from 'react';

interface StreakCardProps {
  currentStreak: number;
  longestStreak: number;
  totalSessions: number;
}

export function StreakCard({ currentStreak, longestStreak, totalSessions }: StreakCardProps) {
  return (
    <div className="card">
      <h3 className="streak-title">🔥 Current Streak</h3>
      <div className="streak-display">
        <div className="streak-number">{currentStreak}</div>
        <div className="streak-label">days</div>
      </div>
      
      <div className="streak-stats">
        <div className="stat-item">
          <span className="stat-label">Longest Streak</span>
          <span className="stat-value">{longestStreak} days</span>
        </div>
        <div className="stat-item">
          <span className="stat-label">Total Sessions</span>
          <span className="stat-value">{totalSessions}</span>
        </div>
      </div>
      
      {currentStreak > 0 && (
        <div className="streak-motivation">
          <p>Keep it up! You're doing amazing! 🚀</p>
        </div>
      )}
    </div>
  );
}
