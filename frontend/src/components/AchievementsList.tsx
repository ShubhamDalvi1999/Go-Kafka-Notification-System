import React from 'react';
import { Achievement } from '../hooks/useEngagementData';

interface AchievementsListProps {
  achievements: Achievement[];
}

export function AchievementsList({ achievements }: AchievementsListProps) {
  if (achievements.length === 0) {
    return (
      <div className="card">
        <h3 className="achievements-title">🏆 Achievements</h3>
        <p className="no-achievements">No achievements unlocked yet. Keep practicing to earn your first badge!</p>
      </div>
    );
  }

  return (
    <div className="card">
      <h3 className="achievements-title">🏆 Achievements</h3>
      <div className="achievements-grid">
        {achievements.map((achievement) => (
          <div key={achievement.id} className="achievement-item">
            <div className="achievement-icon">{achievement.icon}</div>
            <div className="achievement-content">
              <h4 className="achievement-name">{achievement.name}</h4>
              <p className="achievement-description">{achievement.description}</p>
              <span className="achievement-date">
                Unlocked: {new Date(achievement.unlockedAt).toLocaleDateString()}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
